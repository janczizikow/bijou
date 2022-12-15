package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLookupIdentifier(t *testing.T) {
	cases := []struct {
		identifier    string
		expectedToken Token
	}{
		{"var", VAR},
		{"x", Identifier},
	}

	for _, tt := range cases {
		assert.EqualValues(t, tt.expectedToken, LookupIdentifier(tt.identifier))
	}
}
