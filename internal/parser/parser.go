package parser

import (
	"fmt"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/lexer"
)

type (
	SyntaxError = lexer.SyntaxError
	bailout     struct{ err *SyntaxError }
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

// Some identifiers aren't reserved keywords, but have special meaning in some contexts:
//
// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#identifiers_with_special_meanings
func (p *Parser) isSpecialKeyword(keyword string) bool {
	return p.lexer.Token == lexer.Identifier && p.lexer.Literal == keyword
}

func (p *Parser) consumeSemicolon() {
	switch {
	case p.lexer.Token == lexer.Semi:
		p.next()
	case p.lexer.Token == lexer.RBrace, p.lexer.Token == lexer.EOF, p.lexer.HasNewLineBefore:
		// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Lexical_grammar#automatic_semicolon_insertion
	default:
		p.fail("Unexpected token, expected ';'")
	}
}

func (p *Parser) setStatementSpan(start int, stmt ast.Statement) {
	if stmt != nil {
		stmt.SetSpan(lexer.Span{Start: start, End: p.lexer.PrevSpan.End})
	}
}

func (p *Parser) next() {
	p.lexer.Next()
}

func (p *Parser) fail(msg string) {
	if p.lexer.Token == lexer.Illegal {
		panic(bailout{p.lexer.Err})
	}
	panic(bailout{&SyntaxError{Msg: msg, Span: p.lexer.Span}})
}

func (p *Parser) expect(t lexer.Token) {
	if p.lexer.Token != t {
		p.fail(fmt.Sprintf("Unexpected token %q", p.lexer.Token.String()))
	}
	p.next()
}

// #region Statements

func (p *Parser) parseStatement() (stmt ast.Statement) {
	start := p.lexer.Span.Start
	// needs to wrapped in closure to delay reading the stmt until fn exits
	defer func() { p.setStatementSpan(start, stmt) }()
	switch p.lexer.Token {
	case lexer.LBrace:
		p.next()
		stmts := []ast.Statement{}
		for p.lexer.Token != lexer.RBrace && p.lexer.Token != lexer.EOF {
			stmts = append(stmts, p.parseStatement())
			if p.lexer.Token == lexer.Semi {
				p.next()
			}
		}
		p.expect(lexer.RBrace)
		stmt = &ast.BlockStatement{Statements: stmts}
	case lexer.BREAK:
		p.next()
		label := p.parseLabel()
		p.consumeSemicolon()
		stmt = &ast.BreakStatement{Label: label}
	case lexer.CONTINUE:
		p.next()
		label := p.parseLabel()
		p.consumeSemicolon()
		stmt = &ast.ContinueStatement{Label: label}
	case lexer.DO:
		stmt = p.parseDoWhileStatement()
	case lexer.DEBUGGER:
		p.next()
		p.consumeSemicolon()
		stmt = &ast.DebuggerStatement{}
	case lexer.RETURN:
		p.next()
		var ret ast.Expression
		if p.lexer.Token != lexer.Semi && p.lexer.Token != lexer.RBrace &&
			p.lexer.Token != lexer.EOF && !p.lexer.HasNewLineBefore {
			ret = p.parseExpression(Lowest)
		}
		p.consumeSemicolon()
		stmt = &ast.ReturnStatement{ReturnValue: ret}
	case lexer.VAR, lexer.LET, lexer.CONST:
		node := p.parseVariableDeclaration()
		p.consumeSemicolon()
		stmt = node
		return
	case lexer.IF:
		p.next()
		p.expect(lexer.LParen)
		cond := p.parseExpression(Lowest)
		p.expect(lexer.RParen)
		then := p.parseStatement()
		var elseStmt ast.Statement
		if p.lexer.Token == lexer.ELSE {
			p.next()
			elseStmt = p.parseStatement()
		}
		stmt = &ast.IfStatement{Condition: cond, Then: then, Else: elseStmt}
	case lexer.SWITCH:
		stmt = p.parseSwitchStatement()
	case lexer.IMPORT:
		peek := p.lexer.PeekToken()
		if peek == lexer.LParen || peek == lexer.Dot {
			stmt = p.parseExpressionStatement()
			return
		}
		stmt = p.parseImportStatement()
	case lexer.THROW:
		p.next()
		if p.lexer.Token == lexer.EOF || p.lexer.Token == lexer.Semi || p.lexer.HasNewLineBefore {
			p.fail("throw statement requires an argument")
		}
		stmt = &ast.ThrowStatement{Argument: p.parseExpression(Lowest)}
		p.consumeSemicolon()
	case lexer.FOR:
		stmt = p.parseForStatement()
	case lexer.FUNCTION:
		stmt = p.parseFunction(false)
	case lexer.Semi:
		p.next()
		stmt = &ast.EmptyStatement{}
	case lexer.TRY:
		stmt = p.parseTryCatchStatement()
	case lexer.WITH:
		p.next()
		p.expect(lexer.LParen)
		expr := p.parseExpression(Lowest)
		p.expect(lexer.RParen)
		body := p.parseStatement()
		stmt = &ast.WithStatement{Expr: expr, Body: body}
	case lexer.WHILE:
		p.next()
		p.expect(lexer.LParen)
		cond := p.parseExpression(Lowest)
		p.expect(lexer.RParen)
		bodyStmt := p.parseStatement()
		body, ok := bodyStmt.(*ast.BlockStatement)
		if !ok {
			p.fail("while body must be a block statement")
		}
		stmt = &ast.WhileStatement{Condition: cond, Body: body}
	default:
		if p.lexer.Token == lexer.Identifier && p.lexer.PeekToken() == lexer.Colon {
			label := p.lexer.Literal
			p.next()
			p.expect(lexer.Colon)
			stmt = &ast.LabeledStatement{Label: &ast.Identifier{Name: label}, Body: p.parseStatement()}
			return
		}
		stmt = p.parseExpressionStatement()
	}
	return
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	expr := p.parseExpression(Lowest)
	p.consumeSemicolon()
	return &ast.ExpressionStatement{Expr: expr}
}

func (p *Parser) parseImportStatement() ast.Statement {
	p.next()
	var (
		dflt, namespace *ast.ImportSpecifier
		path            ast.Expression
		sawBraces       bool
	)
	specifiers := make([]*ast.ImportSpecifier, 0)
	if p.lexer.Token == lexer.Identifier {
		dflt = &ast.ImportSpecifier{
			Local: &ast.Identifier{Name: p.lexer.Literal},
			Kind:  ast.ImportKindDefault,
		}
		specifiers = append(specifiers, dflt)
		p.next()
	}
	if dflt != nil && p.lexer.Token == lexer.Comma {
		p.next()
		if p.lexer.Token != lexer.Multiplication && p.lexer.Token != lexer.LBrace {
			p.fail("expected '*' or '{' after default import")
		}
	}
	if p.lexer.Token == lexer.Multiplication {
		p.next()
		if !p.isSpecialKeyword("as") {
			p.fail(`Unexpected token, expected "as"`)
		}
		p.next()
		name := p.lexer.Literal
		p.expect(lexer.Identifier)
		namespace = &ast.ImportSpecifier{
			Kind:  ast.ImportKindNamespace,
			Local: &ast.Identifier{Name: name},
		}
		specifiers = append(specifiers, namespace)
	}
	if namespace == nil && p.lexer.Token == lexer.LBrace {
		sawBraces = true
		p.next()
		for p.lexer.Token != lexer.RBrace {
			imported := p.lexer.Literal
			if p.lexer.Token == lexer.DEFAULT {
				imported = "default"
				p.next()
				if !p.isSpecialKeyword("as") {
					p.fail("cannot import as reserved word")
				}
			} else {
				p.expect(lexer.Identifier)
			}
			local := imported
			if p.isSpecialKeyword("as") {
				p.next()
				local = p.lexer.Literal
				p.expect(lexer.Identifier)
			}
			specifiers = append(specifiers, &ast.ImportSpecifier{
				Imported: &ast.Identifier{Name: imported},
				Local:    &ast.Identifier{Name: local},
			})
			if p.lexer.Token != lexer.RBrace {
				p.expect(lexer.Comma)
			}
		}
		p.expect(lexer.RBrace)
	}
	if len(specifiers) != 0 || sawBraces {
		if !p.isSpecialKeyword("from") {
			p.fail(`Unexpected token, expected "from"`)
		}
		p.next()
	}
	if p.lexer.Token != lexer.STRING_LITERAL {
		p.fail("Unexpected token, expected a string literal")
	}
	path = p.parseExpression(Lowest)
	p.consumeSemicolon()
	return &ast.ImportDeclaration{Specifiers: specifiers, Path: path}
}

func (p *Parser) parseDoWhileStatement() ast.Statement {
	p.next()
	blockStmt := p.parseStatement()
	block, ok := blockStmt.(*ast.BlockStatement)
	if !ok {
		p.fail("do body must be a block statement")
	}
	p.expect(lexer.WHILE)
	p.expect(lexer.LParen)
	cond := p.parseExpression(Lowest)
	p.expect(lexer.RParen)
	if p.lexer.Token == lexer.Semi {
		p.next()
	}
	return &ast.DoWhileStatement{
		Condition: cond,
		Body:      block,
	}
}

func (p *Parser) parseTryCatchStatement() ast.Statement {
	p.next()
	block := p.parseStatement()
	blockStmt, ok := block.(*ast.BlockStatement)
	if !ok {
		p.fail("'{' expected")
	}
	var catch *ast.Catch
	if p.lexer.Token == lexer.CATCH {
		catch = &ast.Catch{}
		p.next()
		if p.lexer.Token == lexer.LParen {
			p.next()
			catch.Parameter = p.parseExpression(Lowest)
			p.expect(lexer.RParen)
		}
		catchBlock := p.parseStatement()
		catchStmt, ok := catchBlock.(*ast.BlockStatement)
		if !ok {
			p.fail("'{' expected")
		}
		catch.Block = *catchStmt
	}
	var finally *ast.BlockStatement
	if p.lexer.Token == lexer.FINALLY {
		p.next()
		finallyStmt := p.parseStatement()
		finally, ok = finallyStmt.(*ast.BlockStatement)
		if !ok {
			p.fail("'{' expected")
		}

	}
	if catch == nil && finally == nil {
		p.fail("'catch' or 'finally' expected")
	}
	return &ast.TryCatchStatement{Block: *blockStmt, Catch: catch, Finally: finally}
}

func (p *Parser) parseVariableDeclaration() ast.Statement {
	kind := p.lexer.Token
	p.next()
	if p.lexer.Token != lexer.Identifier {
		// TODO: handle destructuring etc.
		p.fail("variable declaration requires a name")
	}
	name := &ast.Identifier{Name: p.lexer.Literal}
	p.next()
	var value ast.Expression
	if p.lexer.Token == lexer.Assignment {
		p.next() // consume '='
		value = p.parseExpression(Lowest)
	}
	return &ast.VariableDeclaration{
		Token: kind,
		Name:  name,
		Value: value,
	}
}

func (p *Parser) parseForStatement() ast.Statement {
	p.next() // consume "for"
	isAwait := p.isSpecialKeyword("await")
	if isAwait {
		p.next()
	}
	p.expect(lexer.LParen)
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
		p.fail("Unexpected token")
	}

	if p.lexer.Token == lexer.IN {
		if left == nil {
			p.fail("Expression expected")
		}
		p.next() // consume 'in'
		right := p.parseExpression(Lowest)
		p.expect(lexer.RParen)
		body := p.parseStatement()
		return &ast.ForInStatement{Variable: left, Object: right, Body: body}
	}

	if p.isSpecialKeyword("of") {
		if left == nil {
			p.fail("Expression expected")
		}
		p.next()
		right := p.parseExpression(Lowest)
		p.expect(lexer.RParen)
		body := p.parseStatement()
		return &ast.ForOfStatement{Await: isAwait, Variable: left, Iterable: right, Body: body}
	}

	p.expect(lexer.Semi)
	var cond ast.Expression
	if p.lexer.Token == lexer.Semi {
		p.next()
	} else {
		cond = p.parseExpression(Lowest)
		p.expect(lexer.Semi)
	}
	var step ast.Expression
	if p.lexer.Token == lexer.RParen {
		p.next()
	} else {
		step = p.parseExpression(Lowest)
		p.expect(lexer.RParen)
	}
	body := p.parseStatement()
	return &ast.ForStatement{Init: left, Condition: cond, Afterthought: step, Body: body}
}

