package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestFrontendDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
		code   diagnostic.Code
		span   source.ByteSpan
	}{
		{
			"unexpected character", "say @\n",
			lexer.CodeUnexpectedCharacter,
			source.ByteSpan{Start: 4, End: 5},
		},
		{
			"invalid integer", "say 12abc\n",
			lexer.CodeInvalidInteger,
			source.ByteSpan{Start: 4, End: 6},
		},
		{
			"leading zero", "say 012\n",
			lexer.CodeLeadingZero,
			source.ByteSpan{Start: 4, End: 7},
		},
		{
			"unterminated string", "say \"hello\n",
			lexer.CodeUnterminatedString,
			source.ByteSpan{Start: 4, End: 10},
		},
		{
			"unmatched closing parenthesis", "say 1)\n",
			lexer.CodeUnmatchedClosingParen,
			source.ByteSpan{Start: 5, End: 6},
		},
		{
			"missing expression", "say\n",
			CodeUnexpectedToken,
			source.ByteSpan{Start: 3, End: 4},
		},
		{
			"missing declaration initializer", "let value be\n",
			CodeUnexpectedToken,
			source.ByteSpan{Start: 12, End: 13},
		},
		{
			"syntax before lexical error", "say\nsay @\n",
			CodeUnexpectedToken,
			source.ByteSpan{Start: 3, End: 4},
		},
		{
			"lexical before syntax error", "say @\nsay\n",
			lexer.CodeUnexpectedCharacter,
			source.ByteSpan{Start: 4, End: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			program, diag := Parse(tt.source)
			if diag == nil {
				t.Fatal("expected diagnostic")
			}
			if program != nil {
				t.Errorf("program = %+v, want nil", program)
			}
			if diag.Code != tt.code {
				t.Errorf("code = %q, want %q", diag.Code, tt.code)
			}
			if diag.Severity != diagnostic.SeverityError {
				t.Errorf("severity = %v, want error", diag.Severity)
			}
			if diag.Primary.File != 0 {
				t.Errorf("file ID = %d, want 0", diag.Primary.File)
			}
			if diag.Primary.Span != tt.span {
				t.Errorf("span = %+v, want %+v",
					diag.Primary.Span, tt.span)
			}
		})
	}
}
