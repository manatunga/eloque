package ast

import (
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// Arena stores compact AST nodes and their kind-specific payloads.
// Nodes are identified by stable NodeIDs rather than memory addresses.
type Arena struct {
	sourceText string

	nodes       []FlatNode
	strings     []string
	binaries    []BinaryPayload
	unaries     []UnaryPayload
	conversions []ConversionPayload
	bindings    []BindingPayload
	statements  []NodeID
}

// ArenaCheckpoint records the arena's allocation boundaries.
//
// Restoring a checkpoint discards nodes and payloads appended after
// it. Existing node IDs and statement references remain valid.
type ArenaCheckpoint struct {
	nodes       int
	strings     int
	binaries    int
	unaries     int
	conversions int
	bindings    int
	statements  int
	arena       *Arena
}

// NewArena creates an AST arena backed by the given source text.
// The arena retains the source so nodes can reference its byte ranges.
func NewArena(sourceText string) *Arena {
	return &Arena{
		sourceText: sourceText,
	}
}

// Checkpoint captures the current allocation boundaries of the arena.
//
// Creating a checkpoint performs no allocations.
func (a *Arena) Checkpoint() ArenaCheckpoint {
	return ArenaCheckpoint{
		nodes:       len(a.nodes),
		strings:     len(a.strings),
		binaries:    len(a.binaries),
		unaries:     len(a.unaries),
		conversions: len(a.conversions),
		bindings:    len(a.bindings),
		statements:  len(a.statements),
		arena:       a,
	}
}

// Rollback discards allocations and statement references added after
// the supplied checkpoint.
//
// Checkpoints must belong to this arena and must not describe lengths
// greater than the arena's current storage lengths.
func (a *Arena) Rollback(cp ArenaCheckpoint) {
	if cp.arena != a ||
		cp.nodes > len(a.nodes) ||
		cp.strings > len(a.strings) ||
		cp.binaries > len(a.binaries) ||
		cp.unaries > len(a.unaries) ||
		cp.conversions > len(a.conversions) ||
		cp.bindings > len(a.bindings) ||
		cp.statements > len(a.statements) {
		panic("ast: invalid arena checkpoint")
	}

	a.nodes = a.nodes[:cp.nodes]
	a.strings = a.strings[:cp.strings]
	a.binaries = a.binaries[:cp.binaries]
	a.unaries = a.unaries[:cp.unaries]
	a.conversions = a.conversions[:cp.conversions]
	a.bindings = a.bindings[:cp.bindings]
	a.statements = a.statements[:cp.statements]
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

// AddFloat appends a floating-point literal node to the arena.
//
// The node retains the literal's original source span. Numeric
// conversion is deferred to semantic analysis.
func (a *Arena) AddFloat(span source.ByteSpan) NodeID {
	a.SourceText(span) // Validate the source range.

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeFloatLiteral,
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

// AddConversionPayload appends a numeric conversion payload to the arena's
// conversion payload buffer and returns its zero-based index.
//
// This is a low-level storage operation. It does not validate whether
// Operand references an existing expression node, whether TargetType
// identifies a valid source range, or whether the payload is associated
// with a corresponding AST node.
//
// Callers constructing conversion expression nodes should use AddConversion,
// which validates the operand and source spans before storing the payload.
//
// It panics if the next payload index cannot be represented by uint32.
func (a *Arena) AddConversionPayload(payload ConversionPayload) uint32 {
	checkPayloadCapacity(len(a.conversions))
	index := uint32(len(a.conversions))
	a.conversions = append(a.conversions, payload)
	return index
}

// ConversionPayload returns the numeric conversion payload at index.
//
// It panics if index is outside the conversion payload buffer.
func (a *Arena) ConversionPayload(index uint32) ConversionPayload {
	if uint64(index) >= uint64(len(a.conversions)) {
		panic("ast: invalid conversion payload index")
	}
	return a.conversions[index]
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

// AddConversion creates a numeric conversion expression node and its payload.
//
// Operand must identify an existing expression node. TargetType must be a
// valid source span identifying the destination type's spelling. The complete
// expression span must cover the conversion keyword, operand, and destination
// type.
//
// It panics if any source span is invalid, Operand does not identify an
// existing expression node, or the arena exhausts its available node or
// payload indices.
func (a *Arena) AddConversion(
	span source.ByteSpan,
	operand NodeID,
	targetType source.ByteSpan,
) NodeID {
	a.SourceText(span)
	a.SourceText(targetType)
	a.requireExpression(operand)
	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}

	a.requireExpression(operand)
	checkPayloadCapacity(len(a.conversions))

	index := a.AddConversionPayload(ConversionPayload{
		Operand:    operand,
		TargetType: targetType,
	})

	return a.Add(FlatNode{
		Span: span,
		Kind: NodeConversionExpression,
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
// Callers constructing binding statements should use AddLet,
// AddLetWithTypeAnnotation, or AddChange, which validate the relevant
// source spans and value expression before storing the payload.
//
// This is a low-level storage operation. It does not validate the
// binding name's source span, the type annotation's source span,
// the value's NodeID or node kind, or whether the payload is
// associated with a statement node.
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
	return a.addBinding(NodeLetStatement, span, name, source.ByteSpan{}, value)
}

// AddLetWithTypeAnnotation creates a variable declaration node with an
// explicit type annotation and its binding payload.
//
// The annotation identifies the type's spelling in the retained source,
// including an optional suffix such as '?'.
// It panics if the annotation is invalid or empty.
func (a *Arena) AddLetWithTypeAnnotation(
	span source.ByteSpan,
	name source.ByteSpan,
	typeAnnotation source.ByteSpan,
	value NodeID,
) NodeID {
	if typeAnnotation.Start == typeAnnotation.End {
		panic("ast: empty type annotation")
	}
	return a.addBinding(NodeLetStatement, span, name, typeAnnotation, value)
}

// AddChange creates a variable reassignment node and its binding payload.
func (a *Arena) AddChange(
	span source.ByteSpan,
	name source.ByteSpan,
	value NodeID,
) NodeID {
	return a.addBinding(NodeChangeStatement, span, name, source.ByteSpan{}, value)
}

// addBinding creates a statement node with a binding payload.
func (a *Arena) addBinding(
	kind NodeKind,
	span source.ByteSpan,
	name source.ByteSpan,
	typeAnnotation source.ByteSpan,
	value NodeID,
) NodeID {
	a.SourceText(span)
	a.SourceText(name)

	a.SourceText(typeAnnotation)
	if typeAnnotation.Start < typeAnnotation.End {
		if typeAnnotation.Start < span.Start ||
			typeAnnotation.End > span.End {
			panic("ast: type annotation is out of bounds")
		}
		if kind != NodeLetStatement {
			panic("ast: type annotation is only allowed with let statements")
		}
	}

	a.requireExpression(value)
	if uint64(len(a.nodes)) >= uint64(InvalidNodeID) {
		panic("ast: node arena exhausted")
	}
	checkPayloadCapacity(len(a.bindings))

	index := a.AddBindingPayload(BindingPayload{
		Name:           name,
		Value:          value,
		TypeAnnotation: typeAnnotation,
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
