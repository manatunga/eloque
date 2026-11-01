package resolver

import (
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/source"
)

// SymbolID uniquely identifies a declared symbol within a single
// compilation unit.
//
// Symbol IDs are stable for the lifetime of a resolution result,
// but are not guaranteed to remain stable across recompilations.
// They do not represent Monowire registers or runtime storage.
type SymbolID uint32

// InvalidSymbolID represents the absence of a valid symbol binding.
//
// The sentinel is reserved and must never be assigned to a
// successfully declared symbol.
const InvalidSymbolID SymbolID = ^SymbolID(0)

// Symbol describes a declaration discovered during name resolution.
//
// Name retains the declaration's identifier spelling.
// Declaration identifies the AST node introducing the symbol.
// Span identifies the declaration name's half-open byte range.
//
// Symbols contain lexical binding information only. Type checking,
// definite-assignment analysis, and runtime storage allocation are
// responsibilities of later compiler stages.
type Symbol struct {
	Name        string
	Declaration ast.NodeID
	Span        source.ByteSpan
}
