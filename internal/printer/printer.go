package printer

import (
	"fmt"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/parser"
)

type Options struct {
	LineLimit int
	// Print semicolons at the ends of statements.
	Semi        bool
	SingleQuote bool
}

type printer struct {
	content []byte
	options Options
	indent  int
}

func Print(tree *ast.Program, options Options) []byte {
	p := &printer{options: options}
	for _, stmt := range tree.Statements {
		p.printStatement(stmt)
	}

	return p.content
}

func (p *printer) print(text string) {
	p.content = append(p.content, text...)
}

func (p *printer) printStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.BreakStatement:
		p.printIndent()
		p.print("break")
		if s.Label != nil {
			p.printSpace()
			p.print(*s.Label)
		}
		p.printSemicolonAfterStatement()
	case *ast.ContinueStatement:
		p.printIndent()
		p.print("continue")
		if s.Label != nil {
			p.printSpace()
			p.print(*s.Label)
		}
		p.printSemicolonAfterStatement()
	case *ast.DoWhileStatement:
		p.printIndent()
		p.print("do")
		p.printSpace()
		p.printBlock(s.Body, false)
		p.printSpace()
		p.print("while")
		p.printSpace()
		p.print("(")
		p.printExpression(s.Condition, parser.Lowest)
		p.print(")")
		p.printSemicolonAfterStatement()
	case *ast.DebuggerStatement:
		p.printIndent()
		p.print("debugger")
		p.printSemicolonAfterStatement()
	case *ast.BlockStatement:
		p.printBlock(s, true)
	case *ast.ReturnStatement:
		p.printIndent()
		p.print("return")
		if s.ReturnValue != nil {
			p.printSpace()
			p.printExpression(s.ReturnValue, parser.Lowest)
		}
		p.printSemicolonAfterStatement()
	case *ast.VariableDeclaration:
		p.printIndent()
		p.print(s.Token.String())
		p.printSpace()
		p.printExpression(s.Name, parser.Lowest)
		if s.Value != nil {
			p.printSpace()
			p.print("=")
			p.printSpace()
			p.printExpression(s.Value, parser.Lowest)
		}
		p.printSemicolonAfterStatement()
	case *ast.IfStatement:
		p.printIf(s, true)
	case *ast.ExpressionStatement:
		p.printIndent()
		p.printExpression(s.Expression, parser.Lowest)
		p.printSemicolonAfterStatement()
	case *ast.ForStatement:
		p.printIndent()
		p.print("for")
		p.printSpace()
		p.print("(")
		if s.Init != nil {
			p.printStatement(s.Init)
		}
		p.printSemicolonAfterStatement()
		p.printExpression(s.Condition, parser.Lowest)
		p.printSemicolonAfterStatement()
		if s.Afterthought != nil {
			p.printExpression(s.Afterthought, parser.Lowest)
		}
		p.print(")")
		p.printSpace()
		if s.Body != nil {
			p.printStatement(s.Body)
		}
	case *ast.ForInStatement:
		p.printIndent()
		p.print("for")
		p.printSpace()
		p.print("(")
		p.printStatement(s.Variable)
		p.print("in")
		p.printSpace()
		p.printExpression(s.Object, parser.Lowest)
		p.print(")")
		p.printSpace()
		p.printStatement(s.Body)
	case *ast.FunctionDeclaration:
		p.printIndent()
		if s.IsAsync {
			p.print("async")
			p.printSpace()
		}
		p.print("function")
		if s.IsGenerator {
			p.print("*")
		}
		p.printSpace()
		p.printExpression(s.Name, parser.Lowest)
		p.print("(")
		for i, arg := range s.Args {
			if i > 0 {
				p.print(",")
				p.printSpace()
			}
			p.printExpression(arg, parser.Lowest)
		}
		p.print(")")
		p.printSpace()
		p.printBlock(s.Body, true)
	case *ast.TryCatchStatement:
		p.printIndent()
		p.print("try")
		p.printSpace()
		p.printBlock(&s.Block, false)
		if s.Catch != nil {
			p.printIndent()
			p.print("catch")
			if s.Catch.Parameter != nil {
				p.printSpace()
				p.print("(")
				p.printExpression(*s.Catch.Parameter, parser.Lowest)
				p.print(")")
			}
			p.printSpace()
			p.printBlock(&s.Catch.Block, s.Finally == nil)
		}
		if s.Finally != nil {
			p.printIndent()
			p.print("finally")
			p.printSpace()
			p.printBlock(s.Finally, true)
		}
	case *ast.ThrowStatement:
		p.printIndent()
		p.print("throw")
		p.printSpace()
		p.printExpression(s.Argument, parser.Lowest)
		p.printSemicolonAfterStatement()
	case *ast.WithStatement:
		p.printIndent()
		p.print("with")
		p.printSpace()
		p.print("(")
		p.printExpression(s.Expr, parser.Lowest)
		p.print(")")
		p.printSpace()
		p.printStatement(s.Body)
	case *ast.WhileStatement:
		p.printIndent()
		p.print("while")
		p.printSpace()
		p.print("(")
		p.printExpression(s.Condition, parser.Lowest)
		p.print(")")
		p.printSpace()
		p.printBlock(s.Body, true)
	case *ast.EmptyStatement:
		p.printIndent()
		p.print(";")
	default:
		panic(fmt.Sprintf("Unexpected statement type: %T", s))
	}
}

