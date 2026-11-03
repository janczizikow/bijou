package parser_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/lexer"
	"github.com/janczizikow/bijou/internal/parser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stripSpans zeros every lexer.Span recursively
func stripSpans(v any) {
	var walk func(reflect.Value)
	walk = func(rv reflect.Value) {
		switch rv.Kind() {
		case reflect.Pointer, reflect.Interface:
			if rv.IsNil() {
				return
			}
			walk(rv.Elem())
		case reflect.Struct:
			if rv.Type() == reflect.TypeFor[lexer.Span]() {
				rv.Set(reflect.Zero(rv.Type()))
			} else {
				for _, v := range rv.Fields() {
					walk(v)
				}
			}
		case reflect.Slice:
			for i := range rv.Len() {
				walk(rv.Index(i))
			}
		}
	}
	walk(reflect.ValueOf(v))
}

// parseStripped parses input and zeroes all spans, so tests can compare
// against hand-built nodes that don't have any.
func parseStripped(input string) (*ast.Program, error) {
	program, err := parser.Parse(input)
	if err != nil {
		return program, err
	}
	stripSpans(program)
	return program, nil
}

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
			program, err := parseStripped(tc.input)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, program.Statements)
		})
	}
}

func TestAssignments(t *testing.T) {
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

			program, err := parseStripped(input)
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
		{"break label", false, &ast.BreakStatement{Label: &ast.Identifier{Name: "label"}}},
		{"{break}", false, &ast.BlockStatement{Statements: []ast.Statement{&ast.BreakStatement{}}}},
		{"break this", true, nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parseStripped(tc.input)
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
		{"continue label", false, &ast.ContinueStatement{Label: &ast.Identifier{Name: "label"}}},
		{"{continue}", false, &ast.BlockStatement{Statements: []ast.Statement{&ast.ContinueStatement{}}}},
		{"continue this", true, nil},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parseStripped(tc.input)
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
			program, err := parseStripped(tc.input)
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

func TestParseSwitchStatement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{
			name:  "switch - no cases",
			input: `switch (x > 5) {}`,
			expected: &ast.SwitchStatement{
				Test: &ast.BinaryExpression{
					Operator: lexer.GT,
					Left:     &ast.Identifier{Name: "x"},
					Right:    &ast.NumberLiteral{Value: "5"},
				},
				Cases: []ast.Case{},
			},
		},
		{
			name: "switch - cases",
			input: `switch (x > 5) {
		case 0:
			break;
		case 1:
			break;
		default:
			break;
		}`,
			expected: &ast.SwitchStatement{
				Test: &ast.BinaryExpression{Operator: lexer.GT, Left: &ast.Identifier{Name: "x"}, Right: &ast.NumberLiteral{Value: "5"}},
				Cases: []ast.Case{
					{Value: &ast.NumberLiteral{Value: "0"}, Body: []ast.Statement{&ast.BreakStatement{}}},
					{Value: &ast.NumberLiteral{Value: "1"}, Body: []ast.Statement{&ast.BreakStatement{}}},
					{Body: []ast.Statement{&ast.BreakStatement{}}},
				},
			},
		},
		{
			name: "switch - cases fallthrough",
			input: `switch (x > 5) {
		case 0:
			break;
		case 1:
		case 2:
			i++;
		default:
			i++;
			break;
		}`,
			expected: &ast.SwitchStatement{
				Test: &ast.BinaryExpression{Operator: lexer.GT, Left: &ast.Identifier{Name: "x"}, Right: &ast.NumberLiteral{Value: "5"}},
				Cases: []ast.Case{
					{Value: &ast.NumberLiteral{Value: "0"}, Body: []ast.Statement{&ast.BreakStatement{}}},
					{Value: &ast.NumberLiteral{Value: "1"}, Body: []ast.Statement{}},
					{Value: &ast.NumberLiteral{Value: "2"}, Body: []ast.Statement{
						&ast.ExpressionStatement{
							Expr: &ast.UpdateExpression{Operator: lexer.Increment, Argument: &ast.Identifier{Name: "i"}},
						}}},
					{Body: []ast.Statement{
						&ast.ExpressionStatement{
							Expr: &ast.UpdateExpression{Operator: lexer.Increment, Argument: &ast.Identifier{Name: "i"}},
						},
						&ast.BreakStatement{},
					}},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			program, err := parseStripped(tc.input)
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
			program, err := parseStripped(tc.input)
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

func TestParseLabeledStatement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{
			name:  "label",
			input: "foo: x;",
			expected: &ast.LabeledStatement{
				Label: &ast.Identifier{Name: "foo"},
				Body:  &ast.ExpressionStatement{Expr: &ast.Identifier{Name: "x"}},
			},
		},
		{
			name:  "whitespace before colon",
			input: "foo : x",
			expected: &ast.LabeledStatement{
				Label: &ast.Identifier{Name: "foo"},
				Body:  &ast.ExpressionStatement{Expr: &ast.Identifier{Name: "x"}},
			},
		},
		{
			name:  "nested labels",
			input: "foo: bar: x",
			expected: &ast.LabeledStatement{
				Label: &ast.Identifier{Name: "foo"},
				Body: &ast.LabeledStatement{
					Label: &ast.Identifier{Name: "bar"},
					Body:  &ast.ExpressionStatement{Expr: &ast.Identifier{Name: "x"}},
				},
			},
		},
		{
			name:  "labeled loop with break",
			input: "outer: for (;;) { break outer; }",
			expected: &ast.LabeledStatement{
				Label: &ast.Identifier{Name: "outer"},
				Body: &ast.ForStatement{
					Body: &ast.BlockStatement{Statements: []ast.Statement{
						&ast.BreakStatement{Label: &ast.Identifier{Name: "outer"}},
					}},
				},
			},
		},
		{
			name:    "missing body",
			input:   "foo:",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, err := parseStripped(tc.input)
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

func TestParseTryCatch(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{
			name:  "try...catch",
			input: `try { a; } catch (e) { b; }`,
			expected: &ast.TryCatchStatement{
				Block: ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{Expr: &ast.Identifier{Name: "a"}},
				}},
				Catch: &ast.Catch{
					Parameter: &ast.Identifier{Name: "e"},
					Block: ast.BlockStatement{Statements: []ast.Statement{
						&ast.ExpressionStatement{Expr: &ast.Identifier{Name: "b"}},
					}},
				},
			},
		},
		{
			name:  "try...catch without binding",
			input: `try {} catch {}`,
			expected: &ast.TryCatchStatement{
				Block: ast.BlockStatement{Statements: []ast.Statement{}},
				Catch: &ast.Catch{Block: ast.BlockStatement{Statements: []ast.Statement{}}},
			},
		},
		{
			name:  "try...finally",
			input: `try {} finally { a; }`,
			expected: &ast.TryCatchStatement{
				Block: ast.BlockStatement{Statements: []ast.Statement{}},
				Finally: &ast.BlockStatement{Statements: []ast.Statement{
					&ast.ExpressionStatement{Expr: &ast.Identifier{Name: "a"}},
				}},
			},
		},
		{
			name:  "try...catch...finally",
			input: `try {} catch (e) {} finally {}`,
			expected: &ast.TryCatchStatement{
				Block: ast.BlockStatement{Statements: []ast.Statement{}},
				Catch: &ast.Catch{
					Parameter: &ast.Identifier{Name: "e"},
					Block:     ast.BlockStatement{Statements: []ast.Statement{}},
				},
				Finally: &ast.BlockStatement{Statements: []ast.Statement{}},
			},
		},
		{
			name:  "throw inside try",
			input: `try { throw err } catch (e) {}`,
			expected: &ast.TryCatchStatement{
				Block: ast.BlockStatement{Statements: []ast.Statement{
					&ast.ThrowStatement{Argument: &ast.Identifier{Name: "err"}},
				}},
				Catch: &ast.Catch{
					Parameter: &ast.Identifier{Name: "e"},
					Block:     ast.BlockStatement{Statements: []ast.Statement{}},
				},
			},
		},
		{
			name:  "nested try",
			input: `try { try {} finally {} } catch {}`,
			expected: &ast.TryCatchStatement{
				Block: ast.BlockStatement{Statements: []ast.Statement{
					&ast.TryCatchStatement{
						Block:   ast.BlockStatement{Statements: []ast.Statement{}},
						Finally: &ast.BlockStatement{Statements: []ast.Statement{}},
					},
				}},
				Catch: &ast.Catch{Block: ast.BlockStatement{Statements: []ast.Statement{}}},
			},
		},
		{
			name:    "missing catch or finally",
			input:   "try {}",
			wantErr: true,
		},
		{
			name:    "missing try block",
			input:   "try catch (e) {}",
			wantErr: true,
		},
		{
			name:    "missing catch block",
			input:   "try {} catch (e)",
			wantErr: true,
		},
		{
			name:    "missing finally block",
			input:   "try {} finally",
			wantErr: true,
		},
		{
			name:    "unclosed catch parameter",
			input:   "try {} catch (e {}",
			wantErr: true,
		},
		{
			name:    "catch without try",
			input:   "catch (e) {}",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, err := parseStripped(tc.input)
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
		// {
		// 	"import.meta.url",
		// 	false,
		// 	&ast.PropertyAccessExpression{
		// 		Object:   &ast.Identifier{Name: "import"},
		// 		Field:    &ast.Identifier{Name: "meta"},
		// 		Optional: false,
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

			program, err := parseStripped(tc.input)
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

			program, err := parseStripped(tc.input)
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

			program, err := parseStripped(tc.input)
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

func TestArrowFunction(t *testing.T) {
	t.Parallel()
	// TODO:
	t.Skip("unimplemented")

	input := "(a, b) => a + b"
	expected := &ast.ArrowFunction{
		Args: []ast.Expression{
			&ast.Identifier{Name: "a"},
			&ast.Identifier{Name: "b"},
		},
		Body: &ast.BinaryExpression{
			Operator: lexer.Plus,
			Left:     &ast.Identifier{Name: "a"},
			Right:    &ast.Identifier{Name: "b"},
		},
	}
	program, err := parseStripped(input)
	require.NoError(t, err)
	assert.Len(t, program.Statements, 1)

	expr, ok := program.Statements[0].(*ast.ExpressionStatement)
	require.True(t, ok, fmt.Sprintf("Expected *ast.ExpressionStatement, got %T", program.Statements[0]))
	require.NotNil(t, expr)

	got, ok := expr.Expr.(*ast.ArrowFunction)
	require.True(t, ok, fmt.Sprintf("Expected *ast.ArrowFunction, got %T", expr.Expr))

	assert.Equal(t, expected, got)
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

			program, err := parseStripped(tc.input)
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

// named builds a named import specifier: `imported as local`.
func named(imported, local string) *ast.ImportSpecifier {
	return &ast.ImportSpecifier{
		Imported: &ast.Identifier{Name: imported},
		Local:    &ast.Identifier{Name: local},
	}
}

func TestImportStatement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		wantErr  bool
		expected ast.Statement
	}{
		{
			name:     "side effect import",
			input:    `import "module-name";`,
			expected: &ast.ImportDeclaration{Specifiers: []*ast.ImportSpecifier{}, Path: &ast.StringLiteral{Value: "module-name"}},
		},
		{
			name:  "default import",
			input: `import defaultExport from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Kind: ast.ImportKindDefault, Local: &ast.Identifier{Name: "defaultExport"}},
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "namespace import",
			input: `import * as name from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Kind: ast.ImportKindNamespace, Local: &ast.Identifier{Name: "name"}},
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named import",
			input: `import { foo, bar } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Imported: &ast.Identifier{Name: "foo"}, Local: &ast.Identifier{Name: "foo"}},
					{Imported: &ast.Identifier{Name: "bar"}, Local: &ast.Identifier{Name: "bar"}},
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named imports with alias",
			input: `import { foo as f, bar as b } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Imported: &ast.Identifier{Name: "foo"}, Local: &ast.Identifier{Name: "f"}},
					{Imported: &ast.Identifier{Name: "bar"}, Local: &ast.Identifier{Name: "b"}},
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named import - single",
			input: `import { foo } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("foo", "foo")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named import - mixed aliased and plain",
			input: `import { foo, bar as b } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("foo", "foo"), named("bar", "b")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named import - trailing comma",
			input: `import { foo, bar, } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("foo", "foo"), named("bar", "bar")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named import - empty braces",
			input: `import {} from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name: "named import - multiline",
			input: `import {
				foo,
				bar as b,
			} from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("foo", "foo"), named("bar", "b")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "default and named import",
			input: `import defaultExport, { foo, bar as b } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Kind: ast.ImportKindDefault, Local: &ast.Identifier{Name: "defaultExport"}},
					named("foo", "foo"),
					named("bar", "b"),
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "default and namespace import",
			input: `import defaultExport, * as name from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Kind: ast.ImportKindDefault, Local: &ast.Identifier{Name: "defaultExport"}},
					{Kind: ast.ImportKindNamespace, Local: &ast.Identifier{Name: "name"}},
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "single quoted path",
			input: `import { foo } from 'module-name';`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("foo", "foo")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "no semicolon",
			input: `import { foo } from "module-name"`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("foo", "foo")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		// contextual keywords are valid identifiers
		{
			name:  "contextual keyword 'from' as default import",
			input: `import from from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					{Kind: ast.ImportKindDefault, Local: &ast.Identifier{Name: "from"}},
				},
				Path: &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "contextual keyword 'from' as named import",
			input: `import { from } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("from", "from")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "contextual keyword 'as' as named import",
			input: `import { as } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("as", "as")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "contextual keyword 'as' aliased to 'as'",
			input: `import { as as as } from "module-name";`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{named("as", "as")},
				Path:       &ast.StringLiteral{Value: "module-name"},
			},
		},
		{
			name:  "named default",
			input: `import { default as x } from "m"`,
			expected: &ast.ImportDeclaration{
				Specifiers: []*ast.ImportSpecifier{
					named("default", "x"),
				},
				Path: &ast.StringLiteral{Value: "m"},
			},
		},
		// TODO: allow string literals as imported names
		// {`import { "string name" as x } from "m"`}
		// `default` is a reserved word, so it can only be imported with an alias
		{name: "default keyword requires alias", input: `import { default } from "module-name";`, wantErr: true},
		{name: "default keyword requires alias - among others", input: `import { foo, default, bar } from "module-name";`, wantErr: true},
		{name: "missing from - named", input: `import { foo } "module-name";`, wantErr: true},
		{name: "missing from - default", input: `import foo "module-name";`, wantErr: true},
		{name: "missing from - namespace", input: `import * as ns "module-name";`, wantErr: true},
		{name: "missing path", input: `import { foo } from;`, wantErr: true},
		{name: "missing path - EOF", input: `import { foo } from`, wantErr: true},
		{name: "unclosed braces", input: `import { foo from "module-name";`, wantErr: true},
		{name: "unclosed braces - EOF", input: `import { foo`, wantErr: true},
		{name: "unexpected EOF after opening brace", input: `import {`, wantErr: true},
		{name: "missing alias", input: `import { foo as } from "module-name";`, wantErr: true},
		{name: "missing comma between names", input: `import { foo bar } from "module-name";`, wantErr: true},
		{name: "leading comma", input: `import { , foo } from "module-name";`, wantErr: true},
		{name: "double comma", input: `import { foo,, bar } from "module-name";`, wantErr: true},
		{name: "namespace without as", input: `import * from "module-name";`, wantErr: true},
		{name: "namespace without name", input: `import * as from "module-name";`, wantErr: true},
		{name: "namespace inside braces", input: `import { * as ns } from "module-name";`, wantErr: true},
		{name: "namespace before default", input: `import * as ns, def from "module-name";`, wantErr: true},
		{name: "namespace and named", input: `import * as ns, { foo } from "module-name";`, wantErr: true},
		{name: "default after named", input: `import { foo }, def from "module-name";`, wantErr: true},
		{name: "trailing comma after default", input: `import def, from "module-name";`, wantErr: true},
		{name: "alias must be an identifier", input: `import { foo as 1 } from "module-name";`, wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			program, err := parseStripped(tc.input)
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
			program, err := parseStripped(tc.input)
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
		// {"new constructor()", false, &ast.NewExpression{Constructor: &ast.Identifier{Name: "constructor"}, Arguments: []ast.Expression{}}},
		// {"new constructor(a,b)", false, &ast.NewExpression{Constructor: &ast.Identifier{Name: "constructor"}, Arguments: []ast.Expression{
		// 	&ast.Identifier{Name: "a"},
		// 	&ast.Identifier{Name: "b"},
		// }}},
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

			program, err := parseStripped(tc.input)
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
