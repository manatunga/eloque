package parser

import (
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// Diagnostic represents a syntax error detected during parsing.
// Message explains the problem in human-readable form, while
// Location identifies the source range associated with the error.
// Expected lists acceptable token kinds when the error involves
// a token mismatch. Actual identifies the encountered token kind
// when one is relevant. Expected and Actual are optional metadata,
// allowing diagnostics for grammatical errors beyond token mismatches.
type Diagnostic struct {
	Message  string
	Location source.Span
	Expected []lexer.TokenKind
	Actual   *lexer.TokenKind
}
