package parser_test

import (
	"testing"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseJSON(t *testing.T) {
	t.Parallel()
	data := `{"user": {"id": 1, "name": "test", "isVerified": true, "skills": ["foo", "bar", "baz"], "deletedAt": null}}`
	expected := &ast.ObjectLiteralExpression{
		Properties: []ast.ObjectProperty{
			{
				Key: &ast.StringLiteral{Value: "user"},
				Value: &ast.ObjectLiteralExpression{
					Properties: []ast.ObjectProperty{
						{Key: &ast.StringLiteral{Value: "id"}, Value: &ast.NumberLiteral{Value: "1"}},
						{Key: &ast.StringLiteral{Value: "name"}, Value: &ast.StringLiteral{Value: "test"}},
						{Key: &ast.StringLiteral{Value: "isVerified"}, Value: &ast.BooleanLiteral{Value: true}},
						{Key: &ast.StringLiteral{Value: "skills"}, Value: &ast.ArrayLiteralExpression{Elements: []ast.Expression{
							&ast.StringLiteral{Value: "foo"},
							&ast.StringLiteral{Value: "bar"},
							&ast.StringLiteral{Value: "baz"},
						}}},
						{Key: &ast.StringLiteral{Value: "deletedAt"}, Value: &ast.NullLiteral{}},
					},
				},
			},
		},
	}
	got, err := parser.ParseJSON(data)
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
