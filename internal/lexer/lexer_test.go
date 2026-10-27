package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifier(t *testing.T) {
	t.Parallel()
	cases := []struct {
		token    string
		expected string
		valid    bool
	}{
		{"_", "_", true},
		{"$", "$", true},
		{"test", "test", true},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			l := New(tt.token)
			if tt.valid {
				assert.EqualValues(t, Identifier, l.Token)
				assert.Equal(t, tt.expected, l.Literal)
			} else {
				assert.EqualValues(t, SyntaxError, l.Token)
			}
		})
	}
}

func TestStringLiterals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		token    string
		expected string
		valid    bool
	}{
		{"''", "", true},
		{"'hello world'", "hello world", true},
		{"'hello", "", false},
		// {"'hello\nworld'", "", false},
		// {"'hello\nworld'", "", false},
		{`"hello \"mark\""`, `hello \"mark\"`, true},
		{`'hello \'mark\''`, `hello \'mark\'`, true},
		{`"\\"`, "\\\\", true},
		{`'c:\\temp'`, `c:\\temp`, true},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			l := New(tt.token)
			if tt.valid {
				assert.EqualValues(t, STRING_LITERAL, l.Token)
				assert.Equal(t, tt.expected, l.Literal)
			} else {
				assert.EqualValues(t, SyntaxError, l.Token)
			}
		})
	}
}

func TestNumberLiterals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		token    string
		expected string
	}{
		{"0", "0"},
		{"000", "000"},
		{"010", "010"},
		{"123", "123"},
		{"987", "987"},
		{"0000", "0000"},
		{"0123", "0123"},
		{"0123.4567", "0123.4567"},
		{"0987", "0987"},
		{"0987.6543", "0987.6543"},
		{"01289", "01289"},
		{"01289.345", "01289.345"},
		{".123", ".123"},
		{"0.", "0"},
		{"123.", "123"},
		{"999999999", "999999999"},
		{"9999999999", "9999999999"},
		{"99999999999", "99999999999"},
		// binary notation
		{"123456789123456789", "123456789123456789"},
		{"0b00101", "0b00101"},
		{"0b1011101011101011101011101011101011101", "0b1011101011101011101011101011101011101"},
		// octal notation
		{"0o12345", "0o12345"},
		{"0O12345", "0O12345"},
		{"0o1234567654321", "0o1234567654321"},
		{"0O1234567654321", "0O1234567654321"},
		// hexadecimal notation
		{"0x12345678", "0x12345678"},
		{"0xFEDCBA987", "0xFEDCBA987"},
		{"0x000012345678", "0x000012345678"},
		{"0x123456781234", "0x123456781234"},
		// separating with _
		{"123_456", "123_456"},
		{"1_000_000", "1_000_000"},
		{"0b1010_0001", "0b1010_0001"},
		{"0o7_7_7", "0o7_7_7"},
		{"0xDEAD_BEEF", "0xDEAD_BEEF"},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			l := New(tt.token)
			assert.EqualValues(t, NUMERIC_LITERAL, l.Token)
			assert.Equal(t, tt.expected, l.Literal)
		})
	}
}

func TestBigIntLiterals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		token    string
		expected string
	}{
		{"123456789123456789n", "123456789123456789n"},
		{"0n", "0n"},
		{"0o777n", "0o777n"},
		{"0x123n", "0x123n"},
		{"0b101n", "0b101n"},
		{"1_000_000n", "1_000_000n"},
		{"0b1010_0001n", "0b1010_0001n"},
		{"0o7_7_7n", "0o7_7_7n"},
		{"0xDEAD_BEEFn", "0xDEAD_BEEFn"},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			l := New(tt.token)
			assert.EqualValues(t, BigIntLiteral, l.Token)
			assert.Equal(t, tt.expected, l.Literal)
		})
	}
}

func TestInvalidNumberLiterals(t *testing.T) {
	t.Parallel()
	cases := []struct {
		token string
		msg   string
	}{
		{"0b", "SyntaxError: Invalid binary literal"},
		{"0b ", "SyntaxError: Invalid binary literal"},
		{"0b;", "SyntaxError: Invalid binary literal"},
		{"0b2", "SyntaxError: Invalid binary literal"},
		{"0o", "SyntaxError: Invalid octal literal"},
		{"0o ", "SyntaxError: Invalid octal literal"},
		{"0o;", "SyntaxError: Invalid octal literal"},
		{"0o9", "SyntaxError: Invalid octal literal"},
		{"0x", "SyntaxError: Invalid hexadecimal literal"},
		{"0x ", "SyntaxError: Invalid hexadecimal literal"},
		{"0x;", "SyntaxError: Invalid hexadecimal literal"},
		{"0xg", "SyntaxError: Invalid hexadecimal literal"},
		{"1_", "SyntaxError: Numeric separator must be followed by a digit"},
		{"1__0", "SyntaxError: Numeric separator must be followed by a digit"},
		{"0b1_", "SyntaxError: Numeric separator must be followed by a digit"},
		{"0x1_", "SyntaxError: Numeric separator must be followed by a digit"},
		{"0_1", "SyntaxError: Numeric separator can not be used after leading 0."},
		{"07n", "SyntaxError: BigInt literals cannot have a leading zero"},
		{"1.5n", "SyntaxError: Invalid or unexpected token"},
		{"1e2n", "SyntaxError: Invalid or unexpected token"},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			assert.PanicsWithValue(t, LexerPanic{Msg: tt.msg}, func() {
				New(tt.token)
			})
		})
	}
}

