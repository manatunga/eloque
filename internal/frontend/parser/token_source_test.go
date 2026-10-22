package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
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

	t.Run("lexical diagnostic is preserved", func(t *testing.T) {
		s := newStreamingTokenSource("say @")

		s.advance()

		diag := s.lexicalError()
		if diag == nil {
			t.Fatal("expected lexical diagnostic")
		}
		if diag.Code != lexer.CodeUnexpectedCharacter {
			t.Errorf(
				"code = %q, want %q",
				diag.Code,
				lexer.CodeUnexpectedCharacter,
			)
		}
		if diag.Severity != diagnostic.SeverityError {
			t.Errorf(
				"severity = %v, want error",
				diag.Severity,
			)
		}
		if diag.Primary.File != 0 {
			t.Errorf("file ID = %d, want 0", diag.Primary.File)
		}
		if diag.Primary.Span != (source.ByteSpan{Start: 4, End: 5}) {
			t.Errorf(
				"span = %+v, want [4,5)",
				diag.Primary.Span,
			)
		}

		for range 10 {
			s.advance()

			if s.lexicalError() != diag {
				t.Fatal("lexical diagnostic changed")
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
