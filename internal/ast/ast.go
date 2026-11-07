package ast

import "github.com/janczizikow/bijou/internal/lexer"

type Node interface {
	GetSpan() lexer.Span
	SetSpan(lexer.Span)
}

type BaseNode struct{ Span lexer.Span }

func (n *BaseNode) GetSpan() lexer.Span  { return n.Span }
func (n *BaseNode) SetSpan(s lexer.Span) { n.Span = s }

type Statement interface {
	Node
	statement()
}

type Expression interface {
	expression()
}

type Binding interface {
	binding()
}

type Program struct {
	Source     string
	Statements []Statement
}

// #region Bindings

func (ap *AssignmentPattern) binding()       {}
func (ap *ArrayPattern) binding()            {}
func (op *ObjectPattern) binding()           {}
func (i *Identifier) binding()               {}
func (re *RestElement) binding()             {}
func (kv *KeyValuePatternProperty) binding() {}

type RestElement struct {
	Argument Binding
}

type KeyValuePatternProperty struct {
	Key      Expression
	Value    Binding
	Computed bool
}

type AssignmentPattern struct {
	Left  Binding
	Right Expression
}

type ArrayPattern struct {
	Elements []Binding
}

type ObjectPattern struct {
	Properties []Binding
}

// #region Statements

func (bs *BlockStatement) statement()      {}
func (bs *BreakStatement) statement()      {}
func (cs *ClassDeclaration) statement()    {}
func (cs *ContinueStatement) statement()   {}
func (ds *DebuggerStatement) statement()   {}
func (dws *DoWhileStatement) statement()   {}
func (es *EmptyStatement) statement()      {}
func (ed *ExportDeclaration) statement()   {}
func (es *ExpressionStatement) statement() {}
func (fis *ForInStatement) statement()     {}
func (fos *ForOfStatement) statement()     {}
func (fs *ForStatement) statement()        {}
func (fd *FunctionDeclaration) statement() {}
func (is *IfStatement) statement()         {}
func (id *ImportDeclaration) statement()   {}
func (ls *LabeledStatement) statement()    {}
func (rs *ReturnStatement) statement()     {}
func (ss *SwitchStatement) statement()     {}
func (ts *ThrowStatement) statement()      {}
func (ts *TryCatchStatement) statement()   {}
func (vd *VariableDeclaration) statement() {}
func (ws *WhileStatement) statement()      {}
func (ws *WithStatement) statement()       {}

type BlockStatement struct {
	BaseNode
	Statements []Statement
}

type BreakStatement struct {
	BaseNode
	Label *Identifier
}

type ContinueStatement struct {
	BaseNode
	Label *Identifier
}

type ClassDeclaration struct {
	BaseNode
	Label *Identifier
}

type DebuggerStatement struct {
	BaseNode
}

type DoWhileStatement struct {
	BaseNode
	Condition Expression
	Body      *BlockStatement
}

type EmptyStatement struct {
	BaseNode
}

type ExportDeclaration struct {
	BaseNode
}

type ExpressionStatement struct {
	BaseNode
	Expr Expression
}

type ForInStatement struct {
	BaseNode
	Variable Statement
	Object   Expression
	Body     Statement
}

type ForOfStatement struct {
	BaseNode
	Await    bool
	Variable Statement
	Iterable Expression
	Body     Statement
}

type ForStatement struct {
	BaseNode
	Init         Statement
	Condition    Expression
	Afterthought Expression
	Body         Statement
}

type FunctionDeclaration struct {
	BaseNode
	Name        *Identifier
	Params      []Binding
	Body        *BlockStatement
	IsGenerator bool
	IsAsync     bool
}

type IfStatement struct {
	BaseNode
	Condition Expression
	Then      Statement
	Else      Statement
}

type ImportKind uint8

const (
	ImportKindNamed ImportKind = iota
	ImportKindDefault
	ImportKindNamespace
)

type ImportSpecifier struct {
	BaseNode
	Imported, Local *Identifier
	Kind            ImportKind
}

type ImportDeclaration struct {
	BaseNode
	Path       Expression
	Specifiers []*ImportSpecifier
}

type LabeledStatement struct {
	BaseNode
	Label *Identifier
	Body  Statement
}

type ReturnStatement struct {
	BaseNode
	ReturnValue Expression
}

type Case struct {
	BaseNode
	Value Expression
	Body  []Statement
}

type SwitchStatement struct {
	BaseNode
	Test  Expression
	Cases []Case
}

type ThrowStatement struct {
	BaseNode
	Argument Expression
}

type Catch struct {
	BaseNode
	Parameter Expression
	Block     BlockStatement
}

type TryCatchStatement struct {
	BaseNode
	Block   BlockStatement
	Catch   *Catch
	Finally *BlockStatement
}

type VariableDeclaration struct {
	BaseNode
	Token lexer.Token // the first token of the expression
	Name  *Identifier
	Value Expression
}

type WhileStatement struct {
	BaseNode
	Condition Expression
	Body      *BlockStatement
}

type WithStatement struct {
	BaseNode
	Expr Expression
	Body Statement
}

// #region Expressions

func (ale *ArrayLiteralExpression) expression()   {}
func (af *ArrowFunction) expression()             {}
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
func (ie *ImportExpression) expression()          {}
func (ye *YieldExpression) expression()           {}

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

type ObjectProperty struct {
	Key       Expression
	Value     Expression
	Computed  bool
	Shorthand bool
}

type ObjectLiteralExpression struct {
	Properties []ObjectProperty
}

type NewExpression struct {
	Constructor Expression
	Arguments   []Expression
}

type ParenthesizedExpression struct {
	Expr Expression
}

type ArrowFunction struct {
	Args []Expression
	Body Expression
}

type ImportExpression struct {
	Source Expression
}

type YieldExpression struct {
	Expr Expression
}
