package resolver

// scope represents a lexical environment containing locally declared names.
//
// Each scope stores only its own declarations. Name lookup follows the
// parent chain when an identifier is absent from the current scope.
//
// A nil parent identifies the root scope. Child scopes may shadow names
// from enclosing scopes without modifying their bindings.
type scope struct {
	parent *scope
	names  map[string]SymbolID
}

// newScope constructs a lexical scope without allocating its
// declaration map until the first symbol is declared.
func newScope(parent *scope) *scope {
	return &scope{
		parent: parent,
	}
}

// newScopeWithCapacity creates a lexical scope with room for an expected
// number of declarations.
//
// A zero or negative hint leaves the name map unallocated until the
// first declaration. The hint affects allocation only, not lookup or
// declaration semantics.
func newScopeWithCapacity(parent *scope, expectedDeclarations int) *scope {
	if expectedDeclarations <= 0 {
		return newScope(parent)
	}

	return &scope{
		parent: parent,
		names:  make(map[string]SymbolID, expectedDeclarations),
	}
}

// lookupLocal finds a declaration in the current scope only.
//
// It does not search enclosing scopes. If name is absent, it returns
// InvalidSymbolID and false.
func (s *scope) lookupLocal(ident string) (SymbolID, bool) {
	id, exists := s.names[ident]
	if !exists {
		return InvalidSymbolID, false
	}
	return id, true
}

// lookup finds the nearest visible declaration of name.
//
// It searches the current scope before traversing enclosing scopes.
// The first matching declaration wins, allowing inner scopes to shadow
// outer declarations.
//
// If no declaration is found, lookup returns InvalidSymbolID and false.
func (s *scope) lookup(ident string) (SymbolID, bool) {
	current := s
	for current != nil {
		if id, exists := current.lookupLocal(ident); exists {
			return id, true
		}
		current = current.parent
	}
	return InvalidSymbolID, false
}

// declare registers a symbol in the current lexical scope.
//
// It returns false if name is already declared locally, leaving the
// existing binding unchanged. Declarations in enclosing scopes do not
// prevent insertion, allowing lexical shadowing.
//
// Duplicate-declaration errors and shadowing warnings are handled by
// the resolver rather than by the scope itself.
func (s *scope) declare(ident string, id SymbolID) bool {
	if _, exists := s.names[ident]; exists {
		return false
	}

	if s.names == nil {
		s.names = make(map[string]SymbolID)
	}

	s.names[ident] = id

	return true
}
