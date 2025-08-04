package parser

import (
	"errors"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/lexer"
)

type Precedence uint8

// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Operator_precedence
const (
	Lowest Precedence = iota
	Coma
	Spread
	Yield
	Assignment
	Conditional
	NullishCoalescing
	LogicalOr
	LogicalAnd
	BitwiseOr
	BitwiseXor
	BitwiseAnd
	Equality
	Comparison
	Shift
	Add
	Multiply
	Prefix
	Postfix
	LNew
	Call
	Member
)

type Parser struct {
	lexer *lexer.Lexer
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{lexer: l}
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.lexer.Token {
	case lexer.Identifier:
		if p.lexer.IsContextualKeyword("async") {
			p.lexer.Next()
			// FIXME: Check if there was a new line before
			p.lexer.Expect(lexer.FUNCTION)
			return p.parseFunction(true)
		}
		// TODO:?
		// panic(lexer.LexerPanic{Msg: "SyntaxError: expected statement"})
		return &ast.ExpressionStatement{Expression: p.parseExpression(Lowest)}
	case lexer.LBrace:
		p.lexer.Next()
		stmts := []ast.Statement{}
		for p.lexer.Token != lexer.RBrace && p.lexer.Token != lexer.EOF {
			stmts = append(stmts, p.parseStatement())
			if p.lexer.Token == lexer.Semi {
				p.lexer.Next()
			}
		}
		p.lexer.Expect(lexer.RBrace)
		return &ast.BlockStatement{Statements: stmts}
	case lexer.BREAK:
		p.lexer.Next()
		label := p.parseLabel()
		return &ast.BreakStatement{Label: label}
	case lexer.CONTINUE:
		p.lexer.Next()
		label := p.parseLabel()
		return &ast.ContinueStatement{Label: label}
	case lexer.DO:
		p.lexer.Next()
		blockStmt := p.parseStatement()
		block, ok := blockStmt.(*ast.BlockStatement)
		if !ok {
			panic(lexer.LexerPanic{Msg: "do body must be a block statement"})
		}
		p.lexer.Expect(lexer.WHILE)
		p.lexer.Expect(lexer.LParen)
		cond := p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
		return &ast.DoWhileStatement{
			Condition: cond,
			Body:      block,
		}
	case lexer.DEBUGGER:
		p.lexer.Next()
		return &ast.DebuggerStatement{}
	case lexer.RETURN:
		p.lexer.Next()
		return &ast.ReturnStatement{
			Token:       lexer.RETURN,
			ReturnValue: p.parseExpression(Lowest),
		}
	case lexer.VAR, lexer.LET, lexer.CONST:
		kind := p.lexer.Token
		p.lexer.Next()
		return &ast.VariableDeclaration{
			Token: kind,
			Value: p.parseExpression(Lowest),
		}
	case lexer.IF:
		p.lexer.Next()
		p.lexer.Expect(lexer.LParen)
		cond := p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
		stmt := p.parseStatement()
		var elseStmt ast.Statement
		if p.lexer.Token == lexer.ELSE {
			p.lexer.Next()
			elseStmt = p.parseStatement()
		}
		return &ast.IfStatement{Condition: cond, Then: &stmt, Else: &elseStmt}
	case lexer.THROW:
		p.lexer.Next()
		if p.lexer.Token == lexer.Semi {
			panic(lexer.LexerPanic{Msg: "SyntaxError: throw statement requires an argument"})
		}
		return &ast.ThrowStatement{Argument: p.parseExpression(Lowest)}
	case lexer.FOR:
		p.lexer.Next()
		return p.parseForStatement()
	case lexer.FUNCTION:
		p.lexer.Next()
		return p.parseFunction(false)
	case lexer.Semi:
		p.lexer.Next()
		return &ast.EmptyStatement{}
	case lexer.TRY:
		return p.parseTryCatchStatement()
	case lexer.WITH:
		p.lexer.Next()
		p.lexer.Expect(lexer.LParen)
		expr := p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
		body := p.parseStatement()
		return &ast.WithStatement{Expr: expr, Body: body}
	case lexer.WHILE:
		p.lexer.Next()
		p.lexer.Expect(lexer.LParen)
		cond := p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
		bodyStmt := p.parseStatement()
		body, ok := bodyStmt.(*ast.BlockStatement)
		if !ok {
			panic(lexer.LexerPanic{Msg: "while body must be a block statement"})
		}
		return &ast.WhileStatement{Condition: cond, Body: body}
	default:
		expr := p.parseExpression(Lowest)
		if expr == nil {
			panic(lexer.LexerPanic{Msg: "SyntaxError: expected expression"})
		}
		return &ast.ExpressionStatement{Expression: expr}
	}
}

