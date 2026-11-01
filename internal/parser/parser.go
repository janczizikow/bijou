package parser

import (
	"errors"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/lexer"
)

type Parser struct {
	lexer *lexer.Lexer
}

func New(l *lexer.Lexer) *Parser {
	return &Parser{lexer: l}
}

// precedence lookups the precedence of a current token.
func (p *Parser) precedence() Precedence {
	if prec, ok := precedences[p.lexer.Token]; ok {
		return prec
	}
	return Lowest
}

func (p *Parser) isSpecialKeyword(keyword string) bool {
	return p.lexer.Token == lexer.Identifier && p.lexer.Literal == keyword
}

func (p *Parser) semicolon() {
	switch p.lexer.Token {
	case lexer.Semi:
		p.lexer.Next()
		// TODO: expect }, EOF or new line???
		// case lexer.RBrace, lexer.EOF:
		// default:
		// 	panic(lexer.LexerPanic{Msg: "SyntaxError: Unexpected token, expected ';'"})
	}
}

// #region Statements

func (p *Parser) parseStatement() ast.Statement {
	switch p.lexer.Token {
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
		p.semicolon()
		return &ast.BreakStatement{Label: label}
	case lexer.CONTINUE:
		p.lexer.Next()
		label := p.parseLabel()
		p.semicolon()
		return &ast.ContinueStatement{Label: label}
	case lexer.DO:
		return p.parseDoWhileStatement()
	case lexer.DEBUGGER:
		p.lexer.Next()
		p.semicolon()
		return &ast.DebuggerStatement{}
	case lexer.RETURN:
		p.lexer.Next()
		p.semicolon()
		return &ast.ReturnStatement{ReturnValue: p.parseExpression(Lowest)}
	case lexer.VAR, lexer.LET, lexer.CONST:
		node := p.parseVariableDeclaration()
		p.semicolon()
		return node
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
		return &ast.IfStatement{Condition: cond, Then: stmt, Else: elseStmt}
	case lexer.SWITCH:
		return p.parseSwitchStatement()
	case lexer.IMPORT:
		p.lexer.Next()
		var dflt, namespace *ast.Identifier
		var path ast.Expression
		if p.lexer.Token == lexer.Identifier {
			dflt = &ast.Identifier{Name: p.lexer.Literal}
			p.lexer.Next()
		}
		if p.lexer.Token == lexer.Multiplication {
			p.lexer.Next()
			if !p.isSpecialKeyword("as") {
				panic(lexer.LexerPanic{Msg: `SyntaxError: Unexpected token, expected "as"`})
			}
			p.lexer.Next()
			name := p.lexer.Literal
			p.lexer.Expect(lexer.Identifier)
			namespace = &ast.Identifier{Name: name}
		}
		if dflt != nil || namespace != nil {
			// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#identifiers_with_special_meanings
			if !p.isSpecialKeyword("from") {
				panic(lexer.LexerPanic{Msg: `SyntaxError: Unexpected token, expected "from"`})
			}
			p.lexer.Next()
		}
		path = p.parseExpression(Lowest)
		p.semicolon()
		return &ast.ImportStatement{DefaultImport: dflt, NamespaceImport: namespace, Path: path}
	case lexer.THROW:
		p.lexer.Next()
		if p.lexer.Token == lexer.EOF || p.lexer.Token == lexer.Semi {
			panic(lexer.LexerPanic{Msg: "SyntaxError: throw statement requires an argument"})
		}
		p.semicolon()
		return &ast.ThrowStatement{Argument: p.parseExpression(Lowest)}
	case lexer.FOR:
		return p.parseForStatement()
	case lexer.FUNCTION:
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
		if p.lexer.Token == lexer.Identifier && p.lexer.PeekToken() == lexer.Colon {
			label := p.lexer.Literal
			p.lexer.Next()
			p.lexer.Expect(lexer.Colon)
			return &ast.LabeledStatement{Label: &ast.Identifier{Name: label}, Body: p.parseStatement()}
		}
		expr := p.parseExpression(Lowest)
		if expr == nil {
			panic(lexer.LexerPanic{Msg: "SyntaxError: expected expression"})
		}
		p.semicolon()
		return &ast.ExpressionStatement{Expr: expr}
	}
}

