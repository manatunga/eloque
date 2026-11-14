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

// BindingPayload stores the name, value, and optional type annotation of a
// variable declaration or reassignment.
//
// Name identifies the variable's spelling in the retained source.
// Value identifies its expression in the AST arena.
// TypeAnnotation identifies the source span of an explicit type annotation
// on a let declaration, including an optional suffix such as '?'.
// A zero-width span represents the absence of an annotation.
type BindingPayload struct {
	Name           source.ByteSpan
	Value          NodeID
	TypeAnnotation source.ByteSpan
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
