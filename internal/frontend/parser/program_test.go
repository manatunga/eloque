package parser

import (
	"slices"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestFlatProgramEmpty(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty source", ""},
		{"single blank line", "\n"},
		{"multiple blank lines", "\n\n\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFlatStatementTestParser(tt.source)

			program, diagnostic := p.parseFlatProgram()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			if program == nil {
				t.Fatal("expected program, got nil")
			}

			if program.Arena != p.arena {
				t.Error("program does not reference parser arena")
			}

			wantRange := (ast.StatementRange{
				Start: 0,
				End:   0,
			})
			if program.Statements != wantRange {
				t.Errorf("statement range = %+v, want %+v",
					program.Statements, wantRange)
			}

			if got := len(p.arena.StatementRange(program.Statements)); got != 0 {
				t.Errorf("statement count = %d, want 0", got)
			}

			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("next token = %v, want EOF", got)
			}
		})
	}
}

func TestFlatProgramStatements(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantKinds []ast.NodeKind
	}{
		{
			name:      "single statement at EOF",
			source:    "say 42",
			wantKinds: []ast.NodeKind{ast.NodeSayStatement},
		},
		{
			name:      "single statement with trailing newline",
			source:    "say 42\n",
			wantKinds: []ast.NodeKind{ast.NodeSayStatement},
		},
		{
			name:   "mixed statements and blank lines",
			source: "\nlet age be 20\n\nchange age to 21\nsay age\n\n",
			wantKinds: []ast.NodeKind{
				ast.NodeLetStatement,
				ast.NodeChangeStatement,
				ast.NodeSayStatement,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFlatStatementTestParser(tt.source)

			program, diagnostic := p.parseFlatProgram()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			if program == nil {
				t.Fatal("expected program, got nil")
			}

			ids := p.arena.StatementRange(program.Statements)

			if len(ids) != len(tt.wantKinds) {
				t.Fatalf("statement count = %d, want %d",
					len(ids), len(tt.wantKinds))
			}

			for i, id := range ids {
				node := p.arena.Get(id)

				if node.Kind != tt.wantKinds[i] {
					t.Errorf(
						"statement %d kind = %v, want %v",
						i, node.Kind, tt.wantKinds[i],
					)
				}
			}

			if got := p.arena.StatementCount(); got != uint32(len(tt.wantKinds)) {
				t.Errorf("arena statement count = %d, want %d",
					got, len(tt.wantKinds))
			}

			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("next token = %v, want EOF", got)
			}
		})
	}
}

func TestFlatProgramStatementSpans(t *testing.T) {
	sourceText := "\nsay (2 + 3)\n\nsay 4\n"
	p := newFlatStatementTestParser(sourceText)

	program, diagnostic := p.parseFlatProgram()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	ids := p.arena.StatementRange(program.Statements)
	if len(ids) != 2 {
		t.Fatalf("statement count = %d, want 2", len(ids))
	}

	first := p.arena.Get(ids[0])
	second := p.arena.Get(ids[1])

	wantFirst := source.ByteSpan{
		Start: 1,
		End:   12,
	}
	wantSecond := source.ByteSpan{
		Start: 14,
		End:   19,
	}

	if first.Span != wantFirst {
		t.Errorf("first span = %+v, want %+v",
			first.Span, wantFirst)
	}

	if second.Span != wantSecond {
		t.Errorf("second span = %+v, want %+v",
			second.Span, wantSecond)
	}
}

