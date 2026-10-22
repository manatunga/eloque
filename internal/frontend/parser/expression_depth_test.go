package parser

import (
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
)

func TestParseRecoveringExpressionDepthLimit(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name: "grouping",
			source: "say " +
				strings.Repeat("(", maxExpressionDepth+1) +
				"42" +
				strings.Repeat(")", maxExpressionDepth+1) +
				"\n",
		},
		{
			name: "unary",
			source: "say " +
				strings.Repeat("-", maxExpressionDepth+1) +
				"42\n",
		},
		{
			name: "exponentiation",
			source: "say " +
				strings.Repeat("2 ** ", maxExpressionDepth+1) +
				"2\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseRecovering(
				tt.source,
				ParseOptions{},
			)

			if result.Status != ParseRecovered {
				t.Fatalf(
					"status = %v, want ParseRecovered",
					result.Status,
				)
			}

			if len(result.Diagnostics) != 1 {
				t.Fatalf(
					"diagnostics = %d, want 1",
					len(result.Diagnostics),
				)
			}

			if result.Diagnostics[0].Code != CodeExpressionTooDeep {
				t.Fatalf(
					"diagnostic code = %q, want %q",
					result.Diagnostics[0].Code,
					CodeExpressionTooDeep,
				)
			}

			if got := result.Program.Arena.StatementCount(); got != 0 {
				t.Fatalf("statements = %d, want 0", got)
			}
		})
	}
}

func TestParseRecoveringContinuesAfterExpressionDepthLimit(t *testing.T) {
	sourceText := "say " +
		strings.Repeat("-", maxExpressionDepth+1) +
		"42\nsay 7\n"

	result := ParseRecovering(sourceText, ParseOptions{})

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf(
			"diagnostics = %d, want 1",
			len(result.Diagnostics),
		)
	}

	if result.Diagnostics[0].Code != CodeExpressionTooDeep {
		t.Fatalf(
			"diagnostic code = %q, want %q",
			result.Diagnostics[0].Code,
			CodeExpressionTooDeep,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}
}

func TestParseFailFastExpressionDepthLimit(t *testing.T) {
	sourceText := "say " +
		strings.Repeat("-", maxExpressionDepth+1) +
		"42\n"

	program, diag := Parse(sourceText)

	if program != nil {
		t.Fatal("expected nil program")
	}

	if diag == nil {
		t.Fatal("expected diagnostic")
	}

	if diag.Code != CodeExpressionTooDeep {
		t.Fatalf(
			"diagnostic code = %q, want %q",
			diag.Code,
			CodeExpressionTooDeep,
		)
	}

	if diag.Severity != diagnostic.SeverityError {
		t.Fatalf("severity = %v, want error", diag.Severity)
	}
}

func TestParseExpressionDepthBoundary(t *testing.T) {
	// The outermost expression occupies one recursion frame.
	// Each unary operator adds another frame.
	sourceText := "say " +
		strings.Repeat("-", maxExpressionDepth-1) +
		"42\n"

	result := ParseRecovering(sourceText, ParseOptions{})

	if result.Status != ParseComplete {
		t.Fatalf(
			"status = %v, want ParseComplete; diagnostics = %v",
			result.Status,
			result.Diagnostics,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}
}

func TestParseRecoveringExpressionDepthResets(t *testing.T) {
	sourceText := "say " +
		strings.Repeat("-", maxExpressionDepth+1) +
		"42\n" +
		"say " +
		strings.Repeat("-", maxExpressionDepth-1) +
		"7\n"

	result := ParseRecovering(sourceText, ParseOptions{})

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := len(result.Diagnostics); got != 1 {
		t.Fatalf("diagnostics = %d, want 1", got)
	}

	if result.Diagnostics[0].Code != CodeExpressionTooDeep {
		t.Fatalf(
			"diagnostic code = %q, want %q",
			result.Diagnostics[0].Code,
			CodeExpressionTooDeep,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}
}
