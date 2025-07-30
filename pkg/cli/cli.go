package cli

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
}

func New(config api.FormatOptions) CLI {
	return CLI{FormatOptions: config}
}

func (c CLI) Run(args []string) int {
	if err := c.formatFiles(args); err != nil {
		fmt.Println(err.Error())
		return Error
	}
	return Success
}

func (c CLI) formatFiles(paths []string) error {
	// TODO: get other ignore options
	ig := ignore.New([]string{"*/.git", ".svn", ".hg", "**/node_modules"})

	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("No files matching the pattern were found: %s", path)
		}

		if fileInfo.IsDir() {
			err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && !ig.Ignores(p) {
					return c.formatAndWriteFile(p)
				}
				return nil
			})
			return err
		} else if !ig.Ignores(path) {
			return c.formatAndWriteFile(path)
		}
	}

	return nil
}

func (c CLI) formatFile(filePath string) (string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Unable to read file: %s", filePath)
	}
	return api.Format(string(bytes), c.FormatOptions)
}

func (c CLI) formatAndWriteFile(path string) error {
	start := time.Now()
	res, err := c.formatFile(path)
	if err != nil {
		return fmt.Errorf("Error formatting file %s: %w", path, err)
	}

	err = os.WriteFile(path, []byte(res), 0644)
	if err != nil {
		return fmt.Errorf("Unable to write file: %s", path)
	}

	duration := time.Since(start)
	fmt.Printf("%s %dns\n", path, duration.Nanoseconds())

	return nil
}
