package lexer

import (
	"fmt"
	"strings"
	"testing"
)

// compactToken is an experimental token representation.
// Its offsets are byte positions in the original source.
type compactToken struct {
	Start uint32
	End   uint32
	Kind  TokenKind
}

var (
	storedTokens  []Token
	storedCompact []CompactToken
)

// benchmarkSource generates a repeatable, flat Eloque program.
func benchmarkSource(statements int) string {
	return strings.Repeat(
		"let value be 10 + 20\n"+
			"change value to value * 2\n"+
			"say value\n",
		statements/3+1,
	)
}

// collectTokens scans source and retains every token.
func collectTokens(source string) ([]Token, error) {
	l := New(source)
	var tokens []Token

	for {
		token, err := l.nextToken()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
		if token.Kind == TokenEOF {
			return tokens, nil
		}
	}
}

// BenchmarkTokenStorage compares retained-token memory with
// a compact projection of the same token stream.
func BenchmarkTokenStorage(b *testing.B) {
	for _, statements := range []int{300, 3000} {
		source := benchmarkSource(statements)

		b.Run(fmt.Sprintf("Full/%d", statements), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(source)))

			for b.Loop() {
				tokens, err := collectTokens(source)
				if err != nil {
					b.Fatal(err)
				}
				storedTokens = tokens
			}
		})

		b.Run(fmt.Sprintf("CompactProjection/%d", statements), func(b *testing.B) {
			tokens, err := collectTokens(source)
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				compact := make([]CompactToken, len(tokens))

				var offset uint32
				for i, token := range tokens {
					// This is a storage-only projection.
					// It does not reconstruct real source offsets.
					compact[i] = CompactToken{
						Start: offset,
						End:   offset + uint32(len(token.Lexeme)),
						Kind:  token.Kind,
					}
					offset += uint32(len(token.Lexeme))
				}

				storedCompact = compact
			}
		})
	}
}

// collectTokensWithCapacity tokenizes source while reserving
// an initial capacity for the resulting token slice.
func collectTokensWithCapacity(
	source string,
	capacity int,
) ([]Token, error) {
	l := New(source)
	tokens := make([]Token, 0, capacity)

	for {
		token, err := l.nextToken()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
		if token.Kind == TokenEOF {
			return tokens, nil
		}
	}
}

// BenchmarkTokenCapacity compares token collection with
// different initial slice capacities.
func BenchmarkTokenCapacity(b *testing.B) {
	for _, statements := range []int{300, 3000} {
		source := benchmarkSource(statements)

		// Determine the exact token count outside the timed region.
		reference, err := collectTokens(source)
		if err != nil {
			b.Fatal(err)
		}
		tokenCount := len(reference)

		strategies := []struct {
			name     string
			capacity int
		}{
			{"NoCapacity", 0},
			{"HalfCapacity", tokenCount / 2},
			{"ExactCapacity", tokenCount},
			{"DoubleCapacity", tokenCount * 2},
		}

		for _, strategy := range strategies {
			b.Run(
				fmt.Sprintf("%s/%d", strategy.name, statements),
				func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(source)))

					for b.Loop() {
						tokens, err := collectTokensWithCapacity(
							source,
							strategy.capacity,
						)
						if err != nil {
							b.Fatal(err)
						}
						storedTokens = tokens
					}
				},
			)
		}
	}
}

// BenchmarkTokenCapacityHeuristic evaluates source-length-based
// estimates against the existing append-and-grow strategy.
func BenchmarkTokenCapacityHeuristic(b *testing.B) {
	for _, statements := range []int{300, 3000} {
		source := benchmarkSource(statements)

		strategies := []struct {
			name     string
			capacity int
		}{
			{"NoCapacity", 0},
			{"BytesPerToken4", len(source) / 4},
			{"BytesPerToken6", len(source) / 6},
			{"BytesPerToken8", len(source) / 8},
			{"BytesPerToken12", len(source) / 12},
		}

		for _, strategy := range strategies {
			b.Run(
				fmt.Sprintf("%s/%d", strategy.name, statements),
				func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(source)))

					for b.Loop() {
						tokens, err := collectTokensWithCapacity(
							source,
							strategy.capacity,
						)
						if err != nil {
							b.Fatal(err)
						}
						storedTokens = tokens
					}
				},
			)
		}
	}
}

func BenchmarkTokenCapacityWorkloads(b *testing.B) {
	workloads := []struct {
		name   string
		source string
	}{
		{
			name:   "StatementMix",
			source: benchmarkSource(3000),
		},
		{
			name: "TokenDense",
			source: strings.Repeat(
				"say 1+2+3+4+5+6+7+8+9\n", 1000,
			),
		},
		{
			name: "StringHeavy",
			source: strings.Repeat(
				"say \""+strings.Repeat("a", 200)+"\"\n",
				1000,
			),
		},
		{
			name: "CommentHeavy",
			source: strings.Repeat(
				"# "+strings.Repeat("comment ", 25)+"\n"+
					"say 1\n",
				1000,
			),
		},
		{
			name: "Indented",
			source: strings.Repeat(
				"say 1\n"+
					"    say 2\n"+
					"        say 3\n"+
					"    say 4\n"+
					"say 5\n",
				1000,
			),
		},
	}

	for _, workload := range workloads {
		b.Run(workload.name, func(b *testing.B) {
			source := workload.source

			reference, err := collectTokens(source)
			if err != nil {
				b.Fatal(err)
			}

			tokenCount := len(reference)
			bytesPerToken := float64(len(source)) /
				float64(tokenCount)

			b.Logf(
				"source_bytes=%d tokens=%d bytes_per_token=%.2f",
				len(source),
				tokenCount,
				bytesPerToken,
			)

			strategies := []struct {
				name     string
				capacity int
			}{
				{"NoCapacity", 0},
				{"BytesPerToken4", len(source) / 4},
				{"BytesPerToken6", len(source) / 6},
				{"BytesPerToken8", len(source) / 8},
				{"ExactCapacity", tokenCount},
			}

			for _, strategy := range strategies {
				b.Run(strategy.name, func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(source)))

					for b.Loop() {
						tokens, err := collectTokensWithCapacity(
							source,
							strategy.capacity,
						)
						if err != nil {
							b.Fatal(err)
						}

						storedTokens = tokens
					}
				})
			}
		})
	}
}