func TestFlatProgramInvalidSeparator(t *testing.T) {
	p := newFlatStatementTestParser("let age be 20 say age")

	program, diag := p.parseFlatProgram()
	if diag == nil {
		t.Fatal("expected diagnostic, got nil")
	}
	if program != nil {
		t.Errorf("program = %+v, want nil", program)
	}

	if diag.Code != CodeUnexpectedToken {
		t.Errorf("code = %q, want %q", diag.Code, CodeUnexpectedToken)
	}

	if diag.Severity != diagnostic.SeverityError {
		t.Errorf("severity = %v, want error", diag.Severity)
	}

	if diag.Message != "expected a newline or end of file after statement" {
		t.Errorf("message = %q", diag.Message)
	}

	if diag.Primary.Span != (source.ByteSpan{Start: 14, End: 17}) {
		t.Errorf("span = %+v, want [14,17)", diag.Primary.Span)
	}

	if diag.TokenExpectation == nil {
		t.Fatal("missing token expectation")
	}

	wantExpected := []string{
		lexer.TokenNewline.String(),
		lexer.TokenEOF.String(),
	}
	if !slices.Equal(diag.TokenExpectation.Expected, wantExpected) {
		t.Errorf("expected = %v, want %v",
			diag.TokenExpectation.Expected, wantExpected)
	}

	if diag.TokenExpectation.Actual != lexer.TokenSay.String() {
		t.Errorf("actual = %q, want SAY",
			diag.TokenExpectation.Actual)
	}

	if got := p.arena.StatementCount(); got != 0 {
		t.Errorf("stored statements = %d, want 0", got)
	}

	if got := p.peek().Kind; got != lexer.TokenSay {
		t.Errorf("remaining token = %v, want SAY", got)
	}
}

func TestFlatProgramDiagnosticPropagation(t *testing.T) {
	p := newFlatStatementTestParser(
		"let age be 20\nsay age\nchange age to",
	)

	program, diag := p.parseFlatProgram()
	if diag == nil {
		t.Fatal("expected diagnostic, got nil")
	}
	if program != nil {
		t.Errorf("program = %+v, want nil", program)
	}

	if diag.Code != CodeUnexpectedToken {
		t.Errorf("code = %q, want %q", diag.Code, CodeUnexpectedToken)
	}

	if diag.Severity != diagnostic.SeverityError {
		t.Errorf("severity = %v, want error", diag.Severity)
	}

	if diag.Message != "expected an expression" {
		t.Errorf("message = %q, want expected expression", diag.Message)
	}

	if diag.TokenExpectation == nil {
		t.Fatal("missing token expectation")
	}

	if diag.TokenExpectation.Actual != lexer.TokenEOF.String() {
		t.Errorf("actual = %q, want EOF",
			diag.TokenExpectation.Actual)
	}

	if got := p.arena.StatementCount(); got != 2 {
		t.Errorf("stored statements = %d, want 2", got)
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatProgramRangeOffset(t *testing.T) {
	p := newFlatStatementTestParser("say 42")

	existingValue := p.arena.AddInteger(source.ByteSpan{
		Start: 4,
		End:   6,
	})
	existingStatement := p.arena.AddSay(
		source.ByteSpan{Start: 0, End: 6},
		existingValue,
	)
	p.arena.AddStatement(existingStatement)

	program, diagnostic := p.parseFlatProgram()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	wantRange := ast.StatementRange{
		Start: 1,
		End:   2,
	}
	if program.Statements != wantRange {
		t.Errorf("range = %+v, want %+v",
			program.Statements, wantRange)
	}

	ids := p.arena.StatementRange(program.Statements)
	if len(ids) != 1 {
		t.Fatalf("range length = %d, want 1", len(ids))
	}

	if ids[0] == existingStatement {
		t.Error("program range includes the preexisting statement")
	}
}

func TestFlatProgramRootSpan(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   source.ByteSpan
	}{
		{
			name:   "empty source",
			source: "",
			want:   source.ByteSpan{Start: 0, End: 0},
		},
		{
			name:   "blank lines only",
			source: "\n\n",
			want:   source.ByteSpan{Start: 2, End: 2},
		},
		{
			name:   "single statement",
			source: "say 42",
			want:   source.ByteSpan{Start: 0, End: 6},
		},
		{
			name:   "trailing newline",
			source: "say 42\n",
			want:   source.ByteSpan{Start: 0, End: 6},
		},
		{
			name:   "leading and trailing blank lines",
			source: "\n\nsay 42\n\n",
			want:   source.ByteSpan{Start: 2, End: 8},
		},
		{
			name:   "multiple statements",
			source: "\nlet x be 1\n\nsay x\n",
			want:   source.ByteSpan{Start: 1, End: 18},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFlatStatementTestParser(tt.source)

			program, diagnostic := p.parseFlatProgram()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			if program.Span != tt.want {
				t.Errorf(
					"program span = %+v, want %+v",
					program.Span,
					tt.want,
				)
			}
		})
	}
}
