package lexer

import (
	"testing"
	"unsafe"
)

// benchmarkCompactToken is the experimental offset-based token.
type benchmarkCompactToken struct {
	Start uint32
	End   uint32
	Kind  TokenKind
}

var (
	emittedFullToken    Token
	emittedCompactToken benchmarkCompactToken
)

// BenchmarkTokenEmission compares the cost of constructing
// full and compact token records from equivalent scanner state.
func BenchmarkTokenEmission(b *testing.B) {
	source := "example_identifier"

	l := New(source)
	l.start = 0
	l.current = len(source)
	l.startPosition = Position{Line: 1, Column: 1}
	l.line = 1
	l.column = len(source) + 1

	b.Run("Full", func(b *testing.B) {
		for b.Loop() {
			emittedFullToken = l.makeToken(TokenIdent)
		}
	})

	b.Run("Compact", func(b *testing.B) {
		for b.Loop() {
			emittedCompactToken = benchmarkCompactToken{
				Start: uint32(l.start),
				End:   uint32(l.current),
				Kind:  TokenIdent,
			}
		}
	})
}

// TestExperimentalTokenLayout reports the size of each
// representation on the current architecture.
func TestExperimentalTokenLayout(t *testing.T) {
	t.Logf("Full token: %d bytes", unsafe.Sizeof(Token{}))
	t.Logf(
		"Compact token: %d bytes",
		unsafe.Sizeof(benchmarkCompactToken{}),
	)
}
