package ignore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckPattern(t *testing.T) {
	t.Parallel()
	tests := []struct {
		pattern string
		result  bool
	}{
		{pattern: "# serves as a comment", result: false},
		{pattern: "", result: false},
		{pattern: "\n\t   \n", result: false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.result, checkPattern(tt.pattern))
		})
	}
}