func (p *printer) printExpression(expr ast.Expression, presedence parser.Precedence) {
	switch e := expr.(type) {
	case *ast.Identifier:
		name := e.Name
		if e.Private {
			name = "#" + name
		}
		p.print(name)
	case *ast.NumberLiteral:
		p.print(e.Value)
	case *ast.BigIntLiteral:
		p.print(e.Value)
	case *ast.StringLiteral:
		p.printQuoted(e.Value)
	case *ast.BooleanLiteral:
		if e.Value {
			p.print("true")
		} else {
			p.print("false")
		}
	case *ast.NullLiteral:
		p.print("null")
	case *ast.UnaryExpression:
		p.print(e.Operator.String())
		p.printExpression(e.Right, parser.Prefix)
	case *ast.BinaryExpression:
		p.printExpression(e.Left, parser.Lowest)
		p.printSpace()
		p.print(e.Operator.String())
		p.printSpace()
		p.printExpression(e.Right, parser.Lowest)
	case *ast.ArrayLiteralExpression:
		p.print("[")
		for i, el := range e.Elements {
			p.printExpression(el, parser.Comma)
			if i < len(e.Elements)-1 {
				p.print(",")
				p.printSpace()
			}
		}
		p.print("]")
	case *ast.Assignment:
		p.printExpression(e.Left, parser.Lowest)
		p.printSpace()
		p.print(e.Operator.String())
		p.printSpace()
		p.printExpression(e.Right, parser.Lowest)
	case *ast.PropertyAccessExpression:
		op := "."
		if e.Optional {
			op = "?."
		}
		if e.Computed {
			op = "["
		}
		p.printExpression(e.Object, parser.Lowest)
		p.print(op)
		p.printExpression(e.Field, parser.Call)
		if e.Computed {
			p.print("]")
		}
	case *ast.FunctionCall:
		p.printExpression(e.Fn, parser.Lowest)
		p.print("(")
		for _, arg := range e.Args {
			p.printExpression(arg, parser.Comma)
		}
		p.print(")")
	case nil:
		break
	default:
		panic(fmt.Sprintf("Unexpected expression type: %T", e))
	}
}

func (p *printer) printBlock(s *ast.BlockStatement, trailingNewline bool) {
	p.print("{")
	p.printNewLine()
	p.indent++
	for _, stmt := range s.Statements {
		p.printStatement(stmt)
	}
	p.indent--
	p.printIndent()
	p.print("}")
	if trailingNewline {
		p.printNewLine()
	}
}

func (p *printer) printIf(s *ast.IfStatement, indent bool) {
	if indent {
		p.printIndent()
	}
	p.print("if")
	p.printSpace()
	p.print("(")
	p.indent++
	p.printExpression(s.Condition, parser.Lowest)
	p.indent--
	p.print(")")
	p.printSpace()
	if s.Then != nil {
		if thenBlock, ok := (s.Then).(*ast.BlockStatement); ok {
			p.printBlock(thenBlock, s.Else == nil)
		} else {
			p.printStatement(s.Then)
		}
	}
	if s.Else != nil {
		if _, ok := (s.Then).(*ast.BlockStatement); ok {
			p.printSpace()
		} else {
			p.printIndent()
		}
		p.print("else")
		if elseIf, ok := (s.Else).(*ast.IfStatement); ok {
			p.printSpace()
			p.printIf(elseIf, false)
		} else {
			p.printSpace()
			p.printStatement(s.Else)
		}
	}
}

func (p *printer) printSpace() {
	p.print(" ")
}

func (p *printer) printIndent() {
	for i := 0; i < p.indent; i++ {
		p.print("  ")
	}
}

func (p *printer) printSemicolonAfterStatement() {
	if p.options.Semi {
		p.print(";")
	}
	p.printNewLine()
}

func (p *printer) printNewLine() {
	p.print("\n")
}

func (p *printer) printQuoted(s string) {
	quote := "\""
	if p.options.SingleQuote {
		quote = "'"
	}
	p.print(quote)
	p.print(s)
	p.print(quote)
}
