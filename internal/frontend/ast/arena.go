package ast

import (
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// Arena stores compact AST nodes and their kind-specific payloads.
// Nodes are identified by stable NodeIDs rather than memory addresses.
type Arena struct {
	sourceText string

	nodes      []FlatNode
	strings    []string
	binaries   []BinaryPayload
	unaries    []UnaryPayload
	bindings   []BindingPayload
	statements []NodeID
}

// NewArena creates an AST arena backed by the given source text.
// The arena retains the source so nodes can reference its byte ranges.
func NewArena(sourceText string) *Arena {
	return &Arena{
		sourceText: sourceText,
	}
}

// checkPayloadCapacity panics when appending another payload would
// require an index outside the uint32 range.
//
// Payload indices do not reserve a sentinel value, unlike NodeID.
// The check operates on the current length, before appending.
func checkPayloadCapacity(length int) {
	if uint64(length) > uint64(^uint32(0)) {
		panic("ast: payload arena exhausted")
	}
}

// StatementCount returns the number of statement references stored
// in the arena's statement sequence.
//
// The result can be used to construct half-open statement ranges
// without exposing or copying the underlying storage.
func (a *Arena) StatementCount() uint32 {
	return uint32(len(a.statements))
}

// SourceText returns the source bytes covered by span without copying them.
// It panics if the span is outside the retained source.
func (a *Arena) SourceText(span source.ByteSpan) string {
	if span.Start > span.End || uint64(span.End) > uint64(len(a.sourceText)) {
		panic("ast: invalid source span")
	}

	return a.sourceText[int(span.Start):int(span.End)]
}

// requireExpression verifies that id refers to an expression node.
func (a *Arena) requireExpression(id NodeID) {
	if uint64(id) >= uint64(len(a.nodes)) {
		panic("ast: invalid expression ID")
	}

	if !a.nodes[id].Kind.IsExpression() {
		panic("ast: node is not an expression")
	}
}

// Add appends a node and returns its assigned NodeID.
// IDs are assigned sequentially, starting at zero.
// Add panics if the arena exhausts its available node IDs.
func (a *Arena) Add(node FlatNode) NodeID {
	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}

	id := NodeID(len(a.nodes))
	a.nodes = append(a.nodes, node)
	return id
}

// Get returns the node identified by id.
//
// It panics if id is InvalidNodeID or does not reference a node
// currently stored in the arena. The returned pointer must not be
// retained across operations that may append to the node buffer,
// because slice growth can relocate its backing array.
func (a *Arena) Get(id NodeID) *FlatNode {
	if id == InvalidNodeID || uint64(id) >= uint64(len(a.nodes)) {
		panic("ast: invalid node ID")
	}

	return &a.nodes[id]
}

// Len returns the number of nodes stored in the arena.
func (a *Arena) Len() int {
	return len(a.nodes)
}

// AddInteger creates an integer literal node backed by the source text.
// The original digits are preserved for later numeric validation.
func (a *Arena) AddInteger(span source.ByteSpan) NodeID {
	a.SourceText(span) // Validate the source range.

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeIntegerLiteral,
	})
}

// AddIdentifier creates an identifier node backed by the source text.
// Its spelling can be retrieved from the arena using the node's span.
func (a *Arena) AddIdentifier(span source.ByteSpan) NodeID {
	a.SourceText(span) // Validate the source range.

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeIdentifier,
	})
}

// AddString creates a string literal node containing a decoded value.
// The node's Data field indexes the arena's string storage.
func (a *Arena) AddString(
	span source.ByteSpan,
	value string,
) NodeID {
	a.SourceText(span) // Validate the source range.

	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}
	checkPayloadCapacity(len(a.strings))

	index := uint32(len(a.strings))
	a.strings = append(a.strings, value)

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeStringLiteral,
		Data: index,
	})
}

// StringValue returns the decoded contents of a string literal.
//
// It panics if id is invalid, if the referenced node is not a
// string literal, or if its payload index is out of bounds.
func (a *Arena) StringValue(id NodeID) string {
	node := a.Get(id)

	if node.Kind != NodeStringLiteral {
		panic("ast: node is not a string literal")
	}

	if uint64(node.Data) >= uint64(len(a.strings)) {
		panic("ast: invalid string payload index")
	}

	return a.strings[node.Data]
}

// AddBinaryPayload appends a binary expression payload to the arena's
// binary payload buffer and returns its zero-based index.
//
// This is a low-level storage operation. It does not validate whether
// Left and Right reference existing expression nodes, whether Operator
// is a valid binary operator, or whether the payload is associated with
// a corresponding AST node.
//
// Callers constructing binary expression nodes should use AddBinary,
// which validates operand references before storing the payload.
//
// It panics if the next payload index cannot be represented by uint32.
func (a *Arena) AddBinaryPayload(payload BinaryPayload) uint32 {
	checkPayloadCapacity(len(a.binaries))
	index := uint32(len(a.binaries))
	a.binaries = append(a.binaries, payload)
	return index
}

// BinaryPayload returns the binary expression payload at index.
//
// It panics if index is outside the binary payload buffer.
func (a *Arena) BinaryPayload(index uint32) BinaryPayload {
	if uint64(index) >= uint64(len(a.binaries)) {
		panic("ast: invalid binary payload index")
	}

	return a.binaries[index]
}

