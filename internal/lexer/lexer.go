package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer represents a lexical analyzer for JavaScript source code.
// It takes a string as input and converts it into a sequence of tokens.
type Lexer struct {
	input string // original input

	Token   Token  // current token
	Literal string // the token literal value, if any

	current rune
	start   int
	end     int
}

type LexerPanic struct {
	Msg string
}

func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.step()
	l.Next()
	return l
}

func (l *Lexer) step() {
	if l.end < len(l.input) {
		r, width := utf8.DecodeRuneInString(l.input[l.end:])
		if width == 0 {
			l.end = len(l.input)
			l.current = -1
		}
		l.start = l.end
		l.end += width
		l.current = r
	} else {
		l.start = l.end
		l.end = len(l.input)
		l.current = -1
	}
}

func (l *Lexer) peek() rune {
	r, width := utf8.DecodeRuneInString(l.input[l.end:])
	if width == 0 {
		return -1
	}

	return r
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.current) {
		l.step()
	}
}

func (l *Lexer) Next() {
	l.skipWhitespace()

	switch l.current {
	case -1:
		l.Token = EOF

	case '#':
		if l.start == 0 && l.peek() == '!' {
			l.Token = HashBang
			l.Literal = "#!"
			l.step()
			l.step()
		}
		// TODO: syntax error?

	case '!':
		// '!' or '!=' or '!=='
		switch l.peek() {
		case '=':
			l.step()
			switch l.peek() {
			case '=':
				l.step()
				l.Token = StrictInequality
			default:
				l.Token = Inequality
			}
		default:
			l.Token = Exclamation
		}
		l.step()

	case '+':
		// '+' or '+=' or '++'
		switch l.peek() {
		case '+':
			l.step()
			l.Token = Increment
		case '=':
			l.step()
			l.Token = AdditionAssignment
		default:
			l.Token = Plus
		}
		l.step()
	case '-':
		// '-' or '-=' or '--'
		switch l.peek() {
		case '-':
			l.step()
			l.Token = Decrement
		case '=':
			l.step()
			l.Token = SubtractionAssignment
		default:
			l.Token = Subtraction
		}
		l.step()
	case '*':
		// '*' or '*=' or '**' or '**='
		switch l.peek() {
		case '*':
			l.step()
			switch l.peek() {
			case '=':
				l.step()
				l.Token = ExponentiationAssignment
			default:
				l.Token = Exponentiation
			}
		case '=':
			l.step()
			l.Token = MultiplicationAssignment
		default:
			l.Token = Multiplication
		}
		l.step()
	case '/':
		// '/' or '/='
		switch l.peek() {
		case '=':
			l.step()
			l.Token = DivisionAssignment
		default:
			l.Token = Division
		}
		l.step()
	case '%':
		// '%' or '%='
		switch l.peek() {
		case '=':
			l.step()
			l.Token = RemainderAssignment
		default:
			l.Token = Remainder
		}
		l.step()
	case '=':
		// '=' or '=>' or '==' or '==='
		switch l.peek() {
		case '>':
			l.step()
			l.Token = Arrow
		case '=':
			l.step()
			switch l.peek() {
			case '=':
				l.Token = StrictEquality
				l.step()
			default:
				l.Token = Equality
			}
		default:
			l.Token = Assignment
		}
		l.step()
	case '&':
		// '&' or '&=' or '&&' or '&&='
		switch l.peek() {
		case '&':
			l.step()
			switch l.peek() {
			case '=':
				l.step()
				l.Token = LogicalANDAssignment
			default:
				l.Token = LogicalAND
			}
		case '=':
			l.step()
			l.Token = BitwiseANDAssignment
		default:
			l.Token = BitwiseAND
		}
		l.step()
	case '|':
		// '|' or '|=' or '||' or '||='
		switch l.peek() {
		case '|':
			l.step()
			switch l.peek() {
			case '=':
				l.step()
				l.Token = LogicalORAssignment
			default:
				l.Token = LogicalOR
			}
		case '=':
			l.step()
			l.Token = BitwiseORAssignment
		default:
			l.Token = BitwiseOR
		}
		l.step()
	case '^':
		// '^' or '^='
		switch l.peek() {
		case '=':
			l.step()
			l.Token = BitwiseXORAssignment
		default:
			l.Token = BitwiseXOR
		}
		l.step()
	case '?':
		// '?' or '??' or '??=' or '?.'
		switch l.peek() {
		case '?':
			l.step()
			switch l.peek() {
			case '=':
				l.step()
				l.Token = NullishCoalescingAssignment
			case '.':
				l.step()
				// TODO: special token for '?.'
				l.Token = Question
			default:
				l.Token = NullishCoalescing
			}
		default:
			l.Token = Question
		}
		l.step()

	case '~':
		l.Token = BitwiseNOT
		l.step()

	case '<':
		// '<' or '<<' or '<=' or '<<='
		switch l.peek() {
		case '<':
			l.step()
			switch l.peek() {
			case '=':
				l.step()
				l.Token = LeftShiftAssignment
			default:
				l.Token = BitwiseShiftLeft
			}
		case '=':
			l.step()
			l.Token = LTE
		default:
			l.Token = LT
		}
		l.step()

	case '>':
		// '>' or '>>' or '>>>' or '>=' or '>>=' or '>>>='
		switch l.peek() {
		case '>':
			l.step()
			switch l.peek() {
			case '>':
				l.step()
				switch l.peek() {
				case '=':
					l.step()
					l.Token = UnsignedRightShiftAssignment
				default:
					l.Token = BitwiseUnsignedShiftRight
				}
			case '=':
				l.step()
				l.Token = RightShiftAssignment
			default:
				l.Token = BitwiseShiftRight
			}
		case '=':
			l.step()
			l.Token = GTE
		default:
			l.Token = GT
		}
		l.step()
	case ':':
		l.Token = Colon
		l.step()
	case ',':
		l.Token = Comma
		l.step()
	case '.':
		if isDigit(l.peek()) {
			l.Token = NUMERIC_LITERAL
			// append 0
			l.Literal = fmt.Sprintf("0%s", l.scanNumber())
		} else {
			l.Token = Dot
			l.step()
		}
	case '(':
		l.Token = LParen
		l.step()
	case ')':
		l.Token = RParen
		l.step()
	case '[':
		l.Token = LBracket
		l.step()
	case ']':
		l.Token = RBracket
		l.step()
	case '{':
		l.Token = LBrace
		l.step()
	case '}':
		l.Token = RBrace
		l.step()
	case ';':
		l.Token = Semi
		l.step()

	// TODO: template literals?
	case '\'', '"', '`':
		quote := l.current
		l.step()
		start := l.start
	stringLiteral:
		for {
			switch l.current {
			case -1:
				l.Token = SyntaxError
				return
			case '\\':
				l.step() // consume '\'
				l.step() // consume escaped character
				continue
			case quote:
				l.step()
				break stringLiteral
			default:
				l.step()
			}
		}
		l.Token = STRING_LITERAL
		l.Literal = l.input[start : l.start-1]
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		literal := l.scanNumber()
		if l.current == 'n' {
			// BigInts can't have decimals or exponents. The 'e'/'E' check is
			// tricky because 'e' and 'E' are valid hex digits.
			isHex := strings.HasPrefix(literal, "0x") || strings.HasPrefix(literal, "0X")
			if strings.Contains(literal, ".") || (strings.ContainsAny(literal, "eE") && !isHex) {
				panic(LexerPanic{Msg: "SyntaxError: Invalid or unexpected token"})
			}
			// `07n` is invalid, but `0n` is valid. `0o7n` is valid.
			if len(literal) > 1 && literal[0] == '0' {
				if literal[1] != 'x' && literal[1] != 'X' && literal[1] != 'o' && literal[1] != 'O' && literal[1] != 'b' && literal[1] != 'B' {
					panic(LexerPanic{Msg: "SyntaxError: BigInt literals cannot have a leading zero"})
				}
			}
			l.Token = BigIntLiteral
			l.Literal = literal + "n"
			l.step() // consume 'n'
		} else {
			l.Token = NUMERIC_LITERAL
			l.Literal = literal
		}
	default:
		if isIdentifierStart(l.current) {
			l.Literal = l.scanIdentifier()
			l.Token = lookupIdentifier(l.Literal)
		} else {
			l.Literal = string(l.current)
			l.Token = SyntaxError
		}
	}
}

