package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
)

func TestRecoveringFlatParserConstruction(t *testing.T) {
	source := "say 42\n"

	p := newRecoveringFlatParser(source)

	if p == nil || p.Parser == nil {
		t.Fatal("expected initialized parser")
	}

	if p.arena == nil {
		t.Fatal("expected initialized AST arena")
	}

	if p.source != source {
		t.Fatalf("source = %q, want %q", p.source, source)
	}

	if !p.tokenSource.recovering {
		t.Fatal("expected recovery-aware token source")
	}

	if p.tokenSource.scanner == nil {
		t.Fatal("expected streaming lexer")
	}

	if got := p.tokenSource.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("first token = %v, want SAY", got)
	}
}

func TestRecoveringFlatParserInitialDiagnostic(t *testing.T) {
	p := newRecoveringFlatParser("@say 42\n")

	if !p.tokenSource.recovering {
		t.Fatal("expected recovery mode")
	}

	diag := p.tokenSource.takeDiagnostic()

	if diag == nil {
		t.Fatal("expected lexical diagnostic")
	}

	if diag.Code != lexer.CodeUnexpectedCharacter {
		t.Fatalf(
			"diagnostic code = %s, want %s",
			diag.Code,
			lexer.CodeUnexpectedCharacter,
		)
	}

	if got := p.tokenSource.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("token after diagnostic = %v, want SAY", got)
	}
}

func TestFlatParserStillUsesFailFastTokenSource(t *testing.T) {
	p := NewFlatParser("say 42\n")

	if p.tokenSource.recovering {
		t.Fatal("ordinary parser unexpectedly enabled recovery")
	}

	if p.tokenSource.scanner == nil {
		t.Fatal("expected streaming lexer")
	}

	if got := p.tokenSource.peek().Kind; got != lexer.TokenSay {
		t.Fatalf("first token = %v, want SAY", got)
	}
}
