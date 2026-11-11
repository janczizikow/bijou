package ignore

import (
	"path"
	"path/filepath"
	"strings"
)

type pattern struct {
	glob    string
	negated bool
}

type Ignorer struct {
	patterns []pattern
}

var alwaysIgnoredDirectories = []string{".git", ".svn", ".hg", "node_modules"}

func New(pathOrPattern ...string) *Ignorer {
	ig := &Ignorer{patterns: make([]pattern, 0)}
	for _, dir := range alwaysIgnoredDirectories {
		ig.Add(dir)
	}
	for _, pattern := range pathOrPattern {
		if checkPattern(pattern) {
			ig.Add(pattern)
		}
	}
	return ig
}

func checkPattern(pattern string) bool {
	if pattern == "" || strings.HasPrefix(pattern, "#") {
		return false
	}
	_, err := filepath.Match(pattern, "")
	return err == nil
}

func (i *Ignorer) Add(pathOrPattern string) {
	trimmed := strings.TrimSpace(pathOrPattern)
	if checkPattern(trimmed) {
		pat := pattern{glob: trimmed}
		if after, ok := strings.CutPrefix(trimmed, "!"); ok {
			pat = pattern{glob: after, negated: true}
		}
		i.patterns = append(i.patterns, pat)
	}
}

func (i *Ignorer) Ignores(pattern string) bool {
	for part := range strings.SplitSeq(filepath.ToSlash(pattern), "/") {
		ignored := false
		for _, pattern := range i.patterns {
			matches, err := path.Match(pattern.glob, part)
			if err != nil || !matches {
				continue
			}
			ignored = !pattern.negated
		}
		if ignored {
			return true
		}
	}
	return false
}