func (p *Parser) parseParameter() ast.Binding {
	if p.lexer.Token == lexer.Identifier {
		left := &ast.Identifier{Name: p.lexer.Literal}
		p.next()
		switch p.lexer.Token {
		case lexer.Assignment:
			p.next()
			return &ast.AssignmentPattern{Left: left, Right: p.parseExpression(Lowest)}
		// case p.lexer.Token == lexer.Ellipsis && left != nil:
		//  p.lexer.Next()
		//  params = append(params, &ast.ArrayPattern{Argument: left})
		default:
			return left
		}
	}

	if p.lexer.Token == lexer.LBracket {
		p.next()
		arr := &ast.ArrayPattern{}
		for p.lexer.Token != lexer.RBracket {
			if p.lexer.Token == lexer.EOF {
				p.fail("\"]\" expected")
			}
			arr.Elements = append(arr.Elements, p.parseParameter())
			if p.lexer.Token == lexer.Comma {
				p.next()
			} else {
				break
			}
		}
		p.expect(lexer.RBracket)
		return arr
	}

	if p.lexer.Token == lexer.LBrace {
		p.next()
		obj := &ast.ObjectPattern{}
		for p.lexer.Token != lexer.RBrace {
			if p.lexer.Token == lexer.EOF {
				p.fail("\"}\" expected")
			}
			prop := p.parseParameter()
			obj.Properties = append(obj.Properties, prop)
			if p.lexer.Token == lexer.Comma {
				p.next()
			} else {
				break
			}
		}
		p.expect(lexer.RBrace)
		return obj
	}

	if p.lexer.Token == lexer.Ellipsis {
		p.next()
		arg := p.lexer.Literal
		p.expect(lexer.Identifier)
		if p.lexer.Token != lexer.RParen {
			p.fail("expected \")\"")
		}
		return &ast.RestElement{Argument: &ast.Identifier{Name: arg}}
	}

	return nil
}