func (l *Lexer) Expect(t Token) {
	if l.Token != t {
		panic(LexerPanic{Msg: "SyntaxError: '(' expected"})
	}
	l.Next()
}
func (lexer *Lexer) IsContextualKeyword(text string) bool {
	return lexer.Token == Identifier && lexer.Literal == text
}

func (l *Lexer) scanIdentifier() string {
	position := l.start
	for isIdentifierPart(l.current) {
		l.step()
	}
	return l.input[position:l.start]
}

func (l *Lexer) scanNumber() string {
	position := l.start

	if l.current == '0' {
		switch l.peek() {
		case 'x', 'X':
			l.step() // consume '0'
			l.step() // consume 'x' or 'X'
			if !isHexDigit(l.current) {
				panic(LexerPanic{Msg: "SyntaxError: Invalid hexadecimal literal"})
			}
			l.scanNumberPart(isHexDigit)
			return l.input[position:l.start]
		case 'o', 'O':
			l.step() // consume '0'
			l.step() // consume 'o' or 'O'
			if !isOctalDigit(l.current) {
				panic(LexerPanic{Msg: "SyntaxError: Invalid octal literal"})
			}
			l.scanNumberPart(isOctalDigit)
			return l.input[position:l.start]
		case 'b', 'B':
			l.step() // consume '0'
			l.step() // consume 'b' or 'B'
			if !isBinaryDigit(l.current) {
				panic(LexerPanic{Msg: "SyntaxError: Invalid binary literal"})
			}
			l.scanNumberPart(isBinaryDigit)
			return l.input[position:l.start]
		}
		if l.peek() == '_' {
			panic(LexerPanic{Msg: "SyntaxError: Numeric separator can not be used after leading 0."})
		}
	}

	// Integer part
	l.scanNumberPart(isDigit)

	// Fractional part
	if l.current == '.' && isDigit(l.peek()) {
		l.step()
		l.scanNumberPart(isDigit)
	}

	// Exponent part (e.g., e-5, e+10, E10)
	if l.current == 'e' || l.current == 'E' {
		nextRune := l.peek()
		if isDigit(nextRune) {
			l.step() // consume 'e' or 'E'
			l.scanNumberPart(isDigit)
		} else if nextRune == '+' || nextRune == '-' {
			// Peek two characters ahead to ensure a digit follows the sign
			_, width := utf8.DecodeRuneInString(l.input[l.end:])
			if l.end+width < len(l.input) {
				nextNextRune, _ := utf8.DecodeRuneInString(l.input[l.end+width:])
				if isDigit(nextNextRune) {
					l.step() // consume 'e' or 'E'
					l.step() // consume sign
					l.scanNumberPart(isDigit)
				}
			}
		}
	}
	return l.input[position:l.start]
}

