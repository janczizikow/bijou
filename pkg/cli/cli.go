package cli

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/janczizikow/bijou/internal/ignore"
	"github.com/janczizikow/bijou/pkg/api"
)

const (
	SUCCESS = iota
	ERROR
)

func Run(options []string) int {
	if err := formatFiles(options); err != nil {
		fmt.Println(err.Error())
		return ERROR
	}

	return SUCCESS
}

func formatFiles(paths []string) error {
	// TODO: get other ignore options
	ig := ignore.New([]string{"*/.git", ".svn", ".hg", "**/node_modules"})

	for _, path := range paths {
		fileInfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("No files matching the pattern were found: %s", path)
		}

		if fileInfo.IsDir() {
			err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
				if !d.IsDir() && !ig.Ignores(p) {
					res, err := formatFile(path)
					fmt.Println(res)
					return err
				}
				return nil
			})
			return err
		} else if !ig.Ignores(path) {
			res, err := formatFile(path)
			fmt.Println(res)
			return err
		}
	}

	return nil
}

func formatFile(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("Unable to read file: %s", filePath)
	}

	return api.Format(string(bytes))
}
