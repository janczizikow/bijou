package parser

import "github.com/janczizikow/bijou/internal/lexer"

type Precedence uint8

// https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Operators/Operator_precedence
const (
	Lowest Precedence = iota
	Comma
	Assignment
	LogicalOr
	LogicalAnd
	BitwiseOR
	BitwiseXOR
	BitwiseAND
	Equality
	Relational
	BitwiseShift
	Add
	Multiply
	Exponentiation
	Prefix
	Postfix
	PNew
	Call
	Member
)

var precedences = map[lexer.Token]Precedence{
	lexer.Dot:             Call,
	lexer.OptionalChaning: Call,
	lexer.LBracket:        Call,

	// TODO: prefix vs. postfix operators

	lexer.Exponentiation: Exponentiation,

	lexer.Multiplication: Multiply,
	lexer.Division:       Multiply,

	lexer.Plus:  Add,
	lexer.Minus: Add,

	lexer.BitwiseShiftLeft:          BitwiseShift,
	lexer.BitwiseShiftRight:         BitwiseShift,
	lexer.BitwiseUnsignedShiftRight: BitwiseShift,

	lexer.LT:  Relational,
	lexer.LTE: Relational,
	lexer.GT:  Relational,
	lexer.GTE: Relational,

	// TODO: x in y
	// TODO: x instance of y

	lexer.Equality:         Equality,
	lexer.Inequality:       Equality,
	lexer.StrictEquality:   Equality,
	lexer.StrictInequality: Equality,

	lexer.BitwiseAND:        BitwiseAND,
	lexer.BitwiseXOR:        BitwiseXOR,
	lexer.BitwiseOR:         BitwiseOR,
	lexer.LogicalAND:        LogicalAnd,
	lexer.LogicalOR:         LogicalOr,
	lexer.NullishCoalescing: LogicalOr,

	// TODO: spread: Assignment,
	lexer.YIELD:                        Assignment,
	lexer.NullishCoalescingAssignment:  Assignment,
	lexer.LogicalORAssignment:          Assignment,
	lexer.LogicalANDAssignment:         Assignment,
	lexer.BitwiseORAssignment:          Assignment,
	lexer.BitwiseXORAssignment:         Assignment,
	lexer.BitwiseANDAssignment:         Assignment,
	lexer.UnsignedRightShiftAssignment: Assignment,
	lexer.RightShiftAssignment:         Assignment,
	lexer.LeftShiftAssignment:          Assignment,
	lexer.RemainderAssignment:          Assignment,
	lexer.DivisionAssignment:           Assignment,
	lexer.MultiplicationAssignment:     Assignment,
	lexer.ExponentiationAssignment:     Assignment,
	lexer.SubtractionAssignment:        Assignment,
	lexer.AdditionAssignment:           Assignment,
	lexer.Assignment:                   Assignment,

	lexer.Comma: Comma,
}
