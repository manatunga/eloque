package resolver

import "testing"

func TestNewScope(t *testing.T) {
	root := newScope(nil)

	if root == nil {
		t.Fatal("newScope returned nil")
	}
	if root.parent != nil {
		t.Fatal("root scope unexpectedly has a parent")
	}
	if root.names != nil {
		t.Fatal("new scope eagerly allocated a declaration map")
	}

	child := newScope(root)

	if child.parent != root {
		t.Fatal("child scope has the wrong parent")
	}
	if child.names != nil {
		t.Fatal("child scope eagerly allocated a declaration map")
	}
}

func TestScopeLookupLocal(t *testing.T) {
	s := newScope(nil)

	if _, ok := s.lookupLocal("missing"); ok {
		t.Fatal("empty scope unexpectedly resolved a name")
	}

	if !s.declare("x", SymbolID(7)) {
		t.Fatal("failed to declare x")
	}

	id, ok := s.lookupLocal("x")
	if !ok || id != SymbolID(7) {
		t.Fatalf("lookupLocal(x) = (%v, %v), want (7, true)", id, ok)
	}

	if _, ok := s.lookupLocal("missing"); ok {
		t.Fatal("lookupLocal resolved an undeclared name")
	}
}

func TestScopeLookupParentChain(t *testing.T) {
	root := newScope(nil)
	root.declare("global", 0)
	middle := newScope(root)
	middle.declare("function", 1)
	inner := newScope(middle)
	inner.declare("local", 2)
	tests := []struct {
		name string
		want SymbolID
		ok   bool
	}{
		{"local", 2, true},
		{"function", 1, true},
		{"global", 0, true},
		{"missing", InvalidSymbolID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := inner.lookup(tt.name)

			if got != tt.want || ok != tt.ok {
				t.Errorf(
					"lookup(%q) = (%d, %v), want (%d, %v)",
					tt.name, got, ok, tt.want, tt.ok,
				)
			}
		})
	}
}

func TestScopeShadowing(t *testing.T) {
	root := newScope(nil)
	root.declare("score", 0)
	child := newScope(root)
	child.declare("score", 1)
	if id, ok := child.lookup("score"); !ok || id != 1 {
		t.Errorf("child lookup(score) = (%d, %v), want (1, true)", id, ok)
	}

	if id, ok := root.lookup("score"); !ok || id != 0 {
		t.Errorf("root lookup(score) = (%d, %v), want (0, true)", id, ok)
	}
}

func TestScopeSiblingIsolation(t *testing.T) {
	root := newScope(nil)

	first := newScope(root)
	first.declare("value", 0)
	second := newScope(root)
	second.declare("value", 1)
	if id, ok := first.lookup("value"); !ok || id != 0 {
		t.Errorf("first lookup(value) = (%d, %v), want (0, true)", id, ok)
	}

	if id, ok := second.lookup("value"); !ok || id != 1 {
		t.Errorf("second lookup(value) = (%d, %v), want (1, true)", id, ok)
	}

	if id, ok := root.lookup("value"); ok || id != InvalidSymbolID {
		t.Errorf("root lookup(value) = (%d, %v), want (InvalidSymbolID, false)", id, ok)
	}
}

func TestScopeDeclare(t *testing.T) {
	s := newScope(nil)

	if !s.declare("age", 0) {
		t.Fatal("declare(age, 0) = false, want true")
	}

	if !s.declare("score", 1) {
		t.Fatal("declare(score, 1) = false, want true")
	}

	if id, ok := s.lookupLocal("age"); !ok || id != 0 {
		t.Errorf("lookupLocal(age) = (%d, %v), want (0, true)", id, ok)
	}

	if id, ok := s.lookupLocal("score"); !ok || id != 1 {
		t.Errorf("lookupLocal(score) = (%d, %v), want (1, true)", id, ok)
	}

	if len(s.names) != 2 {
		t.Errorf("local declaration count = %d, want 2", len(s.names))
	}
}

func TestScopeDeclareDuplicatePreservesOriginal(t *testing.T) {
	s := newScope(nil)

	if !s.declare("value", 0) {
		t.Fatal("initial declaration failed")
	}

	if s.declare("value", 1) {
		t.Fatal("duplicate declaration succeeded, want rejection")
	}

	id, ok := s.lookupLocal("value")
	if !ok || id != 0 {
		t.Errorf("lookupLocal(value) = (%d, %v), want (0, true)", id, ok)
	}

	if len(s.names) != 1 {
		t.Errorf("local declaration count = %d, want 1", len(s.names))
	}
}

func TestScopeDeclareShadowing(t *testing.T) {
	root := newScope(nil)

	if !root.declare("value", 0) {
		t.Fatal("root declaration failed")
	}

	child := newScope(root)

	if !child.declare("value", 1) {
		t.Fatal("child shadowing declaration failed")
	}

	if id, ok := child.lookup("value"); !ok || id != 1 {
		t.Errorf("child lookup(value) = (%d, %v), want (1, true)", id, ok)
	}

	if id, ok := root.lookup("value"); !ok || id != 0 {
		t.Errorf("root lookup(value) = (%d, %v), want (0, true)", id, ok)
	}
}

