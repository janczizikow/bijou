package parser_test

import (
	"fmt"
	"testing"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/lexer"
	"github.com/janczizikow/bijou/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected []ast.Statement
	}{
		{
			input: "-a * b",
			expected: []ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.BinaryExpression{
						Operator: lexer.Multiplication,
						Left: &ast.UnaryExpression{
							Operator: lexer.Minus,
							Right:    &ast.Identifier{Name: "a"},
						},
						Right: &ast.Identifier{Name: "b"},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, program.Statements)
		})
	}
}

func TestAssignemnts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		operator lexer.Token
	}{
		{"assignment (=)", lexer.Assignment},
		{"multiplication assignment (*=)", lexer.MultiplicationAssignment},
		{"remainder assignment (%=)", lexer.RemainderAssignment},
		{"addition assignment (+=)", lexer.AdditionAssignment},
		{"subtraction assignment (-=)", lexer.SubtractionAssignment},
		{"left shift assignment (<<=)", lexer.LeftShiftAssignment},
		{"right shift assignment (>>=)", lexer.RightShiftAssignment},
		{"unsigned right shift assignment (>>>=)", lexer.UnsignedRightShiftAssignment},
		{"bitwise AND assignment (&=)", lexer.BitwiseANDAssignment},
		{"bitwise XOR assignment (^=)", lexer.BitwiseXORAssignment},
		{"bitwise OR assignment (|=)", lexer.BitwiseORAssignment},
		{"exponentiation assignment (**=)", lexer.ExponentiationAssignment},
		{"logical AND assignment (&&=)", lexer.LogicalANDAssignment},
		{"logical OR assignment (||=)", lexer.LogicalORAssignment},
		{"nullish coalescing (??=)", lexer.NullishCoalescingAssignment},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := fmt.Sprintf("a %s b", tc.operator.String())
			expected := &ast.Assignment{
				Operator: tc.operator,
				Left:     &ast.Identifier{Name: "a"},
				Right:    &ast.Identifier{Name: "b"},
			}

			program, err := parser.Parse(input)
			require.NoError(t, err)
			assert.Len(t, program.Statements, 1)

			expr, ok := program.Statements[0].(*ast.ExpressionStatement)
			require.True(t, ok, fmt.Sprintf("Expected *ast.ExpressionStatement, got %T", program.Statements[0]))
			require.NotNil(t, expr)

			assignment, ok := expr.Expression.(*ast.Assignment)
			require.True(t, ok, fmt.Sprintf("Expected *ast.Assignment, got %T", expr.Expression))

			assert.Equal(t, expected, assignment)
		})
	}
}

// func TestParseErrors(t *testing.T) {
// 	_, err := parser.Parse("{")
// 	require.Error(t, err)

// 	assert.Contains(t, err.Error(), "SyntaxError: '}' expected")
// }

func TestPropertyAccessors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected ast.Expression
	}{
		{
			"object.propertyName",
			&ast.PropertyAccessExpression{
				Object: &ast.Identifier{Name: "object"},
				Field:  &ast.Identifier{Name: "propertyName"},
			},
		},
		{
			"object.#propertyName",
			&ast.PropertyAccessExpression{
				Object: &ast.Identifier{Name: "object"},
				Field:  &ast.Identifier{Name: "propertyName", Private: true},
			},
		},
		{
			"object?.propertyName",
			&ast.PropertyAccessExpression{
				Object:   &ast.Identifier{Name: "object"},
				Field:    &ast.Identifier{Name: "propertyName"},
				Optional: true,
			},
		},
		// TODO:
		// {
		// 	"object[propertyName]",
		// 	&ast.PropertyAccessExpression{
		// 		Object:   &ast.Identifier{Name: "object"},
		// 		Field:    &ast.Identifier{Name: "propertyName"},
		// 		Computed: true,
		// 	},
		// },
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)
			assert.Len(t, program.Statements, 1)

			expr, ok := program.Statements[0].(*ast.ExpressionStatement)
			require.True(t, ok, fmt.Sprintf("Expected *ast.ExpressionStatement, got %T", program.Statements[0]))
			require.NotNil(t, expr)

			got, ok := expr.Expression.(*ast.PropertyAccessExpression)
			require.True(t, ok, fmt.Sprintf("Expected *ast.PropertyAccessExpression, got %T", expr.Expression))

			assert.Equal(t, tc.expected, got)
		})
	}
}
