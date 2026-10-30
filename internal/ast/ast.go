package ast

import "github.com/janczizikow/bijou/internal/lexer"

type Statement interface {
	statement()
}

type Expression interface {
	expression()
}

type Program struct {
	Statements []Statement
}

// #region Statements

func (bs *BlockStatement) statement()      {}
func (bs *BreakStatement) statement()      {}
func (cs *ContinueStatement) statement()   {}
func (ds *DebuggerStatement) statement()   {}
func (dws *DoWhileStatement) statement()   {}
func (es *ExpressionStatement) statement() {}
func (es *EmptyStatement) statement()      {}
func (is *ImportStatement) statement()     {}
func (fs *ForStatement) statement()        {}
func (fis *ForInStatement) statement()     {}
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
	Expr Expression
}

type EmptyStatement struct{}

type ImportStatement struct {
	DefaultImport   *Identifier
	NamespaceImport *Identifier
	Path            Expression
	Assertions      []string
}

type ForStatement struct {
	Init         Statement
	Condition    Expression
	Afterthought Expression
	Body         Statement
}

type ForInStatement struct {
	Variable Statement
	Object   Expression
	Body     Statement
}

type ForOfStatement struct {
	Await    bool
	Variable Statement
	Iterable Expression
	Body     Statement
}

type IfStatement struct {
	Condition Expression
	Then      Statement
	Else      Statement
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

// #region Expressions

func (ale *ArrayLiteralExpression) expression()   {}
func (a *Assignment) expression()                 {}
func (bl *BooleanLiteral) expression()            {}
func (bil *BigIntLiteral) expression()            {}
func (nl *NumberLiteral) expression()             {}
func (sl *StringLiteral) expression()             {}
func (nl *NullLiteral) expression()               {}
func (ole *ObjectLiteralExpression) expression()  {}
func (i *Identifier) expression()                 {}
func (ue *UnaryExpression) expression()           {}
func (ue *BinaryExpression) expression()          {}
func (ce *ConditionalExpression) expression()     {}
func (pae *PropertyAccessExpression) expression() {}
func (pe *ParenthesizedExpression) expression()   {}
func (fc *FunctionCall) expression()              {}
func (ue *UpdateExpression) expression()          {}
func (ne *NewExpression) expression()             {}

type ConditionalExpression struct {
	Condition Expression
	Then      Expression
	Else      Expression
}

type Assignment struct {
	Operator lexer.Token
	Left     Expression
	Right    Expression
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
	Name    string
	Private bool
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

type PropertyAccessExpression struct {
	Object   Expression
	Field    Expression
	Optional bool
	Computed bool
}

type FunctionCall struct {
	Fn   Expression
	Args []Expression
}

type UpdateExpression struct {
	Operator lexer.Token
	Argument Expression
	Prefix   bool
}

type ArrayLiteralExpression struct {
	Elements []Expression
}

type ObjectLiteralExpression struct {
	// TODO: key and values?
}

type NewExpression struct {
	Constructor Expression
	Arguments   []Expression
}

type ParenthesizedExpression struct {
	Expr Expression
}
