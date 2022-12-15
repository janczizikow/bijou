package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTokens(t *testing.T) {
	input := `var x = true;`
	cases := []struct {
		literal  string
		expected Token
	}{
		{"var", VAR},
		{"x", Identifier},
		{"=", Assignment},
		{"true", TRUE},
		{";", Semi},
	}

	l := NewLexer(input)
	assert.Equal(t, l.position, 0)
	assert.Equal(t, l.readPosition, 1)

	for _, tt := range cases {
		t.Run(tt.literal, func(t *testing.T) {
			l.Next()
			assert.EqualValues(t, tt.expected, l.Token)
		})
	}
}
