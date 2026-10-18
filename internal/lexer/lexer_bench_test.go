package lexer

import (
	"fmt"
	"strings"
	"testing"
	"unsafe"
)

// BenchmarkNextToken measures the time and memory required to tokenize
// a representative Eloque source program, including indentation,
// identifiers, integer literals, comments, and dedentation.
func BenchmarkNextToken(b *testing.B) {
	const source = `let count be 10
# Initialize the counter
say count
    let value be 42 # Inline comment
    say value
        say 100
    say count
say 0
`

	b.SetBytes(int64(len(source)))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		l := New(source)

		for {
			token, err := l.NextToken()
			if err != nil {
				b.Fatal(err)
			}

			if token.Kind == TokenEOF {
				break
			}
		}
	}
}

// BenchmarkNextTokenFlat measures tokenization performance for source
// containing identifiers, integer literals, and keywords without
// indentation, comments, or nested scopes.
func BenchmarkNextTokenFlat(b *testing.B) {
	const source = `let count be 10
let value be 42
say count
say value
let result be 100
say result
`

	b.SetBytes(int64(len(source)))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		l := New(source)

		for {
			token, err := l.NextToken()
			if err != nil {
				b.Fatal(err)
			}

			if token.Kind == TokenEOF {
				break
			}
		}
	}
}

// TestTokenMemoryLayout reports the in-memory sizes of the token
// representation and its constituent types.
func TestTokenMemoryLayout(t *testing.T) {
	t.Logf("TokenKind: %d bytes", unsafe.Sizeof(TokenKind(0)))
	t.Logf("Position:  %d bytes", unsafe.Sizeof(Position{}))
	t.Logf("Token:     %d bytes", unsafe.Sizeof(Token{}))
}

// BenchmarkNextTokenScaling measures lexer throughput and memory usage
// as the source length increases while indentation depth stays bounded.
func BenchmarkNextTokenScaling(b *testing.B) {
	const sample = `let count be 10
# Initialize the counter
say count
    let value be 42 # Inline comment
    say value
        say 100
    say count
say 0
`

	for _, repetitions := range []int{1, 10, 100, 1000} {
		b.Run(fmt.Sprintf("x%d", repetitions), func(b *testing.B) {
			source := strings.Repeat(sample, repetitions)

			b.SetBytes(int64(len(source)))
			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				l := New(source)

				for {
					token, err := l.NextToken()
					if err != nil {
						b.Fatal(err)
					}

					if token.Kind == TokenEOF {
						break
					}
				}
			}
		})
	}
}
