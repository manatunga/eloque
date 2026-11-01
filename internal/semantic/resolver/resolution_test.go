package resolver

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/parser"
	"github.com/manatunga/eloque/internal/source"
)

func TestNewResolution(t *testing.T) {
	for _, nodeCount := range []int{0, 1, 10, 1000} {
		r := newResolution(nodeCount, 0)

		if r == nil {
			t.Fatalf("newResolution(%d) returned nil", nodeCount)
		}

		if len(r.bindings) != nodeCount {
			t.Fatalf(
				"newResolution(%d): binding count = %d, want %d",
				nodeCount, len(r.bindings), nodeCount,
			)
		}

		for i, id := range r.bindings {
			if id != InvalidSymbolID {
				t.Errorf(
					"newResolution(%d): binding[%d] = %d, want InvalidSymbolID",
					nodeCount, i, id,
				)
			}
		}

		if len(r.symbols.symbols) != 0 {
			t.Errorf(
				"newResolution(%d): symbol count = %d, want 0",
				nodeCount, len(r.symbols.symbols),
			)
		}
	}
}

func TestNewResolutionIndependentStorage(t *testing.T) {
	first := newResolution(3, 0)
	second := newResolution(3, 0)

	first.bindings[0] = 0

	if second.bindings[0] != InvalidSymbolID {
		t.Fatal("separate resolution results share binding storage")
	}

	first.symbols.add(Symbol{Name: "age"})

	if len(second.symbols.symbols) != 0 {
		t.Fatal("separate resolution results share symbol storage")
	}
}

func TestResolutionBinding(t *testing.T) {
	r := newResolution(5, 0)

	r.bindings[1] = 0
	r.bindings[3] = 2

	tests := []struct {
		name   string
		nodeID ast.NodeID
		wantID SymbolID
		wantOK bool
	}{
		{
			name:   "valid first symbol",
			nodeID: 1,
			wantID: 0,
			wantOK: true,
		},
		{
			name:   "valid later symbol",
			nodeID: 3,
			wantID: 2,
			wantOK: true,
		},
		{
			name:   "unbound node",
			nodeID: 0,
			wantID: InvalidSymbolID,
			wantOK: false,
		},
		{
			name:   "out of range",
			nodeID: 5,
			wantID: InvalidSymbolID,
			wantOK: false,
		},
		{
			name:   "invalid node sentinel",
			nodeID: ast.InvalidNodeID,
			wantID: InvalidSymbolID,
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := r.binding(tt.nodeID)

			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf(
					"binding(%d) = (%d, %t), want (%d, %t)",
					tt.nodeID, gotID, gotOK, tt.wantID, tt.wantOK,
				)
			}
		})
	}
}

func TestResolutionBindingEmpty(t *testing.T) {
	r := newResolution(0, 0)

	gotID, gotOK := r.binding(0)

	if gotID != InvalidSymbolID || gotOK {
		t.Errorf(
			"binding(0) on empty resolution = (%d, %t), want (%d, false)",
			gotID, gotOK, InvalidSymbolID,
		)
	}
}

func TestResolutionBind(t *testing.T) {
	r := newResolution(3, 0)

	if !r.bind(1, 0) {
		t.Fatal("first binding with SymbolID(0) should succeed")
	}

	gotID, ok := r.binding(1)
	if !ok || gotID != 0 {
		t.Fatalf("binding(1) = (%d, %t), want (0, true)", gotID, ok)
	}

	if !r.bind(2, 7) {
		t.Fatal("binding another node should succeed")
	}

	gotID, ok = r.binding(2)
	if !ok || gotID != 7 {
		t.Fatalf("binding(2) = (%d, %t), want (7, true)", gotID, ok)
	}
}

func TestResolutionBindRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name     string
		nodeID   ast.NodeID
		symbolID SymbolID
	}{
		{"out of range", 3, 0},
		{"invalid node sentinel", ast.InvalidNodeID, 0},
		{"invalid symbol sentinel", 1, InvalidSymbolID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newResolution(3, 0)

			if r.bind(tt.nodeID, tt.symbolID) {
				t.Fatal("bind unexpectedly succeeded")
			}

			for i, id := range r.bindings {
				if id != InvalidSymbolID {
					t.Errorf("bindings[%d] changed to %d after rejected bind", i, id)
				}
			}
		})
	}
}

