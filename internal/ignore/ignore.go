package ignore

import (
	"regexp"
	"strings"
)

type Ignorer struct {
	patterns []string
}

func New(patterns []string) *Ignorer {
	return &Ignorer{patterns: patterns}
}

func checkPattern(pattern string) bool {
	onlyWhitespace := regexp.MustCompile(`^\s+$`).MatchString(pattern)
	isBlank := pattern == ""
	isComment := strings.HasPrefix(pattern, "#")

	return !onlyWhitespace && !isBlank && !isComment
}

func (i *Ignorer) Add(pattern string) {
	// TODO: check if pattern is already there?
	if checkPattern(pattern) {
		i.patterns = append(i.patterns, pattern)
	}
}

func (i *Ignorer) Ignores(pattern string) bool {
	// TODO: actual implementation
	return false
}

func (i *Ignorer) Filter(patterns []string) []string {
	filtered := make([]string, 0)

	for _, pattern := range patterns {
		if !i.Ignores(pattern) {
			filtered = append(filtered, pattern)
		}
	}

	return filtered
}
