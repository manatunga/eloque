package ast

import "github.com/manatunga/eloque/internal/source"

// NodeID uniquely identifies a node within an AST arena.
// IDs correspond to indices in the arena's node slice.
type NodeID uint32

// InvalidNodeID is a sentinel indicating that no node is present.
// It must never be assigned to a valid node.
const InvalidNodeID NodeID = ^NodeID(0)

// NodeKind identifies the syntactic category of a node.
type NodeKind uint8

const (
	NodeIntegerLiteral NodeKind = iota
	NodeStringLiteral
	NodeIdentifier
	NodeUnaryExpression
	NodeBinaryExpression
	NodeLetStatement
	NodeChangeStatement
	NodeSayStatement
)

// FlatNode is the compact header of an AST node.
//
// Span records the node's half-open byte range in the source.
// Kind determines how Data is interpreted.
// Data stores either a direct value or an index into a
// kind-specific payload arena.
//
// Nodes reference one another through NodeIDs rather than pointers.
type FlatNode struct {
	Span source.ByteSpan
	Kind NodeKind
	Data uint32
}

// IsExpression reports whether kind identifies an expression node.
func (kind NodeKind) IsExpression() bool {
	switch kind {
	case NodeIntegerLiteral,
		NodeStringLiteral,
		NodeIdentifier,
		NodeUnaryExpression,
		NodeBinaryExpression:
		return true
	default:
		return false
	}
}

// IsStatement reports whether kind identifies a statement node.
func (kind NodeKind) IsStatement() bool {
	switch kind {
	case NodeLetStatement,
		NodeChangeStatement,
		NodeSayStatement:
		return true
	default:
		return false
	}
}
