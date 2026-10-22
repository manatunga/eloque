package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
)

func TestParseRecoveringValidProgram(t *testing.T) {
	result := ParseRecovering("say 1\nsay 2\n", ParseOptions{})

	if result.Status != ParseComplete {
		t.Fatalf("status = %v, want ParseComplete", result.Status)
	}
	if len(result.Diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %v", result.Diagnostics)
	}
	if result.Program == nil {
		t.Fatal("expected program")
	}
	if got := result.Program.Arena.StatementCount(); got != 2 {
		t.Fatalf("statement count = %d, want 2", got)
	}
}

func TestParseRecoveringPreservesValidStatements(t *testing.T) {
	result := ParseRecovering(
		"say 1\nsay +\nsay 3\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
	}
	if got := result.Program.Arena.StatementCount(); got != 2 {
		t.Fatalf("statement count = %d, want 2", got)
	}

	ids := result.Program.Arena.StatementRange(
		ast.StatementRange{
			Start: result.Program.Statements.Start,
			End:   result.Program.Statements.End,
		},
	)

	if len(ids) != 2 {
		t.Fatalf("committed statements = %d, want 2", len(ids))
	}
}

func TestParseRecoveringRespectsErrorBudget(t *testing.T) {
	result := ParseRecovering(
		"say +\nsay +\nsay 3\n",
		ParseOptions{MaxErrors: 1},
	)

	if result.Status != ParseAborted {
		t.Fatalf("status = %v, want ParseAborted", result.Status)
	}
	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
	}
}

func TestParseRecoveringPreservesStatementAfterBrokenGrouping(t *testing.T) {
	result := ParseRecovering(
		"say (1 +\nsay 42\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}
	if len(result.Diagnostics) == 0 {
		t.Fatal("expected syntax diagnostic")
	}
	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}

func TestParseRecoveringValidMultilineExpression(t *testing.T) {
	result := ParseRecovering(
		"say (\n1 +\n2\n)\nsay 3\n",
		ParseOptions{},
	)

	if result.Status != ParseComplete {
		t.Fatalf("status = %v, want ParseComplete; diagnostics: %v",
			result.Status, result.Diagnostics)
	}
	if got := result.Program.Arena.StatementCount(); got != 2 {
		t.Fatalf("statement count = %d, want 2", got)
	}
}

func TestParseRecoveringConsecutiveLexicalErrors(t *testing.T) {
	result := ParseRecovering(
		"say @ @\nsay 3\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if len(result.Diagnostics) != 2 {
		t.Fatalf(
			"diagnostics = %d, want 2: %v",
			len(result.Diagnostics),
			result.Diagnostics,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}

func TestParseRecoveringLexicalErrorInsideExpression(t *testing.T) {
	result := ParseRecovering(
		"say 1 + @\nsay 2\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf(
			"diagnostics = %d, want 1: %v",
			len(result.Diagnostics),
			result.Diagnostics,
		)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}

func TestParseRecoveringRepeatedInvalidTokens(t *testing.T) {
	result := ParseRecovering(
		"@ @ @\nsay 42\n",
		ParseOptions{MaxErrors: -1},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if len(result.Diagnostics) == 0 {
		t.Fatal("expected diagnostics")
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}

func TestParseRecoveringUnexpectedStatementStarter(t *testing.T) {
	result := ParseRecovering(
		"1\n2\nsay 3\n",
		ParseOptions{MaxErrors: -1},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if len(result.Diagnostics) != 2 {
		t.Fatalf("diagnostics = %d, want 2", len(result.Diagnostics))
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}

func TestParseRecoveringUnexpectedTokenAfterStatement(t *testing.T) {
	result := ParseRecovering(
		"say 1 2\nsay 3\n",
		ParseOptions{MaxErrors: -1},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1", len(result.Diagnostics))
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}

func TestParseRecoveringRollsBackMalformedStatement(t *testing.T) {
	result := ParseRecovering(
		"say 1 +\nsay 2\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}

	// The malformed statement may have allocated an integer node
	// before failing. Rollback must remove that abandoned node.
	// The valid "say 2" statement needs exactly two nodes:
	// one integer literal and one say statement.
	if got := result.Program.Arena.Len(); got != 2 {
		t.Fatalf("arena nodes = %d, want 2", got)
	}
}

func TestParseRecoveringRollsBackInvalidTerminator(t *testing.T) {
	result := ParseRecovering(
		"say 1 2\nsay 3\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := result.Program.Arena.Len(); got != 2 {
		t.Fatalf("arena nodes = %d, want 2", got)
	}
}

func TestParseRecoveringRollbackPreservesEarlierStatements(t *testing.T) {
	result := ParseRecovering(
		"say 1\nsay 2 +\nsay 3\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := result.Program.Arena.StatementCount(); got != 2 {
		t.Fatalf("statement count = %d, want 2", got)
	}

	if got := result.Program.Arena.Len(); got != 4 {
		t.Fatalf("arena nodes = %d, want 4", got)
	}
}

func TestParseRecoveringMixedErrors(t *testing.T) {
	result := ParseRecovering(
		"say +\nsay @\nsay 42\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := len(result.Diagnostics); got != 2 {
		t.Fatalf("diagnostics = %d, want 2: %v",
			got, result.Diagnostics)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}

	if got := result.Program.Arena.Len(); got != 2 {
		t.Fatalf("arena nodes = %d, want 2", got)
	}

	if result.Diagnostics[0].Primary.Span.Start >=
		result.Diagnostics[1].Primary.Span.Start {
		t.Fatal("diagnostics are not in source order")
	}
}

func TestParseRecoveringTerminalLexicalError(t *testing.T) {
	result := ParseRecovering(
		"say 1\nsay \"unterminated",
		ParseOptions{},
	)

	if result.Status != ParseAborted {
		t.Fatalf("status = %v, want ParseAborted", result.Status)
	}

	if got := len(result.Diagnostics); got != 1 {
		t.Fatalf("diagnostics = %d, want 1: %v",
			got, result.Diagnostics)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}

	if got := result.Program.Arena.Len(); got != 2 {
		t.Fatalf("arena nodes = %d, want 2", got)
	}
}

func TestParseRecoveringUnlimitedErrors(t *testing.T) {
	result := ParseRecovering(
		"say +\nsay +\nsay +\nsay 7\n",
		ParseOptions{MaxErrors: -1},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered", result.Status)
	}

	if got := len(result.Diagnostics); got != 3 {
		t.Fatalf("diagnostics = %d, want 3", got)
	}

	if got := result.Program.Arena.StatementCount(); got != 1 {
		t.Fatalf("statements = %d, want 1", got)
	}
}
