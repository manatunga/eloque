package resolver

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
)

// resolver holds the mutable state of a single lexical name-resolution pass.
//
// It traverses an existing flat AST without modifying its nodes, records
// declaration and reference bindings in a Resolution, and accumulates
// structured diagnostics.
//
// Each resolver instance is confined to one compilation. Lexical scopes
// are maintained independently of the AST and symbol arena.
type resolver struct {
	arena        *ast.Arena
	currentScope *scope
	result       *Resolution
	diagnostics  []diagnostic.Diagnostic
}

// newResolver constructs the working state for a single semantic resolution
// pass without reserving storage for declarations.
//
// The resolver owns its lexical scope hierarchy, symbol bindings, and
// diagnostic collection. Its AST arena is borrowed from the caller and
// remains the source of all syntax nodes and source spans.
//
// This constructor is primarily used when declaration counts are unknown
// or when callers do not require capacity planning. Symbol storage and
// scope maps grow as declarations are encountered.
//
// Each invocation creates independent resolution state. Symbol identities
// and lexical scopes are never shared between resolver instances.
func newResolver(arena *ast.Arena) *resolver {
	return newResolverWithCapacity(arena, 0)
}

// newResolverWithCapacity constructs the working state for a semantic
// resolution pass with storage reserved for expected declarations.
//
// expectedDeclarations is a capacity hint rather than a declaration limit.
// The resolver may process more declarations than anticipated without
// invalidating existing SymbolIDs or changing name-resolution semantics.
//
// A positive hint reserves storage in two independent structures:
//
//   - The symbol arena, which stores symbols under stable numeric IDs.
//   - The root lexical scope, which maps declared names to those IDs.
//
// When expectedDeclarations is zero or negative, the root scope retains
// lazy map allocation. This avoids allocating a name map for programs
// that contain no declarations.
//
// Capacity planning affects allocation behaviour only. It must not alter
// duplicate detection, lexical lookup, shadowing, binding identities,
// or diagnostic results.
//
// The supplied AST arena remains caller-owned; all newly constructed
// resolver state is private to this resolution pass.
func newResolverWithCapacity(
	arena *ast.Arena,
	expectedDeclarations int,
) *resolver {
	root := newScope(nil)

	if expectedDeclarations > 0 {
		root.names = make(map[string]SymbolID, expectedDeclarations)
	}

	return &resolver{
		arena:        arena,
		currentScope: root,
		result:       newResolution(arena.Len(), expectedDeclarations),
	}
}

// addDiagnostic appends a resolver diagnostic, growing storage geometrically
// when the current capacity is exhausted. Programs without diagnostics retain
// a nil diagnostics slice and incur no diagnostic-buffer allocation.
func (r *resolver) addDiagnostic(d diagnostic.Diagnostic) {
	if len(r.diagnostics) == cap(r.diagnostics) {
		newCapacity := 8
		if current := cap(r.diagnostics); current > 0 {
			newCapacity = current * 2
		}

		grown := make([]diagnostic.Diagnostic, len(r.diagnostics), newCapacity)
		copy(grown, r.diagnostics)
		r.diagnostics = grown
	}

	r.diagnostics = append(r.diagnostics, d)
}

// Resolve performs lexical name resolution for a parsed Eloque program.
//
// It associates declarations and identifier references with compilation-local
// symbols, preserving the original AST and collecting structured diagnostics.
//
// Declarations are processed in source order. Name resolution does not perform
// type checking or definite-assignment analysis.
//
// The caller must provide a non-nil program with a non-nil AST arena.
func Resolve(program *ast.FlatProgram) (*Resolution, []diagnostic.Diagnostic) {
	statementIDs := program.Arena.StatementRange(program.Statements)

	declarationCount := 0
	for _, nodeID := range statementIDs {
		if program.Arena.Get(nodeID).Kind == ast.NodeLetStatement {
			declarationCount++
		}
	}

	r := newResolverWithCapacity(program.Arena, declarationCount)

	for _, nodeID := range statementIDs {
		r.resolveStatement(nodeID)
	}

	return r.result, r.diagnostics
}

