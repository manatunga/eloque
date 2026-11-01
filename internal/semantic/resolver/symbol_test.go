package resolver

import "testing"

func TestSymbolArenaSequentialIDs(t *testing.T) {
	var arena symbolArena

	symbols := []Symbol{
		{Name: "age"},
		{Name: "score"},
		{Name: "total"},
	}

	for i, want := range symbols {
		id := arena.add(want)

		if id != SymbolID(i) {
			t.Fatalf("add() = %d, want %d", id, i)
		}

		got, ok := arena.lookup(id)
		if !ok {
			t.Fatalf("lookup(%d) failed", id)
		}

		if got != want {
			t.Errorf("lookup(%d) = %+v, want %+v", id, got, want)
		}
	}

	if len(arena.symbols) != len(symbols) {
		t.Errorf("symbol count = %d, want %d", len(arena.symbols), len(symbols))
	}
}

func TestSymbolArenaInvalidLookup(t *testing.T) {
	var arena symbolArena

	if got, ok := arena.lookup(0); ok || got != (Symbol{}) {
		t.Errorf("lookup(0) on empty arena = (%+v, %v), want zero value and false", got, ok)
	}

	arena.add(Symbol{Name: "age"})

	for _, id := range []SymbolID{1, 100, InvalidSymbolID} {
		got, ok := arena.lookup(id)

		if ok || got != (Symbol{}) {
			t.Errorf("lookup(%d) = (%+v, %v), want zero value and false", id, got, ok)
		}
	}
}

func TestSymbolArenaIndependentIDs(t *testing.T) {
	var first symbolArena
	var second symbolArena

	firstID := first.add(Symbol{Name: "age"})
	secondID := second.add(Symbol{Name: "score"})

	if firstID != 0 || secondID != 0 {
		t.Fatalf("first IDs = (%d, %d), want (0, 0)", firstID, secondID)
	}

	firstSymbol, firstOK := first.lookup(firstID)
	secondSymbol, secondOK := second.lookup(secondID)

	if !firstOK || firstSymbol.Name != "age" {
		t.Errorf("first arena lookup = (%+v, %v)", firstSymbol, firstOK)
	}

	if !secondOK || secondSymbol.Name != "score" {
		t.Errorf("second arena lookup = (%+v, %v)", secondSymbol, secondOK)
	}

	first.add(Symbol{Name: "total"})

	if len(second.symbols) != 1 {
		t.Errorf("second arena contains %d symbols, want 1", len(second.symbols))
	}
}

func TestNewSymbolArenaCapacity(t *testing.T) {
	for _, capacity := range []int{0, 1, 8, 100} {
		arena := newSymbolArena(capacity)

		if len(arena.symbols) != 0 {
			t.Fatalf("capacity %d: initial length = %d, want 0",
				capacity, len(arena.symbols))
		}

		if cap(arena.symbols) != capacity {
			t.Fatalf("requested capacity %d, got %d",
				capacity, cap(arena.symbols))
		}
	}
}

func TestSymbolArenaPreallocatedIDs(t *testing.T) {
	arena := newSymbolArena(4)

	for i := 0; i < 4; i++ {
		id := arena.add(Symbol{Name: "symbol"})

		if id != SymbolID(i) {
			t.Fatalf("add() = %d, want %d", id, i)
		}
	}

	if len(arena.symbols) != 4 {
		t.Fatalf("symbol count = %d, want 4", len(arena.symbols))
	}

	if cap(arena.symbols) != 4 {
		t.Fatalf("capacity changed before reservation was exhausted: %d",
			cap(arena.symbols))
	}

	id := arena.add(Symbol{Name: "overflow"})

	if id != 4 {
		t.Fatalf("fifth symbol ID = %d, want 4", id)
	}

	for i := SymbolID(0); i <= 4; i++ {
		if _, ok := arena.lookup(i); !ok {
			t.Errorf("symbol %d became inaccessible after growth", i)
		}
	}
}

func TestNewSymbolArenaNegativeCapacity(t *testing.T) {
	arena := newSymbolArena(-1)

	if len(arena.symbols) != 0 || cap(arena.symbols) != 0 {
		t.Fatalf("negative capacity produced len=%d cap=%d, want 0, 0",
			len(arena.symbols), cap(arena.symbols))
	}

	if id := arena.add(Symbol{Name: "first"}); id != 0 {
		t.Fatalf("first symbol ID = %d, want 0", id)
	}
}
