package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
)

func TestRecoveringTokenSourceContinuesAfterLexicalError(t *testing.T) {
	s := newRecoveringTokenSource("@say 42\n")

	if s.err == nil {
		t.Fatal("expected initial lexical diagnostic")
	}

	if s.cached.Kind != lexer.TokenInvalid {
		t.Fatalf("cached = %v, want INVALID", s.cached.Kind)
	}

	diag := s.takeDiagnostic()
	if diag == nil || diag.Code != lexer.CodeUnexpectedCharacter {
		t.Fatalf("diagnostic = %v, want unexpected-character diagnostic", diag)
	}

	for _, want := range []lexer.TokenKind{
		lexer.TokenSay,
		lexer.TokenInt,
		lexer.TokenNewline,
		lexer.TokenEOF,
	} {
		if s.err != nil {
			t.Fatalf("unexpected diagnostic before %v: %v", want, s.err)
		}

		if got := s.peek().Kind; got != want {
			t.Fatalf("token = %v, want %v", got, want)
		}

		s.advance()
	}
}

func TestRecoveringTokenSourceMultipleDiagnostics(t *testing.T) {
	s := newRecoveringTokenSource("@say 007\n)say 1\n")

	wantEvents := []struct {
		token *lexer.TokenKind
		code  string
	}{
		{code: string(lexer.CodeUnexpectedCharacter)},
		{token: tokenKindPtr(lexer.TokenSay)},
		{code: string(lexer.CodeLeadingZero)},
		{token: tokenKindPtr(lexer.TokenNewline)},
		{code: string(lexer.CodeUnmatchedClosingParen)},
		{token: tokenKindPtr(lexer.TokenSay)},
		{token: tokenKindPtr(lexer.TokenInt)},
		{token: tokenKindPtr(lexer.TokenNewline)},
		{token: tokenKindPtr(lexer.TokenEOF)},
	}

	for i, want := range wantEvents {
		if want.token == nil {
			diag := s.takeDiagnostic()
			if diag == nil {
				t.Fatalf("event %d: expected diagnostic", i)
			}
			if string(diag.Code) != want.code {
				t.Fatalf("event %d: code = %s, want %s",
					i, diag.Code, want.code)
			}
			continue
		}

		if s.err != nil {
			t.Fatalf("event %d: unexpected diagnostic %v", i, s.err)
		}

		if got := s.peek().Kind; got != *want.token {
			t.Fatalf("event %d: token = %v, want %v",
				i, got, *want.token)
		}

		s.advance()
	}
}

func TestRecoveringTokenSourceTerminalDiagnostic(t *testing.T) {
	s := newRecoveringTokenSource("\"unterminated")

	if !s.terminal {
		t.Fatal("expected terminal token source")
	}

	diag := s.takeDiagnostic()
	if diag == nil || diag.Code != lexer.CodeUnterminatedString {
		t.Fatalf("diagnostic = %v, want unterminated-string diagnostic", diag)
	}

	if second := s.takeDiagnostic(); second != nil {
		t.Fatalf("unexpected repeated diagnostic: %v", second)
	}

	if !s.terminal {
		t.Fatal("terminal state was cleared")
	}

	if s.err != nil {
		t.Fatalf("pending diagnostic was not cleared: %v", s.err)
	}
}

func TestRecoveringTokenSourceAdvanceDoesNotConsumeDiagnostic(t *testing.T) {
	s := newRecoveringTokenSource("@say 1")

	before := s.scanner

	for i := 0; i < 3; i++ {
		token := s.advance()

		if token.Kind != lexer.TokenInvalid {
			t.Fatalf("call %d: token = %v, want INVALID", i, token.Kind)
		}

		if s.err == nil {
			t.Fatalf("call %d: pending diagnostic was lost", i)
		}

		if s.scanner != before {
			t.Fatalf("call %d: scanner changed", i)
		}
	}

	diag := s.takeDiagnostic()
	if diag == nil {
		t.Fatal("expected diagnostic")
	}

	if got := s.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("token after diagnostic = %v, want SAY", got)
	}
}

func TestStreamingTokenSourceRemainsFailFast(t *testing.T) {
	s := newStreamingTokenSource("@say 1")

	if s.err == nil {
		t.Fatal("expected fail-fast lexical diagnostic")
	}

	if s.recovering {
		t.Fatal("ordinary streaming source unexpectedly enabled recovery")
	}

	if diag := s.takeDiagnostic(); diag != nil {
		t.Fatalf("fail-fast diagnostic unexpectedly consumed: %v", diag)
	}

	if s.err == nil {
		t.Fatal("fail-fast diagnostic was cleared")
	}
}

func tokenKindPtr(kind lexer.TokenKind) *lexer.TokenKind {
	return &kind
}

func TestSynchronizeLinePreservesCachedNewline(t *testing.T) {
	s := newRecoveringTokenSource("say 1\nsay 2\n")

	s.advance() // SAY
	s.advance() // INT

	if got := s.peek().Kind; got != lexer.TokenNewline {
		t.Fatalf("cached token = %v, want NEWLINE", got)
	}

	if !s.synchronizeLine(false) {
		t.Fatal("expected cached newline boundary")
	}

	if got := s.peek().Kind; got != lexer.TokenNewline {
		t.Fatalf("cached token = %v, want NEWLINE", got)
	}

	s.advance()

	if got := s.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("next token = %v, want SAY", got)
	}
}