// resolveStatement processes a statement node during lexical name resolution.
//
// It records declaration and reference bindings, resolves expressions
// contained in the statement, and emits diagnostics for invalid names.
//
// Statement-specific handling is implemented as the resolver gains
// support for each AST statement kind.
func (r *resolver) resolveStatement(nodeID ast.NodeID) {
	node := r.arena.Get(nodeID)
	switch node.Kind {
	case ast.NodeLetStatement:
		payload := r.arena.BindingPayload(node.Data)
		name := r.arena.SourceText(payload.Name)

		existingID, exists := r.currentScope.lookupLocal(name)
		if exists {
			original, _ := r.result.Symbol(existingID)

			diagnostic := diagnostic.Diagnostic{
				Code:     CodeDuplicateDeclaration,
				Severity: diagnostic.SeverityError,
				Message:  "duplicate declaration: " + name,
				Primary: diagnostic.Location{
					File: 0,
					Span: payload.Name,
				},
				Related: []diagnostic.RelatedLocation{
					{
						Location: diagnostic.Location{
							File: 0,
							Span: original.Span,
						},
						Message: "first declared here",
					},
				},
			}
			r.addDiagnostic(diagnostic)
			r.resolveExpression(payload.Value)
			return
		}

		if r.currentScope.parent != nil {
			existingID, exists := r.currentScope.parent.lookup(name)
			if exists {
				existingSymbol, _ := r.result.Symbol(existingID)

				diagnostic := diagnostic.Diagnostic{
					Code:     CodeShadowedDeclaration,
					Severity: diagnostic.SeverityWarning,
					Message:  "declaration shadows outer variable: " + name,
					Primary: diagnostic.Location{
						File: 0,
						Span: payload.Name,
					},
					Related: []diagnostic.RelatedLocation{
						{
							Location: diagnostic.Location{
								File: 0,
								Span: existingSymbol.Span,
							},
							Message: "outer declaration here",
						},
					},
				}
				r.addDiagnostic(diagnostic)
			}
		}

		symbol := Symbol{
			Name:        name,
			Declaration: nodeID,
			Span:        payload.Name,
		}

		symbolID := r.result.symbols.add(symbol)
		r.currentScope.declare(name, symbolID)
		r.result.bind(nodeID, symbolID)
		r.resolveExpression(payload.Value)

	case ast.NodeSayStatement:
		expressionID := ast.NodeID(node.Data)
		r.resolveExpression(expressionID)

	case ast.NodeChangeStatement:
		payload := r.arena.BindingPayload(node.Data)
		name := r.arena.SourceText(payload.Name)

		symbolID, exists := r.currentScope.lookup(name)
		if exists {
			r.result.bind(nodeID, symbolID)
		} else {
			diagnostic := diagnostic.Diagnostic{
				Code:     CodeUndefinedName,
				Severity: diagnostic.SeverityError,
				Message:  "undefined variable: " + name,
				Primary: diagnostic.Location{
					File: 0,
					Span: payload.Name,
				},
			}
			r.addDiagnostic(diagnostic)
		}

		r.resolveExpression(payload.Value)
	}
}

// resolveExpression resolves identifier references within an expression.
//
// Literal expressions require no bindings. Identifier expressions are
// associated with visible declarations, while compound expressions
// recursively resolve their operands.
//
// Name resolution does not evaluate expressions or check their types.
func (r *resolver) resolveExpression(nodeID ast.NodeID) {
	node := r.arena.Get(nodeID)
	switch node.Kind {
	case ast.NodeIntegerLiteral, ast.NodeStringLiteral:
		return

	case ast.NodeIdentifier:
		ident := r.arena.SourceText(node.Span)
		symbolID, exists := r.currentScope.lookup(ident)
		if !exists {
			diagnostic := diagnostic.Diagnostic{
				Code:     CodeUndefinedName,
				Severity: diagnostic.SeverityError,
				Message:  "undefined variable: " + ident,
				Primary: diagnostic.Location{
					File: 0,
					Span: node.Span,
				},
			}
			r.addDiagnostic(diagnostic)
			return
		}
		r.result.bind(nodeID, symbolID)
		return

	case ast.NodeUnaryExpression:
		payload := r.arena.UnaryPayload(node.Data)
		r.resolveExpression(payload.Operand)
		return

	case ast.NodeBinaryExpression:
		payload := r.arena.BinaryPayload(node.Data)
		r.resolveExpression(payload.Left)
		r.resolveExpression(payload.Right)
		return
	}
}