func (p *Parser) parseParameters() []ast.Binding {
	p.expect(lexer.LParen)
	params := []ast.Binding{}
	for p.lexer.Token != lexer.RParen {
		params = append(params, p.parseParameter())
		if p.lexer.Token == lexer.Comma {
			p.next()
		} else {
			break
		}
	}
	p.expect(lexer.RParen)
	return params
}

func (p *Parser) parseFunction(isAsync bool) ast.Statement {
	p.next()
	isGenerator := false
	if p.lexer.Token == lexer.Multiplication {
		isGenerator = true
		p.next() // consume '*'
	}

	if p.lexer.Token != lexer.Identifier {
		p.fail("function statement requires a name")
	}
	name := &ast.Identifier{Name: p.lexer.Literal}
	p.next() // consume identifier

	params := p.parseParameters()
	bodyStmt := p.parseStatement()
	body, ok := bodyStmt.(*ast.BlockStatement)
	if !ok {
		p.fail("function body must be a block statement")
	}
	return &ast.FunctionDeclaration{
		Name:        name,
		Params:      params,
		Body:        body,
		IsGenerator: isGenerator,
		IsAsync:     isAsync,
	}
}

func (p *Parser) parseSwitchStatement() ast.Statement {
	p.next()
	p.expect(lexer.LParen)
	cond := p.parseExpression(Lowest)
	p.expect(lexer.RParen)
	p.expect(lexer.LBrace)
	cases := make([]ast.Case, 0)
	for p.lexer.Token != lexer.RBrace {
		switch p.lexer.Token {
		case lexer.CASE:
			p.next()
			expr := p.parseExpression(Lowest)
			p.expect(lexer.Colon)
			stmts := make([]ast.Statement, 0)
			for p.lexer.Token != lexer.CASE && p.lexer.Token != lexer.DEFAULT && p.lexer.Token != lexer.RBrace {
				stmts = append(stmts, p.parseStatement())
			}
			cases = append(cases, ast.Case{Value: expr, Body: stmts})
		case lexer.DEFAULT:
			p.next()
			p.expect(lexer.Colon)
			stmts := make([]ast.Statement, 0)
			for p.lexer.Token != lexer.CASE && p.lexer.Token != lexer.DEFAULT && p.lexer.Token != lexer.RBrace {
				stmts = append(stmts, p.parseStatement())
			}
			cases = append(cases, ast.Case{Body: stmts})
		default:
			p.fail("unexpected token")
		}
	}
	p.expect(lexer.RBrace)
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
		expr = infix
	}

	return expr
}

