package printer

import (
	"fmt"
	"strings"

	"github.com/janczizikow/bijou/internal/ast"
	"github.com/janczizikow/bijou/internal/parser"
)

type Options struct {
	UseTabs   bool
	TabWidth  int
	LineLimit int
	// Print semicolons at the ends of statements.
	Semi           bool
	SingleQuote    bool
	BracketSpacing bool
}

type printer struct {
	source  string
	content []byte
	options Options
	indent  int
}

func Print(program *ast.Program, options Options) []byte {
	p := &printer{source: program.Source, options: options}
	p.printStatements(program.Statements)
	return p.content
}

func (p *printer) print(text string) {
	p.content = append(p.content, text...)
}

func (p *printer) printStatements(stmts []ast.Statement) {
	for i, stmt := range stmts {
		if i > 0 && p.blankLineBetween(stmts[i-1], stmt) {
			p.printNewLine()
		}
		p.printStatement(stmt)
	}
}

// blankLineBetween reports whether the source has at least one empty line
// between the end of prev and the start of next.
func (p *printer) blankLineBetween(prev, next ast.Statement) bool {
	start, end := prev.GetSpan().End, next.GetSpan().Start
	if start > end || end > len(p.source) {
		return false
	}
	return countLineBreaks(p.source[start:end]) >= 2
}

// countLineBreaks counts \n, \r\n and lone \r as one line break each.
func countLineBreaks(s string) int {
	n := strings.Count(s, "\n")
	for i := 0; i < len(s); i++ {
		if s[i] == '\r' && (i+1 == len(s) || s[i+1] != '\n') {
			n++
		}
	}
	return n
}

func (p *printer) printStatement(stmt ast.Statement) {
	switch s := stmt.(type) {
	case *ast.BlockStatement:
		p.printBlock(s, true)
	case *ast.BreakStatement:
		p.printIndent()
		p.print("break")
		if s.Label != nil {
			p.printSpace()
			p.print(s.Label.Name)
		}
		p.printSemicolonAfterStatement()
	case *ast.ClassDeclaration:
		// TODO:
		panic("*ast.ClassDeclaration unimplemented")
	case *ast.ContinueStatement:
		p.printIndent()
		p.print("continue")
		if s.Label != nil {
			p.printSpace()
			p.print(s.Label.Name)
		}
		p.printSemicolonAfterStatement()
	case *ast.DebuggerStatement:
		p.printIndent()
		p.print("debugger")
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
	case *ast.EmptyStatement:
		p.print(";")
	case *ast.ExportDeclaration:
		// TODO:
		panic("*ast.ExportStatement unimplemented")
	case *ast.ExpressionStatement:
		p.printIndent()
		p.printExpression(s.Expr, parser.Lowest)
		p.printSemicolonAfterStatement()
	case *ast.ForInStatement:
		p.printForIn(s)
	case *ast.ForOfStatement:
		p.printForOf(s)
	case *ast.ForStatement:
		p.printFor(s)
	case *ast.FunctionDeclaration:
		p.printFunctionDeclaration(s)
	case *ast.IfStatement:
		p.printIf(s, true)
	case *ast.ImportDeclaration:
		p.print("import")
		p.printSpace()
		var seenDefault bool
		var named []*ast.ImportSpecifier
		for _, ss := range s.Specifiers {
			if ss.Kind == ast.ImportKindDefault {
				seenDefault = true
				p.print(ss.Local.Name)
			}
			if ss.Kind == ast.ImportKindNamespace {
				p.print("*")
				p.printSpace()
				p.print("as")
				p.printSpace()
				p.print(ss.Local.Name)
			}
			if ss.Kind == ast.ImportKindNamed {
				named = append(named, ss)
			}
		}
		if len(named) > 0 {
			if seenDefault {
				p.print(",")
				p.printSpace()
			}
			p.print("{")
			p.printSpace()
			for i, spec := range named {
				if i != 0 {
					p.print(",")
					p.printSpace()
				}
				if spec.Imported.Name == spec.Local.Name {
					p.print(spec.Imported.Name)
				} else {
					p.print(spec.Imported.Name)
					p.printSpace()
					p.print("as")
					p.printSpace()
					p.print(spec.Local.Name)
				}
			}
			p.printSpace()
			p.print("}")
		}
		if len(s.Specifiers) > 0 {
			p.printSpace()
			p.print("from")
			p.printSpace()
		}
		p.printExpression(s.Path, parser.Lowest)
		p.printSemicolonAfterStatement()
	case *ast.LabeledStatement:
		p.printIndent()
		p.print(s.Label.Name)
		p.print(":")
		p.printSpace()
		p.printStatement(s.Body)
	case *ast.ReturnStatement:
		p.printIndent()
		p.print("return")
		if s.ReturnValue != nil {
			p.printSpace()
			p.printExpression(s.ReturnValue, parser.Lowest)
		}
		p.printSemicolonAfterStatement()
	case *ast.SwitchStatement:
		p.printSwitch(s)
	case *ast.ThrowStatement:
		p.printIndent()
		p.print("throw")
		p.printSpace()
		p.printExpression(s.Argument, parser.Lowest)
		p.printSemicolonAfterStatement()
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
				p.printExpression(s.Catch.Parameter, parser.Lowest)
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
	case *ast.VariableDeclaration:
		p.printVariableDeclaration(s)
		p.printSemicolonAfterStatement()
	case *ast.WhileStatement:
		p.printIndent()
		p.print("while")
		p.printSpace()
		p.print("(")
		p.printExpression(s.Condition, parser.Lowest)
		p.print(")")
		p.printSpace()
		p.printBlock(s.Body, true)
	case *ast.WithStatement:
		p.printIndent()
		p.print("with")
		p.printSpace()
		p.print("(")
		p.printExpression(s.Expr, parser.Lowest)
		p.print(")")
		p.printSpace()
		p.printStatement(s.Body)
	default:
		panic(fmt.Sprintf("Unexpected statement type: %T", s))
	}
}

