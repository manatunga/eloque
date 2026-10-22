package parser

import (
	"slices"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestProgramConformanceValid(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantKinds []ast.NodeKind
		wantSpan  source.ByteSpan
	}{
		{"empty source", "", nil, source.ByteSpan{Start: 0, End: 0}},
		{"blank lines", "\n\n", nil, source.ByteSpan{Start: 2, End: 2}},
		{"integer output", "say 42",
			[]ast.NodeKind{ast.NodeSayStatement}, source.ByteSpan{Start: 0, End: 6}},
		{"variable declaration", "let age be 20",
			[]ast.NodeKind{ast.NodeLetStatement}, source.ByteSpan{Start: 0, End: 13}},
		{"variable reassignment", "change age to 21",
			[]ast.NodeKind{ast.NodeChangeStatement}, source.ByteSpan{Start: 0, End: 16}},
		{"grouped arithmetic", "say (2 + 3) * 4",
			[]ast.NodeKind{ast.NodeSayStatement}, source.ByteSpan{Start: 0, End: 15}},
		{"unary exponentiation", "say -2 ** 3",
			[]ast.NodeKind{ast.NodeSayStatement}, source.ByteSpan{Start: 0, End: 11}},
		{"multiline program", "let x be 1\nsay x",
			[]ast.NodeKind{ast.NodeLetStatement, ast.NodeSayStatement},
			source.ByteSpan{Start: 0, End: 16}},
		{"surrounding blank lines", "\nlet x be 1\n\nsay x\n",
			[]ast.NodeKind{ast.NodeLetStatement, ast.NodeSayStatement},
			source.ByteSpan{Start: 1, End: 18}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFlatParser(tt.source)
			program, diag := p.parseFlatProgram()
			if diag != nil {
				t.Fatalf("unexpected diagnostic: %+v", diag)
			}
			if program == nil {
				t.Fatal("expected program")
			}

			ids := program.Arena.StatementRange(program.Statements)
			if len(ids) != len(tt.wantKinds) {
				t.Fatalf("statement count = %d, want %d",
					len(ids), len(tt.wantKinds))
			}

			for i, id := range ids {
				node := program.Arena.Get(id)
				if node.Kind != tt.wantKinds[i] {
					t.Errorf("statement %d kind = %v, want %v",
						i, node.Kind, tt.wantKinds[i])
				}
			}

			if program.Span != tt.wantSpan {
				t.Errorf("program span = %+v, want %+v",
					program.Span, tt.wantSpan)
			}
			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("remaining token = %v, want EOF", got)
			}
		})
	}
}

func TestProgramConformanceInvalid(t *testing.T) {
	expressionStarters := []lexer.TokenKind{
		lexer.TokenInt, lexer.TokenString, lexer.TokenIdent,
		lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus,
	}

	tests := []struct {
		name     string
		source   string
		message  string
		span     source.ByteSpan
		expected []lexer.TokenKind
		actual   lexer.TokenKind
	}{
		{
			"missing declaration name", "let be 20",
			"expected a variable name after 'let'",
			source.ByteSpan{Start: 4, End: 6},
			[]lexer.TokenKind{lexer.TokenIdent}, lexer.TokenBe,
		},
		{
			"missing be keyword", "let x 20",
			"expected 'be' after variable name to initialize a variable value",
			source.ByteSpan{Start: 6, End: 8},
			[]lexer.TokenKind{lexer.TokenBe}, lexer.TokenInt,
		},
		{
			"missing to keyword", "change x 10",
			"expected 'to' after variable name to assign a new variable value",
			source.ByteSpan{Start: 9, End: 11},
			[]lexer.TokenKind{lexer.TokenTo}, lexer.TokenInt,
		},
		{
			"missing output expression", "say",
			"expected an expression",
			source.ByteSpan{Start: 3, End: 3}, expressionStarters, lexer.TokenEOF,
		},
		{
			"missing grouped operand", "say (1 + )",
			"expected an expression",
			source.ByteSpan{Start: 9, End: 10}, expressionStarters, lexer.TokenRightParen,
		},
		{
			"trailing operator", "say 1 +",
			"expected an expression",
			source.ByteSpan{Start: 7, End: 7}, expressionStarters, lexer.TokenEOF,
		},
		{
			"missing statement separator", "say 1 say 2",
			"expected a newline or end of file after statement",
			source.ByteSpan{Start: 6, End: 9},
			[]lexer.TokenKind{lexer.TokenNewline, lexer.TokenEOF},
			lexer.TokenSay,
		},
		{
			"invalid second statement", "let x be 1\nsay (2 + )",
			"expected an expression",
			source.ByteSpan{Start: 20, End: 21}, expressionStarters, lexer.TokenRightParen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFlatParser(tt.source)
			program, diag := p.parseFlatProgram()

			if program != nil {
				t.Errorf("malformed source returned a program: %+v", program)
			}
			assertConformanceDiagnostic(
				t, diag, tt.message, tt.span, tt.expected, tt.actual,
			)
		})
	}
}

