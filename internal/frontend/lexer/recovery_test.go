package lexer

import (
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
)

func TestNextRecoverableValidTokens(t *testing.T) {
	l := New("say 42\n", 0)

	expected := []TokenKind{
		TokenSay,
		TokenInt,
		TokenNewline,
		TokenEOF,
	}

	for i, kind := range expected {
		result := l.NextRecoverable()

		if result.Status != RecoveryReady {
			t.Fatalf("token %d: status = %v, want RecoveryReady",
				i, result.Status)
		}

		if result.Diagnostic != nil {
			t.Fatalf("token %d: unexpected diagnostic: %+v",
				i, result.Diagnostic)
		}

		if result.Token.Kind != kind {
			t.Fatalf("token %d: kind = %v, want %v",
				i, result.Token.Kind, kind)
		}
	}
}

// TestNextRecoverableSimpleErrors verifies that lexical errors with
// already-consumed offending input allow subsequent scanning.
func TestNextRecoverableSimpleErrors(t *testing.T) {
	tests := []struct {
		name string
		src  string
		code diagnostic.Code
	}{
		{
			name: "unexpected character",
			src:  "@say 42\n",
			code: CodeUnexpectedCharacter,
		},
		{
			name: "unmatched closing parenthesis",
			src:  ")say 42\n",
			code: CodeUnmatchedClosingParen,
		},
		{
			name: "leading zero",
			src:  "007 say 42\n",
			code: CodeLeadingZero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.src, 0)

			first := l.NextRecoverable()

			if first.Status != RecoveryResynchronized {
				t.Fatalf("status = %v, want RecoveryResynchronized",
					first.Status)
			}

			if first.Diagnostic == nil {
				t.Fatal("expected lexical diagnostic")
			}

			if first.Diagnostic.Code != tt.code {
				t.Fatalf("code = %s, want %s",
					first.Diagnostic.Code, tt.code)
			}

			next := l.NextRecoverable()

			if next.Status != RecoveryReady {
				t.Fatalf("next status = %v, want RecoveryReady",
					next.Status)
			}

			if next.Diagnostic != nil {
				t.Fatalf("unexpected follow-up diagnostic: %+v",
					next.Diagnostic)
			}

			if next.Token.Kind != TokenSay {
				t.Fatalf("next token = %v, want TokenSay",
					next.Token.Kind)
			}
		})
	}
}

func TestNextRecoverableTerminalFallback(t *testing.T) {
	l := New("say \"unterminated", 0)

	if result := l.NextRecoverable(); result.Token.Kind != TokenSay {
		t.Fatalf("first token = %v, want TokenSay", result.Token.Kind)
	}

	result := l.NextRecoverable()

	if result.Status != RecoveryTerminal {
		t.Fatalf("status = %v, want RecoveryTerminal", result.Status)
	}

	if result.Diagnostic == nil {
		t.Fatal("expected terminal diagnostic")
	}

	if result.Diagnostic.Code != CodeUnterminatedString {
		t.Fatalf("code = %s, want %s",
			result.Diagnostic.Code, CodeUnterminatedString)
	}
}

func TestNextRecoverableInvalidInteger(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		wantEnd uint32
	}{
		{
			name:    "alphabetic suffix",
			source:  "123abc456 say 1",
			wantEnd: 9,
		},
		{
			name:    "underscore suffix",
			source:  "123abc_def say 1",
			wantEnd: 10,
		},
		{
			name:    "suffix at EOF",
			source:  "123abc",
			wantEnd: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source, 0)

			result := l.NextRecoverable()

			if result.Status != RecoveryResynchronized {
				t.Fatalf("status = %v, want RecoveryResynchronized",
					result.Status)
			}

			if result.Diagnostic == nil {
				t.Fatal("expected diagnostic")
			}

			if result.Diagnostic.Code != CodeInvalidInteger {
				t.Fatalf("code = %s, want %s",
					result.Diagnostic.Code, CodeInvalidInteger)
			}

			if got := result.Diagnostic.Primary.Span.End; got != tt.wantEnd {
				t.Errorf("span end = %d, want %d", got, tt.wantEnd)
			}

			next := l.NextRecoverable()

			wantKind := TokenSay
			if tt.name == "suffix at EOF" {
				wantKind = TokenEOF
			}

			if next.Status != RecoveryReady || next.Token.Kind != wantKind {
				t.Errorf("next = (%v, %v), want (RecoveryReady, %v)",
					next.Status, next.Token.Kind, wantKind)
			}
		})
	}
}

