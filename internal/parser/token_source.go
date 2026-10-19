package parser

import "github.com/manatunga/eloque/internal/lexer"

// tokenSource supplies tokens to the parser through a shared navigation
// interface.
//
// Buffered sources read from a prepared token slice. Streaming sources
// request tokens incrementally from the lexer.
//
// Both implementations preserve the terminal EOF token so repeated
// inspection after the end of input remains safe.
type tokenSource struct {
	tokens  []lexer.Token
	scanner *lexer.Lexer
	current int
	cached  lexer.Token
	err     error
}

// newBufferedTokenSource constructs a token source backed by an existing
// token slice.
//
// The caller must provide a nonempty token sequence ending with EOF.
func newBufferedTokenSource(tokens []lexer.Token) tokenSource {
	return tokenSource{tokens: tokens}
}

// newStreamingTokenSource constructs a token source that reads directly
// from the lexer.
//
// The first token is prefetched so peek can inspect the current token
// without advancing the lexer.
func newStreamingTokenSource(source string) tokenSource {
	s := tokenSource{scanner: lexer.New(source)}
	s.refill()
	return s
}

// refill requests the next token from the streaming lexer.
//
// Lexical errors are retained and represented by TokenInvalid so that
// parser navigation can stop without discarding the original error.
//
// Once an error or EOF has been reached, no further tokens are requested.
func (s *tokenSource) refill() {
	if s.err != nil || s.cached.Kind == lexer.TokenEOF {
		return
	}

	token, err := s.scanner.NextToken()
	if err != nil {
		s.err = err
		s.cached = lexer.Token{Kind: lexer.TokenInvalid}
		return
	}

	s.cached = token
}

// peek returns the current token without consuming it.
func (s *tokenSource) peek() lexer.Token {
	if s.scanner != nil {
		return s.cached
	}

	return s.tokens[s.current]
}

// advance consumes and returns the current token.
//
// EOF remains available after consumption. A streaming lexical failure
// also remains available until the frontend returns the original error.
func (s *tokenSource) advance() lexer.Token {
	token := s.peek()

	if s.scanner != nil {
		if s.err == nil && token.Kind != lexer.TokenEOF {
			s.refill()
		}
		return token
	}

	if token.Kind != lexer.TokenEOF {
		s.current++
	}

	return token
}

// lexicalError returns the first error reported by the streaming lexer.
//
// Buffered token sources do not perform lexical analysis and therefore
// return nil.
func (s *tokenSource) lexicalError() error {
	return s.err
}
