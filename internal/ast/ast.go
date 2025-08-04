package ast

import "github.com/janczizikow/bijou/internal/lexer"

type Precedence uint8

type Statement interface {
	statement()
}

type Expression interface {
	expression()
}

type Program struct {
	Statements []Statement
}

// Statements

func (bs *BlockStatement) statement()      {}
func (bs *BreakStatement) statement()      {}
func (cs *ContinueStatement) statement()   {}
func (ds *DebuggerStatement) statement()   {}
func (dws *DoWhileStatement) statement()   {}
func (es *ExpressionStatement) statement() {}
func (es *EmptyStatement) statement()      {}
func (fs *ForStatement) statement()        {}
func (fos *ForOfStatement) statement()     {}
func (is *IfStatement) statement()         {}
func (vd *VariableDeclaration) statement() {}
func (rs *ReturnStatement) statement()     {}
func (fd *FunctionDeclaration) statement() {}
func (ts *ThrowStatement) statement()      {}
func (ts *TryCatchStatement) statement()   {}
func (ws *WithStatement) statement()       {}
func (ws *WhileStatement) statement()      {}

type BlockStatement struct {
	Statements []Statement
}

type DebuggerStatement struct{}

type DoWhileStatement struct {
	Condition Expression
	Body      *BlockStatement
}

type ContinueStatement struct {
	Label *string
}

type BreakStatement struct {
	Label *string
}

type ExpressionStatement struct {
	Expression Expression
}

type EmptyStatement struct{}

type ForStatement struct {
	Init         *Statement
	Condition    *Expression
	Afterthought *Expression
	Body         *Statement
}

type ForOfStatement struct {
	Variable *Statement
	Iterable Expression
	Body     *Statement
}

type IfStatement struct {
	Condition Expression
	Then      *Statement
	Else      *Statement
}

type VariableDeclaration struct {
	Token lexer.Token // the first token of the expression
	Name  *Identifier
	Value Expression
}

type FunctionDeclaration struct {
	Name        *Identifier
	Args        []*Identifier
	Body        *BlockStatement
	IsGenerator bool
	IsAsync     bool
}

type ReturnStatement struct {
	Token       lexer.Token // the 'return' token
	ReturnValue Expression
}

type ThrowStatement struct {
	Argument Expression
}

type TryCatchStatement struct {
	Block   BlockStatement
	Catch   *Catch
	Finally *BlockStatement
}

type Catch struct {
	Parameter *Expression
	Block     BlockStatement
}

type WithStatement struct {
	Expr Expression
	Body Statement
}

type WhileStatement struct {
	Condition Expression
	Body      *BlockStatement
}

// Expressions

func (a *Assignment) expression()        {}
func (bl *BooleanLiteral) expression()   {}
func (bil *BigIntLiteral) expression()   {}
func (nl *NumberLiteral) expression()    {}
func (sl *StringLiteral) expression()    {}
func (nl *NullLiteral) expression()      {}
func (i *Identifier) expression()        {}
func (ue *UnaryExpression) expression()  {}
func (ue *BinaryExpression) expression() {}

type Assignment struct {
	Left     Expression
	Right    Expression
	Operator lexer.Token
}

type BooleanLiteral struct {
	Value bool
}

type BigIntLiteral struct {
	Value string
}

type NumberLiteral struct {
	Value string
}

type StringLiteral struct {
	Value string
}

type NullLiteral struct{}

type Identifier struct {
	Name  string
	Token lexer.Token
}

type UnaryExpression struct {
	Operator lexer.Token
	Right    Expression
}

type BinaryExpression struct {
	Left     Expression
	Operator lexer.Token
	Right    Expression
}
