package lexer

import (
	"strings"
	"testing"
)

// expectedToken describes the readable token representation used by tests.
// It is deliberately separate from the compact production Token.
type expectedToken struct {
	Kind   TokenKind
	Lexeme string
	Span   Span
}

// tokenPosition converts a source byte offset into a one-based line
// and column position for test assertions.
func tokenPosition(source string, offset uint32) Position {
	prefix := source[:int(offset)]

	line := strings.Count(prefix, "\n") + 1
	lastNewline := strings.LastIndexByte(prefix, '\n')

	return Position{
		Line:   line,
		Column: len(prefix) - lastNewline,
	}
}

// assertToken compares a compact token with its expected lexical
// category, source lexeme, and source span.
func assertToken(t *testing.T, source string, got Token, want expectedToken) {
	t.Helper()

	if got.Kind != want.Kind {
		t.Errorf("kind = %s, want %s", got.Kind, want.Kind)
	}

	if got.Start > got.End || int(got.End) > len(source) {
		t.Fatalf("invalid token offsets: [%d, %d)", got.Start, got.End)
	}

	lexeme := source[got.Start:got.End]
	if lexeme != want.Lexeme {
		t.Errorf("lexeme = %q, want %q", lexeme, want.Lexeme)
	}

	span := Span{
		Start: tokenPosition(source, got.Start),
		End:   tokenPosition(source, got.End),
	}

	if span != want.Span {
		t.Errorf("span = %+v, want %+v", span, want.Span)
	}
}
