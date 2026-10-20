package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestExpectSuccess(t *testing.T) {
	p := New("be", []lexer.Token{
		{Kind: lexer.TokenBe, Start: 0, End: 2},
		{Kind: lexer.TokenEOF, Start: 2, End: 2},
	})

	token, diagnostic := p.expect(
		lexer.TokenBe,
		"expected 'be' after variable name",
	)

	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	if token.Kind != lexer.TokenBe {
		t.Errorf("token kind = %v, want BE", token.Kind)
	}

	if bufferedCursor(p) != 1 {
		t.Errorf("current = %d, want 1", bufferedCursor(p))
	}
}

func TestExpectFailure(t *testing.T) {
	span := source.Span{
		Start: source.Position{Line: 1, Column: 12},
		End:   source.Position{Line: 1, Column: 14},
	}

	sourceText := "           42"

	p := New(sourceText, []lexer.Token{
		{Kind: lexer.TokenInt, Start: 11, End: 13},
		{Kind: lexer.TokenEOF, Start: 13, End: 13},
	})

	token, diagnostic := p.expect(
		lexer.TokenBe,
		"expected 'be' after variable name",
	)

	if diagnostic == nil {
		t.Fatal("expected a syntax diagnostic")
	}

	if token != (lexer.Token{}) {
		t.Errorf("token = %+v, want zero-value token", token)
	}

	if diagnostic.Message != "expected 'be' after variable name" {
		t.Errorf("message = %q", diagnostic.Message)
	}

	if diagnostic.Location != span {
		t.Errorf("location = %+v, want %+v", diagnostic.Location, span)
	}

	if len(diagnostic.Expected) != 1 ||
		diagnostic.Expected[0] != lexer.TokenBe {
		t.Errorf("expected = %v, want [BE]", diagnostic.Expected)
	}

	if diagnostic.Actual == nil ||
		*diagnostic.Actual != lexer.TokenInt {
		t.Errorf("actual = %v, want INT", diagnostic.Actual)
	}

	if bufferedCursor(p) != 0 {
		t.Error("failed expect() advanced the cursor")
	}
}

func TestExpectUnexpectedEOF(t *testing.T) {
	eofSpan := source.Span{
		Start: source.Position{Line: 2, Column: 1},
		End:   source.Position{Line: 2, Column: 1},
	}

	p := New("\n", []lexer.Token{
		{Kind: lexer.TokenEOF, Start: 1, End: 1},
	})

	_, diagnostic := p.expect(
		lexer.TokenRightParen,
		"expected ')' to close expression",
	)

	if diagnostic == nil {
		t.Fatal("expected a diagnostic at EOF")
	}

	if diagnostic.Location != eofSpan {
		t.Errorf("location = %+v, want %+v", diagnostic.Location, eofSpan)
	}

	if diagnostic.Actual == nil ||
		*diagnostic.Actual != lexer.TokenEOF {
		t.Errorf("actual = %v, want EOF", diagnostic.Actual)
	}

	if bufferedCursor(p) != 0 {
		t.Error("unexpected EOF advanced the cursor")
	}
}
