package parser_test

import (
	"testing"

	"github.com/janczizikow/bijou/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	program, err := parser.Parse(`throw "hello";`)
	require.NoError(t, err)

	assert.Len(t, program.Statements, 1)
}

func TestParseErrors(t *testing.T) {
	_, err := parser.Parse("{")
	require.Error(t, err)

	assert.Contains(t, err.Error(), "SyntaxError: '(' expected")
}
