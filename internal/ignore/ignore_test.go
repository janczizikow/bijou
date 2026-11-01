package ignore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIgnores(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path    string
		ignored bool
	}{
		{"node_modules/foo.js", true},
		{"a/b/node_modules/foo.js", true},
		{"./node_modules/foo.js", true},
		{"foo/.git/config", true},
		{".svn/entries", true},
		{"src/.hg/x", true},
		{"src/node_modules_utils/foo.ts", false},
		{"src/index.js", false},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()
			ig := New()
			assert.Equal(t, tc.ignored, ig.Ignores(tc.path))
		})
	}
}

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	cases := []struct {
		pattern string
		result  bool
	}{
		{pattern: "# serves as a comment", result: false},
		{pattern: "", result: false},
		{pattern: "\n\t   \n", result: false},
	}

	for _, tc := range cases {
		t.Run(tc.pattern, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.result, checkPattern(tc.pattern))
		})
	}
}
