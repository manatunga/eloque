package ast

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
)

func TestBinaryPayloadStorage(t *testing.T) {
	var arena Arena

	want := BinaryPayload{
		Left:     0,
		Right:    1,
		Operator: lexer.TokenPlus,
	}

	index := arena.AddBinaryPayload(want)

	if index != 0 {
		t.Fatalf("first binary payload index = %d, want 0", index)
	}

	if got := arena.BinaryPayload(index); got != want {
		t.Errorf("BinaryPayload(%d) = %+v, want %+v", index, got, want)
	}
}

func TestUnaryPayloadStorage(t *testing.T) {
	var arena Arena

	want := UnaryPayload{
		Operand:  0,
		Operator: lexer.TokenMinus,
	}

	index := arena.AddUnaryPayload(want)

	if index != 0 {
		t.Fatalf("first unary payload index = %d, want 0", index)
	}

	if got := arena.UnaryPayload(index); got != want {
		t.Errorf("UnaryPayload(%d) = %+v, want %+v", index, got, want)
	}
}

func TestPayloadGrowth(t *testing.T) {
	var arena Arena

	const count = 10_000

	for i := 0; i < count; i++ {
		binary := BinaryPayload{
			Left:     NodeID(i),
			Right:    NodeID(i + 1),
			Operator: lexer.TokenStar,
		}

		unary := UnaryPayload{
			Operand:  NodeID(i),
			Operator: lexer.TokenMinus,
		}

		if index := arena.AddBinaryPayload(binary); index != uint32(i) {
			t.Fatalf("binary index = %d, want %d", index, i)
		}

		if index := arena.AddUnaryPayload(unary); index != uint32(i) {
			t.Fatalf("unary index = %d, want %d", index, i)
		}
	}

	for i := 0; i < count; i++ {
		binary := arena.BinaryPayload(uint32(i))
		unary := arena.UnaryPayload(uint32(i))

		if binary.Left != NodeID(i) ||
			binary.Right != NodeID(i+1) ||
			binary.Operator != lexer.TokenStar {
			t.Fatalf("binary payload %d corrupted: %+v", i, binary)
		}

		if unary.Operand != NodeID(i) ||
			unary.Operator != lexer.TokenMinus {
			t.Fatalf("unary payload %d corrupted: %+v", i, unary)
		}
	}
}

func TestPayloadStorageIndependent(t *testing.T) {
	var arena Arena

	binaryIndex := arena.AddBinaryPayload(BinaryPayload{
		Left:     1,
		Right:    2,
		Operator: lexer.TokenPlus,
	})

	unaryIndex := arena.AddUnaryPayload(UnaryPayload{
		Operand:  3,
		Operator: lexer.TokenMinus,
	})

	if binaryIndex != 0 || unaryIndex != 0 {
		t.Fatalf(
			"payload indices = (%d, %d), want (0, 0)",
			binaryIndex,
			unaryIndex,
		)
	}
}