func (p *Parser) parsePrefix() ast.Expression {
	switch p.lexer.Token {
	case lexer.Identifier, lexer.PrivateIdentifier:
		tkn := p.lexer.Token
		name := p.lexer.Literal
		p.next()
		// TODO: nested switch?
		return &ast.Identifier{Name: name, Private: tkn == lexer.PrivateIdentifier}
	case lexer.IMPORT:
		p.next()
		p.expect(lexer.LParen)
		src := p.parseExpression(Comma)
		p.expect(lexer.RParen)
		return &ast.ImportExpression{Source: src}
	case lexer.Exclamation, lexer.Plus, lexer.Minus:
		tkn := p.lexer.Token
		p.next()
		// TODO: syntax check?
		return &ast.UnaryExpression{Operator: tkn, Right: p.parseExpression(Prefix)}
	case lexer.TRUE, lexer.FALSE:
		value := p.lexer.Token == lexer.TRUE
		p.next()
		return &ast.BooleanLiteral{Value: value}
	case lexer.NUMERIC_LITERAL:
		v := p.lexer.Literal
		p.next()
		return &ast.NumberLiteral{Value: v}
	case lexer.BigIntLiteral:
		v := p.lexer.Literal
		p.next()
		return &ast.BigIntLiteral{Value: v}
	case lexer.STRING_LITERAL:
		v := p.lexer.Literal
		p.next()
		return &ast.StringLiteral{Value: v}
	case lexer.NULL:
		p.next()
		return &ast.NullLiteral{}
	case lexer.LBracket:
		p.next()
		elements := make([]ast.Expression, 0)
		for p.lexer.Token != lexer.RBracket {
			if p.lexer.Token == lexer.EOF {
				p.fail("']' expected.")
			}
			elements = append(elements, p.parseExpression(Comma))
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
		for p.lexer.Token != lexer.RBrace {
			if p.lexer.Token == lexer.EOF || p.lexer.Token == lexer.Semi {
				p.fail("'}' expected.")
			}
			// TODO: handle key value pairs??
			if p.lexer.Token == lexer.Comma {
				p.next()
			} else {
				break
			}
		}
		p.expect(lexer.RBrace)
		return &ast.ObjectLiteralExpression{}
	case lexer.Increment, lexer.Decrement:
		op := p.lexer.Token
		p.next()
		arg := p.parseExpression(Prefix)
		switch arg.(type) {
		case *ast.Identifier, *ast.PropertyAccessExpression:
			return &ast.UpdateExpression{Operator: op, Argument: arg, Prefix: true}
		default:
			p.fail("Invalid left-hand side in prefix operation.")
		}
	case lexer.NEW:
		p.next()
		// FIXME: https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/new.target
		// if p.lexer.Token == lexer.Dot {
		// }
		constructor := p.parsePrefix()
	inner:
		for {
			switch p.lexer.Token {
			case lexer.Dot:
				p.next()
				// TODO: keywords are also allowed e.g.: object.this
				if p.lexer.Token != lexer.Identifier && p.lexer.Token != lexer.PrivateIdentifier {
					p.fail("Unexpected token")
				}
				field := p.parsePrefix()
				constructor = &ast.PropertyAccessExpression{
					Object: constructor,
					Field:  field,
				}
			case lexer.LBracket:
				p.next()
				field := p.parseExpression(Lowest)
				p.expect(lexer.RBracket)
				constructor = &ast.PropertyAccessExpression{
					Object:   constructor,
					Field:    field,
					Computed: true,
				}
			default:
				break inner
			}
		}
		var args []ast.Expression
		if p.lexer.Token == lexer.LParen {
			args = p.parseArguments()
		}
		return &ast.NewExpression{Constructor: constructor, Arguments: args}
	case lexer.LParen:
		p.next()
		expr := p.parseExpression(Comma)
		p.expect(lexer.RParen)
		return &ast.ParenthesizedExpression{Expr: expr}
	case lexer.YIELD:
		p.next()
		var expr ast.Expression
		if p.lexer.Token != lexer.Semi && p.lexer.Token != lexer.RBrace &&
			p.lexer.Token != lexer.EOF && !p.lexer.HasNewLineBefore {
			expr = p.parseExpression(Lowest)
		}
		return &ast.YieldExpression{Expr: expr}
	default:
		p.fail("expected expression")
	}
	return nil
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
		p.next()
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
		p.next()
		if op == lexer.Exponentiation {
			prec -= 1
		}
		return &ast.BinaryExpression{
			Operator: op,
			Left:     left,
			Right:    p.parseExpression(prec),
		}
	case lexer.Dot, lexer.OptionalChaining, lexer.LBracket:
		optional := p.lexer.Token == lexer.OptionalChaining
		computed := p.lexer.Token == lexer.LBracket
		p.next()
		var field ast.Expression
		if computed {
			field = p.parseExpression(Lowest)
			p.expect(lexer.RBracket)
		} else {
			// TODO: keywords are also allowed e.g.: object.this
			if p.lexer.Token != lexer.Identifier && p.lexer.Token != lexer.PrivateIdentifier {
				p.fail("Unexpected token")
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
		if p.lexer.HasNewLineBefore {
			return nil
		}
		op := p.lexer.Token
		p.next()
		switch left.(type) {
		case *ast.Identifier, *ast.PropertyAccessExpression:
			return &ast.UpdateExpression{Operator: op, Argument: left}
		default:
			p.fail("Invalid left-hand side in postfix operation.")
		}
	case lexer.LParen:
		return &ast.FunctionCall{Fn: left, Args: p.parseArguments()}
	case lexer.Question:
		p.next()
		then := p.parseExpression(prec - 1)
		p.expect(lexer.Colon)
		els := p.parseExpression(prec - 1)
		return &ast.ConditionalExpression{Condition: left, Then: then, Else: els}
	default:
		return nil
	}
	return nil
}

func (p *Parser) parseLabel() *ast.Identifier {
	if p.lexer.Token == lexer.Semi || p.lexer.Token == lexer.EOF || p.lexer.Token == lexer.RBrace || p.lexer.HasNewLineBefore {
		return nil
	}
	name := p.lexer.Literal
	p.expect(lexer.Identifier)

	return &ast.Identifier{Name: name}
}

// parseArguments parses expressions wrapped in () separated by ","
func (p *Parser) parseArguments() []ast.Expression {
	p.next() // consume '('
	args := make([]ast.Expression, 0)
	for p.lexer.Token != lexer.RParen {
		if p.lexer.Token == lexer.EOF {
			p.fail("')' expected")
		}
		args = append(args, p.parseExpression(Comma))
		if p.lexer.Token == lexer.Comma {
			p.next()
		} else {
			break
		}
	}
	p.expect(lexer.RParen)
	return args
}

func Parse(input string) (program *ast.Program, err error) {
	p := New(lexer.New(input))
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		b, ok := r.(bailout)
		if !ok {
			panic(r)
		}
		b.err.Line, b.err.Col = position(input, b.err.Span.Start)
		program, err = nil, b.err
	}()
	program = &ast.Program{Source: input}

	for p.lexer.Token != lexer.EOF {
		stmt := p.parseStatement()
		program.Statements = append(program.Statements, stmt)
	}

	return
}

func position(src string, offset int) (line, col int) {
	line, col = 1, 0
	for i, r := range src {
		if i >= offset {
			break
		}
		switch r {
		case '\n', '\u2028', '\u2029':
			line++
			col = 0
		case '\r':
			// treat \r\n as one terminator: skip the \r, let \n count
			if i+1 < len(src) && src[i+1] == '\n' {
				continue
			}
			line++
			col = 0
		default:
			col++
		}
	}
	return
}
