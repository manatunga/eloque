package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/lexer"
)

// tokenSource supplies tokens to the parser through a shared navigation
// interface.
//
// Buffered sources read from a prepared token slice. Streaming sources
// request tokens incrementally from the lexer.
//
// Both implementations preserve the terminal EOF token so repeated
// inspection after the end of input remains safe.
//
// In recovery mode, it also exposes lexical diagnostics as consumable
// events and supports parser-directed synchronization.
type tokenSource struct {
	tokens               []lexer.Token
	scanner              *lexer.Lexer
	current              int
	cached               lexer.Token
	err                  *diagnostic.Diagnostic
	recovering           bool
	terminal             bool
	cachedCrossedNewline bool
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
	s := tokenSource{scanner: lexer.New(source, 0)}
	s.refill()
	return s
}

// newRecoveringTokenSource constructs a streaming token source that
// supports lexical diagnostic collection and parser-directed recovery.
//
// It does not buffer the entire source or modify the normal
// fail-fast tokenization path.
func newRecoveringTokenSource(source string) tokenSource {
	s := tokenSource{
		scanner:    lexer.New(source, 0),
		recovering: true,
	}
	s.refill()
	return s
}

// refill obtains the next token or lexical diagnostic from the lexer.
//
// In recovery mode, diagnostics are retained until consumed, and
// terminal failures prevent further scanning. Normal mode preserves
// the existing fail-fast behavior.
func (s *tokenSource) refill() {
	if s.err != nil || s.cached.Kind == lexer.TokenEOF || s.terminal {
		return
	}
	if !s.recovering {
		token, err := s.scanner.NextToken()
		if err != nil {
			s.err = err
			s.cached = lexer.Token{Kind: lexer.TokenInvalid}
			return
		}
		s.cached = token
		return
	}

	result := s.scanner.NextRecoverable()
	s.cachedCrossedNewline = s.scanner.RecoveryCrossedNewline()

	switch result.Status {
	case lexer.RecoveryReady:
		s.cached = result.Token

	case lexer.RecoveryResynchronized:
		s.err = result.Diagnostic
		s.cached = lexer.Token{Kind: lexer.TokenInvalid}

	case lexer.RecoveryTerminal:
		s.err = result.Diagnostic
		s.terminal = true
		s.cached = lexer.Token{Kind: lexer.TokenInvalid}
	}
}

// takeDiagnostic returns and clears the pending lexical diagnostic.
//
// In recovery mode, consuming a recoverable diagnostic resumes token
// acquisition. Terminal diagnostics are returned once without
// restarting the lexer.
//
// In fail-fast mode, this method must not be used.
func (s *tokenSource) takeDiagnostic() *diagnostic.Diagnostic {
	if !s.recovering {
		return nil
	}

	diag := s.err
	if diag == nil {
		return nil
	}

	s.err = nil

	if !s.terminal {
		s.refill()
	}

	return diag
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
		if s.recovering && s.err != nil {
			return token
		}
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

// lexicalError returns the currently retained lexical diagnostic.
//
// In fail-fast mode, this is the first lexical error and remains
// available until parsing stops. In recovery mode, it is the pending
// diagnostic, which can be consumed through takeDiagnostic.
//
// Buffered token sources return nil.
func (s *tokenSource) lexicalError() *diagnostic.Diagnostic {
	return s.err
}

// synchronizeLine abandons the current malformed statement and
// positions the token source at the next physical newline.
//
// Pending lexical diagnostics must be consumed before synchronization.
// Cached newlines are preserved. When requested, grouping state is
// reset so physical newlines regain their structural meaning.
//
// The method returns true when a newline boundary is available,
// or false when synchronization reaches EOF or cannot proceed.
func (s *tokenSource) synchronizeLine(resetGrouping bool) bool {
	if !s.recovering || s.scanner == nil {
		return false
	}
	if s.err != nil || s.terminal {
		return false
	}
	if s.atRecoveredStatementBoundary() {
		s.scanner.ResetRecoveryGrouping()
		s.cachedCrossedNewline = false
		return true
	}
	if s.cached.Kind == lexer.TokenNewline {
		if resetGrouping {
			s.scanner.ResetRecoveryGrouping()
		}
		return true
	}
	if s.cached.Kind == lexer.TokenEOF {
		return false
	}
	foundNewline := s.scanner.RecoverToLineBoundary(resetGrouping)
	s.cached = lexer.Token{Kind: lexer.TokenInvalid}
	s.cachedCrossedNewline = false
	s.refill()

	return foundNewline
}

// atRecoveredStatementBoundary reports whether the cached token
// begins a supported statement after a suppressed physical newline.
//
// This is a recovery heuristic, not a general statement-boundary
// rule. It is used only after the parser has reported an error.
func (s *tokenSource) atRecoveredStatementBoundary() bool {
	if !s.recovering || !s.cachedCrossedNewline || s.err != nil {
		return false
	}

	switch s.cached.Kind {
	case lexer.TokenLet, lexer.TokenChange, lexer.TokenSay:
		return true
	default:
		return false
	}
}

// finishLineSynchronization consumes a cached newline after recovery.
//
// A preserved statement starter is left untouched so the parser can
// begin parsing it on the next iteration.
func (s *tokenSource) finishLineSynchronization() {
	if s.err == nil && s.cached.Kind == lexer.TokenNewline {
		s.advance()
	}
}
