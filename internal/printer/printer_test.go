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
			input:    "for(;;);",
			expected: "for (;;);\n",
			options:  printer.Options{Semi: true},
		},
		{
			input:    "for(true;;);",
			expected: "for (true;;);\n",
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
