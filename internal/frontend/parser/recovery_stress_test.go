package parser

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseRecoveringDeepGrouping(t *testing.T) {
	for _, depth := range []int{10, 100, 1000} {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			sourceText := "say " +
				strings.Repeat("(", depth) +
				"42" +
				strings.Repeat(")", depth) +
				"\n"

			result := ParseRecovering(sourceText, ParseOptions{})

			if result.Status != ParseComplete {
				t.Fatalf(
					"depth %d: status = %v, want ParseComplete; diagnostics = %v",
					depth,
					result.Status,
					result.Diagnostics,
				)
			}

			if got := result.Program.Arena.StatementCount(); got != 1 {
				t.Fatalf(
					"depth %d: statements = %d, want 1",
					depth,
					got,
				)
			}
		})
	}
}

func TestParseRecoveringManyMalformedStatements(t *testing.T) {
	const malformedCount = 10_000

	var source strings.Builder
	source.Grow(malformedCount*6 + 16)

	for range malformedCount {
		source.WriteString("say +\n")
	}

	source.WriteString("say 42\n")

	result := ParseRecovering(
		source.String(),
		ParseOptions{MaxErrors: -1},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := len(result.Diagnostics); got != malformedCount {
		t.Fatalf(
			"diagnostics = %d, want %d",
			got,
			malformedCount,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}
}

func TestParseRecoveringLargeErrorBudget(t *testing.T) {
	const (
		malformedCount = 1000
		errorBudget    = 100
	)

	sourceText := strings.Repeat("say +\n", malformedCount) +
		"say 42\n"

	result := ParseRecovering(
		sourceText,
		ParseOptions{MaxErrors: errorBudget},
	)

	if result.Status != ParseAborted {
		t.Fatalf("status = %v, want ParseAborted", result.Status)
	}

	if got := len(result.Diagnostics); got != errorBudget {
		t.Fatalf(
			"diagnostics = %d, want %d",
			got,
			errorBudget,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 0 {
		t.Fatalf("statements = %d, want 0", got)
	}
}

func TestParseRecoveringAlternatingValidAndInvalid(t *testing.T) {
	const pairs = 1000

	var source strings.Builder
	source.Grow(pairs * 16)

	for range pairs {
		source.WriteString("say 42\n")
		source.WriteString("say +\n")
	}

	result := ParseRecovering(
		source.String(),
		ParseOptions{MaxErrors: -1},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := len(result.Diagnostics); got != pairs {
		t.Fatalf("diagnostics = %d, want %d", got, pairs)
	}

	if got := result.Program.Arena.StatementCount(); got != pairs {
		t.Fatalf("statements = %d, want %d", got, pairs)
	}
}
