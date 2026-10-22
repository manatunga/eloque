package parser

import (
	"math/rand"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
)

func TestParseRecoveringRandomMalformedInputs(t *testing.T) {
	const (
		seed       = int64(20261022)
		iterations = 1000
		maxLength  = 256
	)

	rng := rand.New(rand.NewSource(seed))

	alphabet := []byte(
		"abcdefghijklmnopqrstuvwxyz" +
			"0123456789" +
			" \t\n+-*/()=\"'@#$\\",
	)

	for i := 0; i < iterations; i++ {
		length := rng.Intn(maxLength + 1)
		input := make([]byte, length)

		for j := range input {
			input[j] = alphabet[rng.Intn(len(alphabet))]
		}

		sourceText := string(input)

		result := ParseRecovering(
			sourceText,
			ParseOptions{MaxErrors: -1},
		)

		if result.Program == nil {
			t.Fatalf("iteration %d: nil program for %q", i, sourceText)
		}

		if result.Status == ParseComplete && len(result.Diagnostics) != 0 {
			t.Fatalf(
				"iteration %d: complete parse with diagnostics: %q",
				i, sourceText,
			)
		}

		if result.Status == ParseRecovered && len(result.Diagnostics) == 0 {
			t.Fatalf(
				"iteration %d: recovered without diagnostics: %q",
				i, sourceText,
			)
		}

		statements := result.Program.Arena.StatementRange(
			ast.StatementRange{
				Start: result.Program.Statements.Start,
				End:   result.Program.Statements.End,
			},
		)

		for _, id := range statements {
			node := result.Program.Arena.Get(id)

			if !node.Kind.IsStatement() {
				t.Fatalf(
					"iteration %d: non-statement node committed: %q",
					i, sourceText,
				)
			}

			if node.Span.Start > node.Span.End ||
				int(node.Span.End) > len(sourceText) {
				t.Fatalf(
					"iteration %d: invalid statement span: %q",
					i, sourceText,
				)
			}
		}

		for _, diag := range result.Diagnostics {
			if diag == nil {
				t.Fatalf("iteration %d: nil diagnostic", i)
			}

			span := diag.Primary.Span

			if span.Start > span.End ||
				int(span.End) > len(sourceText) {
				t.Fatalf(
					"iteration %d: invalid diagnostic span: %q",
					i, sourceText,
				)
			}
		}
	}
}

func FuzzParseRecovering(f *testing.F) {
	seeds := []string{
		"",
		"say 1\n",
		"say +\nsay 2\n",
		"say (1 +\nsay 42\n",
		"say @ @\nsay 3\n",
		"say \"unterminated",
		"let x be 1\nchange x be 2\n",
		"say (\n1 +\n2\n)\n",
		"say 1 2\nsay 3\n",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, sourceText string) {
		// Bound input size to keep individual executions manageable.
		if len(sourceText) > 4096 {
			t.Skip()
		}

		result := ParseRecovering(
			sourceText,
			ParseOptions{MaxErrors: -1},
		)

		if result.Program == nil {
			t.Fatal("nil program")
		}

		if result.Status == ParseComplete &&
			len(result.Diagnostics) != 0 {
			t.Fatal("complete parse contains diagnostics")
		}

		if result.Status == ParseRecovered &&
			len(result.Diagnostics) == 0 {
			t.Fatal("recovered parse contains no diagnostics")
		}
	})
}