func TestPropertyAccessors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected []Token
	}{
		{"object.propertyName", []Token{Identifier, Dot, Identifier}},
		{"object?.propertyName", []Token{Identifier, OptionalChaning, Identifier}},
		{`object[expression]`, []Token{Identifier, LBracket, Identifier, RBracket}},
		{`object.#privateProperty`, []Token{Identifier, Dot, PrivateIdentifier}},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			l := New(tc.input)
			out := make([]Token, 0)
			for l.Token != EOF {
				out = append(out, l.Token)
				l.Next()
			}
			assert.Equal(t, tc.expected, out)
		})
	}
}

func TestTokens(t *testing.T) {
	t.Parallel()
	cases := []struct {
		token    string
		expected Token
	}{
		{"", EOF},
		{"#!", HashBang},
		{"x", Identifier},
		// Arithmetic operators
		{"**", Exponentiation},
		{"*", Multiplication},
		{"/", Division},
		{"%", Remainder},
		{"+", Plus},
		{"-", Minus},
		{"++", Increment},
		{"--", Decrement},
		// Assignment operators
		{"=", Assignment},
		{"*=", MultiplicationAssignment},
		{"**=", ExponentiationAssignment},
		{"/=", DivisionAssignment},
		{"%=", RemainderAssignment},
		{"+=", AdditionAssignment},
		{"-=", SubtractionAssignment},
		{"&=", BitwiseANDAssignment},
		{"|=", BitwiseORAssignment},
		{"^=", BitwiseXORAssignment},
		{"<<=", LeftShiftAssignment},
		{">>=", RightShiftAssignment},
		{">>>=", UnsignedRightShiftAssignment},
		{"&&=", LogicalANDAssignment},
		{"||=", LogicalORAssignment},
		{"??=", NullishCoalescingAssignment},
		// Equality operators
		{"==", Equality},
		{"===", StrictEquality},
		{"!=", Inequality},
		{"!==", StrictInequality},
		// Arrow fn notation
		{"=>", Arrow},
		// Bitwise shift operators
		{"<<", BitwiseShiftLeft},
		{">>", BitwiseShiftRight},
		{">>>", BitwiseUnsignedShiftRight},
		// Binary bitwise operators
		{"&", BitwiseAND},
		{"|", BitwiseOR},
		{"^", BitwiseXOR},
		{"~", BitwiseNOT},
		// Binary logical operators
		{"&&", LogicalAND},
		{"||", LogicalOR},
		{"??", NullishCoalescing},
		// Relational operators
		{"<", LT},
		{">", GT},
		{"<=", LTE},
		{">=", GTE},
		// punctuation
		{":", Colon},
		{",", Comma},
		{".", Dot},
		{"!", Exclamation},
		{"?", Question},
		{"(", LParen},
		{")", RParen},
		{"{", LBrace},
		{"}", RBrace},
		{"[", LBracket},
		{"]", RBracket},
		{";", Semi},
		// keywords
		{"break", BREAK},
		{"case", CASE},
		{"catch", CATCH},
		{"class", CLASS},
		{"const", CONST},
		{"continue", CONTINUE},
		{"debugger", DEBUGGER},
		{"default", DEFAULT},
		{"delete", DELETE},
		{"do", DO},
		{"else", ELSE},
		{"export", EXPORT},
		{"extends", EXTENDS},
		{"false", FALSE},
		{"finally", FINALLY},
		{"for", FOR},
		{"function", FUNCTION},
		{"if", IF},
		{"import", IMPORT},
		{"in", IN},
		{"instanceof", INSTANCEOF},
		{"new", NEW},
		{"null", NULL},
		{"return", RETURN},
		{"super", SUPER},
		{"switch", SWITCH},
		{"this", THIS},
		{"throw", THROW},
		{"true", TRUE},
		{"try", TRY},
		{"typeof", TYPEOF},
		{"var", VAR},
		{"void", VOID},
		{"while", WHILE},
		{"with", WITH},
		// strict mode reserved keywords
		{"let", LET},
		{"static", STATIC},
		{"yield", YIELD},
		{"implements", IMPLEMENTS},
		{"interface", INTERFACE},
		{"package", PACKAGE},
		{"private", PRIVATE},
		{"protected", PROTECTED},
		{"public", PUBLIC},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			l := New(tt.token)
			assert.Equal(t, tt.expected, l.Token)
		})
	}
}