func (p *printer) printFunctionDeclaration(s *ast.FunctionDeclaration) {
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
	for i, param := range s.Params {
		if i > 0 {
			p.print(",")
			p.printSpace()
		}
		p.printBinding(param)
	}
	p.print(")")
	p.printSpace()
	p.printBlock(s.Body, true)
}

func (p *printer) printBinding(b ast.Binding) {
	switch v := b.(type) {
	case *ast.Identifier:
		p.print(v.Name)
	case *ast.AssignmentPattern:
		p.printBinding(v.Left)
		p.print(" = ")
		p.printExpression(v.Right, parser.Lowest)
	case *ast.RestElement:
		p.print("...")
		p.printBinding(v.Argument)
	case *ast.KeyValuePatternProperty:
		if v.Computed {
			p.print("[")
			p.printExpression(v.Key, parser.Lowest)
			p.print("]")
			p.print(":")
			p.printSpace()
			p.printBinding(v.Value)
		}
	case *ast.ArrayPattern:
		p.print("[")
		for i, el := range v.Elements {
			p.printBinding(el)
			if i < len(v.Elements)-1 {
				p.print(",")
				p.printSpace()
			}
		}
		p.print("]")
	case *ast.ObjectPattern:
		p.print("{")
		for i, el := range v.Properties {
			p.printBinding(el)
			if i < len(v.Properties)-1 {
				p.print(",")
				p.printSpace()
			}
		}
		p.print("}")
	case nil:
		p.printSpace()
	}
}

func (p *printer) printFor(s *ast.ForStatement) {
	p.printIndent()
	p.print("for")
	p.printSpace()
	p.print("(")
	if s.Init != nil {
		switch v := s.Init.(type) {
		case *ast.VariableDeclaration:
			p.printVariableDeclaration(v)
		case *ast.ExpressionStatement:
			p.printExpression(v.Expr, parser.Lowest)
		default:
			p.printStatement(s.Init)
		}
	}
	p.print(";")
	if s.Condition != nil {
		p.printSpace()
		p.printExpression(s.Condition, parser.Lowest)
	}
	p.print(";")
	if s.Afterthought != nil {
		p.printSpace()
		p.printExpression(s.Afterthought, parser.Lowest)
	}
	p.print(")")
	switch v := s.Body.(type) {
	case *ast.EmptyStatement:
		p.printStatement(v)
		p.printNewLine()
	case *ast.BlockStatement:
		p.printSpace()
		p.printBlock(v, false)
		p.printSemicolonAfterStatement()
	default:
		p.printSpace()
		p.printStatement(v)
	}
}