func (p *Parser) parseDoWhileStatement() ast.Statement {
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
	p.semicolon()
	return &ast.DoWhileStatement{
		Condition: cond,
		Body:      block,
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
			p.lexer.Next()
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

func (p *Parser) parseVariableDeclaration() ast.Statement {
	kind := p.lexer.Token
	p.lexer.Next()
	if p.lexer.Token != lexer.Identifier {
		// TODO: handle destructuring etc.
		panic(lexer.LexerPanic{Msg: "SyntaxError: variable declaration requires a name"})
	}
	name := &ast.Identifier{Name: p.lexer.Literal}
	p.lexer.Next()
	var value ast.Expression
	if p.lexer.Token == lexer.Assignment {
		p.lexer.Next() // consume '='
		value = p.parseExpression(Lowest)
	}
	return &ast.VariableDeclaration{
		Token: kind,
		Name:  name,
		Value: value,
	}
}

func (p *Parser) parseForStatement() ast.Statement {
	p.lexer.Next() // consume "for"
	isAwait := p.isSpecialKeyword("await")
	if isAwait {
		p.lexer.Next()
	}
	p.lexer.Expect(lexer.LParen)
	var left ast.Statement
	if p.lexer.Token != lexer.Semi {
		switch p.lexer.Token {
		case lexer.VAR, lexer.LET, lexer.CONST:
			left = p.parseVariableDeclaration() // no terminator consumed
		default:
			left = &ast.ExpressionStatement{Expr: p.parseExpression(Lowest)}
		}
	}

	if isAwait && !p.isSpecialKeyword("of") {
		panic(lexer.LexerPanic{Msg: "Unexpected token"})
	}

	if p.lexer.Token == lexer.IN {
		if left == nil {
			panic(lexer.LexerPanic{Msg: "Expression expected"})
		}
		p.lexer.Next() // consume 'in'
		right := p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
		body := p.parseStatement()
		return &ast.ForInStatement{Variable: left, Object: right, Body: body}
	}

	if p.isSpecialKeyword("of") {
		if left == nil {
			panic(lexer.LexerPanic{Msg: "Expression expected"})
		}
		p.lexer.Next()
		right := p.parseExpression(Lowest)
		p.lexer.Expect(lexer.RParen)
		body := p.parseStatement()
		return &ast.ForOfStatement{Await: isAwait, Variable: left, Iterable: right, Body: body}
	}

	p.lexer.Expect(lexer.Semi)
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
	return &ast.ForStatement{Init: left, Condition: cond, Afterthought: step, Body: body}
}

func (p *Parser) parseFunction(isAsync bool) ast.Statement {
	p.lexer.Next()
	isGenerator := false
	if p.lexer.Token == lexer.Multiplication {
		isGenerator = true
		p.lexer.Next() // consume '*'
	}

	if p.lexer.Token != lexer.Identifier {
		panic(lexer.LexerPanic{Msg: "SyntaxError: function statement requires a name"})
	}
	name := &ast.Identifier{Name: p.lexer.Literal}
	p.lexer.Next() // consume identifier

	p.lexer.Expect(lexer.LParen)
	args := []*ast.Identifier{}

	for p.lexer.Token != lexer.RParen {
		if p.lexer.Token != lexer.Identifier {
			panic(lexer.LexerPanic{Msg: "SyntaxError: expected identifier in function arguments"})
		}
		args = append(args, &ast.Identifier{Name: p.lexer.Literal})
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

func (p *Parser) parseSwitchStatement() ast.Statement {
	p.lexer.Next()
	p.lexer.Expect(lexer.LParen)
	cond := p.parseExpression(Lowest)
	p.lexer.Expect(lexer.RParen)
	p.lexer.Expect(lexer.LBrace)
	cases := make([]ast.Case, 0)
	for p.lexer.Token != lexer.RBrace {
		switch p.lexer.Token {
		case lexer.CASE:
			p.lexer.Next()
			expr := p.parseExpression(Lowest)
			p.lexer.Expect(lexer.Colon)
			stmts := make([]ast.Statement, 0)
			for p.lexer.Token != lexer.CASE && p.lexer.Token != lexer.DEFAULT && p.lexer.Token != lexer.RBrace {
				stmts = append(stmts, p.parseStatement())
			}
			cases = append(cases, ast.Case{Value: expr, Body: stmts})
		case lexer.DEFAULT:
			p.lexer.Next()
			p.lexer.Expect(lexer.Colon)
			stmts := make([]ast.Statement, 0)
			for p.lexer.Token != lexer.CASE && p.lexer.Token != lexer.DEFAULT && p.lexer.Token != lexer.RBrace {
				stmts = append(stmts, p.parseStatement())
			}
			cases = append(cases, ast.Case{Body: stmts})
		// case lexer.Semi:
		// 	p.lexer.Next()
		case lexer.EOF:
			panic(lexer.LexerPanic{Msg: "unexpected EOF"})
		}
	}
	p.lexer.Expect(lexer.RBrace)
	return &ast.SwitchStatement{Test: cond, Cases: cases}
}

// #region Expressions

func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	expr := p.parsePrefix()

	for precedence < p.precedence() {
		infix := p.parseInfix(expr)
		if infix == nil {
			break
		}
		// p.lexer.Next() // consume semi???
		expr = infix
	}

	return expr
}

func (p *Parser) parsePrefix() ast.Expression {
	switch p.lexer.Token {
	case lexer.Identifier, lexer.PrivateIdentifier:
		tkn := p.lexer.Token
		name := p.lexer.Literal
		p.lexer.Next()
		// TODO: nested switch?
		return &ast.Identifier{Name: name, Private: tkn == lexer.PrivateIdentifier}
	case lexer.Exclamation, lexer.Plus, lexer.Minus:
		tkn := p.lexer.Token
		p.lexer.Next()
		// TODO: syntax check?
		return &ast.UnaryExpression{Operator: tkn, Right: p.parseExpression(Prefix)}
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
	case lexer.LBracket:
		p.lexer.Next()
		elements := make([]ast.Expression, 0)
		for p.lexer.Token != lexer.RBracket {
			if p.lexer.Token == lexer.EOF {
				panic(lexer.LexerPanic{Msg: "SyntaxError: ']' expected."})
			}
			elements = append(elements, p.parseExpression(Comma))
			if p.lexer.Token == lexer.Comma {
				p.lexer.Next()
			} else {
				break
			}
		}
		p.lexer.Expect(lexer.RBracket)
		return &ast.ArrayLiteralExpression{Elements: elements}
	case lexer.LBrace:
		p.lexer.Next()
		for p.lexer.Token != lexer.RBrace {
			if p.lexer.Token == lexer.EOF || p.lexer.Token == lexer.Semi {
				panic(lexer.LexerPanic{Msg: "SyntaxError: '}' expected."})
			}
			// TODO: handle key value pairs??
			if p.lexer.Token == lexer.Comma {
				p.lexer.Next()
			} else {
				break
			}
		}
		p.lexer.Expect(lexer.RBrace)
		return &ast.ObjectLiteralExpression{}
	case lexer.Increment, lexer.Decrement:
		op := p.lexer.Token
		p.lexer.Next()
		arg := p.parseExpression(Prefix)
		switch arg.(type) {
		case *ast.Identifier, *ast.PropertyAccessExpression:
			return &ast.UpdateExpression{Operator: op, Argument: arg, Prefix: true}
		default:
			panic(lexer.LexerPanic{Msg: "SyntaxError: Invalid left-hand side in prefix operation."})
		}
	case lexer.NEW:
		p.lexer.Next()
		constructor := p.parseExpression(PNew)
		// hasArgs := p.lexer.Token == lexer.LParen
		var args []ast.Expression
		if p.lexer.Token == lexer.LParen {
			args = p.parseArguments()
		}
		return &ast.NewExpression{Constructor: constructor, Arguments: args}
	case lexer.LParen:
		p.lexer.Next()
		expr := p.parseExpression(Comma)
		p.lexer.Expect(lexer.RParen)
		return &ast.ParenthesizedExpression{Expr: expr}
	default:
		// p.lexer.Next()
		// TODO: check?
		return nil
	}
}

func (p *Parser) parseInfix(left ast.Expression) ast.Expression {
	prec := p.precedence()
	switch p.lexer.Token {
	case lexer.Assignment, lexer.MultiplicationAssignment,
		lexer.DivisionAssignment, lexer.RemainderAssignment,
		lexer.AdditionAssignment, lexer.SubtractionAssignment,
		lexer.LeftShiftAssignment, lexer.RightShiftAssignment,
		lexer.UnsignedRightShiftAssignment, lexer.BitwiseANDAssignment,
		lexer.BitwiseXORAssignment, lexer.BitwiseORAssignment,
		lexer.ExponentiationAssignment, lexer.LogicalANDAssignment,
		lexer.LogicalORAssignment, lexer.NullishCoalescingAssignment:
		op := p.lexer.Token
		p.lexer.Next()
		return &ast.Assignment{
			Operator: op,
			Left:     left,
			Right:    p.parseExpression(prec - 1),
		}
	case
		// arithmetic
		lexer.Exponentiation, lexer.Multiplication, lexer.Division,
		lexer.Remainder, lexer.Plus, lexer.Minus,
		// relational
		lexer.LT, lexer.LTE, lexer.GT, lexer.GTE, lexer.INSTANCEOF, lexer.IN,
		// equality
		lexer.Equality, lexer.Inequality,
		lexer.StrictEquality, lexer.StrictInequality,
		// bitwise shift
		lexer.BitwiseShiftLeft, lexer.BitwiseShiftRight, lexer.BitwiseUnsignedShiftRight,
		// binary bitwise
		lexer.BitwiseAND, lexer.BitwiseOR, lexer.BitwiseXOR,
		// logical
		lexer.LogicalAND, lexer.LogicalOR, lexer.NullishCoalescing:
		op := p.lexer.Token
		p.lexer.Next()
		if op == lexer.Exponentiation {
			prec -= 1
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     left,
			Right:    p.parseExpression(prec),
		}
	case lexer.Dot, lexer.OptionalChaning, lexer.LBracket:
		optional := p.lexer.Token == lexer.OptionalChaning
		computed := p.lexer.Token == lexer.LBracket
		p.lexer.Next()
		var field ast.Expression
		if computed {
			field = p.parseExpression(Lowest)
			p.lexer.Expect(lexer.RBracket)
		} else {
			// TODO: keywords are also allowed e.g.: object.this
			if p.lexer.Token != lexer.Identifier && p.lexer.Token != lexer.PrivateIdentifier {
				panic(lexer.LexerPanic{Msg: "SyntaxError: Unexpected token"})
			}
			field = p.parseExpression(Member)
		}
		return &ast.PropertyAccessExpression{
			Object:   left,
			Field:    field,
			Optional: optional,
			Computed: computed,
		}
	case lexer.Increment, lexer.Decrement:
		op := p.lexer.Token
		p.lexer.Next()
		switch left.(type) {
		case *ast.Identifier, *ast.PropertyAccessExpression:
			return &ast.UpdateExpression{Operator: op, Argument: left}
		default:
			panic(lexer.LexerPanic{Msg: "SyntaxError: Invalid left-hand side in postfix operation."})
		}
	case lexer.LParen:
		return &ast.FunctionCall{Fn: left, Args: p.parseArguments()}
	case lexer.Question:
		p.lexer.Next()
		then := p.parseExpression(prec - 1)
		p.lexer.Expect(lexer.Colon)
		els := p.parseExpression(prec - 1)
		return &ast.ConditionalExpression{Condition: left, Then: then, Else: els}
	default:
		return nil
	}
}

func (p *Parser) parseLabel() *ast.Identifier {
	if p.lexer.Token == lexer.Semi || p.lexer.Token == lexer.EOF || p.lexer.Token == lexer.RBrace {
		return nil
	}
	name := p.lexer.Literal
	p.lexer.Expect(lexer.Identifier)

	return &ast.Identifier{Name: name}
}

// parseArguments parses expressions wrapped in () separated by ","
func (p *Parser) parseArguments() []ast.Expression {
	p.lexer.Next() // consume '('
	args := make([]ast.Expression, 0)
	for p.lexer.Token != lexer.RParen {
		if p.lexer.Token == lexer.EOF {
			panic(lexer.LexerPanic{Msg: "SyntaxError: ')' expected"})
		}
		args = append(args, p.parseExpression(Comma))
		if p.lexer.Token == lexer.Comma {
			p.lexer.Next()
		} else {
			break
		}
	}
	p.lexer.Expect(lexer.RParen)
	return args
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

	p := New(lexer.New(input))

	for p.lexer.Token != lexer.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
	}

	return
}