func TestSynchronizeLineDiscardsCachedToken(t *testing.T) {
	s := newRecoveringTokenSource("say 1 + 2\nsay 3\n")

	s.advance() // SAY

	if got := s.peek().Kind; got != lexer.TokenInt {
		t.Fatalf("cached token = %v, want INT", got)
	}

	if !s.synchronizeLine(false) {
		t.Fatal("expected newline boundary")
	}

	if got := s.peek().Kind; got != lexer.TokenNewline {
		t.Fatalf("cached token = %v, want NEWLINE", got)
	}

	s.advance()

	if got := s.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("next token = %v, want SAY", got)
	}
}

func TestSynchronizeLineRejectsPendingDiagnostic(t *testing.T) {
	s := newRecoveringTokenSource("@say 1\n")

	if s.synchronizeLine(false) {
		t.Fatal("synchronization unexpectedly succeeded")
	}

	if s.err == nil {
		t.Fatal("pending lexical diagnostic was lost")
	}
}

func TestSynchronizeLineRejectsTerminalState(t *testing.T) {
	s := newRecoveringTokenSource("\"unterminated")

	if s.synchronizeLine(true) {
		t.Fatal("terminal source unexpectedly synchronized")
	}

	if !s.terminal {
		t.Fatal("terminal state was cleared")
	}
}

func TestSynchronizeLineAtEOF(t *testing.T) {
	s := newRecoveringTokenSource("say 1")

	s.advance() // SAY

	if s.synchronizeLine(false) {
		t.Fatal("expected EOF, not newline")
	}

	if got := s.peek().Kind; got != lexer.TokenEOF {
		t.Fatalf("token = %v, want EOF", got)
	}
}

func TestRecoveringTokenSourceTracksSuppressedNewline(t *testing.T) {
	s := newRecoveringTokenSource("say (1 +\n2)\n")

	for {
		token := s.peek()

		if s.err != nil {
			t.Fatalf("unexpected lexical diagnostic: %v", s.err)
		}

		if token.Kind == lexer.TokenInt &&
			s.cachedCrossedNewline {
			if token.Start != uint32(len("say (1 +\n")) {
				t.Fatalf("unexpected token offset: %d", token.Start)
			}

			s.advance()

			if s.cachedCrossedNewline {
				t.Fatal("newline metadata was not reset for next token")
			}

			return
		}

		if token.Kind == lexer.TokenEOF {
			t.Fatal("expected token after suppressed newline")
		}

		s.advance()
	}
}

func TestRecoveringTokenSourceTracksMalformedGroupingBoundary(t *testing.T) {
	s := newRecoveringTokenSource("say (1 +\nsay 42\n")

	for {
		if s.err != nil {
			t.Fatalf("unexpected lexical diagnostic: %v", s.err)
		}

		token := s.peek()

		if token.Kind == lexer.TokenSay &&
			token.Start == uint32(len("say (1 +\n")) {

			if !s.cachedCrossedNewline {
				t.Fatal("expected suppressed-newline boundary")
			}

			if s.scanner == nil {
				t.Fatal("expected recovery-aware lexer")
			}

			return
		}

		if token.Kind == lexer.TokenEOF {
			t.Fatal("expected second SAY token")
		}

		s.advance()
	}
}

func TestSynchronizeLinePreservesStatementAfterSuppressedNewline(t *testing.T) {
	s := newRecoveringTokenSource("say (1 +\nsay 42\n")

	for s.peek().Start < uint32(len("say (1 +\n")) {
		if s.err != nil {
			t.Fatalf("unexpected diagnostic: %v", s.err)
		}
		s.advance()
	}

	if !s.atRecoveredStatementBoundary() {
		t.Fatal("expected recoverable statement boundary")
	}

	if !s.synchronizeLine(true) {
		t.Fatal("expected synchronization to succeed")
	}

	if got := s.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("cached token = %v, want SAY", got)
	}

	if s.cachedCrossedNewline {
		t.Fatal("stale newline metadata after synchronization")
	}

	s.advance()

	if got := s.peek().Kind; got != lexer.TokenInt {
		t.Fatalf("next token = %v, want INT", got)
	}
}

func TestSynchronizeLineDoesNotPreserveSameLineStatementStarter(t *testing.T) {
	s := newRecoveringTokenSource("say 1 say 2\nsay 3\n")

	s.advance() // First SAY
	s.advance() // INT

	if s.atRecoveredStatementBoundary() {
		t.Fatal("same-line SAY must not be treated as a new boundary")
	}

	if !s.synchronizeLine(false) {
		t.Fatal("expected synchronization to next newline")
	}

	if got := s.peek().Kind; got != lexer.TokenNewline {
		t.Fatalf("cached token = %v, want NEWLINE", got)
	}

	s.advance()

	if got := s.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("next statement = %v, want SAY", got)
	}
}
