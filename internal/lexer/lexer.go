package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Span struct{ Start, End int }

func (s Span) Len() int {
	return s.End - s.Start
}

type SyntaxError struct {
	Msg  string
	Span Span
	Line int
	Col  int
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("SyntaxError: %s (%d:%d)", e.Msg, e.Line, e.Col)
}

// Lexer represents a lexical analyzer for JavaScript source code.
// It takes a string as input and converts it into a sequence of tokens.
type Lexer struct {
	source string // original input

	HasNewLineBefore bool
	Token            Token // current token
	Span             Span
	PrevSpan         Span
	Literal          string // the token literal value, if any

	current rune // current code point
	start   int  // byte offset of current code point
	end     int  // byte offset past current code point

	Err *SyntaxError
}

func New(source string) *Lexer {
	l := &Lexer{source: source}
	l.step()
	l.Next()
	return l
}

func (l *Lexer) errorf(format string, args ...any) {
	l.Token = Illegal
	l.Err = &SyntaxError{
		Msg:  fmt.Sprintf(format, args...),
		Span: Span{l.start, l.end},
	}
}

func (l *Lexer) step() {
	if l.end < len(l.source) {
		r, width := utf8.DecodeRuneInString(l.source[l.end:])
		if width == 0 {
			l.end = len(l.source)
			l.current = -1
		}
		l.start = l.end
		l.end += width
		l.current = r
	} else {
		l.start = l.end
		l.end = len(l.source)
		l.current = -1
	}
}

func (l *Lexer) peek() rune {
	r, width := utf8.DecodeRuneInString(l.source[l.end:])
	if width == 0 {
		return -1
	}

	return r
}

func (l *Lexer) skipWhitespace() {
	for isWhitespace(l.current) {
		if isLineTerminator(l.current) {
			l.HasNewLineBefore = true
		}
		l.step()
	}
}

func (l *Lexer) Next() {
	l.HasNewLineBefore = false
	l.PrevSpan = l.Span
	l.skipWhitespace()
	start := l.start
	l.scan()
	l.Span = Span{Start: start, End: l.start}
}

func (l *Lexer) scan() {
	switch l.current {
	case -1:
		l.Token = EOF
	case '#':
		if l.start == 0 && l.peek() == '!' {
			// TODO: parse whole HashBang
			l.step()
			l.step()
			l.Token = HashBang
			l.Literal = "#!"
		} else {
			// private identifier e.g.: #foo
			l.step()
			l.Literal = l.scanIdentifier()
			l.Token = PrivateIdentifier
		}
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
			l.Token = Minus
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
			default:
				l.Token = NullishCoalescing
			}
		case '.':
			l.step()
			l.Token = OptionalChaining
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
		if strings.HasPrefix(l.source[l.start:], "...") {
			l.Token = Ellipsis
			l.step()
			l.step()
			l.step()
		} else if isDigit(l.peek()) {
			l.Token = NUMERIC_LITERAL
			l.Literal = l.scanNumber()
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
		literal := l.scanString()
		if l.Err != nil {
			return
		}
		l.Token = STRING_LITERAL
		l.Literal = literal
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		literal := l.scanNumber()
		if l.Err != nil {
			return
		}
		if l.current == 'n' {
			// BigInts can't have decimals or exponents. The 'e'/'E' check is
			// tricky because 'e' and 'E' are valid hex digits.
			isHex := strings.HasPrefix(literal, "0x") || strings.HasPrefix(literal, "0X")
			if strings.Contains(literal, ".") || (strings.ContainsAny(literal, "eE") && !isHex) {
				l.errorf("Invalid BigIntLiteral")
				return
			}
			// `07n` is invalid, but `0n` is valid. `0o7n` is valid.
			if len(literal) > 1 && literal[0] == '0' {
				if literal[1] != 'x' && literal[1] != 'X' && literal[1] != 'o' && literal[1] != 'O' && literal[1] != 'b' && literal[1] != 'B' {
					l.errorf("BigInt literals cannot have a leading zero")
					return
				}
			}
			l.Token = BigIntLiteral
			l.Literal = literal + "n"
			l.step() // consume 'n'
		} else {
			l.Token = NUMERIC_LITERAL
			l.Literal = strings.TrimSuffix(literal, ".")
		}
	default:
		if isIdentifierStart(l.current) {
			l.Literal = l.scanIdentifier()
			l.Token = lookupIdentifier(l.Literal)
		} else {
			l.Literal = string(l.current)
			l.errorf("Unexpected character %q", l.current)
		}
	}
}