func TestScopeDeclareSiblingIsolation(t *testing.T) {
	root := newScope(nil)
	first := newScope(root)
	second := newScope(root)

	if !first.declare("item", 0) {
		t.Fatal("first sibling declaration failed")
	}

	if !second.declare("item", 1) {
		t.Fatal("second sibling declaration failed")
	}

	if id, ok := first.lookup("item"); !ok || id != 0 {
		t.Errorf("first lookup(item) = (%d, %v), want (0, true)", id, ok)
	}

	if id, ok := second.lookup("item"); !ok || id != 1 {
		t.Errorf("second lookup(item) = (%d, %v), want (1, true)", id, ok)
	}

	if id, ok := root.lookup("item"); ok || id != InvalidSymbolID {
		t.Errorf("root lookup(item) = (%d, %v), want (InvalidSymbolID, false)", id, ok)
	}
}

func TestScopeLazyMapInitialization(t *testing.T) {
	s := newScope(nil)

	if s.names != nil {
		t.Fatal("new scope allocated a declaration map")
	}

	if _, ok := s.lookupLocal("missing"); ok {
		t.Fatal("empty scope unexpectedly resolved a name")
	}

	if !s.declare("x", SymbolID(0)) {
		t.Fatal("first declaration failed")
	}

	if s.names == nil {
		t.Fatal("first declaration did not initialize the map")
	}

	if id, ok := s.lookupLocal("x"); !ok || id != SymbolID(0) {
		t.Fatal("declared symbol could not be resolved")
	}

	if s.declare("x", SymbolID(1)) {
		t.Fatal("duplicate declaration was accepted")
	}
}

func TestNewScopeWithCapacity(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
	}{
		{"negative", -1},
		{"zero", 0},
		{"one", 1},
		{"multiple", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := newScope(nil)
			child := newScopeWithCapacity(parent, tt.capacity)

			if child == nil {
				t.Fatal("newScopeWithCapacity returned nil")
			}

			if child.parent != parent {
				t.Fatal("scope has incorrect parent")
			}

			if tt.capacity <= 0 && child.names != nil {
				t.Fatal("nonpositive capacity allocated a map")
			}

			if tt.capacity > 0 && child.names == nil {
				t.Fatal("positive capacity did not allocate a map")
			}

			if len(child.names) != 0 {
				t.Fatalf("initial declaration count = %d, want 0",
					len(child.names))
			}
		})
	}
}

func TestScopeWithCapacityPreservesDeclarations(t *testing.T) {
	s := newScopeWithCapacity(nil, 2)

	if !s.declare("first", SymbolID(0)) {
		t.Fatal("first declaration failed")
	}

	if !s.declare("second", SymbolID(1)) {
		t.Fatal("second declaration failed")
	}

	if s.declare("first", SymbolID(2)) {
		t.Fatal("duplicate declaration was accepted")
	}

	for _, tt := range []struct {
		name string
		want SymbolID
	}{
		{"first", 0},
		{"second", 1},
	} {
		got, ok := s.lookupLocal(tt.name)

		if !ok || got != tt.want {
			t.Errorf("lookupLocal(%q) = (%d, %v), want (%d, true)",
				tt.name, got, ok, tt.want)
		}
	}

	if !s.declare("third", SymbolID(2)) {
		t.Fatal("declaration beyond initial capacity failed")
	}

	got, ok := s.lookupLocal("third")
	if !ok || got != 2 {
		t.Fatalf("lookupLocal(third) = (%d, %v), want (2, true)",
			got, ok)
	}
}

func TestScopeWithCapacityParentLookup(t *testing.T) {
	parent := newScopeWithCapacity(nil, 1)
	child := newScopeWithCapacity(parent, 2)

	if !parent.declare("outer", SymbolID(0)) {
		t.Fatal("parent declaration failed")
	}

	if !child.declare("inner", SymbolID(1)) {
		t.Fatal("child declaration failed")
	}

	if got, ok := child.lookup("outer"); !ok || got != 0 {
		t.Fatalf("child lookup(outer) = (%d, %v), want (0, true)",
			got, ok)
	}

	if got, ok := child.lookup("inner"); !ok || got != 1 {
		t.Fatalf("child lookup(inner) = (%d, %v), want (1, true)",
			got, ok)
	}

	if _, ok := parent.lookup("inner"); ok {
		t.Fatal("parent unexpectedly resolved child declaration")
	}

	if !child.declare("outer", SymbolID(2)) {
		t.Fatal("child shadowing declaration failed")
	}

	if got, ok := child.lookup("outer"); !ok || got != 2 {
		t.Fatalf("shadowed lookup = (%d, %v), want (2, true)",
			got, ok)
	}

	if got, ok := parent.lookup("outer"); !ok || got != 0 {
		t.Fatalf("parent lookup changed = (%d, %v), want (0, true)",
			got, ok)
	}
}
