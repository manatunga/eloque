package resolver

import (
	"github.com/manatunga/eloque/internal/frontend/ast"
)

// Resolution contains the semantic bindings produced by lexical name
// resolution for a single compilation.
//
// Symbols have compilation-local identities and are stored independently
// of the AST. Later compiler stages use the resolution result to retrieve
// declaration metadata and resolved identifier bindings.
//
// A Resolution must not outlive the AST source storage referenced by
// its symbols.
type Resolution struct {
	symbols  symbolArena
	bindings []SymbolID
}

// newResolution creates an empty resolution result for an AST containing
// nodeCount nodes, reserving space for expectedDeclarations symbols.
//
// All node bindings begin unbound. Symbol capacity is only a hint;
// symbols are still assigned sequential IDs as declarations are registered.
func newResolution(nodeCount int, expectedDeclarations int) *Resolution {
	bindings := make([]SymbolID, nodeCount)
	for i := range bindings {
		bindings[i] = InvalidSymbolID
	}

	return &Resolution{
		symbols:  newSymbolArena(expectedDeclarations),
		bindings: bindings,
	}
}

// binding retrieves the symbol associated with an AST node.
//
// It returns the symbol ID and true when the node has a valid binding.
// Unbound nodes and out-of-range node IDs return InvalidSymbolID
// and false without panicking.
func (r *Resolution) binding(nodeID ast.NodeID) (SymbolID, bool) {
	if uint64(nodeID) >= uint64(len(r.bindings)) {
		return InvalidSymbolID, false
	}

	symbolID := r.bindings[nodeID]

	if symbolID == InvalidSymbolID {
		return InvalidSymbolID, false
	}

	return symbolID, true
}

// bind associates an AST node with a resolved symbol.
//
// A node may be bound only once. The method rejects out-of-range
// node IDs, InvalidSymbolID, and attempts to overwrite an existing
// binding, returning false without modifying the table.
//
// The caller is responsible for supplying an allocated symbol ID.
func (r *Resolution) bind(nodeID ast.NodeID, symbolID SymbolID) bool {
	if uint64(nodeID) >= uint64(len(r.bindings)) {
		return false
	}
	if symbolID == InvalidSymbolID {
		return false
	}
	if r.bindings[nodeID] != InvalidSymbolID {
		return false
	}

	r.bindings[nodeID] = symbolID
	return true
}

// Symbol retrieves declaration metadata for a symbol ID.
//
// It returns the symbol and true when id identifies a declaration
// in this resolution result. Invalid or out-of-range IDs return
// the zero-value Symbol and false.
func (r *Resolution) Symbol(id SymbolID) (Symbol, bool) {
	return r.symbols.lookup(id)
}

// Declaration retrieves the symbol associated with a declaration node.
//
// It returns the symbol ID and true when the node has a recorded
// declaration binding. Unbound or out-of-range nodes return
// InvalidSymbolID and false.
//
// The resolver is responsible for recording declaration bindings
// only for AST nodes that introduce symbols.
func (r *Resolution) Declaration(nodeID ast.NodeID) (SymbolID, bool) {
	return r.binding(nodeID)
}

// Reference retrieves the symbol associated with an identifier use
// or assignment target.
//
// It returns the symbol ID and true when the node has a recorded
// reference binding. Unbound or out-of-range nodes return
// InvalidSymbolID and false.
//
// The resolver is responsible for recording reference bindings
// only for AST nodes that use existing symbols.
func (r *Resolution) Reference(nodeID ast.NodeID) (SymbolID, bool) {
	return r.binding(nodeID)
}