func TestNextRecoverableInvalidEscape(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantNext TokenKind
	}{
		{
			name:     "simple invalid escape",
			source:   `"hello\q" say 1`,
			wantNext: TokenSay,
		},
		{
			name:     "escaped quote after invalid escape",
			source:   `"hello\q\"world" say 1`,
			wantNext: TokenSay,
		},
		{
			name:     "unterminated at newline",
			source:   "\"hello\\q\nsay 1",
			wantNext: TokenNewline,
		},
		{
			name:     "unterminated at EOF",
			source:   `"hello\q`,
			wantNext: TokenEOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source, 0)

			result := l.NextRecoverable()

			if result.Status != RecoveryResynchronized {
				t.Fatalf("status = %v, want RecoveryResynchronized",
					result.Status)
			}

			if result.Diagnostic == nil {
				t.Fatal("expected diagnostic")
			}

			if result.Diagnostic.Code != CodeInvalidEscape {
				t.Fatalf("code = %s, want %s",
					result.Diagnostic.Code, CodeInvalidEscape)
			}

			next := l.NextRecoverable()

			if next.Status != RecoveryReady {
				t.Fatalf("next status = %v, want RecoveryReady",
					next.Status)
			}

			if next.Token.Kind != tt.wantNext {
				t.Errorf("next token = %v, want %v",
					next.Token.Kind, tt.wantNext)
			}
		})
	}
}

func TestNextRecoverableTerminalStatePersists(t *testing.T) {
	l := New("say \"unterminated", 0)

	first := l.NextRecoverable()
	if first.Status != RecoveryReady || first.Token.Kind != TokenSay {
		t.Fatalf("first result = (%v, %v), want (RecoveryReady, SAY)",
			first.Status, first.Token.Kind)
	}

	terminal := l.NextRecoverable()
	if terminal.Status != RecoveryTerminal {
		t.Fatalf("terminal status = %v, want RecoveryTerminal", terminal.Status)
	}
	if terminal.Diagnostic == nil {
		t.Fatal("expected diagnostic on first terminal failure")
	}
	if terminal.Diagnostic.Code != CodeUnterminatedString {
		t.Fatalf("diagnostic code = %s, want %s",
			terminal.Diagnostic.Code, CodeUnterminatedString)
	}

	position := l.current

	for i := 0; i < 3; i++ {
		result := l.NextRecoverable()

		if result.Status != RecoveryTerminal {
			t.Fatalf("call %d: status = %v, want RecoveryTerminal",
				i+1, result.Status)
		}
		if result.Diagnostic != nil {
			t.Fatalf("call %d: unexpected repeated diagnostic: %v",
				i+1, result.Diagnostic)
		}
		if l.current != position {
			t.Fatalf("call %d: current = %d, want %d",
				i+1, l.current, position)
		}
	}
}

func TestNextRecoverableTerminalStateIsPerLexer(t *testing.T) {
	broken := New("\"unterminated", 0)
	valid := New("say 42", 0)

	if result := broken.NextRecoverable(); result.Status != RecoveryTerminal {
		t.Fatalf("broken lexer status = %v, want RecoveryTerminal", result.Status)
	}

	result := valid.NextRecoverable()
	if result.Status != RecoveryReady || result.Token.Kind != TokenSay {
		t.Fatalf("valid lexer result = (%v, %v), want (RecoveryReady, SAY)",
			result.Status, result.Token.Kind)
	}
}

func TestNextRecoverableUnterminatedStringAtNewline(t *testing.T) {
	l := New("say \"hello\nsay 42\n", 0)

	first := l.NextRecoverable()
	if first.Status != RecoveryReady || first.Token.Kind != TokenSay {
		t.Fatalf("first result = (%v, %v), want (RecoveryReady, SAY)",
			first.Status, first.Token.Kind)
	}

	result := l.NextRecoverable()

	if result.Status != RecoveryResynchronized {
		t.Fatalf("status = %v, want RecoveryResynchronized", result.Status)
	}

	if result.Diagnostic == nil ||
		result.Diagnostic.Code != CodeUnterminatedString {
		t.Fatalf("diagnostic = %v, want CodeUnterminatedString",
			result.Diagnostic)
	}

	for _, want := range []TokenKind{
		TokenNewline,
		TokenSay,
		TokenInt,
		TokenNewline,
		TokenEOF,
	} {
		next := l.NextRecoverable()

		if next.Status != RecoveryReady {
			t.Fatalf("status = %v, want RecoveryReady", next.Status)
		}

		if next.Token.Kind != want {
			t.Fatalf("token = %v, want %v", next.Token.Kind, want)
		}
	}
}

func TestNextRecoverableUnterminatedStringAtEOF(t *testing.T) {
	l := New("\"hello", 0)

	first := l.NextRecoverable()

	if first.Status != RecoveryTerminal {
		t.Fatalf("status = %v, want RecoveryTerminal", first.Status)
	}

	if first.Diagnostic == nil ||
		first.Diagnostic.Code != CodeUnterminatedString {
		t.Fatalf("diagnostic = %v, want CodeUnterminatedString",
			first.Diagnostic)
	}

	position := l.current

	second := l.NextRecoverable()

	if second.Status != RecoveryTerminal {
		t.Fatalf("second status = %v, want RecoveryTerminal", second.Status)
	}

	if second.Diagnostic != nil {
		t.Fatalf("unexpected repeated diagnostic: %v", second.Diagnostic)
	}

	if l.current != position {
		t.Fatalf("lexer advanced after terminal failure")
	}
}

