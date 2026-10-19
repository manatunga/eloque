package parser

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

// TestParseASTEquivalence verifies that the production streaming
// frontend produces the same AST as the buffered reference frontend.
func TestParseASTEquivalence(t *testing.T) {
	for _, workload := range frontendWorkloads {
		t.Run(workload.name, func(t *testing.T) {
			for _, size := range []int{1, 10, 100, 1000} {
				sourceText := workload.build(size)

				want, err := parseFrontend(sourceText)
				if err != nil {
					t.Fatalf("buffered frontend: %v", err)
				}

				got, err := Parse(sourceText)
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

// TestParseLexicalError verifies that lexical failures retain
// their original error type.
func TestParseLexicalError(t *testing.T) {
	_, err := Parse("say @")

	var lexErr *lexer.LexError
	if !errors.As(err, &lexErr) {
		t.Fatalf("expected *lexer.LexError, got %v", err)
	}
}

// TestParseSyntaxError verifies that malformed syntax produces
// an error rather than a successful AST.
func TestParseSyntaxError(t *testing.T) {
	program, err := Parse("let age be")

	if err == nil {
		t.Fatal("expected a syntax error")
	}
	if program != nil {
		t.Fatal("expected nil program on syntax error")
	}
	if !strings.Contains(err.Error(), "syntax error:") {
		t.Fatalf("unexpected syntax error: %v", err)
	}
}

// TestParseEmptyProgram verifies that empty input produces
// a valid program with no statements.
func TestParseEmptyProgram(t *testing.T) {
	program, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program == nil {
		t.Fatal("expected non-nil program")
	}
	if len(program.Statements) != 0 {
		t.Fatalf(
			"expected zero statements, got %d",
			len(program.Statements),
		)
	}
}

// TestParseTrailingNewline verifies that a program ending in
// a newline is accepted.
func TestParseTrailingNewline(t *testing.T) {
	program, err := Parse("let age be 20\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if program == nil || len(program.Statements) != 1 {
		t.Fatalf("expected one statement, got %#v", program)
	}
}

// TestParseFirstEncounteredError verifies that an earlier syntax
// error is not replaced by a later lexical error.
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