// AddUnaryPayload appends a unary expression payload to the arena's
// unary payload buffer and returns its zero-based index.
//
// This is a low-level storage operation. It does not validate whether
// Operand references an existing expression node, whether Operator
// is a valid unary operator, or whether the payload is associated with
// a corresponding AST node.
//
// Callers constructing unary expression nodes should use AddUnary,
// which validates the operand reference before storing the payload.
//
// It panics if the next payload index cannot be represented by uint32.
func (a *Arena) AddUnaryPayload(payload UnaryPayload) uint32 {
	checkPayloadCapacity(len(a.unaries))
	index := uint32(len(a.unaries))
	a.unaries = append(a.unaries, payload)
	return index
}

// UnaryPayload returns the unary expression payload at index.
//
// It panics if index is outside the unary payload buffer.
func (a *Arena) UnaryPayload(index uint32) UnaryPayload {
	if uint64(index) >= uint64(len(a.unaries)) {
		panic("ast: invalid unary payload index")
	}

	return a.unaries[index]
}

// AddBinary creates a binary expression node and its payload.
// Left and right must identify existing nodes in the arena.
func (a *Arena) AddBinary(
	span source.ByteSpan,
	left NodeID,
	operator lexer.TokenKind,
	right NodeID,
) NodeID {
	a.SourceText(span)
	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}
	a.requireExpression(left)
	a.requireExpression(right)
	checkPayloadCapacity(len(a.binaries))

	index := a.AddBinaryPayload(BinaryPayload{
		Left:     left,
		Right:    right,
		Operator: operator,
	})

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeBinaryExpression,
		Data: index,
	})
}

// AddUnary creates a unary expression node and its payload.
// Operand must identify an existing node in the arena.
func (a *Arena) AddUnary(
	span source.ByteSpan,
	operator lexer.TokenKind,
	operand NodeID,
) NodeID {
	a.SourceText(span)
	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}
	a.requireExpression(operand)
	checkPayloadCapacity(len(a.unaries))

	index := a.AddUnaryPayload(UnaryPayload{
		Operand:  operand,
		Operator: operator,
	})

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeUnaryExpression,
		Data: index,
	})
}

// AddBindingPayload appends a binding payload to the arena's
// binding payload buffer and returns its zero-based index.
//
// Binding payloads store the source span of a binding name and
// the NodeID of its associated value expression.
//
// This is a low-level storage operation. It does not validate the
// binding name's source span, the value's NodeID or node kind,
// or whether the payload is associated with a statement node.
//
// Callers constructing binding statements should use AddLet or
// AddChange, which validate the name span and value expression
// before storing the payload.
//
// It panics if the next payload index cannot be represented by uint32.
func (a *Arena) AddBindingPayload(payload BindingPayload) uint32 {
	checkPayloadCapacity(len(a.bindings))

	index := uint32(len(a.bindings))
	a.bindings = append(a.bindings, payload)
	return index
}

// BindingPayload returns the binding payload at index.
//
// It panics if index is outside the binding payload buffer.
func (a *Arena) BindingPayload(index uint32) BindingPayload {
	if uint64(index) >= uint64(len(a.bindings)) {
		panic("ast: invalid binding payload index")
	}

	return a.bindings[index]
}

// AddLet creates a variable declaration node and its binding payload.
func (a *Arena) AddLet(
	span source.ByteSpan,
	name source.ByteSpan,
	value NodeID,
) NodeID {
	return a.addBinding(NodeLetStatement, span, name, value)
}

// AddChange creates a variable reassignment node and its binding payload.
func (a *Arena) AddChange(
	span source.ByteSpan,
	name source.ByteSpan,
	value NodeID,
) NodeID {
	return a.addBinding(NodeChangeStatement, span, name, value)
}

// addBinding creates a statement node with a binding payload.
func (a *Arena) addBinding(
	kind NodeKind,
	span source.ByteSpan,
	name source.ByteSpan,
	value NodeID,
) NodeID {
	a.SourceText(span)
	a.SourceText(name)

	a.requireExpression(value)
	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}
	checkPayloadCapacity(len(a.bindings))

	index := a.AddBindingPayload(BindingPayload{
		Name:  name,
		Value: value,
	})

	return a.Add(FlatNode{
		Span: span,
		Kind: kind,
		Data: index,
	})
}

// AddSay creates a say statement containing an expression.
func (a *Arena) AddSay(
	span source.ByteSpan,
	value NodeID,
) NodeID {
	a.SourceText(span)
	a.requireExpression(value)
	return a.Add(FlatNode{
		Span: span,
		Kind: NodeSayStatement,
		Data: uint32(value),
	})
}

// AddStatement appends a statement node ID to the ordered statement buffer.
// It panics if id does not identify a statement node.
func (a *Arena) AddStatement(id NodeID) {
	if uint64(id) >= uint64(len(a.nodes)) {
		panic("ast: invalid statement ID")
	}

	if !a.nodes[id].Kind.IsStatement() {
		panic("ast: node is not a statement")
	}

	if uint64(len(a.statements)) >= uint64(^uint32(0)) {
		panic("ast: statement buffer exhausted")
	}

	a.statements = append(a.statements, id)
}

// StatementRange returns the statement IDs within the specified
// half-open range [Start, End).
//
// It panics if the range is reversed or extends beyond the arena's
// statement buffer. The returned slice aliases arena-owned storage;
// callers must not modify its contents.
func (a *Arena) StatementRange(r StatementRange) []NodeID {
	if r.Start > r.End || uint64(r.End) > uint64(len(a.statements)) {
		panic("ast: invalid statement range")
	}

	return a.statements[r.Start:r.End]
}
