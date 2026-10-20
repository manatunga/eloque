package ast

import "github.com/manatunga/eloque/internal/source"

// FlatProgram represents the root of a parsed Eloque source file.
//
// Arena owns the AST nodes and their associated payloads.
// Statements identifies an ordered, half-open range within the
// arena's statement buffer.
//
// Span identifies the source range occupied by the program.
// For nonempty programs, it extends from the first statement's
// beginning to the last statement's end. Empty programs have
// a zero-width span at EOF.
type FlatProgram struct {
	Arena      *Arena
	Statements StatementRange
	Span       source.ByteSpan
}

// StatementRange identifies a half-open range of statement IDs
// within the arena's contiguous statement buffer.
type StatementRange struct {
	Start uint32
	End   uint32
}