func (p *Parser) parseTryCatchStatement() ast.Statement {
	p.lexer.Next()
	block := p.parseStatement()
	blockStmt, ok := block.(*ast.BlockStatement)
	if !ok {
		panic(lexer.LexerPanic{Msg: "SyntaxError: '{' expected"})
	}
	var catch *ast.Catch
	if p.lexer.Token == lexer.CATCH {
		catch = &ast.Catch{}
		p.lexer.Next()
		if p.lexer.Token == lexer.LParen {
			param := p.parseExpression(Lowest)
			catch.Parameter = &param
			p.lexer.Expect(lexer.RParen)
		}
		catchBlock := p.parseStatement()
		catchStmt, ok := catchBlock.(*ast.BlockStatement)
		if !ok {
			panic(lexer.LexerPanic{Msg: "SyntaxError: '{' expected"})
		}
		catch.Block = *catchStmt
	}
	var finally *ast.BlockStatement
	if p.lexer.Token == lexer.FINALLY {
		p.lexer.Next()
		finallyStmt := p.parseStatement()
		finally, ok = finallyStmt.(*ast.BlockStatement)
		if !ok {
			panic(lexer.LexerPanic{Msg: "SyntaxError: '{' expected"})
		}

	}
	if catch == nil && finally == nil {
		panic(lexer.LexerPanic{Msg: "SyntaxError: 'catch' or 'finally' expected"})
	}
	return &ast.TryCatchStatement{Block: *blockStmt, Catch: catch, Finally: finally}
}

func (p *Parser) parseForStatement() ast.Statement {
	p.lexer.Expect(lexer.LParen)
	var init ast.Statement
	if p.lexer.Token == lexer.Semi {
		p.lexer.Next()
	} else {
		init = p.parseStatement()
	}
	var cond ast.Expression
	if p.lexer.Token == lexer.Semi {
		p.lexer.Next()
	} else {
		cond = p.parseExpression(Lowest)
		p.lexer.Expect(lexer.Semi)
	}
	var step ast.Expression
	if p.lexer.Token == lexer.RParen {
		p.lexer.Next()
	} else {
		step = p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
	}
	body := p.parseStatement()
	return &ast.ForStatement{Init: &init, Condition: &cond, Afterthought: &step, Body: &body}
}

