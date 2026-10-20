package ast

import (
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// assertArenaPanic verifies that an invalid arena operation panics
// with the expected descriptive error message.
func assertArenaPanic(t *testing.T, want string, fn func()) {
	t.Helper()

	defer func() {
		recovered := recover()

		if recovered == nil {
			t.Fatalf("expected panic containing %q", want)
		}

		message, ok := recovered.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T", recovered)
		}

		if !strings.Contains(message, want) {
			t.Errorf("panic = %q, want substring %q", message, want)
		}
	}()

	fn()
}

func TestArenaGetInvalidIDs(t *testing.T) {
	arena := NewArena("42")
	arena.AddInteger(source.ByteSpan{Start: 0, End: 2})

	tests := []struct {
		name string
		id   NodeID
	}{
		{"first nonexistent ID", 1},
		{"reserved invalid ID", InvalidNodeID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertArenaPanic(t, "ast: invalid node ID", func() {
				arena.Get(tt.id)
			})
		})
	}
}

func TestArenaPayloadBounds(t *testing.T) {
	arena := NewArena("")

	tests := []struct {
		name string
		read func()
		want string
	}{
		{
			name: "binary payload",
			read: func() { arena.BinaryPayload(0) },
			want: "ast: invalid binary payload index",
		},
		{
			name: "unary payload",
			read: func() { arena.UnaryPayload(0) },
			want: "ast: invalid unary payload index",
		},
		{
			name: "binding payload",
			read: func() { arena.BindingPayload(0) },
			want: "ast: invalid binding payload index",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertArenaPanic(t, tt.want, tt.read)
		})
	}
}

func TestArenaStringValueValidation(t *testing.T) {
	arena := NewArena("42")
	integer := arena.AddInteger(source.ByteSpan{Start: 0, End: 2})

	t.Run("wrong node kind", func(t *testing.T) {
		assertArenaPanic(t, "ast: node is not a string literal", func() {
			arena.StringValue(integer)
		})
	})

	t.Run("invalid node ID", func(t *testing.T) {
		assertArenaPanic(t, "ast: invalid node ID", func() {
			arena.StringValue(InvalidNodeID)
		})
	})

	t.Run("invalid payload index", func(t *testing.T) {
		id := arena.Add(FlatNode{
			Kind: NodeStringLiteral,
			Data: 0,
		})

		assertArenaPanic(t, "ast: invalid string payload index", func() {
			arena.StringValue(id)
		})
	})
}

func TestArenaStatementRangeValidation(t *testing.T) {
	arena := NewArena("say 1")

	value := arena.AddInteger(source.ByteSpan{
		Start: 4,
		End:   5,
	})

	statement := arena.AddSay(
		source.ByteSpan{Start: 0, End: 5},
		value,
	)

	arena.AddStatement(statement)

	tests := []struct {
		name       string
		rangeValue StatementRange
		valid      bool
	}{
		{
			name:       "full range",
			rangeValue: StatementRange{Start: 0, End: 1},
			valid:      true,
		},
		{
			name:       "empty range at end",
			rangeValue: StatementRange{Start: 1, End: 1},
			valid:      true,
		},
		{
			name:       "reversed range",
			rangeValue: StatementRange{Start: 1, End: 0},
		},
		{
			name:       "end beyond buffer",
			rangeValue: StatementRange{Start: 0, End: 2},
		},
		{
			name:       "start beyond buffer",
			rangeValue: StatementRange{Start: 2, End: 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				arena.StatementRange(tt.rangeValue)
				return
			}

			assertArenaPanic(t, "ast: invalid statement range", func() {
				arena.StatementRange(tt.rangeValue)
			})
		})
	}
}

func TestArenaAddStatementRejectsExpression(t *testing.T) {
	arena := NewArena("42")

	expression := arena.AddInteger(source.ByteSpan{
		Start: 0,
		End:   2,
	})

	assertArenaPanic(t, "ast: node is not a statement", func() {
		arena.AddStatement(expression)
	})

	if got := arena.StatementCount(); got != 0 {
		t.Errorf("statement count = %d, want 0", got)
	}
}

func TestArenaExpressionConstructorsRejectInvalidSpans(t *testing.T) {
	arena := NewArena("1 + 2")

	left := arena.AddInteger(source.ByteSpan{Start: 0, End: 1})
	right := arena.AddInteger(source.ByteSpan{Start: 4, End: 5})

	invalid := source.ByteSpan{Start: 0, End: 6}

	t.Run("binary expression", func(t *testing.T) {
		before := arena.Len()

		assertArenaPanic(t, "ast: invalid source span", func() {
			arena.AddBinary(invalid, left, lexer.TokenPlus, right)
		})

		if got := arena.Len(); got != before {
			t.Errorf("node count = %d, want %d", got, before)
		}
	})

	t.Run("unary expression", func(t *testing.T) {
		before := arena.Len()

		assertArenaPanic(t, "ast: invalid source span", func() {
			arena.AddUnary(invalid, lexer.TokenMinus, left)
		})

		if got := arena.Len(); got != before {
			t.Errorf("node count = %d, want %d", got, before)
		}
	})
}

func TestArenaExpressionConstructorsRejectStatementOperands(t *testing.T) {
	arena := NewArena("say 1")

	value := arena.AddInteger(source.ByteSpan{Start: 4, End: 5})
	statement := arena.AddSay(
		source.ByteSpan{Start: 0, End: 5},
		value,
	)

	t.Run("binary expression", func(t *testing.T) {
		assertArenaPanic(t, "ast: node is not an expression", func() {
			arena.AddBinary(
				source.ByteSpan{Start: 0, End: 5},
				statement,
				lexer.TokenPlus,
				value,
			)
		})
	})

	t.Run("unary expression", func(t *testing.T) {
		assertArenaPanic(t, "ast: node is not an expression", func() {
			arena.AddUnary(
				source.ByteSpan{Start: 0, End: 5},
				lexer.TokenMinus,
				statement,
			)
		})
	})
}

func TestArenaConstructorRejectsInvalidExpressionID(t *testing.T) {
	arena := NewArena("say 1")

	assertArenaPanic(t, "ast: invalid expression ID", func() {
		arena.AddSay(
			source.ByteSpan{Start: 0, End: 5},
			InvalidNodeID,
		)
	})
}

func TestPayloadCapacityBoundaries(t *testing.T) {
	max := uint64(^uint32(0))

	tests := []struct {
		name      string
		length    uint64
		wantPanic bool
	}{
		{"empty", 0, false},
		{"ordinary", 100, false},
		{"maximum minus one", max - 1, false},
		{"maximum index", max, false},
		{"beyond maximum", max + 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This boundary requires a 64-bit int.
			if tt.length > uint64(^uint(0)>>1) {
				t.Skip("boundary not representable by int")
			}

			check := func() {
				checkPayloadCapacity(int(tt.length))
			}

			if tt.wantPanic {
				assertArenaPanic(t, "ast: payload arena exhausted", check)
				return
			}

			check()
		})
	}
}
