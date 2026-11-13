package ast

import (
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// BinaryPayload stores the operands and operator of a binary expression.
// Left and Right identify nodes within the same AST arena.
type BinaryPayload struct {
	Left     NodeID
	Right    NodeID
	Operator lexer.TokenKind
}

// UnaryPayload stores the operator and operand of a unary expression.
type UnaryPayload struct {
	Operand  NodeID
	Operator lexer.TokenKind
}

// BindingPayload stores the name and value of a variable
// declaration or reassignment.
//
// Name identifies the variable's spelling in the retained source.
// Value identifies its expression in the AST arena.
type BindingPayload struct {
	Name  source.ByteSpan
	Value NodeID
}

// ConversionPayload stores the operand and destination type of a
// numeric conversion expression.
//
// Operand identifies the expression being converted. TargetType
// identifies the destination type's spelling in the retained source.
type ConversionPayload struct {
	Operand    NodeID
	TargetType source.ByteSpan
}