func TestProgramConformanceDiagnosticEdges(t *testing.T) {
	expressionStarters := []lexer.TokenKind{
		lexer.TokenInt, lexer.TokenString, lexer.TokenIdent,
		lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus,
	}

	tests := []struct {
		name     string
		source   string
		message  string
		span     source.ByteSpan
		expected []lexer.TokenKind
		actual   lexer.TokenKind
	}{
		{
			"missing let identifier at EOF", "let",
			"expected a variable name after 'let'",
			source.ByteSpan{Start: 3, End: 3},
			[]lexer.TokenKind{lexer.TokenIdent}, lexer.TokenEOF,
		},
		{
			"missing be at EOF", "let value",
			"expected 'be' after variable name to initialize a variable value",
			source.ByteSpan{Start: 9, End: 9},
			[]lexer.TokenKind{lexer.TokenBe}, lexer.TokenEOF,
		},
		{
			"missing change identifier at EOF", "change",
			"expected a variable name after 'change'",
			source.ByteSpan{Start: 6, End: 6},
			[]lexer.TokenKind{lexer.TokenIdent}, lexer.TokenEOF,
		},
		{
			"missing to at EOF", "change value",
			"expected 'to' after variable name to assign a new variable value",
			source.ByteSpan{Start: 12, End: 12},
			[]lexer.TokenKind{lexer.TokenTo}, lexer.TokenEOF,
		},
		{
			"missing initializer on second line", "say 1\nlet value be",
			"expected an expression",
			source.ByteSpan{Start: 18, End: 18},
			expressionStarters, lexer.TokenEOF,
		},
		{
			"missing operand before newline", "say 1 +\nsay 2",
			"expected an expression",
			source.ByteSpan{Start: 7, End: 8},
			expressionStarters, lexer.TokenNewline,
		},
		{
			"nested grouping error on third line",
			"say 1\nsay 2\nsay (3 * (4 + ))",
			"expected an expression",
			source.ByteSpan{Start: 26, End: 27},
			expressionStarters, lexer.TokenRightParen,
		},
		{
			"first of two syntax errors", "say 1 +\nlet be 2",
			"expected an expression",
			source.ByteSpan{Start: 7, End: 8},
			expressionStarters, lexer.TokenNewline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFlatParser(tt.source)
			program, diag := p.parseFlatProgram()

			if program != nil {
				t.Errorf("malformed source returned a program: %+v", program)
			}
			assertConformanceDiagnostic(
				t, diag, tt.message, tt.span, tt.expected, tt.actual,
			)
		})
	}
}

func TestProgramConformanceLexicalBoundaries(t *testing.T) {
	tests := []struct {
		name    string
		source  string
		code    diagnostic.Code
		message string
		span    source.ByteSpan
	}{
		{
			"unclosed parenthesis", "say (1 + 2",
			lexer.CodeUnclosedOpeningParen,
			"unclosed opening parenthesis",
			source.ByteSpan{Start: 4, End: 5},
		},
		{
			"unmatched closing parenthesis", "say 1)",
			lexer.CodeUnmatchedClosingParen,
			"unmatched closing parenthesis",
			source.ByteSpan{Start: 5, End: 6},
		},
		{
			"invalid integer suffix", "say 123abc",
			lexer.CodeInvalidInteger,
			"invalid integer literal",
			source.ByteSpan{Start: 4, End: 7},
		},
		{
			"leading zero", "say 007",
			lexer.CodeLeadingZero,
			"integer literal cannot have leading zeros",
			source.ByteSpan{Start: 4, End: 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.New(tt.source, 0)
			var diag *diagnostic.Diagnostic

			for {
				token, nextDiag := scanner.NextToken()
				if nextDiag != nil {
					diag = nextDiag
					break
				}
				if token.Kind == lexer.TokenEOF {
					break
				}
			}

			if diag == nil {
				t.Fatal("expected lexical diagnostic")
			}
			if diag.Code != tt.code {
				t.Errorf("code = %q, want %q", diag.Code, tt.code)
			}
			if diag.Severity != diagnostic.SeverityError {
				t.Errorf("severity = %v, want error", diag.Severity)
			}
			if diag.Message != tt.message {
				t.Errorf("message = %q, want %q", diag.Message, tt.message)
			}
			if diag.Primary.File != 0 {
				t.Errorf("file = %d, want 0", diag.Primary.File)
			}
			if diag.Primary.Span != tt.span {
				t.Errorf("span = %+v, want %+v",
					diag.Primary.Span, tt.span)
			}
		})
	}
}

func assertConformanceDiagnostic(
	t *testing.T,
	diag *diagnostic.Diagnostic,
	message string,
	span source.ByteSpan,
	expected []lexer.TokenKind,
	actual lexer.TokenKind,
) {
	t.Helper()

	if diag == nil {
		t.Fatal("expected syntax diagnostic")
	}
	if diag.Code != CodeUnexpectedToken {
		t.Errorf("code = %q, want %q", diag.Code, CodeUnexpectedToken)
	}
	if diag.Severity != diagnostic.SeverityError {
		t.Errorf("severity = %v, want error", diag.Severity)
	}
	if diag.Message != message {
		t.Errorf("message = %q, want %q", diag.Message, message)
	}
	if diag.Primary.File != 0 {
		t.Errorf("file = %d, want 0", diag.Primary.File)
	}
	if diag.Primary.Span != span {
		t.Errorf("span = %+v, want %+v", diag.Primary.Span, span)
	}
	if diag.TokenExpectation == nil {
		t.Fatal("missing token expectation")
	}

	wantExpected := make([]string, len(expected))
	for i, kind := range expected {
		wantExpected[i] = kind.String()
	}

	if !slices.Equal(diag.TokenExpectation.Expected, wantExpected) {
		t.Errorf("expected = %v, want %v",
			diag.TokenExpectation.Expected, wantExpected)
	}
	if diag.TokenExpectation.Actual != actual.String() {
		t.Errorf("actual = %q, want %q",
			diag.TokenExpectation.Actual, actual.String())
	}
}
