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
	cases := []struct {
		name, input string
		wantErr     bool
		expected    ast.Expression
	}{
		{
			name:  "valid JSON",
			input: `{"user": {"id": 1, "name": "test", "isVerified": true, "skills": ["foo", "bar", "baz"], "deletedAt": null}}`,
			expected: &ast.ObjectLiteralExpression{
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
			},
		},
		{
			name:  "negative number - valid",
			input: `{"test": -42}`,
			expected: &ast.ObjectLiteralExpression{
				Properties: []ast.ObjectProperty{
					{Key: &ast.StringLiteral{Value: "test"}, Value: &ast.NumberLiteral{Value: "-42"}},
				},
			},
		},
		{
			name:  "floating point and exponential notation - valid",
			input: `{"test": 1.23e+5}`,
			expected: &ast.ObjectLiteralExpression{
				Properties: []ast.ObjectProperty{
					{Key: &ast.StringLiteral{Value: "test"}, Value: &ast.NumberLiteral{Value: "1.23e+5"}},
				},
			},
		},
		{
			name:    "trailing comma in object - errors",
			input:   `{"test": true,}`,
			wantErr: true,
		},
		{
			name:    "trailing comma in array - errors",
			input:   `{"test": [1,2,]}`,
			wantErr: true,
		},
		// numbers
		{
			name:    "positive sign on number - errors",
			input:   `{"test": +5}`,
			wantErr: true,
		},
		{
			name:    "leading zero - errors",
			input:   `{"test": 05}`,
			wantErr: true,
		},
		{
			name:    "leading decimal point - errors",
			input:   `{"test": .5}`,
			wantErr: true,
		},
		{
			name:    "trailing decimal point - errors",
			input:   `{"test": 5.}`,
			wantErr: true,
		},
		{
			name:    "unquoted key - errors",
			input:   `{test: 123}`,
			wantErr: true,
		},
		{
			name:    "unassigned/undefined value - errors",
			input:   `{"test": undefined}`,
			wantErr: true,
		},
		{
			name:    "single quotes - errors",
			input:   "{'test': true}",
			wantErr: true,
		},
		{
			name:    "binary number - errors",
			input:   `{"test": 0b00101}`,
			wantErr: true,
		},
		{
			name:    "octal number - errors",
			input:   `{"test": 0o12345}`,
			wantErr: true,
		},
		{
			name:    "hexadecimal number - errors",
			input:   `{"test": 0xDEAD_BEEF}`,
			wantErr: true,
		},
		{
			name:    "number with numeric seprator - errors",
			input:   `{"test": 1_000}`,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parser.ParseJSON(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}