// buildLineOffsets records the byte offset at which each source line begins.
func buildLineOffsets(source string) []int {
	offsets := []int{0}

	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			offsets = append(offsets, i+1)
		}
	}

	return offsets
}

// positionOffset converts a one-based source position to a byte offset.
// This is valid for the lexer's current byte-oriented column semantics.
func positionOffset(pos Position, lineOffsets []int) int {
	return lineOffsets[pos.Line-1] + pos.Column - 1
}

// compactTokens converts tokens into an offset-based representation.
func compactTokens(
	tokens []Token,
	lineOffsets []int,
) []CompactToken {
	result := make([]CompactToken, len(tokens))

	for i, token := range tokens {
		result[i] = CompactToken{
			Kind: token.Kind,
			Start: uint32(positionOffset(
				token.Span.Start,
				lineOffsets,
			)),
			End: uint32(positionOffset(
				token.Span.End,
				lineOffsets,
			)),
		}
	}

	return result
}

// BenchmarkCompactTokenConversion measures conversion overhead,
// including construction of the source line index.
func BenchmarkCompactTokenConversion(b *testing.B) {
	for _, statements := range []int{300, 3000} {
		source := benchmarkSource(statements)

		tokens, err := collectTokens(source)
		if err != nil {
			b.Fatal(err)
		}

		b.Run(fmt.Sprintf("%d", statements), func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(source)))

			for b.Loop() {
				lineOffsets := buildLineOffsets(source)
				compact := compactTokens(tokens, lineOffsets)
				storedCompact = compact
			}
		})
	}
}

// offsetPosition reconstructs a source position from a byte offset.
func offsetPosition(offset uint32, lineOffsets []int) Position {
	target := int(offset)

	line := 0
	low, high := 0, len(lineOffsets)

	for low < high {
		mid := low + (high-low)/2
		if lineOffsets[mid] <= target {
			line = mid
			low = mid + 1
		} else {
			high = mid
		}
	}

	return Position{
		Line:   line + 1,
		Column: target - lineOffsets[line] + 1,
	}
}

var storedPosition Position

// BenchmarkCompactPositionLookup measures source-position
// reconstruction using a line-start index.
func BenchmarkCompactPositionLookup(b *testing.B) {
	source := benchmarkSource(3000)

	tokens, err := collectTokens(source)
	if err != nil {
		b.Fatal(err)
	}

	lineOffsets := buildLineOffsets(source)
	compact := compactTokens(tokens, lineOffsets)

	b.ReportAllocs()

	b.ResetTimer()

	for b.Loop() {
		for _, token := range compact {
			storedPosition = offsetPosition(token.Start, lineOffsets)
		}
	}
}

// collectNativeCompactTokens lexes source directly into compact tokens.
func collectNativeCompactTokens(source string) ([]CompactToken, error) {
	l := NewCompact(source)

	var tokens []CompactToken

	for {
		token, err := l.NextToken()
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)

		if token.Kind == TokenEOF {
			return tokens, nil
		}
	}
}

// BenchmarkNativeCompactLexer compares complete token collection
// using the existing and experimental lexer implementations.
func BenchmarkNativeCompactLexer(b *testing.B) {
	workloads := []struct {
		name   string
		source string
	}{
		{"StatementMix", benchmarkSource(3000)},
		{"TokenDense", strings.Repeat("say 1+2+3+4+5+6+7+8+9\n", 1000)},
		{"StringHeavy", strings.Repeat("say \""+strings.Repeat("a", 200)+"\"\n", 1000)},
		{"CommentHeavy", strings.Repeat("# "+strings.Repeat("comment ", 25)+"\n"+"say 1\n", 1000)},
		{"Indented", strings.Repeat("say 1\n"+"    say 2\n"+"        say 3\n"+"    say 4\n"+"say 5\n", 1000)},
	}

	for _, workload := range workloads {
		b.Run(workload.name, func(b *testing.B) {
			b.Run("Full", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(workload.source)))

				for b.Loop() {
					tokens, err := collectTokens(workload.source)
					if err != nil {
						b.Fatal(err)
					}
					storedTokens = tokens
				}
			})

			b.Run("Compact", func(b *testing.B) {
				b.ReportAllocs()
				b.SetBytes(int64(len(workload.source)))

				for b.Loop() {
					tokens, err := collectNativeCompactTokens(workload.source)
					if err != nil {
						b.Fatal(err)
					}
					storedCompact = tokens
				}
			})
		})
	}
}