func TestRecoverToLineBoundaryPreservesNewline(t *testing.T) {
	l := New("invalid contents\nsay 42\n", 0)

	found := l.RecoverToLineBoundary(false)

	if !found {
		t.Fatal("expected to find a physical newline")
	}

	if l.peek() != '\n' {
		t.Fatalf("peek = %q, want newline", l.peek())
	}

	next := l.NextRecoverable()
	if next.Status != RecoveryReady || next.Token.Kind != TokenNewline {
		t.Fatalf("next = (%v, %v), want (RecoveryReady, NEWLINE)",
			next.Status, next.Token.Kind)
	}

	next = l.NextRecoverable()
	if next.Status != RecoveryReady || next.Token.Kind != TokenSay {
		t.Fatalf("next = (%v, %v), want (RecoveryReady, SAY)",
			next.Status, next.Token.Kind)
	}
}

func TestRecoverToLineBoundaryResetsGrouping(t *testing.T) {
	l := New("(invalid contents\nsay 42\n", 0)

	first := l.NextRecoverable()
	if first.Status != RecoveryReady || first.Token.Kind != TokenLeftParen {
		t.Fatalf("first = (%v, %v), want (RecoveryReady, LEFT_PAREN)",
			first.Status, first.Token.Kind)
	}

	if l.parenDepth != 1 {
		t.Fatalf("parenDepth = %d, want 1", l.parenDepth)
	}

	found := l.RecoverToLineBoundary(true)
	if !found {
		t.Fatal("expected physical newline")
	}

	if l.parenDepth != 0 || len(l.parenPositions) != 0 {
		t.Fatalf("grouping not reset: depth=%d positions=%d",
			l.parenDepth, len(l.parenPositions))
	}

	next := l.NextRecoverable()
	if next.Status != RecoveryReady || next.Token.Kind != TokenNewline {
		t.Fatalf("next = (%v, %v), want (RecoveryReady, NEWLINE)",
			next.Status, next.Token.Kind)
	}

	next = l.NextRecoverable()
	if next.Status != RecoveryReady || next.Token.Kind != TokenSay {
		t.Fatalf("next = (%v, %v), want (RecoveryReady, SAY)",
			next.Status, next.Token.Kind)
	}
}

func TestRecoverToLineBoundaryAtEOF(t *testing.T) {
	l := New("invalid contents", 0)

	found := l.RecoverToLineBoundary(false)
	if found {
		t.Fatal("unexpected newline at EOF")
	}

	next := l.NextRecoverable()
	if next.Status != RecoveryReady || next.Token.Kind != TokenEOF {
		t.Fatalf("next = (%v, %v), want (RecoveryReady, EOF)",
			next.Status, next.Token.Kind)
	}
}

func TestNextRecoverableInvalidIndentationContinues(t *testing.T) {
	l := New("say 1\n  say 2\nsay 3\n", 0)

	for _, want := range []TokenKind{
		TokenSay,
		TokenInt,
		TokenNewline,
	} {
		result := l.NextRecoverable()
		if result.Status != RecoveryReady || result.Token.Kind != want {
			t.Fatalf("got (%v, %v), want (RecoveryReady, %v)",
				result.Status, result.Token.Kind, want)
		}
	}

	result := l.NextRecoverable()

	if result.Status != RecoveryResynchronized {
		t.Fatalf("status = %v, want RecoveryResynchronized", result.Status)
	}
	if result.Diagnostic == nil ||
		result.Diagnostic.Code != CodeInvalidIndentation {
		t.Fatalf("diagnostic = %v, want CodeInvalidIndentation",
			result.Diagnostic)
	}

	for _, want := range []TokenKind{
		TokenNewline,
		TokenSay,
		TokenInt,
		TokenNewline,
		TokenEOF,
	} {
		next := l.NextRecoverable()

		if next.Status != RecoveryReady || next.Token.Kind != want {
			t.Fatalf("got (%v, %v), want (RecoveryReady, %v)",
				next.Status, next.Token.Kind, want)
		}
	}
}

func TestRecoverableScanTracksSuppressedNewline(t *testing.T) {
	l := New("say (1 +\n2)\n", 0)

	for {
		result := l.NextRecoverable()

		if result.Status != RecoveryReady {
			t.Fatalf("unexpected recovery status: %v", result.Status)
		}

		if result.Token.Kind == TokenInt &&
			l.RecoveryCrossedNewline() {
			if result.Token.Start != uint32(len("say (1 +\n")) {
				t.Fatalf("unexpected token position: %d", result.Token.Start)
			}
			return
		}

		if result.Token.Kind == TokenEOF {
			t.Fatal("expected suppressed newline metadata")
		}
	}
}

func TestRecoverableScanResetsNewlineMetadata(t *testing.T) {
	l := New("say (1 +\n2)\n", 0)

	found := false

	for {
		result := l.NextRecoverable()

		if result.Status != RecoveryReady {
			t.Fatalf("unexpected recovery status: %v", result.Status)
		}

		if result.Token.Kind == TokenInt &&
			l.RecoveryCrossedNewline() {
			found = true
			continue
		}

		if found {
			if l.RecoveryCrossedNewline() {
				t.Fatal("suppressed newline flag was not reset")
			}
			return
		}

		if result.Token.Kind == TokenEOF {
			t.Fatal("did not encounter expected multiline token")
		}
	}
}