func TestResolutionBindRejectsRebinding(t *testing.T) {
	tests := []struct {
		name        string
		replacement SymbolID
	}{
		{"same symbol", 0},
		{"different symbol", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newResolution(2, 0)

			if !r.bind(1, 0) {
				t.Fatal("initial binding failed")
			}

			if r.bind(1, tt.replacement) {
				t.Fatal("rebinding should be rejected")
			}

			gotID, ok := r.binding(1)
			if !ok || gotID != 0 {
				t.Errorf(
					"original binding was not preserved: got (%d, %t), want (0, true)",
					gotID, ok,
				)
			}
		})
	}
}

func TestResolutionSymbol(t *testing.T) {
	r := newResolution(3, 0)

	want := Symbol{
		Name:        "age",
		Declaration: ast.NodeID(1),
		Span:        source.ByteSpan{Start: 4, End: 7},
	}

	id := r.symbols.add(want)

	got, ok := r.Symbol(id)
	if !ok {
		t.Fatal("Symbol returned false for a valid symbol ID")
	}

	if got != want {
		t.Errorf("Symbol(%d) = %+v, want %+v", id, got, want)
	}
}

func TestResolutionSymbolInvalidIDs(t *testing.T) {
	r := newResolution(2, 0)
	r.symbols.add(Symbol{Name: "age"})

	tests := []struct {
		name string
		id   SymbolID
	}{
		{"invalid sentinel", InvalidSymbolID},
		{"out of range", 1},
		{"larger out of range", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := r.Symbol(tt.id)

			if ok {
				t.Fatalf("Symbol(%d) unexpectedly succeeded", tt.id)
			}

			if got != (Symbol{}) {
				t.Errorf("Symbol(%d) = %+v, want zero-value Symbol", tt.id, got)
			}
		})
	}
}

func TestResolutionSymbolReturnsValue(t *testing.T) {
	r := newResolution(1, 0)
	id := r.symbols.add(Symbol{Name: "original"})

	symbol, ok := r.Symbol(id)
	if !ok {
		t.Fatal("initial symbol lookup failed")
	}

	symbol.Name = "modified"

	stored, ok := r.Symbol(id)
	if !ok {
		t.Fatal("second symbol lookup failed")
	}

	if stored.Name != "original" {
		t.Errorf("stored symbol was modified: got %q, want %q", stored.Name, "original")
	}
}

func TestResolutionDeclaration(t *testing.T) {
	r := newResolution(4, 0)

	if !r.bind(1, 0) {
		t.Fatal("failed to record declaration binding")
	}

	if !r.bind(3, 2) {
		t.Fatal("failed to record second declaration binding")
	}

	tests := []struct {
		name   string
		nodeID ast.NodeID
		wantID SymbolID
		wantOK bool
	}{
		{"first declaration", 1, 0, true},
		{"second declaration", 3, 2, true},
		{"unbound node", 0, InvalidSymbolID, false},
		{"out of range", 4, InvalidSymbolID, false},
		{"invalid node sentinel", ast.InvalidNodeID, InvalidSymbolID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := r.Declaration(tt.nodeID)

			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf(
					"Declaration(%d) = (%d, %t), want (%d, %t)",
					tt.nodeID, gotID, gotOK, tt.wantID, tt.wantOK,
				)
			}
		})
	}
}

func TestResolutionDeclarationEmpty(t *testing.T) {
	r := newResolution(0, 0)

	gotID, ok := r.Declaration(0)
	if gotID != InvalidSymbolID || ok {
		t.Errorf(
			"Declaration(0) = (%d, %t), want (%d, false)",
			gotID, ok, InvalidSymbolID,
		)
	}
}

func TestResolutionReference(t *testing.T) {
	r := newResolution(5, 0)

	if !r.bind(1, 0) {
		t.Fatal("failed to record first reference")
	}

	if !r.bind(3, 2) {
		t.Fatal("failed to record second reference")
	}

	tests := []struct {
		name   string
		nodeID ast.NodeID
		wantID SymbolID
		wantOK bool
	}{
		{"first symbol", 1, 0, true},
		{"later symbol", 3, 2, true},
		{"unbound node", 0, InvalidSymbolID, false},
		{"out of range", 5, InvalidSymbolID, false},
		{"invalid node sentinel", ast.InvalidNodeID, InvalidSymbolID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := r.Reference(tt.nodeID)

			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf(
					"Reference(%d) = (%d, %t), want (%d, %t)",
					tt.nodeID, gotID, gotOK, tt.wantID, tt.wantOK,
				)
			}
		})
	}
}

