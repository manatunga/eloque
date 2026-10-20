package ast

import (
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// IntegerLiteral represents an integer literal in Eloque source code.
// Value preserves the original decimal digits without converting them
// into a numeric type. Numeric range validation is handled by a later
// compilation stage. Location spans the complete literal.
type IntegerLiteral struct {
	Value    string
	Location source.Span
}

// StringLiteral represents a quoted string literal in Eloque source code.
// Value contains the decoded string contents, excluding the surrounding
// quotation marks. Escape sequences are interpreted during literal
// processing. Location spans the complete quoted literal.
type StringLiteral struct {
	Value    string
	Location source.Span
}

// Identifier represents a reference to a named binding in an expression.
// Name preserves the identifier's source spelling. Resolving that name
// to a declaration is the responsibility of the name resolver.
// Location spans the identifier itself.
type Identifier struct {
	Name     string
	Location source.Span
}

// UnaryExpression represents a prefix operator applied to an expression.
// Operator identifies the unary operation, while Operand contains the
// expression to which it applies. The parser determines their grammatical
// relationship without evaluating the result.
// Location spans the operator and its complete operand.
type UnaryExpression struct {
	Operator lexer.TokenKind
	Operand  Expression
	Location source.Span
}

// BinaryExpression represents an operator applied to two expressions.
// Left and Right preserve the operand expressions, while Operator
// identifies the operation between them. The parser determines the
// tree structure according to operator precedence and associativity.
// Location spans the complete expression, including both operands.
type BinaryExpression struct {
	Left     Expression
	Operator lexer.TokenKind
	Right    Expression
	Location source.Span
}

// Span returns the source range occupied by the complete integer literal.
// The range includes the literal's digits and excludes adjacent syntax.
func (n *IntegerLiteral) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the complete string literal.
// The range includes the surrounding quotation marks.
func (n *StringLiteral) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the identifier.
// It identifies the source occurrence, not the declaration it references.
func (n *Identifier) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the complete unary expression.
// The range includes the prefix operator and its operand.
func (n *UnaryExpression) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the complete binary expression.
// The range includes both operands and the operator between them.
func (n *BinaryExpression) Span() source.Span {
	return n.Location
}

func (n *IntegerLiteral) expressionNode() {}

func (n *StringLiteral) expressionNode() {}

func (n *Identifier) expressionNode() {}

func (n *UnaryExpression) expressionNode() {}

func (n *BinaryExpression) expressionNode() {}