func (l *Lexer) scanNumberPart(isDigitFunc func(rune) bool) {
	for {
		if isDigitFunc(l.current) {
			l.step()
		} else if l.current == '_' {
			l.step() // consume '_'
			if !isDigitFunc(l.current) {
				panic(LexerPanic{Msg: "SyntaxError: Numeric separator must be followed by a digit"})
			}
		} else {
			break
		}
	}
}

func isWhitespace(chr rune) bool {
	return unicode.IsSpace(chr)
}

func isIdentifierStart(chr rune) bool {
	return chr == '$' || chr == '_' ||
		'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z'
}

func isIdentifierPart(chr rune) bool {
	return chr == '$' || chr == '_' ||
		'a' <= chr && chr <= 'z' || 'A' <= chr && chr <= 'Z' ||
		'0' <= chr && chr <= '9'
}

func isDigit(chr rune) bool {
	return '0' <= chr && chr <= '9'
}

func isHexDigit(chr rune) bool {
	return isDigit(chr) || ('a' <= chr && chr <= 'f') || ('A' <= chr && chr <= 'F')
}

func isOctalDigit(chr rune) bool {
	return '0' <= chr && chr <= '7'
}

func isBinaryDigit(chr rune) bool {
	return chr == '0' || chr == '1'
}