func TestResolutionReferenceEmpty(t *testing.T) {
	r := newResolution(0, 0)

	gotID, ok := r.Reference(0)

	if gotID != InvalidSymbolID || ok {
		t.Errorf(
			"Reference(0) = (%d, %t), want (%d, false)",
			gotID, ok, InvalidSymbolID,
		)
	}
}

func TestResolutionReferenceSharedSymbol(t *testing.T) {
	r := newResolution(4, 0)

	if !r.bind(1, 0) {
		t.Fatal("failed to bind first reference")
	}

	if !r.bind(3, 0) {
		t.Fatal("failed to bind second reference")
	}

	first, firstOK := r.Reference(1)
	second, secondOK := r.Reference(3)

	if !firstOK || !secondOK {
		t.Fatal("expected both references to resolve")
	}

	if first != second {
		t.Errorf(
			"references resolve to different symbols: %d and %d",
			first, second,
		)
	}
}

func TestResolveBindingsReferenceExistingSymbols(t *testing.T) {
	program, err := parser.Parse(
		"let x be 10\n" +
			"let y be x + 5\n" +
			"change x to y + missing\n" +
			"say x + y\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1", len(diagnostics))
	}
	if diagnostics[0].Code != CodeUndefinedName {
		t.Fatalf("diagnostic code = %q, want %q",
			diagnostics[0].Code, CodeUndefinedName)
	}

	boundCount := 0

	for nodeID, symbolID := range result.bindings {
		if symbolID == InvalidSymbolID {
			continue
		}

		boundCount++

		symbol, ok := result.Symbol(symbolID)
		if !ok {
			t.Errorf(
				"node %d references nonexistent symbol %d",
				nodeID, symbolID,
			)
			continue
		}

		if symbol.Name == "" {
			t.Errorf(
				"node %d references symbol %d with an empty name",
				nodeID, symbolID,
			)
		}
	}

	if boundCount != 7 {
		t.Errorf("bound node count = %d, want 7", boundCount)
	}
}

func TestNewResolutionSymbolCapacity(t *testing.T) {
	tests := []struct {
		name                 string
		nodeCount            int
		expectedDeclarations int
	}{
		{"empty", 0, 0},
		{"no declarations", 100, 0},
		{"one declaration", 100, 1},
		{"multiple declarations", 100, 25},
		{"capacity exceeds node count", 5, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newResolution(tt.nodeCount, tt.expectedDeclarations)

			if len(r.bindings) != tt.nodeCount {
				t.Fatalf("binding count = %d, want %d",
					len(r.bindings), tt.nodeCount)
			}

			if len(r.symbols.symbols) != 0 {
				t.Fatalf("initial symbol count = %d, want 0",
					len(r.symbols.symbols))
			}

			if cap(r.symbols.symbols) != tt.expectedDeclarations {
				t.Fatalf("symbol capacity = %d, want %d",
					cap(r.symbols.symbols), tt.expectedDeclarations)
			}

			for i, id := range r.bindings {
				if id != InvalidSymbolID {
					t.Fatalf("binding[%d] = %d, want InvalidSymbolID", i, id)
				}
			}
		})
	}
}

func TestResolutionPreallocationPreservesBindings(t *testing.T) {
	r := newResolution(5, 3)

	first := r.symbols.add(Symbol{Name: "first"})
	second := r.symbols.add(Symbol{Name: "second"})
	third := r.symbols.add(Symbol{Name: "third"})

	if first != 0 || second != 1 || third != 2 {
		t.Fatalf("symbol IDs = (%d, %d, %d), want (0, 1, 2)",
			first, second, third)
	}

	if !r.bind(0, first) || !r.bind(2, second) || !r.bind(4, third) {
		t.Fatal("failed to record bindings")
	}

	for _, tt := range []struct {
		node ast.NodeID
		want SymbolID
	}{
		{0, first},
		{2, second},
		{4, third},
	} {
		got, ok := r.binding(tt.node)
		if !ok || got != tt.want {
			t.Errorf("binding(%d) = (%d, %v), want (%d, true)",
				tt.node, got, ok, tt.want)
		}
	}

	if _, ok := r.binding(1); ok {
		t.Fatal("unbound node unexpectedly resolved")
	}
}
