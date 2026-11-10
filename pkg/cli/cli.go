package cli

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/janczizikow/bijou/internal/ignore"
	"github.com/janczizikow/bijou/pkg/api"
)

const (
	ExitSuccess = iota
	ExitError
)

type CLI struct {
	version string
	opts    api.FormatOptions
	ig      *ignore.Ignorer
}

func New(version string) *CLI {
	return &CLI{
		version: version,
		opts:    api.DefaultOptions(),
		ig:      ignore.New(),
	}
}

func (c *CLI) Run(args []string) int {
	fs := flag.NewFlagSet("bijou", flag.ExitOnError)

	fs.Var(&arrowParensValue{target: &c.opts.ArrowParens}, "arrow-parens", "Include parentheses around a sole arrow function parameter")
	fs.BoolVar(&c.opts.UseTabs, "use-tabs", c.opts.UseTabs, "Indent with tabs instead of spaces")
	fs.IntVar(&c.opts.TabWidth, "tab-width", c.opts.TabWidth, "Number of spaces per indentation level")
	fs.IntVar(&c.opts.PrintWidth, "print-width", c.opts.PrintWidth, "The line length to wrap")
	fs.BoolVar(&c.opts.SingleQuote, "single-quote", c.opts.SingleQuote, "Use single quotes instead of double quotes")
	fs.Var(&trailingCommaValue{target: &c.opts.TrailingComma}, "trailing-comma", "Print trailing commas wherever possible when multi-line")
	fs.BoolVar(&c.opts.BracketSpacing, "bracket-spacing", c.opts.BracketSpacing, "Print spaces between brackets")
	fs.BoolVar(&c.opts.JsxBracketSameLine, "jsx-bracket-same-line", c.opts.JsxBracketSameLine, "Print > on the last line instead of at a new line")

	noSemi := fs.Bool("no-semi", false, "Do not print semicolons, except at the beginning of lines which may need them")
	printVersion := fs.Bool("version", false, "print bijou version and exit")
	fs.BoolVar(printVersion, "v", false, "print bijou version and exit")

	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "Usage: bijou [options] [file/dir/glob ...]\n\noptions:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return ExitError
	}

	if *printVersion {
		fmt.Fprintln(fs.Output(), c.version)
		return ExitSuccess
	}

	if *noSemi {
		c.opts.Semi = false
	}

	paths := fs.Args()
	if len(paths) == 0 {
		fs.Usage()
		return ExitError
	}

	if errs := c.formatFiles(paths); len(errs) > 0 {
		for _, err := range errs {
			fmt.Fprintln(fs.Output(), err.Error())
		}
		return ExitError
	}
	return ExitSuccess
}

func (c *CLI) formatFiles(paths []string) []error {
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)
	workers := runtime.GOMAXPROCS(0)
	// Buffer 2x worker count to stage the next batch while workers are busy,
	// preventing worker starvation between disk reads without unbounded memory use.
	jobs := make(chan string, workers*2)
	for range workers {
		wg.Go(func() {
			for path := range jobs {
				if err := c.formatAndWriteFile(path); err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}
		})
	}

	walkErr := c.walk(paths, jobs)
	close(jobs)
	wg.Wait()
	// TODO: report errors as they are encountered in formatFile??
	if walkErr != nil {
		errs = append([]error{walkErr}, errs...)
	}

	return errs
}

func (c *CLI) walk(paths []string, jobs chan string) error {
	seen := map[string]struct{}{}
	enqueue := func(path string) {
		path = filepath.Clean(path)
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		jobs <- path
	}
	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("no files matching the pattern were found: %s", path)
		}

		if fileInfo.IsDir() {
			err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() && c.ig.Ignores(p) {
					return filepath.SkipDir
				}
				if !d.IsDir() && !c.ig.Ignores(p) {
					enqueue(p)
					return nil
				}
				return nil
			})
			if err != nil {
				return err
			}
		} else if !c.ig.Ignores(path) {
			enqueue(path)
			return nil
		}
	}
	return nil
}

func (c *CLI) formatFile(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s", filePath)
	}
	switch strings.ToLower(filepath.Ext(filePath)) {
	// TODO: ".json5"
	case ".json":
		return api.FormatJSON(string(bytes), c.opts)
	default:
		return api.Format(string(bytes), c.opts)
	}
}

func (c *CLI) formatAndWriteFile(path string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic while formatting file %s: %v", path, r)
		}
	}()
	start := time.Now()
	res, err := c.formatFile(path)
	if err != nil {
		return fmt.Errorf("error formatting file %s: %w", path, err)
	}

	err = os.WriteFile(path, []byte(res), 0644)
	if err != nil {
		return fmt.Errorf("unable to write file: %s", path)
	}

	duration := time.Since(start)
	fmt.Printf("%s %dns\n", path, duration.Nanoseconds())

	return nil
}
