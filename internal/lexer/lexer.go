package lexer

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type Lexer struct {
	input string // original input

	Token   Token  // current token
	Literal string // the token literal value, if any

	chr          rune // current character
	position     int  // current position in input (points to current char)
	readPosition int  // current reading position in input (after current char)
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input}
	l.step()
	return l
}

func (l *Lexer) step() {
	if l.readPosition >= len(l.input) {
		l.chr = -1 // Use -1 to indicate end of file
		l.readPosition = len(l.input)
	} else {
		r, width := utf8.DecodeRuneInString(l.input[l.position:])
		if width == 0 {
			l.chr = -1 // Use -1 to indicate end of file
		}

		l.chr = r
		l.position = l.readPosition
		l.readPosition += width
	}
}

func (l *Lexer) skipWhitespace() {
	// TODO: improve this
	switch l.chr {
	case ' ':
		l.step()
	}
}

func (l *Lexer) Next() {
	l.skipWhitespace()

	switch l.chr {
	case -1:
		l.Token = EOF
		l.Literal = ""
	case '=':
		// TODO: check if it could be == or ===
		l.Token = Assignment
		l.Literal = "="
		l.step()
	case ';':
		l.Token = Semi
		l.Literal = ";"
		l.step()
	default:
		if isIdentifierStart(l.chr) {
			l.Literal = l.scanIdentifier()
			l.Token = LookupIdentifier(l.Literal)
			fmt.Println(l.Literal)
		} else {
			l.Token = SyntaxError
		}
	}
}

func (l *Lexer) scanIdentifier() string {
	position := l.position
	for isIdentifierStart(l.chr) {
		l.step()
	}
	return l.input[position:l.position]
}

func (l *Lexer) print() {
	fmt.Printf("input: %s char: %v start: %v end %v\n", l.input, string(l.chr), l.position, l.readPosition)
}

// func isWhitespace() bool {

// }

func isLetter(ch rune) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

func isIdentifierStart(chr rune) bool {
	return chr == '$' || chr == '_' || chr == '\\' ||
		'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z' ||
		chr >= utf8.RuneSelf && unicode.IsLetter(chr)
}