func (p *Parser) parseFunction(isAsync bool) ast.Statement {
	isGenerator := false
	if p.lexer.Token == lexer.Multiplication {
		isGenerator = true
		p.lexer.Next() // consume '*'
	}

	if p.lexer.Token != lexer.Identifier {
		panic(lexer.LexerPanic{Msg: "SyntaxError: function statement requires a name"})
	}
	name := &ast.Identifier{Name: p.lexer.Literal, Token: lexer.Identifier}
	p.lexer.Next() // consume identifier

	p.lexer.Expect(lexer.LParen)
	args := []*ast.Identifier{}

	for p.lexer.Token != lexer.RParen {
		if p.lexer.Token != lexer.Identifier {
			panic(lexer.LexerPanic{Msg: "SyntaxError: expected identifier in function arguments"})
		}
		args = append(args, &ast.Identifier{Name: p.lexer.Literal, Token: lexer.Identifier})
		p.lexer.Next()

		if p.lexer.Token == lexer.Comma {
			p.lexer.Next()
		} else {
			break
		}
	}

	p.lexer.Expect(lexer.RParen)
	bodyStmt := p.parseStatement()
	body, ok := bodyStmt.(*ast.BlockStatement)
	if !ok {
		panic(lexer.LexerPanic{Msg: "function body must be a block statement"})
	}
	return &ast.FunctionDeclaration{
		Name:        name,
		Args:        args,
		Body:        body,
		IsGenerator: isGenerator,
		IsAsync:     isAsync,
	}
}

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	prefix := p.parsePrefix(precedence)

	// switch p.lexer.Token {
	// case lexer.NUMBER_LITERAL:
	// 	v := p.lexer.Literal
	// 	p.lexer.Next()
	// 	return &ast.NumberLiteral{
	// 		Value: v,
	// 	}
	// case lexer.Identifier:
	// 	name := p.lexer.Literal
	// 	p.lexer.Next()
	// 	return &ast.Identifier{
	// 		Name:  name,
	// 		Token: lexer.Identifier,
	// 	}
	// default:
	// 	return &ast.Assignment{
	// 		Left:     p.parseExpression(),
	// 		Right:    p.parseExpression(),
	// 		Operator: p.lexer.Token,
	// 	}
	// }

	return p.parseSuffix(prefix, precedence)
}

func (p *Parser) parsePrefix(precedence Precedence) ast.Expression {
	switch p.lexer.Token {
	case lexer.Identifier:
		name := p.lexer.Literal
		p.lexer.Next()
		// TODO: nested switch?
		return &ast.Identifier{
			Name:  name,
			Token: lexer.Identifier,
		}
	case lexer.Exclamation:
		p.lexer.Next()
		// TODO: syntax check?
		return &ast.UnaryExpression{Operator: lexer.Exclamation, Right: p.parseExpression(Prefix)}
	case lexer.TRUE, lexer.FALSE:
		value := p.lexer.Token == lexer.TRUE
		p.lexer.Next()
		return &ast.BooleanLiteral{Value: value}
	case lexer.NUMERIC_LITERAL:
		v := p.lexer.Literal
		p.lexer.Next()
		return &ast.NumberLiteral{Value: v}
	case lexer.BigIntLiteral:
		v := p.lexer.Literal
		p.lexer.Next()
		return &ast.BigIntLiteral{Value: v}
	case lexer.STRING_LITERAL:
		v := p.lexer.Literal
		p.lexer.Next()
		return &ast.StringLiteral{Value: v}
	case lexer.NULL:
		p.lexer.Next()
		return &ast.NullLiteral{}
	default:
		p.lexer.Next()
		// TODO: check?
		return nil
	}
}

func (p *Parser) parseSuffix(left ast.Expression, precedence Precedence) ast.Expression {
	switch p.lexer.Token {
	case lexer.Assignment:
		p.lexer.Next()
		return &ast.Assignment{
			Left:     left,
			Right:    p.parseExpression(Assignment),
			Operator: lexer.Assignment,
		}
	case lexer.Plus:
		op := p.lexer.Token
		p.lexer.Next()
		return &ast.BinaryExpression{
			Left:     left,
			Right:    p.parseExpression(Lowest),
			Operator: op,
		}
	default:
		return left
	}
}

func (p *Parser) parseLabel() *string {
	if p.lexer.Token == lexer.Identifier {
		name := p.lexer.Literal
		p.lexer.Next()
		return &name
	}

	return nil
}

func Parse(input string) (program *ast.Program, err error) {
	program = &ast.Program{}
	defer func() {
		r := recover()
		if p, isLexerPanic := r.(lexer.LexerPanic); isLexerPanic {
			err = errors.New(p.Msg)
		} else if r != nil {
			panic(r)
		}
	}()

	l := lexer.New(input)
	p := New(l)

	for p.lexer.Token != lexer.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)

		// consume semicolon if present
		if p.lexer.Token == lexer.Semi {
			p.lexer.Next()
		}
	}

	return
}
