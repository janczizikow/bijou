package printer_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/janczizikow/bijou/internal/parser"
	"github.com/janczizikow/bijou/internal/printer"
)

func TestPrint(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected string
		options  printer.Options
	}{
		{
			input: `function xyz() {
  let x = 5;
  if (true) {
x = 4;
  }
  return x;
}`,
			expected: `function xyz() {
  let x = 5;
  if (true) {
    x = 4;
  }
  return x;
}
`,
			options: printer.Options{Semi: true},
		},
		{
			input:    "return;",
			expected: "return\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "let x;",
			expected: "let x\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "let x = 5;",
			expected: "let x = 5\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "x ? y : z",
			expected: "x ? y : z\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "{}",
			expected: "{\n}\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "if (true) { let x = 5; }",
			expected: "if (true) {\n  let x = 5\n}\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "if (true) { let x = 5; }",
			expected: "if (true) {\n  let x = 5\n}\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "!true;",
			expected: "!true\n",
			options:  printer.Options{Semi: false},
		},
		{
			input:    "function add(a, b) { return a + b; }",
			expected: "function add(a, b) {\n  return a + b;\n}\n",
			options:  printer.Options{Semi: true},
		},
		{
			input:    "1 + 2 * 3 - 4",
			expected: "1 + 2 * 3 - 4\n",
			options:  printer.Options{Semi: false},
		},
	}

	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tt.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tt.options)
			assert.Equal(t, tt.expected, string(bytes))
		})
	}
}

