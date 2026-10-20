package ast

import (
	"testing"

	"github.com/manatunga/eloque/internal/source"
)

func TestFlatProgramStatementOrder(t *testing.T) {
	arena := NewArena("let age be 20\nsay age")

	integer := arena.AddInteger(
		source.ByteSpan{Start: 11, End: 13},
	)

	identifier := arena.AddIdentifier(
		source.ByteSpan{Start: 18, End: 21},
	)

	letID := arena.AddLet(
		source.ByteSpan{Start: 0, End: 13},
		source.ByteSpan{Start: 4, End: 7},
		integer,
	)

	sayID := arena.AddSay(
		source.ByteSpan{Start: 14, End: 21},
		identifier,
	)

	arena.AddStatement(letID)
	arena.AddStatement(sayID)

	program := FlatProgram{
		Arena: arena,
		Statements: StatementRange{
			Start: 0,
			End:   2,
		},
	}

	got := program.Arena.StatementRange(program.Statements)
	want := []NodeID{letID, sayID}

	if len(got) != len(want) {
		t.Fatalf("statement count = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("statement[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestArenaRejectsExpressionAsStatement(t *testing.T) {
	arena := NewArena("42")

	expression := arena.AddInteger(
		source.ByteSpan{Start: 0, End: 2},
	)

	defer func() {
		if recover() == nil {
			t.Error("expected panic for expression used as statement")
		}

		if len(arena.statements) != 0 {
			t.Error("statement buffer modified after invalid append")
		}
	}()

	arena.AddStatement(expression)
}

func TestArenaRejectsInvalidStatementID(t *testing.T) {
	var arena Arena

	defer func() {
		if recover() == nil {
			t.Error("expected panic for invalid statement ID")
		}
	}()

	arena.AddStatement(InvalidNodeID)
}

func TestArenaStatementRange(t *testing.T) {
	arena := NewArena("say 1\nsay 2")

	firstValue := arena.AddInteger(
		source.ByteSpan{Start: 4, End: 5},
	)
	secondValue := arena.AddInteger(
		source.ByteSpan{Start: 10, End: 11},
	)

	first := arena.AddSay(
		source.ByteSpan{Start: 0, End: 5},
		firstValue,
	)
	second := arena.AddSay(
		source.ByteSpan{Start: 6, End: 11},
		secondValue,
	)

	arena.AddStatement(first)
	arena.AddStatement(second)

	got := arena.StatementRange(StatementRange{
		Start: 1,
		End:   2,
	})

	if len(got) != 1 || got[0] != second {
		t.Errorf("statement range = %v, want [%d]", got, second)
	}
}
