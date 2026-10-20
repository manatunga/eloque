package parser

import (
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/source"
)

// FlatParser maintains the state required to construct Eloque's
// arena-backed abstract syntax tree.
//
// It embeds the existing Parser to reuse token navigation, source
// access, operator precedence helpers, and positioned diagnostics.
// This preserves M1 parsing behavior while allowing H2 to construct
// compact nodes identified by ast.NodeID values.
//
// The arena owns all constructed nodes and their associated payloads.
// Source text is retained by the arena so identifiers and integer
// literals can be recovered from byte spans without storing separate
// strings.
type FlatParser struct {
	*Parser
	arena *ast.Arena
}

// flatExpressionResult represents the result of parsing an H2 expression.
//
// id identifies the semantic expression node stored in the AST arena.
// extent records the complete half-open byte range consumed by parsing.
//
// These ranges may differ when parentheses surround an expression.
// Grouping parentheses extend the syntactic extent but do not introduce
// an additional AST node or modify the inner node's source span.
//
// Enclosing unary and binary expressions use extent to calculate their
// own source spans accurately.
type flatExpressionResult struct {
	id     ast.NodeID
	extent source.ByteSpan
}

// NewFlatParser constructs an arena-backed parser using Eloque's
// streaming lexer.
//
// The parser requests tokens incrementally and stores constructed
// AST nodes in an arena.
//
// This constructor initializes parser state without parsing the
// complete source. The production Parse function performs that
// operation and returns a FlatProgram.
func NewFlatParser(sourceText string) *FlatParser {
	return &FlatParser{
		Parser: newStreamingParser(sourceText),
		arena:  ast.NewArena(sourceText),
	}
}
