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
			input: "(a + b) * c",
			expected: []ast.Statement{
				&ast.ExpressionStatement{
					Expr: &ast.BinaryExpression{
						Operator: lexer.Multiplication,
						Left: &ast.ParenthesizedExpression{
							Expr: &ast.BinaryExpression{
								Operator: lexer.Plus,
								Left:     &ast.Identifier{Name: "a"},
								Right:    &ast.Identifier{Name: "b"},
							},
						},
						Right: &ast.Identifier{Name: "c"},
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

			assignment, ok := expr.Expr.(*ast.Assignment)
			require.True(t, ok, fmt.Sprintf("Expected *ast.Assignment, got %T", expr.Expr))

			assert.Equal(t, expected, assignment)
		})
	}
}

func TestParseBreak(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{"break", false, &ast.BreakStatement{}},
		{"break label", false, &ast.BreakStatement{Label: new("label")}},
		{"{break}", false, &ast.BlockStatement{Statements: []ast.Statement{&ast.BreakStatement{}}}},
		{"break this", true, nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)
				assert.Equal(t, tc.expected, program.Statements[0])
			}
		})
	}
}

func TestParseContinue(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{"continue", false, &ast.ContinueStatement{}},
		{"continue label", false, &ast.ContinueStatement{Label: new("label")}},
		{"{continue}", false, &ast.BlockStatement{Statements: []ast.Statement{&ast.ContinueStatement{}}}},
		{"continue this", true, nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)
				assert.Equal(t, tc.expected, program.Statements[0])
			}
		})
	}
}

func TestParseThrow(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{"throw;", true, nil},
		{"throw", true, nil},
		{"throw err", false, &ast.ThrowStatement{Argument: &ast.Identifier{Name: "err"}}},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)
				assert.Equal(t, tc.expected, program.Statements[0])
			}
		})
	}
}

func TestParseDoWhile(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{
			name:  "valid do...while",
			input: `do {i += 1;result += i;} while (i < 5);`,
			expected: &ast.DoWhileStatement{
				Condition: &ast.BinaryExpression{
					Operator: lexer.LT,
					Left:     &ast.Identifier{Name: "i"},
					Right:    &ast.NumberLiteral{Value: "5"},
				},
				Body: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expr: &ast.Assignment{
								Operator: lexer.AdditionAssignment,
								Left:     &ast.Identifier{Name: "i"},
								Right:    &ast.NumberLiteral{Value: "1"},
							},
						},
						&ast.ExpressionStatement{
							Expr: &ast.Assignment{
								Operator: lexer.AdditionAssignment,
								Left:     &ast.Identifier{Name: "result"},
								Right:    &ast.Identifier{Name: "i"},
							},
						},
					},
				},
			},
		},
		{
			name:    "missing block",
			input:   "do while (i < 5)",
			wantErr: true,
		},
		{
			name:    "missing while",
			input:   "do {} (i < 5)",
			wantErr: true,
		},
		{
			name:    "missing condition",
			input:   "do {} while i < 5",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)
				assert.Equal(t, tc.expected, program.Statements[0])
			}
		})
	}
}

