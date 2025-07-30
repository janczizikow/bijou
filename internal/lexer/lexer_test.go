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
		// {"''", "", true},
		// {"'hello world'\n", "hello world", true},
		// {"'hello", "", false},
		{"hey \"mark\"", "hey \"mark\"", true},
	}

	for _, tt := range cases {
		t.Run(tt.token, func(t *testing.T) {
			t.Parallel()
			l := New(tt.token)
			if tt.valid {
				assert.EqualValues(t, STRING_LITERAL, l.Token)
			} else {
				assert.EqualValues(t, SyntaxError, l.Token)
			}
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
