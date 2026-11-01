package ignore

import (
	"path/filepath"
	"regexp"
	"strings"
)

type Ignorer struct {
	patterns    []string
	directories map[string]struct{}
}

var alwaysIgnoredDirectories = []string{".git", ".svn", ".hg", "node_modules"}

func New(patterns ...string) *Ignorer {
	dirs := make(map[string]struct{}, 0)
	for _, dir := range alwaysIgnoredDirectories {
		dirs[dir] = struct{}{}
	}
	for _, pattern := range patterns {
		if checkPattern(pattern) {
			dirs[pattern] = struct{}{}
		}
	}
	return &Ignorer{directories: dirs}
}

func checkPattern(pattern string) bool {
	onlyWhitespace := regexp.MustCompile(`^\s+$`).MatchString(pattern)
	isBlank := pattern == ""
	isComment := strings.HasPrefix(pattern, "#")

	return !onlyWhitespace && !isBlank && !isComment
}

func (i *Ignorer) Add(pattern string) {
	// TODO: glob patterns
	if checkPattern(pattern) {
		i.patterns = append(i.patterns, pattern)
	}
}

func (i *Ignorer) Ignores(pattern string) bool {
	for path := range strings.SplitSeq(filepath.ToSlash(pattern), "/") {
		if _, ok := i.directories[path]; ok {
			return true
		}
	}
	return false
}
