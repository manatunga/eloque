package parser

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
)

// frontendProgramSink prevents benchmark results from being eliminated.
var frontendProgramSink *ast.Program

// collectFullTokens tokenizes source with the production lexer.
func collectFullTokens(source string) ([]lexer.Token, error) {
	scanner := lexer.New(source)
	tokens := make([]lexer.Token, 0, 64)

	for {
		token, err := scanner.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)

		if token.Kind == lexer.TokenEOF {
			return tokens, nil
		}
	}
}

// collectCompactTokens tokenizes source with the compact lexer.
func collectCompactTokens(source string) ([]lexer.CompactToken, error) {
	scanner := lexer.NewCompact(source)
	tokens := make([]lexer.CompactToken, 0, 64)

	for {
		token, err := scanner.NextToken()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)

		if token.Kind == lexer.TokenEOF {
			return tokens, nil
		}
	}
}

// parseFullFrontend runs the production lexer and parser.
func parseFullFrontend(source string) (*ast.Program, *Diagnostic, error) {
	tokens, err := collectFullTokens(source)
	if err != nil {
		return nil, nil, err
	}
	program, diagnostic := New(tokens).parseProgram()
	return program, diagnostic, nil
}

// parseCompactFrontend runs the compact lexer and parser.
func parseCompactFrontend(source string) (*ast.Program, *Diagnostic, error) {
	tokens, err := collectCompactTokens(source)
	if err != nil {
		return nil, nil, err
	}
	program, diagnostic := NewCompact(source, tokens).parseCompactProgram()
	return program, diagnostic, nil
}

// TestCompactFrontendEquivalence checks complete AST and diagnostic
// equivalence between the production and compact frontends.
func TestCompactFrontendEquivalence(t *testing.T) {
	cases := []struct {
		name   string
		source string
	}{
		{"empty", ""},
		{"blank_lines", "\n\n"},
		{"integer", "say 42"},
		{"string", "say \"hello\\nworld\""},
		{"identifier", "say result"},
		{"declaration", "let age be 20"},
		{"reassignment", "change age to 21"},
		{"multiple_statements", "let age be 20\nchange age to 21\nsay age"},
		{"unary", "say -2"},
		{"precedence", "say 2 + 3 * 4"},
		{"exponentiation", "say 2 ** 3 ** 2"},
		{"parentheses", "say (2 + 3) * 4"},
		{"multiline_parentheses", "say (2 +\n    3)"},
		{"comments", "# comment\nlet x be 5 # another comment\nsay x"},
		{"nested_unary", "say --5"},
		{"floor_division", "say 10 // 3"},
		{"modulo", "say 10 % 3"},
		{"empty_expression", "say"},
		{"missing_identifier", "let be 5"},
		{"missing_be", "let x 5"},
		{"missing_to", "change x 5"},
		{"missing_operand", "say 2 +"},
		{"missing_parenthesis", "say (2 + 3"},
		{"unexpected_token", "42"},
		{"extra_token", "say 1 2"},
		{"invalid_character", "say @"},
		{"invalid_integer", "say 012"},
		{"unterminated_string", "say \"hello"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fullAST, fullDiagnostic, fullErr := parseFullFrontend(tc.source)
			compactAST, compactDiagnostic, compactErr := parseCompactFrontend(tc.source)

			if (fullErr == nil) != (compactErr == nil) {
				t.Fatalf(
					"lexical error mismatch:\nfull: %v\ncompact: %v",
					fullErr, compactErr,
				)
			}

			if fullErr != nil {
				if !reflect.DeepEqual(fullErr, compactErr) {
					t.Fatalf(
						"lexical errors differ:\nfull: %#v\ncompact: %#v",
						fullErr, compactErr,
					)
				}
				return
			}

			if !reflect.DeepEqual(fullDiagnostic, compactDiagnostic) {
				t.Fatalf(
					"diagnostics differ:\nfull: %#v\ncompact: %#v",
					fullDiagnostic, compactDiagnostic,
				)
			}

			if !reflect.DeepEqual(fullAST, compactAST) {
				t.Fatalf(
					"ASTs differ:\nfull: %#v\ncompact: %#v",
					fullAST, compactAST,
				)
			}
		})
	}
}

// frontendBenchmarkSource constructs a valid program containing
// declarations, reassignments, expressions and output statements.
func frontendBenchmarkSource(repetitions int) string {
	var builder strings.Builder
	builder.Grow(repetitions * 90)

	for i := range repetitions {
		fmt.Fprintf(&builder, "let value_%d be (%d + 3) * 4\n", i, i)
		fmt.Fprintf(&builder, "change value_%d to value_%d ** 2\n", i, i)
		fmt.Fprintf(&builder, "say value_%d\n", i)
	}

	return builder.String()
}

// BenchmarkFrontendFull measures production lexing and parsing.
func BenchmarkFrontendFull(b *testing.B) {
	source := frontendBenchmarkSource(300)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		program, diagnostic, err := parseFullFrontend(source)
		if err != nil || diagnostic != nil {
			b.Fatalf("frontend failed: lexical=%v syntax=%v", err, diagnostic)
		}
		frontendProgramSink = program
	}
}

// BenchmarkFrontendCompact measures compact lexing and parsing.
func BenchmarkFrontendCompact(b *testing.B) {
	source := frontendBenchmarkSource(300)
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		program, diagnostic, err := parseCompactFrontend(source)
		if err != nil || diagnostic != nil {
			b.Fatalf("frontend failed: lexical=%v syntax=%v", err, diagnostic)
		}
		frontendProgramSink = program
	}
}
