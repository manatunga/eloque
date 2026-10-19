package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

// bufferedCursor returns the current token index of a buffered parser.
//
// This helper allows legacy parser tests to inspect token consumption
// without requiring Parser to maintain a duplicate cursor field.
func bufferedCursor(p *Parser) int {
	return p.tokenSource.current
}

func TestParserPeek(t *testing.T) {
	tokens := []lexer.Token{
		{Kind: lexer.TokenInt, Start: 0, End: 2},
		{Kind: lexer.TokenEOF, Start: 2, End: 2},
	}

	p := New("42", tokens)

	if got := p.peek(); got.Kind != lexer.TokenInt {
		t.Errorf("peek() = %v, want INT", got.Kind)
	}

	if bufferedCursor(p) != 0 {
		t.Errorf("current = %d, want 0", bufferedCursor(p))
	}
}

func TestParserAdvance(t *testing.T) {
	tokens := []lexer.Token{
		{Kind: lexer.TokenInt, Start: 0, End: 2},
		{Kind: lexer.TokenNewline, Start: 2, End: 3},
		{Kind: lexer.TokenEOF, Start: 3, End: 3},
	}

	p := New("42\n", tokens)

	if got := p.advance(); got.Kind != lexer.TokenInt {
		t.Errorf("first advance() = %v, want INT", got.Kind)
	}

	if got := p.advance(); got.Kind != lexer.TokenNewline {
		t.Errorf("second advance() = %v, want NEWLINE", got.Kind)
	}

	if got := p.advance(); got.Kind != lexer.TokenEOF {
		t.Errorf("third advance() = %v, want EOF", got.Kind)
	}

	if got := p.advance(); got.Kind != lexer.TokenEOF {
		t.Errorf("repeated advance() = %v, want EOF", got.Kind)
	}

	if bufferedCursor(p) != 2 {
		t.Errorf("current = %d, want 2", bufferedCursor(p))
	}
}

func TestParserCheck(t *testing.T) {
	p := New("let", []lexer.Token{
		{Kind: lexer.TokenLet, Start: 0, End: 3},
		{Kind: lexer.TokenEOF, Start: 3, End: 3},
	})

	if !p.check(lexer.TokenLet) {
		t.Error("check(LET) = false, want true")
	}

	if p.check(lexer.TokenSay) {
		t.Error("check(SAY) = true, want false")
	}

	if bufferedCursor(p) != 0 {
		t.Error("check() unexpectedly consumed a token")
	}
}

func TestParserMatch(t *testing.T) {
	p := New("+5", []lexer.Token{
		{Kind: lexer.TokenPlus, Start: 0, End: 1},
		{Kind: lexer.TokenInt, Start: 1, End: 2},
		{Kind: lexer.TokenEOF, Start: 2, End: 2},
	})

	if p.match(lexer.TokenMinus, lexer.TokenStar) {
		t.Error("match() succeeded for an absent token kind")
	}

	if bufferedCursor(p) != 0 {
		t.Error("failed match() advanced the cursor")
	}

	if !p.match(lexer.TokenMinus, lexer.TokenPlus) {
		t.Error("match() failed to consume PLUS")
	}

	if bufferedCursor(p) != 1 {
		t.Errorf("current = %d, want 1", bufferedCursor(p))
	}

	if p.match(lexer.TokenPlus) {
		t.Error("match() consumed an unexpected token")
	}

	if got := p.peek(); got.Kind != lexer.TokenInt {
		t.Errorf("peek() = %v, want INT", got.Kind)
	}
}

func TestParserMatchNoKinds(t *testing.T) {
	p := New("42", []lexer.Token{
		{Kind: lexer.TokenInt, Start: 0, End: 2},
		{Kind: lexer.TokenEOF, Start: 2, End: 2},
	})

	if p.match() {
		t.Error("match() succeeded without any token kinds")
	}

	if bufferedCursor(p) != 0 {
		t.Errorf("current = %d, want 0", bufferedCursor(p))
	}
}

func TestParserMatchEOF(t *testing.T) {
	p := New("", []lexer.Token{
		{Kind: lexer.TokenEOF, Start: 0, End: 0},
	})

	if !p.match(lexer.TokenEOF) {
		t.Fatal("match(EOF) = false, want true")
	}

	if bufferedCursor(p) != 0 {
		t.Errorf("current = %d, want 0", bufferedCursor(p))
	}

	if got := p.peek(); got.Kind != lexer.TokenEOF {
		t.Errorf("peek() = %v, want EOF", got.Kind)
	}
}
