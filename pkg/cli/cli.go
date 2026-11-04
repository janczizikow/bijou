package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/janczizikow/bijou/internal/ignore"
	"github.com/janczizikow/bijou/pkg/api"
)

const (
	Success = iota
	Error
)

type CLI struct {
	FormatOptions api.FormatOptions
	ig            *ignore.Ignorer
}

func New(config api.FormatOptions) CLI {
	return CLI{FormatOptions: config, ig: ignore.New()}
}

func (c CLI) Run(args []string) int {
	if errs := c.formatFiles(args); len(errs) > 0 {
		for _, err := range errs {
			fmt.Println(err.Error())
		}
		return Error
	}
	return Success
}

func (c CLI) formatFiles(paths []string) []error {
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

func (c CLI) walk(paths []string, jobs chan string) error {
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

func (c CLI) formatFile(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to read file: %s", filePath)
	}
	return api.Format(string(bytes), c.FormatOptions)
}

func (c CLI) formatAndWriteFile(path string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic while formatting file %s: %v", r)
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