func TestPrint_PropertyAccess(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected string
		options  printer.Options
	}{
		{"object.propertyName", "object.propertyName;\n", printer.Options{Semi: true}},
		{"object?.propertyName", "object?.propertyName;\n", printer.Options{Semi: true}},
		{"object.#propertyName", "object.#propertyName;\n", printer.Options{Semi: true}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}

func TestPrint_ImportStatment(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		expected string
		options  printer.Options
	}{
		{
			name:     "side effect import",
			input:    `import "module-name"`,
			expected: "import \"module-name\";\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "default import",
			input:    `import defaultExport from "module-name"`,
			expected: "import defaultExport from \"module-name\";\n",
			options:  printer.Options{Semi: true},
		},

		{
			name:     "namespace import with alias",
			input:    `import * as name from "module-name"`,
			expected: "import * as name from \"module-name\";\n",
			options:  printer.Options{Semi: true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}

func TestPrint_Arrays(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input    string
		expected string
		options  printer.Options
	}{
		{"[1,2,3]", "[1, 2, 3];\n", printer.Options{Semi: true}},
		{"[[1,2],3]", "[[1, 2], 3];\n", printer.Options{Semi: true}},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}

func TestPrint_ForLoops(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		expected string
		options  printer.Options
	}{
		{
			name:     "empty for loop",
			input:    "for(;;);",
			expected: "for (;;);\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "for only initalizer",
			input:    "for(true;;);",
			expected: "for (true;;);\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "classic for loop",
			input:    "for (let i = 0; i < arr.length; i++) {}",
			expected: "for (let i = 0; i < arr.length; i++) {\n};\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "for...in loop",
			input:    `for (const property in object) { console.log(property); }`,
			expected: "for (const property in object) {\n  console.log(property);\n}\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "for...of loop",
			input:    `for (const element of array) { console.log(element); }`,
			expected: "for (const element of array) {\n  console.log(element);\n}\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "for await...of loop",
			input:    `for await (const element of array) { console.log(element); }`,
			expected: "for await (const element of array) {\n  console.log(element);\n}\n",
			options:  printer.Options{Semi: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}

func TestPrint_SwithStatment(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		expected string
		options  printer.Options
	}{
		{
			name: "simple",
			input: `switch (expr) {
  case "Oranges":
    console.log("Oranges are $0.59 a pound.");
    break;
  case "Mangoes":
  case "Papayas":
    console.log("Mangoes and papayas are $2.79 a pound.");
    break;
  default:
    console.log("Out of stock");
}`,
			expected: `switch (expr) {
  case "Oranges":
    console.log("Oranges are $0.59 a pound.");
    break;
  case "Mangoes":
  case "Papayas":
    console.log("Mangoes and papayas are $2.79 a pound.");
    break;
  default:
    console.log("Out of stock");
}
`,
			options: printer.Options{Semi: true},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}

func TestPrint_LabeledStatement(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		input    string
		expected string
		options  printer.Options
	}{
		{
			name:     "label",
			input:    "foo: x;",
			expected: "foo: x;\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "label without semicolons",
			input:    "foo: x;",
			expected: "foo: x\n",
			options:  printer.Options{Semi: false},
		},
		{
			name:     "nested labels",
			input:    "foo: bar: x",
			expected: "foo: bar: x;\n",
			options:  printer.Options{Semi: true},
		},
		{
			name:     "labeled loop with break and continue",
			input:    "outer: for (const x of xs) { break outer; continue outer; }",
			expected: "outer: for (const x of xs) {\n  break outer;\n  continue outer;\n}\n",
			options:  printer.Options{Semi: true},
		},
		// FIXME: idents
		// {
		// 	name:     "label inside block",
		// 	input:    "{ foo: x; }",
		// 	expected: "{\n  foo: x;\n}\n",
		// 	options:  printer.Options{Semi: true},
		// },
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}

func TestPrint_BlankLines(t *testing.T) {
	t.Parallel()
	semi := printer.Options{Semi: true}
	cases := []struct {
		name     string
		input    string
		expected string
		options  printer.Options
	}{
		{
			name:     "no blank line stays none",
			input:    "let a = 1;\nlet b = 2;",
			expected: "let a = 1;\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "one blank line is kept",
			input:    "let a = 1;\n\nlet b = 2;",
			expected: "let a = 1;\n\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "several blank lines collapse to one",
			input:    "let a = 1;\n\n\n\nlet b = 2;",
			expected: "let a = 1;\n\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "whitespace-only line counts as blank",
			input:    "let a = 1;\n   \nlet b = 2;",
			expected: "let a = 1;\n\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "statements on one line are not separated",
			input:    "let a = 1; let b = 2;",
			expected: "let a = 1;\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "no blank line at block start or end",
			input:    "if (a) {\n\n  b;\n\n}",
			expected: "if (a) {\n  b;\n}\n",
			options:  semi,
		},
		{
			name:     "indentation alone is not a blank line",
			input:    "let a = 1;\n    let b = 2;",
			expected: "let a = 1;\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "CRLF: a single line break is not a blank line",
			input:    "let a = 1;\r\nlet b = 2;",
			expected: "let a = 1;\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "CRLF: a blank line is kept",
			input:    "let a = 1;\r\n\r\nlet b = 2;",
			expected: "let a = 1;\n\nlet b = 2;\n",
			options:  semi,
		},
		{
			name:     "leading whitespace does not add a blank line",
			input:    "  let a = 1;",
			expected: "let a = 1;\n",
			options:  semi,
		},
		{
			name:     "leading blank lines are dropped",
			input:    "\n\nlet a = 1;",
			expected: "let a = 1;\n",
			options:  semi,
		},
		{
			name:     "statement right after a block gets no blank line",
			input:    "if (a) {\n  b;\n}\nlet d = 4;",
			expected: "if (a) {\n  b;\n}\nlet d = 4;\n",
			options:  semi,
		},
		{
			name:     "blank line after a block is kept",
			input:    "if (a) {\n  b;\n}\n\nlet d = 4;",
			expected: "if (a) {\n  b;\n}\n\nlet d = 4;\n",
			options:  semi,
		},
		{
			name:     "blank line inside a block is kept",
			input:    "if (a) {\n  b;\n\n  c;\n}",
			expected: "if (a) {\n  b;\n\n  c;\n}\n",
			options:  semi,
		},
		{
			name:     "blank line inside a function body is kept",
			input:    "function f() {\n  let x = 1;\n\n  return x;\n}",
			expected: "function f() {\n  let x = 1;\n\n  return x;\n}\n",
			options:  semi,
		},
		{
			name:     "blank line inside a case body is kept",
			input:    "switch (x) {\ncase 1:\n  a;\n\n  b;\n}",
			expected: "switch (x) {\n  case 1:\n    a;\n\n    b;\n}\n",
			options:  semi,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			program, err := parser.Parse(tc.input)
			require.NoError(t, err)

			bytes := printer.Print(program, tc.options)
			assert.Equal(t, tc.expected, string(bytes))
		})
	}
}
