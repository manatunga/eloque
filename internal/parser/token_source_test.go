package parser

import (
	"errors"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

func TestStreamingTokenSource(t *testing.T) {
	t.Run("peek is stable", func(t *testing.T) {
		s := newStreamingTokenSource("say 42")

		first := s.peek()

		if first.Kind != lexer.TokenSay {
			t.Fatalf("expected SAY, got %s", first.Kind)
		}

		for range 10 {
			if s.peek() != first {
				t.Fatal("peek unexpectedly advanced")
			}
		}

		if s.advance() != first {
			t.Fatal("advance returned a different token")
		}

		if got := s.peek().Kind; got != lexer.TokenInt {
			t.Fatalf("expected INT, got %s", got)
		}
	})

	t.Run("EOF is stable", func(t *testing.T) {
		s := newStreamingTokenSource("")

		if got := s.peek().Kind; got != lexer.TokenEOF {
			t.Fatalf("expected EOF, got %s", got)
		}

		for range 10 {
			s.advance()

			if s.peek().Kind != lexer.TokenEOF {
				t.Fatal("EOF did not remain stable")
			}
		}
	})

	t.Run("lexical error is preserved", func(t *testing.T) {
		s := newStreamingTokenSource("say @")

		s.advance()

		var lexErr *lexer.LexError
		if !errors.As(s.lexicalError(), &lexErr) {
			t.Fatalf(
				"expected LexError, got %v",
				s.lexicalError(),
			)
		}

		original := s.lexicalError()

		for range 10 {
			s.advance()

			if s.lexicalError() != original {
				t.Fatal("lexical error changed")
			}
		}
	})

	t.Run("buffered tokens remain ordered", func(t *testing.T) {
		tokens := []lexer.Token{
			{Kind: lexer.TokenSay},
			{Kind: lexer.TokenInt},
			{Kind: lexer.TokenEOF},
		}

		s := newBufferedTokenSource(tokens)

		for _, expected := range tokens {
			if got := s.peek().Kind; got != expected.Kind {
				t.Fatalf(
					"expected %s, got %s",
					expected.Kind,
					got,
				)
			}

			s.advance()
		}

		if s.peek().Kind != lexer.TokenEOF {
			t.Fatal("buffered EOF did not remain stable")
		}
	})
}
