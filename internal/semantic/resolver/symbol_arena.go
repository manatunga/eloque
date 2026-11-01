package resolver

// symbolArena stores symbols in a contiguous, append-only collection.
//
// Each symbol's index corresponds to its SymbolID. Appending symbols
// preserves existing identities, even if the slice's backing array
// is reallocated.
//
// The arena is owned by a single compilation and does not share
// symbol identities with other compilations.
type symbolArena struct {
	symbols []Symbol
}

// newSymbolArena creates a symbol arena with capacity for the expected
// number of declarations. Capacity is a performance hint and does not
// affect symbol identities or the number of registered symbols.
func newSymbolArena(expectedDeclarations int) symbolArena {
	if expectedDeclarations < 0 {
		expectedDeclarations = 0
	}

	return symbolArena{
		symbols: make([]Symbol, 0, expectedDeclarations),
	}
}

// add appends a symbol to the arena and returns its assigned SymbolID.
//
// Symbol IDs are assigned sequentially from zero. Previously assigned
// IDs remain stable after insertion.
//
// add panics if the arena has exhausted its available valid symbol IDs.
func (a *symbolArena) add(symbol Symbol) SymbolID {
	if uint64(len(a.symbols)) >= uint64(InvalidSymbolID) {
		panic("resolver: symbol arena exhausted")
	}
	symbolID := SymbolID(len(a.symbols))
	a.symbols = append(a.symbols, symbol)
	return symbolID
}

// lookup retrieves the symbol identified by id.
//
// It returns the symbol and true when id identifies an existing entry.
// InvalidSymbolID and out-of-range IDs return the zero-value Symbol
// and false without panicking.
func (a *symbolArena) lookup(id SymbolID) (Symbol, bool) {
	if uint64(id) >= uint64(len(a.symbols)) {
		return Symbol{}, false
	}
	return a.symbols[id], true
}
