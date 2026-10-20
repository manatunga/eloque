package parser

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
)

func TestLegacyStreamingASTEquivalence(t *testing.T) {
	for _, workload := range frontendWorkloads {
		t.Run(workload.name, func(t *testing.T) {
			for _, size := range []int{1, 10, 100, 1000} {
				sourceText := workload.build(size)

				want, err := parseFrontend(sourceText)
				if err != nil {
					t.Fatalf("buffered frontend: %v", err)
				}

				got, err := parseLegacy(sourceText)
				if err != nil {
					t.Fatalf("streaming frontend: %v", err)
				}

				if !reflect.DeepEqual(got, want) {
					t.Fatalf(
						"AST mismatch for %s with %d repetitions",
						workload.name,
						size,
					)
				}
			}
		})
	}
}

func TestParseLexicalError(t *testing.T) {
	program, err := Parse("say @")

	if program != nil {
		t.Fatal("expected nil program on lexical error")
	}

	var lexErr *lexer.LexError
	if !errors.As(err, &lexErr) {
		t.Fatalf("expected *lexer.LexError, got %v", err)
	}

	if strings.HasPrefix(err.Error(), "syntax error: ") {
		t.Fatalf("lexical error incorrectly formatted as syntax error: %v", err)
	}

	_, legacyErr := parseLegacy("say @")
	if legacyErr == nil {
		t.Fatal("legacy parser unexpectedly accepted invalid character")
	}

	if reflect.TypeOf(err) != reflect.TypeOf(legacyErr) {
		t.Errorf(
			"lexical error type differs: H2=%T, M1=%T",
			err,
			legacyErr,
		)
	}
}

func TestParseSyntaxError(t *testing.T) {
	program, err := Parse("let age be")

	if err == nil {
		t.Fatal("expected a syntax error")
	}
	if program != nil {
		t.Fatal("expected nil program on syntax error")
	}
	if !strings.HasPrefix(err.Error(), "syntax error:") {
		t.Fatalf("unexpected syntax error: %v", err)
	}
}

func TestParseEmptyProgram(t *testing.T) {
	program, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program == nil {
		t.Fatal("expected non-nil program")
	}
	if program.Arena == nil {
		t.Fatal("expected non-nil arena")
	}
	if got := len(program.Arena.StatementRange(program.Statements)); got != 0 {
		t.Fatalf("expected zero statements, got %d", got)
	}
	if program.Span.Start != 0 || program.Span.End != 0 {
		t.Errorf(
			"empty program span = %+v, want [0,0)",
			program.Span,
		)
	}
}

func TestParseTrailingNewline(t *testing.T) {
	program, err := Parse("let age be 20\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program == nil {
		t.Fatal("expected non-nil program")
	}

	if got := len(program.Arena.StatementRange(program.Statements)); got != 1 {
		t.Fatalf("expected one statement, got %d", got)
	}
}

func TestParseStatementOrder(t *testing.T) {
	const sourceText = "let age be 20\nsay age\nchange age to 21\n"

	program, err := Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if program == nil {
		t.Fatal("Parse() returned nil program")
	}
	if program.Arena == nil {
		t.Fatal("Parse() returned program with nil arena")
	}

	ids := program.Arena.StatementRange(program.Statements)

	wantKinds := []ast.NodeKind{
		ast.NodeLetStatement,
		ast.NodeSayStatement,
		ast.NodeChangeStatement,
	}

	if len(ids) != len(wantKinds) {
		t.Fatalf(
			"statement count = %d, want %d",
			len(ids),
			len(wantKinds),
		)
	}

	for i, id := range ids {
		if got := program.Arena.Get(id).Kind; got != wantKinds[i] {
			t.Errorf(
				"statement %d kind = %v, want %v",
				i,
				got,
				wantKinds[i],
			)
		}
	}
}

func TestParseFirstEncounteredError(t *testing.T) {
	_, err := Parse("let age be\nsay @")

	if err == nil {
		t.Fatal("expected an error")
	}

	var lexErr *lexer.LexError
	if errors.As(err, &lexErr) {
		t.Fatalf(
			"expected earlier syntax error, got lexical error: %v",
			err,
		)
	}

	if !strings.Contains(err.Error(), "syntax error:") {
		t.Fatalf("expected syntax error, got %v", err)
	}
}