func TestParsePropertyAccessors(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		wantErr  bool
		expected ast.Expression
	}{
		{
			"object.propertyName",
			false,
			&ast.PropertyAccessExpression{
				Object: &ast.Identifier{Name: "object"},
				Field:  &ast.Identifier{Name: "propertyName"},
			},
		},
		{
			"object.#propertyName",
			false,
			&ast.PropertyAccessExpression{
				Object: &ast.Identifier{Name: "object"},
				Field:  &ast.Identifier{Name: "propertyName", Private: true},
			},
		},
		{
			"object?.propertyName",
			false,
			&ast.PropertyAccessExpression{
				Object:   &ast.Identifier{Name: "object"},
				Field:    &ast.Identifier{Name: "propertyName"},
				Optional: true,
			},
		},
		// TODO: allow keywords
		// {
		// 	"object.this",
		// 	false,
		// 	&ast.PropertyAccessExpression{
		// 		Object: &ast.PropertyAccessExpression{
		// 			Object: &ast.Identifier{Name: "object"},
		// 			Field:  nil, // TODO: new node??
		// 		},
		// 		Field: &ast.Identifier{Name: "otherName"},
		// 	},
		// },
		{
			"object.propertyName.otherName",
			false,
			&ast.PropertyAccessExpression{
				Object: &ast.PropertyAccessExpression{
					Object: &ast.Identifier{Name: "object"},
					Field:  &ast.Identifier{Name: "propertyName"},
				},
				Field: &ast.Identifier{Name: "otherName"},
			},
		},
		{
			"object.propertyName?.otherName",
			false,
			&ast.PropertyAccessExpression{
				Object: &ast.PropertyAccessExpression{
					Object: &ast.Identifier{Name: "object"},
					Field:  &ast.Identifier{Name: "propertyName"},
				},
				Field:    &ast.Identifier{Name: "otherName"},
				Optional: true,
			},
		},
		{
			`object["propertyName"]`,
			false,
			&ast.PropertyAccessExpression{
				Object:   &ast.Identifier{Name: "object"},
				Field:    &ast.StringLiteral{Value: "propertyName"},
				Computed: true,
			},
		},
		{
			"object[x + y]",
			false,
			&ast.PropertyAccessExpression{
				Object: &ast.Identifier{Name: "object"},
				Field: &ast.BinaryExpression{
					Operator: lexer.Plus,
					Left:     &ast.Identifier{Name: "x"},
					Right:    &ast.Identifier{Name: "y"},
				},
				Computed: true,
			},
		},
		{"a.-b", true, nil},
		{"a.[1]", true, nil},
		// {"a.1", true, nil}, // FIXME: fails because ".1" is lexed as a numeric literal
		{"a.", true, nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)

				expr, ok := program.Statements[0].(*ast.ExpressionStatement)
				require.True(t, ok, fmt.Sprintf("Expected *ast.ExpressionStatement, got %T", program.Statements[0]))
				require.NotNil(t, expr)

				got, ok := expr.Expr.(*ast.PropertyAccessExpression)
				require.True(t, ok, fmt.Sprintf("Expected *ast.PropertyAccessExpression, got %T", expr.Expr))

				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestParseUpdateExpression(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		wantErr  bool
		expected ast.Expression
	}{
		{"x++", false, &ast.UpdateExpression{Operator: lexer.Increment, Argument: &ast.Identifier{Name: "x"}}},
		{"--x", false, &ast.UpdateExpression{Operator: lexer.Decrement, Argument: &ast.Identifier{Name: "x"}, Prefix: true}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)

				expr, ok := program.Statements[0].(*ast.ExpressionStatement)
				require.True(t, ok, fmt.Sprintf("Expected *ast.ExpressionStatement, got %T", program.Statements[0]))
				require.NotNil(t, expr)

				got, ok := expr.Expr.(*ast.UpdateExpression)
				require.True(t, ok, fmt.Sprintf("Expected *ast.UpdateExpression, got %T", expr.Expr))

				assert.Equal(t, tc.expected, got)
			}
		})
	}
}

func TestParseFunctionCall(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected ast.Expression
	}{
		{
			input: "square(5)",
			expected: &ast.FunctionCall{
				Fn:   &ast.Identifier{Name: "square"},
				Args: []ast.Expression{&ast.NumberLiteral{Value: "5"}},
			},
		},
		{
			input: "apply(1,2,3)",
			expected: &ast.FunctionCall{
				Fn: &ast.Identifier{Name: "apply"},
				Args: []ast.Expression{
					&ast.NumberLiteral{Value: "1"},
					&ast.NumberLiteral{Value: "2"},
					&ast.NumberLiteral{Value: "3"},
				},
			},
		},
		{
			input: "console.log('test')",
			expected: &ast.FunctionCall{
				Fn: &ast.PropertyAccessExpression{
					Object: &ast.Identifier{Name: "console"},
					Field:  &ast.Identifier{Name: "log"},
				},
				Args: []ast.Expression{
					&ast.StringLiteral{Value: "test"},
				},
			},
		},
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

			got, ok := expr.Expr.(*ast.FunctionCall)
			require.True(t, ok, fmt.Sprintf("Expected *ast.FunctionCall, got %T", expr.Expr))

			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestParseArrays(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected ast.Expression
	}{
		{"[1, 2, 3]", &ast.ArrayLiteralExpression{Elements: []ast.Expression{
			&ast.NumberLiteral{Value: "1"},
			&ast.NumberLiteral{Value: "2"},
			&ast.NumberLiteral{Value: "3"},
		}}},
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

			got, ok := expr.Expr.(*ast.ArrayLiteralExpression)
			require.True(t, ok, fmt.Sprintf("Expected *ast.ArrayLiteralExpression, got %T", expr.Expr))

			assert.Equal(t, tc.expected, got)
		})
	}

	t.Run("unexpected EOF", func(t *testing.T) {
		t.Parallel()
		_, err := parser.Parse("[1,2")
		require.Error(t, err)
		assert.ErrorContains(t, err, "SyntaxError")
	})
}

