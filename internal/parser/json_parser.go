package parser

import (
	"fmt"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/lexer"
)

type jsonParser struct {
	lexer *lexer.Lexer
}

func newJsonParser(lexer *lexer.Lexer) *jsonParser {
	return &jsonParser{lexer: lexer}
}

func (p *jsonParser) next() {
	p.lexer.Next()
}

func (p *jsonParser) expect(t lexer.Token) {
	if p.lexer.Token != t {
		panic(fmt.Sprintf("Unexpected token %q", p.lexer.Token.String()))
	}
	p.next()
}

func (p *jsonParser) parseExpression() ast.Expression {
	// TODO: error tracking
	// start := p.lexer.Span.Start
	// defer func() { p.setStatementSpan(start, stmt) }()
	switch p.lexer.Token {
	case lexer.STRING_LITERAL:
		value := p.lexer.Literal
		p.next()
		return &ast.StringLiteral{Value: value}
	case lexer.NUMERIC_LITERAL:
		value := p.lexer.Literal
		p.next()
		return &ast.NumberLiteral{Value: value}
	case lexer.TRUE, lexer.FALSE:
		value := p.lexer.Token == lexer.TRUE
		p.next()
		return &ast.BooleanLiteral{Value: value}
	case lexer.NULL:
		p.next()
		return &ast.NullLiteral{}
	case lexer.LBracket:
		p.next()
		elements := make([]ast.Expression, 0)
		for p.lexer.Token != lexer.RBracket {
			if p.lexer.Token == lexer.EOF {
				panic("unexpected EOF")
			}
			elements = append(elements, p.parseExpression())
			if p.lexer.Token == lexer.Comma {
				p.next()
			} else {
				break
			}
		}
		p.expect(lexer.RBracket)
		return &ast.ArrayLiteralExpression{Elements: elements}
	case lexer.LBrace:
		p.next()
		properties := make([]ast.ObjectProperty, 0)
		for p.lexer.Token != lexer.RBrace {
			if p.lexer.Token == lexer.EOF {
				panic("unexpected EOF")
			}
			key := p.parseExpression()
			p.expect(lexer.Colon)
			value := p.parseExpression()
			properties = append(properties, ast.ObjectProperty{Key: key, Value: value})
			if p.lexer.Token == lexer.Comma {
				p.next()
			} else {
				break
			}
		}
		p.expect(lexer.RBrace)
		return &ast.ObjectLiteralExpression{Properties: properties}
	default:
		panic(fmt.Sprintf("unexpected token %q", p.lexer.Token.String()))
	}
}

func ParseJSON(input string) (ast.Expression, error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		b, ok := r.(bailout)
		if !ok {
			panic(b)
		}
		b.err.Line, b.err.Col = position(input, b.err.Span.Start)
	}()
	p := newJsonParser(lexer.New(input))
	var expr ast.Expression
	for p.lexer.Token != lexer.EOF {
		expr = p.parseExpression()
	}
	return expr, nil
}
