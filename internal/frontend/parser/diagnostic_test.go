package parser

import (
	"slices"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestExpectSuccess(t *testing.T) {
	p := New("be", []lexer.Token{
		{Kind: lexer.TokenBe, Start: 0, End: 2},
		{Kind: lexer.TokenEOF, Start: 2, End: 2},
	})

	token, diag := p.expect(
		lexer.TokenBe,
		"expected 'be' after variable name",
	)

	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	if token.Kind != lexer.TokenBe {
		t.Errorf("token kind = %v, want BE", token.Kind)
	}

	if bufferedCursor(p) != 1 {
		t.Errorf("current = %d, want 1", bufferedCursor(p))
	}
}

func TestExpectFailure(t *testing.T) {
	sourceText := "           42"

	p := New(sourceText, []lexer.Token{
		{Kind: lexer.TokenInt, Start: 11, End: 13},
		{Kind: lexer.TokenEOF, Start: 13, End: 13},
	})

	token, diag := p.expect(
		lexer.TokenBe,
		"expected 'be' after variable name",
	)

	if diag == nil {
		t.Fatal("expected a syntax diagnostic")
	}

	if token != (lexer.Token{}) {
		t.Errorf("token = %+v, want zero-value token", token)
	}

	if diag.Code != CodeUnexpectedToken {
		t.Errorf(
			"code = %q, want %q",
			diag.Code,
			CodeUnexpectedToken,
		)
	}

	if diag.Severity != diagnostic.SeverityError {
		t.Errorf(
			"severity = %v, want error",
			diag.Severity,
		)
	}

	if diag.Message != "expected 'be' after variable name" {
		t.Errorf("message = %q", diag.Message)
	}

	wantSpan := source.ByteSpan{Start: 11, End: 13}

	if diag.Primary.File != 0 {
		t.Errorf("file ID = %d, want 0", diag.Primary.File)
	}

	if diag.Primary.Span != wantSpan {
		t.Errorf(
			"span = %+v, want %+v",
			diag.Primary.Span,
			wantSpan,
		)
	}

	if diag.TokenExpectation == nil {
		t.Fatal("missing token expectation")
	}

	wantExpected := []string{lexer.TokenBe.String()}

	if !slices.Equal(diag.TokenExpectation.Expected, wantExpected) {
		t.Errorf(
			"expected = %v, want %v",
			diag.TokenExpectation.Expected,
			wantExpected,
		)
	}

	if diag.TokenExpectation.Actual != lexer.TokenInt.String() {
		t.Errorf(
			"actual = %q, want %q",
			diag.TokenExpectation.Actual,
			lexer.TokenInt.String(),
		)
	}

	if bufferedCursor(p) != 0 {
		t.Error("failed expect() advanced the cursor")
	}
}

func TestExpectUnexpectedEOF(t *testing.T) {
	p := New("\n", []lexer.Token{
		{Kind: lexer.TokenEOF, Start: 1, End: 1},
	})

	_, diag := p.expect(
		lexer.TokenRightParen,
		"expected ')' to close expression",
	)

	if diag == nil {
		t.Fatal("expected a diagnostic at EOF")
	}

	if diag.Code != CodeUnexpectedToken {
		t.Errorf(
			"code = %q, want %q",
			diag.Code,
			CodeUnexpectedToken,
		)
	}

	if diag.Severity != diagnostic.SeverityError {
		t.Errorf(
			"severity = %v, want error",
			diag.Severity,
		)
	}

	wantSpan := source.ByteSpan{Start: 1, End: 1}

	if diag.Primary.File != 0 {
		t.Errorf("file ID = %d, want 0", diag.Primary.File)
	}

	if diag.Primary.Span != wantSpan {
		t.Errorf(
			"span = %+v, want %+v",
			diag.Primary.Span,
			wantSpan,
		)
	}

	if diag.TokenExpectation == nil {
		t.Fatal("missing token expectation")
	}

	wantExpected := []string{lexer.TokenRightParen.String()}

	if !slices.Equal(diag.TokenExpectation.Expected, wantExpected) {
		t.Errorf(
			"expected = %v, want %v",
			diag.TokenExpectation.Expected,
			wantExpected,
		)
	}

	if diag.TokenExpectation.Actual != lexer.TokenEOF.String() {
		t.Errorf(
			"actual = %q, want %q",
			diag.TokenExpectation.Actual,
			lexer.TokenEOF.String(),
		)
	}

	if bufferedCursor(p) != 0 {
		t.Error("unexpected EOF advanced the cursor")
	}
}
