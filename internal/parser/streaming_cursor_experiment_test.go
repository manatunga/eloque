package parser

import (
	"errors"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

// streamingCursorExperiment provides one-token lookahead over
// Eloque's existing pull-based lexer.
//
// The current token is cached, so repeated peek calls never
// advance the underlying lexer.
//
// A lexical error is sticky: once scanning fails, subsequent
// operations preserve the original error.
type streamingCursorExperiment struct {
	scanner *lexer.Lexer
	current lexer.Token
	err     error
}

// newStreamingCursorExperiment creates and primes the cursor.
// Priming reads the first token, including EOF for empty source.
func newStreamingCursorExperiment(
	source string,
) *streamingCursorExperiment {
	cursor := &streamingCursorExperiment{
		scanner: lexer.New(source),
	}

	cursor.refill()
	return cursor
}

// refill reads exactly one token unless an error has already
// occurred or EOF has already been reached.
func (c *streamingCursorExperiment) refill() {
	if c.err != nil || c.current.Kind == lexer.TokenEOF {
		return
	}

	token, err := c.scanner.NextToken()
	if err != nil {
		c.err = err
		c.current = lexer.Token{}
		return
	}

	c.current = token
}

// peek returns the cached token without scanning.
func (c *streamingCursorExperiment) peek() lexer.Token {
	return c.current
}

// advance returns the current token and refills the cache.
//
// EOF remains stable: advancing at EOF does not request
// additional tokens from the lexer.
func (c *streamingCursorExperiment) advance() lexer.Token {
	token := c.current

	if c.err == nil && token.Kind != lexer.TokenEOF {
		c.refill()
	}

	return token
}

// error returns the first lexical error encountered.
func (c *streamingCursorExperiment) error() error {
	return c.err
}

func TestStreamingCursorExperiment(t *testing.T) {
	t.Run("peek does not consume", func(t *testing.T) {
		cursor := newStreamingCursorExperiment("say 42")

		first := cursor.peek()

		if first.Kind != lexer.TokenSay {
			t.Fatalf("expected SAY, got %v", first.Kind)
		}

		for range 10 {
			if cursor.peek() != first {
				t.Fatal("peek changed the current token")
			}
		}

		if cursor.advance() != first {
			t.Fatal("advance did not return the cached token")
		}

		if cursor.peek().Kind != lexer.TokenInt {
			t.Fatalf(
				"expected INT after SAY, got %v",
				cursor.peek().Kind,
			)
		}
	})

	t.Run("EOF remains stable", func(t *testing.T) {
		cursor := newStreamingCursorExperiment("")

		if cursor.error() != nil {
			t.Fatal(cursor.error())
		}

		if cursor.peek().Kind != lexer.TokenEOF {
			t.Fatalf("expected EOF, got %v", cursor.peek().Kind)
		}

		first := cursor.peek()

		for range 10 {
			if cursor.advance() != first {
				t.Fatal("EOF changed after advance")
			}
		}
	})

	t.Run("lexical error is sticky", func(t *testing.T) {
		cursor := newStreamingCursorExperiment("say @")

		if cursor.peek().Kind != lexer.TokenSay {
			t.Fatal("expected SAY before invalid character")
		}

		cursor.advance()

		var lexicalError *lexer.LexError
		if !errors.As(cursor.error(), &lexicalError) {
			t.Fatalf(
				"expected LexError, got %v",
				cursor.error(),
			)
		}

		originalError := cursor.error()

		for range 10 {
			cursor.advance()

			if cursor.error() != originalError {
				t.Fatal("lexical error was replaced")
			}
		}
	})
}