func TestImportStatement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		expected ast.Statement
	}{
		{
			name:     "side effect import",
			input:    `import "module-name";`,
			expected: &ast.ImportStatement{Path: &ast.StringLiteral{Value: "module-name"}},
		},
		{
			name:  "default import",
			input: `import defaultExport from "module-name";`,
			expected: &ast.ImportStatement{
				DefaultImport: &ast.Identifier{Name: "defaultExport"},
				Path:          &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "namespace import",
			input: `import * as name from "module-name";`,
			expected: &ast.ImportStatement{
				NamespaceImport: &ast.Identifier{Name: "name"},
				Path:            &ast.StringLiteral{Value: "module-name"},
			},
		},
		// TODO:
		// {
		// 	name:  "named import",
		// 	input: `import { foo, bar } from "module-name";`,
		// 	expected: &ast.ImportStatement{
		// 		Path: &ast.StringLiteral{Value: "module-name"},
		// 	},
		// },
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			require.NoError(t, err)
			assert.Len(t, program.Statements, 1)
			assert.Equal(t, tc.expected, program.Statements[0])
		})
	}
}

func TestParseForStatements(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		expected ast.Statement
	}{
		{
			name:     "empty for loop",
			input:    "for (;;);",
			expected: &ast.ForStatement{Body: &ast.EmptyStatement{}},
		},
		{
			name:  "for loop",
			input: `for (let i = 0; i < arr.length; i++) {}`,
			expected: &ast.ForStatement{
				Init: &ast.VariableDeclaration{
					Token: lexer.LET,
					Name:  &ast.Identifier{Name: "i"},
					Value: &ast.NumberLiteral{Value: "0"},
				},
				Condition: &ast.BinaryExpression{
					Operator: lexer.LT,
					Left:     &ast.Identifier{Name: "i"},
					Right: &ast.PropertyAccessExpression{
						Object: &ast.Identifier{Name: "arr"},
						Field:  &ast.Identifier{Name: "length"},
					},
				},
				Afterthought: &ast.UpdateExpression{
					Operator: lexer.Increment,
					Argument: &ast.Identifier{Name: "i"},
				},
				Body: &ast.BlockStatement{Statements: []ast.Statement{}},
			},
		},
		{
			name:  "for...in",
			input: `for (const property in object) {}`,
			expected: &ast.ForInStatement{
				Variable: &ast.VariableDeclaration{
					Token: lexer.CONST,
					Name:  &ast.Identifier{Name: "property"},
				},
				Object: &ast.Identifier{Name: "object"},
				Body:   &ast.BlockStatement{Statements: []ast.Statement{}},
			},
		},
		{
			name:  "for...of",
			input: `for (const element of array) {}`,
			expected: &ast.ForOfStatement{
				Variable: &ast.VariableDeclaration{
					Token: lexer.CONST,
					Name:  &ast.Identifier{Name: "element"},
				},
				Iterable: &ast.Identifier{Name: "array"},
				Body:     &ast.BlockStatement{Statements: []ast.Statement{}},
			},
		},
		{
			name:  "for await...of",
			input: `for await (const element of array) {}`,
			expected: &ast.ForOfStatement{
				Await: true,
				Variable: &ast.VariableDeclaration{
					Token: lexer.CONST,
					Name:  &ast.Identifier{Name: "element"},
				},
				Iterable: &ast.Identifier{Name: "array"},
				Body:     &ast.BlockStatement{Statements: []ast.Statement{}},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, err := parser.Parse(tc.input)
			require.NoError(t, err)
			assert.Len(t, program.Statements, 1)
			assert.Equal(t, tc.expected, program.Statements[0])
		})
	}
}

func TestParseNewExpression(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		wantErr  bool
		expected ast.Expression
	}{
		{"new", false, &ast.NewExpression{}},
		{"new constructor", false, &ast.NewExpression{Constructor: &ast.Identifier{Name: "constructor"}}},
		{"new constructor()", false, &ast.NewExpression{Constructor: &ast.Identifier{Name: "constructor"}, Arguments: []ast.Expression{}}},
		{"new constructor(a,b)", false, &ast.NewExpression{Constructor: &ast.Identifier{Name: "constructor"}, Arguments: []ast.Expression{
			&ast.Identifier{Name: "a"},
			&ast.Identifier{Name: "b"},
		}}},
		// {"new object.property(a,b)", false,
		// 	&ast.NewExpression{
		// 		Constructor: &ast.PropertyAccessExpression{
		// 			Object: &ast.Identifier{Name: "object"},
		// 			Field:  &ast.Identifier{Name: "property"},
		// 		},
		// 		Arguments: []ast.Expression{
		// 			&ast.Identifier{Name: "a"},
		// 			&ast.Identifier{Name: "b"},
		// 		},
		// 	},
		// },
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			if tc.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "SyntaxError")
			} else {
				require.NoError(t, err)
				assert.Len(t, program.Statements, 1)

				expr, ok := program.Statements[0].(*ast.ExpressionStatement)
				require.True(t, ok, fmt.Sprintf("Expected *ast.ExpressionStatement, got %T", program.Statements[0]))
				require.NotNil(t, expr)

				got, ok := expr.Expr.(*ast.NewExpression)
				require.True(t, ok, fmt.Sprintf("Expected *ast.NewExpression, got %T", expr.Expr))

				assert.Equal(t, tc.expected, got)
			}
		})
	}
}
