package lexer

import (
	"fmt"
)

type Token uint8

const (
	Illegal Token = iota
	EOF

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#hashbang_comments
	HashBang
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#identifiers
	Identifier
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Classes/Private_elements
	PrivateIdentifier

	// Arithmetic operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#arithmetic_operators

	Exponentiation
	Multiplication
	Division
	Remainder
	Plus
	Minus

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#increment_and_decrement

	Increment
	Decrement

	// Assignment operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#assignment_operators

	Assignment
	MultiplicationAssignment
	DivisionAssignment
	RemainderAssignment
	AdditionAssignment
	SubtractionAssignment
	LeftShiftAssignment
	RightShiftAssignment
	UnsignedRightShiftAssignment
	BitwiseANDAssignment
	BitwiseORAssignment
	BitwiseXORAssignment
	BitwiseShiftLeftAssignment
	BitwiseShiftRightsAssignment
	BitwiseUnsignedShiftRightAssignment
	ExponentiationAssignment
	LogicalANDAssignment
	LogicalORAssignment
	NullishCoalescingAssignment

	// Logical operators

	// Equality operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#equality_operators

	Equality
	StrictEquality
	Inequality
	StrictInequality

	// Arrow function notation "=>"
	Arrow

	// Bitwise shift operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#bitwise_shift_operators

	BitwiseShiftLeft
	BitwiseShiftRight
	BitwiseUnsignedShiftRight

	// Binary bitwise operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#binary_bitwise_operators

	BitwiseAND
	BitwiseOR
	BitwiseXOR
	BitwiseNOT

	// Binary logical operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#binary_logical_operators

	LogicalAND
	LogicalOR
	NullishCoalescing

	// Relational operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#relational_operators

	LT
	GT
	LTE
	GTE

	// punctuation / delimiters

	Semi
	Colon
	Comma
	Dot
	LParen
	RParen
	LBrace
	RBrace
	LBracket
	RBracket
	Exclamation
	Question
	OptionalChaining

	// reserved keywords
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#keywords

	BREAK
	CASE
	CATCH
	CLASS
	CONST
	CONTINUE
	DEBUGGER
	DEFAULT
	DELETE
	DO
	ELSE
	EXPORT
	EXTENDS
	FALSE
	FINALLY
	FOR
	FUNCTION
	IF
	IMPORT
	IN
	INSTANCEOF
	NEW
	NULL
	RETURN
	SUPER
	SWITCH
	THIS
	THROW
	TRUE
	TRY
	TYPEOF
	VAR
	VOID
	WHILE
	WITH

	// strict mode reserved keywords

	LET
	STATIC
	YIELD
	AWAIT
	IMPLEMENTS
	INTERFACE
	PACKAGE
	PRIVATE
	PROTECTED
	PUBLIC

	// literals
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#literals

	NUMERIC_LITERAL
	BigIntLiteral
	STRING_LITERAL
	REGEXP_LITERAL
)

// reserved keywords
var keywords = map[string]Token{
	"break":      BREAK,
	"case":       CASE,
	"catch":      CATCH,
	"class":      CLASS,
	"const":      CONST,
	"continue":   CONTINUE,
	"debugger":   DEBUGGER,
	"default":    DEFAULT,
	"delete":     DELETE,
	"do":         DO,
	"else":       ELSE,
	"export":     EXPORT,
	"extends":    EXTENDS,
	"false":      FALSE,
	"finally":    FINALLY,
	"for":        FOR,
	"function":   FUNCTION,
	"if":         IF,
	"import":     IMPORT,
	"in":         IN,
	"instanceof": INSTANCEOF,
	"new":        NEW,
	"null":       NULL,
	"return":     RETURN,
	"super":      SUPER,
	"switch":     SWITCH,
	"this":       THIS,
	"throw":      THROW,
	"true":       TRUE,
	"try":        TRY,
	"typeof":     TYPEOF,
	"var":        VAR,
	"void":       VOID,
	"while":      WHILE,
	"with":       WITH,
}

// strict mode reserved keywords
var strictModeKeywords = map[string]Token{
	"let":        LET,
	"static":     STATIC,
	"yield":      YIELD,
	"implements": IMPLEMENTS,
	"interface":  INTERFACE,
	"package":    PACKAGE,
	"private":    PRIVATE,
	"protected":  PROTECTED,
	"public":     PUBLIC,
}

func lookupIdentifier(identifier string) Token {
	if token, exist := keywords[identifier]; exist {
		return token
	}

	if token, exist := strictModeKeywords[identifier]; exist {
		return token
	}

	return Identifier
}

func (t Token) String() string {
	if s, ok := tokenToString[t]; ok {
		return s
	}

	return fmt.Sprintf("unknown_token(%d)", t)
}
