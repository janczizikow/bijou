package ignore

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnores(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path     string
		patterns []string
		ignored  bool
	}{
		{"node_modules/foo.js", []string{}, true},
		{"a/b/node_modules/foo.js", []string{}, true},
		{"./node_modules/foo.js", []string{}, true},
		{"foo/.git/config", []string{}, true},
		{".svn/entries", []string{}, true},
		{"src/.hg/x", []string{}, true},
		{"src/node_modules_utils/foo.ts", []string{}, false},
		{"src/index.js", []string{}, false},
		{".env", []string{".env*"}, true},
		{".env.example", []string{".env*", "!.env.example"}, false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			ig := New(tc.patterns...)
			assert.Equal(t, tc.ignored, ig.Ignores(tc.path))
		})
	}
}

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		pattern string
		result  bool
	}{
		{name: "comment", pattern: "# serves as a comment", result: false},
		{name: "empty", pattern: "", result: false},
		{name: "only white space", pattern: "\n\t   \n", result: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.result, checkPattern(strings.TrimSpace(tc.pattern)))
		})
	}
}
