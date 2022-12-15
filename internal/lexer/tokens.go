package lexer

import "fmt"

type Token uint8

const (
	EOF = iota
	SyntaxError

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#hashbang_comments
	HashBang
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#identifiers
	Identifier

	// arithmetic operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#arithmetic_operators
	Exponentiation
	Multiplication
	Division
	Remainder
	Plus
	Subtraction

	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#increment_and_decrement
	Increment
	Decrement

	// arithmetic operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#assignment_operators
	Assignment
	Equality

	// Relational operators
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators#relational_operators
	LT
	GT
	LTE
	GTE

	// punctuation
	Semi

	// keywords
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

	// literals
	// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#literals
	// TODO:
)

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

func LookupIdentifier(identifier string) Token {
	fmt.Printf("looking up if %v is a keyword\n", identifier)
	if token, exist := keywords[identifier]; exist {
		return token
	}

	return Identifier
}
