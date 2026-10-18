package parser

import (
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// positionBenchmarkSpanSink prevents the compiler from discarding
// the reconstructed source spans.
var positionBenchmarkSpanSink source.Span

// positionBenchmarkOffsetsSink prevents the compiler from discarding
// the byte-offset values.
var positionBenchmarkOffsetsSink uint32

// positionBenchmarkSource creates a representative multiline Eloque
// program containing declarations and output statements.
func positionBenchmarkSource(lines int) string {
	var builder strings.Builder
	builder.Grow(lines * 32)

	for range lines {
		builder.WriteString("let value be (12 + 34) * 56\n")
	}

	return builder.String()
}

// positionBenchmarkTokens collects compact tokens from the source.
func positionBenchmarkTokens(b *testing.B, text string) []lexer.CompactToken {
	b.Helper()

	scanner := lexer.NewCompact(text)
	tokens := make([]lexer.CompactToken, 0, len(text)/4)

	for {
		token, err := scanner.NextToken()
		if err != nil {
			b.Fatal(err)
		}

		tokens = append(tokens, token)

		if token.Kind == lexer.TokenEOF {
			return tokens
		}
	}
}

// BenchmarkCompactTokenSpan measures reconstruction of complete
// line/column spans from compact token byte offsets.
func BenchmarkCompactTokenSpan(b *testing.B) {
	text := positionBenchmarkSource(300)
	tokens := positionBenchmarkTokens(b, text)
	parser := NewCompact(text, tokens)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var span source.Span

		for _, token := range tokens {
			span = parser.tokenSpan(token)
		}

		positionBenchmarkSpanSink = span
	}
}

// BenchmarkCompactOffsetAccess measures reading the same tokens'
// byte offsets without reconstructing line/column positions.
func BenchmarkCompactOffsetAccess(b *testing.B) {
	text := positionBenchmarkSource(300)
	tokens := positionBenchmarkTokens(b, text)

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var end uint32

		for _, token := range tokens {
			end = token.End
		}

		positionBenchmarkOffsetsSink = end
	}
}
