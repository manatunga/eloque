package ast

import "github.com/manatunga/eloque/internal/source"

// Node represents any node in Eloque's abstract syntax tree.
// Every node preserves its location in the original source.
type Node interface {
	Span() source.Span
}

// Expression represents a construct that produces a value.
type Expression interface {
	Node
	expressionNode()
}

// Statement represents a construct that performs a statement-level
// operation, such as declaring a variable or producing output.
type Statement interface {
	Node
	statementNode()
}
