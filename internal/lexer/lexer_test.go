package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifier(t *testing.T) {
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
			assert.EqualValues(t, NUMBER_LITERAL, l.Token)
			assert.Equal(t, tt.expected, l.Literal)
		})
	}
}

func TestInvalidNumberLiterals(t *testing.T) {
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

func TestTokens(t *testing.T) {
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
		{"-", Subtraction},
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
