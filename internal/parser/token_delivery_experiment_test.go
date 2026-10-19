package parser

import (
	"errors"
	"fmt"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

// These experiments compare token delivery only.
// They do not measure AST construction or parser integration.

const (
	checksumOffset uint64 = 14695981039346656037
	checksumPrime  uint64 = 1099511628211
)

// checksumToken incorporates token kind and byte offsets into an
// order-sensitive rolling checksum.
func checksumToken(hash uint64, token lexer.Token) uint64 {
	values := [...]uint64{
		uint64(token.Kind),
		uint64(token.Start),
		uint64(token.End),
	}

	for _, value := range values {
		hash ^= value
		hash *= checksumPrime
	}

	return hash
}

// consumeBuffered lexes the complete source into a token slice
// before traversing the tokens.
func consumeBuffered(source string) (uint64, int, error) {
	scanner := lexer.New(source)
	tokens := make([]lexer.Token, 0, len(source)/4)

	for {
		token, err := scanner.NextToken()
		if err != nil {
			return 0, 0, err
		}

		tokens = append(tokens, token)

		if token.Kind == lexer.TokenEOF {
			break
		}
	}

	hash := checksumOffset

	for _, token := range tokens {
		hash = checksumToken(hash, token)
	}

	return hash, len(tokens), nil
}

// consumeStreaming lexes and consumes one token at a time,
// without retaining a complete token sequence.
func consumeStreaming(source string) (uint64, int, error) {
	scanner := lexer.New(source)
	hash := checksumOffset
	count := 0

	for {
		token, err := scanner.NextToken()
		if err != nil {
			return 0, 0, err
		}

		hash = checksumToken(hash, token)
		count++

		if token.Kind == lexer.TokenEOF {
			break
		}
	}

	return hash, count, nil
}

func TestTokenDeliveryEquivalence(t *testing.T) {
	validSources := []struct {
		name   string
		source string
	}{
		{"empty", ""},
		{"declaration", "let age be 20"},
		{"multiline", "let x be 1\nsay x\n"},
		{"arithmetic", "say (2 + 3) * 4"},
		{"comments", "# comment\n\nsay 42\n"},
		{"indentation", "say 1\n    say 2\nsay 3\n"},
		{"nested indentation", "say 1\n    say 2\n        say 3\nsay 4\n"},
	}

	for _, tc := range validSources {
		t.Run(tc.name, func(t *testing.T) {
			bufferedHash, bufferedCount, bufferedErr :=
				consumeBuffered(tc.source)

			streamingHash, streamingCount, streamingErr :=
				consumeStreaming(tc.source)

			if bufferedErr != nil {
				t.Fatalf("buffered delivery failed: %v", bufferedErr)
			}

			if streamingErr != nil {
				t.Fatalf("streaming delivery failed: %v", streamingErr)
			}

			if bufferedCount != streamingCount {
				t.Errorf(
					"token count: buffered=%d streaming=%d",
					bufferedCount,
					streamingCount,
				)
			}

			if bufferedHash != streamingHash {
				t.Errorf(
					"checksum: buffered=%d streaming=%d",
					bufferedHash,
					streamingHash,
				)
			}
		})
	}
}

func TestTokenDeliveryLexicalErrors(t *testing.T) {
	invalidSources := []struct {
		name   string
		source string
	}{
		{"unterminated string", "say \"hello"},
		{"invalid integer", "let x be 123abc"},
		{"leading zero", "let x be 0123"},
		{"unmatched parenthesis", "say 1)"},
		{"unclosed parenthesis", "say (1 + 2"},
		{"invalid indentation", "say 1\n  say 2"},
	}

	for _, tc := range invalidSources {
		t.Run(tc.name, func(t *testing.T) {
			_, _, bufferedErr := consumeBuffered(tc.source)
			_, _, streamingErr := consumeStreaming(tc.source)

			var bufferedLexErr *lexer.LexError
			var streamingLexErr *lexer.LexError

			if !errors.As(bufferedErr, &bufferedLexErr) {
				t.Fatalf(
					"expected buffered lexical error, got %v",
					bufferedErr,
				)
			}

			if !errors.As(streamingErr, &streamingLexErr) {
				t.Fatalf(
					"expected streaming lexical error, got %v",
					streamingErr,
				)
			}

			if *bufferedLexErr != *streamingLexErr {
				t.Errorf(
					"lexical errors differ:\nbuffered: %+v\nstreaming: %+v",
					bufferedLexErr,
					streamingLexErr,
				)
			}
		})
	}
}

// Keep benchmark outputs observable to prevent dead-code elimination.
var tokenDeliveryHashSink uint64
var tokenDeliveryCountSink int

func BenchmarkTokenDeliveryExperiment(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	candidates := []struct {
		name    string
		consume func(string) (uint64, int, error)
	}{
		{"Buffered", consumeBuffered},
		{"Streaming", consumeStreaming},
	}

	for _, candidate := range candidates {
		for _, workload := range frontendWorkloads {
			for _, size := range sizes {
				source := workload.build(size)

				b.Run(
					fmt.Sprintf(
						"%s/%s/N%d",
						candidate.name,
						workload.name,
						size,
					),
					func(b *testing.B) {
						// Verify source and consumer before timing.
						_, _, err := candidate.consume(source)
						if err != nil {
							b.Fatalf("invalid workload: %v", err)
						}

						b.ReportAllocs()
						b.SetBytes(int64(len(source)))
						b.ResetTimer()

						for b.Loop() {
							hash, count, err := candidate.consume(source)
							if err != nil {
								b.Fatal(err)
							}

							tokenDeliveryHashSink = hash
							tokenDeliveryCountSink = count
						}
					},
				)
			}
		}
	}
}