// PeekToken returns the token after the current one without consuming anything.
func (l *Lexer) PeekToken() Token {
	saved := *l
	l.Next()
	t := l.Token
	*l = saved
	return t
}

func (l *Lexer) scanIdentifier() string {
	position := l.start
	for isIdentifierPart(l.current) {
		l.step()
	}
	return l.source[position:l.start]
}

func (l *Lexer) scanString() string {
	quote := l.current
	quoteStart := l.start
	l.step() // consume opening quote
	position := l.start
	for {
		switch l.current {
		case -1:
			l.errorf("Unterminated string constant")
			l.Err.Span = Span{quoteStart, l.start}
			return ""
		case '\n', '\r':
			if quote != '`' {
				l.errorf("Unterminated string constant")
				l.Err.Span = Span{quoteStart, l.start}
				return ""
			}
			l.step()
		case '\\':
			l.step() // consume '\'
			l.step() // consume escaped character
		case quote:
			literal := l.source[position:l.start]
			l.step() // consume closing quote
			return literal
		default:
			l.step()
		}
	}
}

func (l *Lexer) scanNumber() string {
	position := l.start

	if l.current == '0' {
		switch l.peek() {
		case 'x', 'X':
			l.step() // consume '0'
			l.step() // consume 'x' or 'X'
			if !isHexDigit(l.current) {
				l.errorf("Invalid hexadecimal literal")
				return ""
			}
			l.scanNumberPart(isHexDigit)
			return l.source[position:l.start]
		case 'o', 'O':
			l.step() // consume '0'
			l.step() // consume 'o' or 'O'
			if !isOctalDigit(l.current) {
				l.errorf("Invalid octal literal")
				return ""
			}
			l.scanNumberPart(isOctalDigit)
			return l.source[position:l.start]
		case 'b', 'B':
			l.step() // consume '0'
			l.step() // consume 'b' or 'B'
			if !isBinaryDigit(l.current) {
				l.errorf("Invalid binary literal")
			}
			l.scanNumberPart(isBinaryDigit)
			return l.source[position:l.start]
		}
		if l.peek() == '_' {
			l.errorf("Numeric separator can not be used after leading 0.")
		}
	}

	// Integer part
	l.scanNumberPart(isDigit)

	// Fractional part
	if l.current == '.' {
		l.step() // consume '.'
		if isDigit(l.current) {
			l.scanNumberPart(isDigit)
		}
	}

	// Exponent part (e.g., e-5, e+10, E10)
	if l.current == 'e' || l.current == 'E' {
		nextRune := l.peek()
		if isDigit(nextRune) {
			l.step() // consume 'e' or 'E'
			l.scanNumberPart(isDigit)
		} else if nextRune == '+' || nextRune == '-' {
			// Peek two characters ahead to ensure a digit follows the sign
			_, width := utf8.DecodeRuneInString(l.source[l.end:])
			if l.end+width < len(l.source) {
				nextNextRune, _ := utf8.DecodeRuneInString(l.source[l.end+width:])
				if isDigit(nextNextRune) {
					l.step() // consume 'e' or 'E'
					l.step() // consume sign
					l.scanNumberPart(isDigit)
				}
			}
		}
	}
	return l.source[position:l.start]
}

func (l *Lexer) scanNumberPart(isDigitFunc func(rune) bool) {
	for {
		if isDigitFunc(l.current) {
			l.step()
		} else if l.current == '_' {
			l.step() // consume '_'
			if !isDigitFunc(l.current) {
				l.errorf("Numeric separator must be followed by a digit")
				return
			}
		} else {
			break
		}
	}
}

// isLineTerminator reports whether chr is a JavaScript line terminator.
//
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#line_terminators
func isLineTerminator(chr rune) bool {
	return chr == '\n' || chr == '\r' || chr == 0x2028 || chr == 0x2029
}

// isWhitespace reports whether chr is a JavaScript white space.
//
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#white_space
func isWhitespace(chr rune) bool {
	switch chr {
	case ' ', '\t', '\v', '\f', '\r', '\n':
		return true
	case 0x00A0, 0xFEFF:
		return true
	default:
		return chr == 0x2028 || chr == 0x2029 || unicode.Is(unicode.Zs, chr)
	}
}

func isIdentifierStart(chr rune) bool {
	return chr == '$' || chr == '_' || unicode.IsLetter(chr)
}

func isIdentifierPart(chr rune) bool {
	return isIdentifierStart(chr) || isDigit(chr)
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