func (p *printer) printForOf(s *ast.ForOfStatement) {
	p.printIndent()
	p.print("for")
	p.printSpace()
	if s.Await {
		p.print("await")
		p.printSpace()
	}
	p.print("(")
	switch v := s.Variable.(type) {
	case *ast.VariableDeclaration:
		p.printVariableDeclaration(v)
	case *ast.ExpressionStatement:
		p.printExpression(v.Expr, parser.Lowest)
	default:
		p.printStatement(s.Variable)
	}
	p.printSpace()
	p.print("of")
	p.printSpace()
	p.printExpression(s.Iterable, parser.Lowest)
	p.print(")")
	p.printSpace()
	p.printStatement(s.Body)
}

func (p *printer) printForIn(s *ast.ForInStatement) {
	p.printIndent()
	p.print("for")
	p.printSpace()
	p.print("(")
	switch v := s.Variable.(type) {
	case *ast.VariableDeclaration:
		p.printVariableDeclaration(v)
	case *ast.ExpressionStatement:
		p.printExpression(v.Expr, parser.Lowest)
	default:
		p.printStatement(s.Variable)
	}
	p.printSpace()
	p.print("in")
	p.printSpace()
	p.printExpression(s.Object, parser.Lowest)
	p.print(")")
	p.printSpace()
	p.printStatement(s.Body)
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
	case *ast.ConditionalExpression:
		p.printExpression(e.Condition, parser.Lowest)
		p.printSpace()
		p.print("?")
		p.printSpace()
		p.printExpression(e.Then, parser.Lowest)
		p.printSpace()
		p.print(":")
		p.printSpace()
		p.printExpression(e.Else, parser.Lowest)
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
	case *ast.ParenthesizedExpression:
		p.print("(")
		p.printExpression(e.Expr, parser.Comma)
		p.print(")")
	case *ast.FunctionCall:
		p.printExpression(e.Fn, parser.Lowest)
		p.print("(")
		for _, arg := range e.Args {
			p.printExpression(arg, parser.Comma)
		}
		p.print(")")
	case *ast.ArrowFunction:
		p.print("(")
		for _, arg := range e.Args {
			p.printExpression(arg, parser.Comma)
		}
		p.print(")")
		p.printSpace()
		p.printExpression(e.Body, parser.Lowest)
	case *ast.UpdateExpression:
		if e.Prefix {
			p.print(e.Operator.String())
		}
		p.printExpression(e.Argument, parser.Prefix)
		if !e.Prefix {
			p.print(e.Operator.String())
		}
	case *ast.NewExpression:
		p.print("new")
		p.printSpace()
		p.printExpression(e.Constructor, parser.PNew)

		for _, arg := range e.Arguments {
			p.printExpression(arg, parser.Comma)
		}
	case *ast.ImportExpression:
		p.print("import")
		p.print("(")
		p.printExpression(e.Source, parser.Comma)
		p.print(")")
	case *ast.YieldExpression:
		p.print("yield")
		if e.Expr != nil {
			p.printSpace()
			p.printExpression(e.Expr, parser.Lowest)
		}
	case nil:
		break
	default:
		panic(fmt.Sprintf("Unexpected expression type: %T", e))
	}
}

func (p *printer) printVariableDeclaration(s *ast.VariableDeclaration) {
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
}

func (p *printer) printBlock(s *ast.BlockStatement, trailingNewline bool) {
	p.print("{")
	p.printNewLine()
	p.indent++
	p.printStatements(s.Statements)
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

func (p *printer) printSwitchCase(s ast.Case) {
	p.printIndent()
	if s.Value != nil {
		p.print("case")
		p.printSpace()
		p.printExpression(s.Value, parser.Lowest)
		p.print(":")
		p.printNewLine()
	} else {
		p.print("default:")
		p.printNewLine()
	}
	p.indent++
	p.printStatements(s.Body)
	p.indent--
}

func (p *printer) printSwitch(s *ast.SwitchStatement) {
	p.printIndent()
	p.print("switch")
	p.printSpace()
	p.print("(")
	p.printExpression(s.Test, parser.Lowest)
	p.print(")")
	p.printSpace()
	p.print("{")
	p.printNewLine()
	p.indent++
	for _, c := range s.Cases {
		p.printSwitchCase(c)
	}
	p.indent--
	p.print("}")
	p.printNewLine()
}

func (p *printer) printSpace() {
	p.print(" ")
}

func (p *printer) printIndent() {
	for i := 0; i < p.indent; i++ {
		if p.options.UseTabs {
			p.print("\t")
		} else {
			p.print(strings.Repeat(" ", p.options.TabWidth))
		}
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
