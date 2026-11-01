package resolver

import (
	"reflect"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/parser"
)

func TestNewResolver(t *testing.T) {
	arena := ast.NewArena("")

	r := newResolver(arena)
	if r == nil {
		t.Fatal("newResolver returned nil")
	}
	if r.arena != arena {
		t.Fatal("resolver retained the wrong arena")
	}
	if r.currentScope == nil {
		t.Fatal("resolver did not create a root scope")
	}
	if r.currentScope.parent != nil {
		t.Fatal("resolver root scope unexpectedly has a parent")
	}
	if r.currentScope.names != nil {
		t.Fatal("resolver eagerly allocated the root scope map")
	}
	if r.result == nil {
		t.Fatal("resolver did not initialize its resolution result")
	}
	if len(r.diagnostics) != 0 {
		t.Fatal("new resolver unexpectedly contains diagnostics")
	}
}

func TestNewResolverIndependentState(t *testing.T) {
	arena := new(ast.Arena)

	first := newResolver(arena)
	second := newResolver(arena)

	if first.currentScope == second.currentScope {
		t.Fatal("resolvers share their root scope")
	}

	if first.result == second.result {
		t.Fatal("resolvers share their resolution result")
	}

	first.currentScope.declare("age", 0)

	if _, exists := second.currentScope.lookup("age"); exists {
		t.Fatal("declaration leaked into another resolver")
	}

	first.result.symbols.add(Symbol{Name: "age"})

	if len(second.result.symbols.symbols) != 0 {
		t.Fatal("symbols leaked into another resolver")
	}
}

func TestResolveParsedProgram(t *testing.T) {
	sourceText := "let x be 10\nlet y be x + 5\nsay y * 2\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}

	statements := program.Arena.StatementRange(program.Statements)
	if len(statements) != 3 {
		t.Fatalf("statement count = %d, want 3", len(statements))
	}

	xDecl := statements[0]
	yDecl := statements[1]

	xID, ok := result.Declaration(xDecl)
	if !ok {
		t.Fatal("x declaration was not bound")
	}

	yID, ok := result.Declaration(yDecl)
	if !ok {
		t.Fatal("y declaration was not bound")
	}

	if xID == yID {
		t.Fatal("x and y declarations share a symbol ID")
	}

	xSymbol, ok := result.Symbol(xID)
	if !ok || xSymbol.Name != "x" || xSymbol.Declaration != xDecl {
		t.Fatalf("unexpected x symbol: %+v, found=%t", xSymbol, ok)
	}

	ySymbol, ok := result.Symbol(yID)
	if !ok || ySymbol.Name != "y" || ySymbol.Declaration != yDecl {
		t.Fatalf("unexpected y symbol: %+v, found=%t", ySymbol, ok)
	}

	yPayload := program.Arena.BindingPayload(program.Arena.Get(yDecl).Data)
	yInitializer := program.Arena.Get(yPayload.Value)
	if yInitializer.Kind != ast.NodeBinaryExpression {
		t.Fatalf("y initializer kind = %v, want binary expression", yInitializer.Kind)
	}

	yBinary := program.Arena.BinaryPayload(yInitializer.Data)
	xRefID, ok := result.Reference(yBinary.Left)
	if !ok || xRefID != xID {
		t.Fatalf("x reference = (%d, %t), want (%d, true)", xRefID, ok, xID)
	}

	sayNode := program.Arena.Get(statements[2])
	if sayNode.Kind != ast.NodeSayStatement {
		t.Fatalf("third statement kind = %v, want say", sayNode.Kind)
	}

	sayExpression := program.Arena.Get(ast.NodeID(sayNode.Data))
	if sayExpression.Kind != ast.NodeBinaryExpression {
		t.Fatalf("say expression kind = %v, want binary", sayExpression.Kind)
	}

	sayBinary := program.Arena.BinaryPayload(sayExpression.Data)
	yRefID, ok := result.Reference(sayBinary.Left)
	if !ok || yRefID != yID {
		t.Fatalf("y reference = (%d, %t), want (%d, true)", yRefID, ok, yID)
	}
}

func TestResolveForwardReference(t *testing.T) {
	program, err := parser.Parse("say x\nlet x be 10\n")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, _ := Resolve(program)
	statements := program.Arena.StatementRange(program.Statements)

	sayNode := program.Arena.Get(statements[0])
	referenceNodeID := ast.NodeID(sayNode.Data)

	if id, ok := result.Reference(referenceNodeID); ok {
		t.Fatalf("forward reference unexpectedly resolved to symbol %d", id)
	}

	if _, ok := result.Declaration(statements[1]); !ok {
		t.Fatal("later declaration was not registered")
	}
}

func TestResolveSelfReference(t *testing.T) {
	program, err := parser.Parse("let x be x + 1\n")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}

	statements := program.Arena.StatementRange(program.Statements)
	declarationID := statements[0]

	symbolID, ok := result.Declaration(declarationID)
	if !ok {
		t.Fatal("declaration was not bound")
	}

	payload := program.Arena.BindingPayload(program.Arena.Get(declarationID).Data)
	binary := program.Arena.BinaryPayload(program.Arena.Get(payload.Value).Data)

	referenceID, ok := result.Reference(binary.Left)
	if !ok || referenceID != symbolID {
		t.Fatalf(
			"self-reference = (%d, %t), want (%d, true)",
			referenceID, ok, symbolID,
		)
	}
}

func TestResolveChangeStatement(t *testing.T) {
	program, err := parser.Parse(
		"let count be 10\nchange count to count + 5\nsay count\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	if len(statements) != 3 {
		t.Fatalf("statement count = %d, want 3", len(statements))
	}

	declarationID, ok := result.Declaration(statements[0])
	if !ok {
		t.Fatal("count declaration was not bound")
	}

	changeID, ok := result.Reference(statements[1])
	if !ok || changeID != declarationID {
		t.Fatalf(
			"assignment target = (%d, %t), want (%d, true)",
			changeID, ok, declarationID,
		)
	}

	changeNode := arena.Get(statements[1])
	if changeNode.Kind != ast.NodeChangeStatement {
		t.Fatalf("statement kind = %v, want NodeChangeStatement", changeNode.Kind)
	}

	payload := arena.BindingPayload(changeNode.Data)
	expression := arena.Get(payload.Value)

	if expression.Kind != ast.NodeBinaryExpression {
		t.Fatalf("assignment expression kind = %v, want binary", expression.Kind)
	}

	binary := arena.BinaryPayload(expression.Data)

	rhsID, ok := result.Reference(binary.Left)
	if !ok || rhsID != declarationID {
		t.Fatalf(
			"assignment RHS reference = (%d, %t), want (%d, true)",
			rhsID, ok, declarationID,
		)
	}

	sayNode := arena.Get(statements[2])
	sayID, ok := result.Reference(ast.NodeID(sayNode.Data))
	if !ok || sayID != declarationID {
		t.Fatalf(
			"say reference = (%d, %t), want (%d, true)",
			sayID, ok, declarationID,
		)
	}
}

func TestResolveUndefinedChangeTargetStillResolvesRHS(t *testing.T) {
	program, err := parser.Parse(
		"let count be 10\nchange missing to count + 5\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, _ := Resolve(program)
	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	declarationID, ok := result.Declaration(statements[0])
	if !ok {
		t.Fatal("count declaration was not bound")
	}

	if id, ok := result.Reference(statements[1]); ok {
		t.Fatalf("undefined assignment target unexpectedly bound to %d", id)
	}

	changeNode := arena.Get(statements[1])
	payload := arena.BindingPayload(changeNode.Data)
	expression := arena.Get(payload.Value)
	binary := arena.BinaryPayload(expression.Data)

	rhsID, ok := result.Reference(binary.Left)
	if !ok || rhsID != declarationID {
		t.Fatalf(
			"RHS reference = (%d, %t), want (%d, true)",
			rhsID, ok, declarationID,
		)
	}
}

func TestResolveUndefinedIdentifierDiagnostic(t *testing.T) {
	sourceText := "let count be 10\nsay missing\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %+v", len(diagnostics), diagnostics)
	}

	diag := diagnostics[0]

	if diag.Code != CodeUndefinedName {
		t.Errorf("code = %q, want %q", diag.Code, CodeUndefinedName)
	}
	if diag.Severity != diagnostic.SeverityError {
		t.Errorf("severity = %v, want error", diag.Severity)
	}
	if diag.Message != "undefined variable: missing" {
		t.Errorf("message = %q", diag.Message)
	}
	if diag.Primary.File != 0 {
		t.Errorf("file ID = %d, want 0", diag.Primary.File)
	}
	if got := program.Arena.SourceText(diag.Primary.Span); got != "missing" {
		t.Errorf("highlighted text = %q, want missing", got)
	}

	statements := program.Arena.StatementRange(program.Statements)
	sayNode := program.Arena.Get(statements[1])
	referenceID := ast.NodeID(sayNode.Data)

	if id, ok := result.Reference(referenceID); ok {
		t.Errorf("undefined identifier unexpectedly bound to %d", id)
	}
}

func TestResolveMultipleUndefinedNames(t *testing.T) {
	sourceText := "say first + second\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 2 {
		t.Fatalf("diagnostic count = %d, want 2: %+v", len(diagnostics), diagnostics)
	}

	for i, want := range []string{"first", "second"} {
		diag := diagnostics[i]

		if diag.Code != CodeUndefinedName {
			t.Errorf("diagnostic %d code = %q", i, diag.Code)
		}
		if got := program.Arena.SourceText(diag.Primary.Span); got != want {
			t.Errorf("diagnostic %d highlights %q, want %q", i, got, want)
		}
	}

	statements := program.Arena.StatementRange(program.Statements)
	sayNode := program.Arena.Get(statements[0])
	expression := program.Arena.Get(ast.NodeID(sayNode.Data))
	binary := program.Arena.BinaryPayload(expression.Data)

	for _, id := range []ast.NodeID{binary.Left, binary.Right} {
		if symbolID, ok := result.Reference(id); ok {
			t.Errorf("undefined reference bound to %d", symbolID)
		}
	}
}

func TestResolveUndefinedChangeTargetDiagnostic(t *testing.T) {
	sourceText := "let count be 10\nchange missing to count + unknown\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 2 {
		t.Fatalf("diagnostic count = %d, want 2: %+v", len(diagnostics), diagnostics)
	}

	for i, want := range []string{"missing", "unknown"} {
		diag := diagnostics[i]

		if diag.Code != CodeUndefinedName {
			t.Errorf("diagnostic %d code = %q", i, diag.Code)
		}
		if diag.Severity != diagnostic.SeverityError {
			t.Errorf("diagnostic %d severity = %v", i, diag.Severity)
		}
		if got := program.Arena.SourceText(diag.Primary.Span); got != want {
			t.Errorf("diagnostic %d highlights %q, want %q", i, got, want)
		}
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	countID, ok := result.Declaration(statements[0])
	if !ok {
		t.Fatal("count declaration was not bound")
	}

	if id, ok := result.Reference(statements[1]); ok {
		t.Errorf("undefined assignment target bound to %d", id)
	}

	changeNode := arena.Get(statements[1])
	payload := arena.BindingPayload(changeNode.Data)
	expression := arena.Get(payload.Value)
	binary := arena.BinaryPayload(expression.Data)

	rhsCountID, ok := result.Reference(binary.Left)
	if !ok || rhsCountID != countID {
		t.Errorf("RHS count reference = (%d, %t), want (%d, true)",
			rhsCountID, ok, countID)
	}

	if id, ok := result.Reference(binary.Right); ok {
		t.Errorf("undefined RHS identifier bound to %d", id)
	}
}

func TestResolveContinuesAfterUndefinedName(t *testing.T) {
	sourceText := "say missing\nlet count be 10\nsay count\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %+v", len(diagnostics), diagnostics)
	}

	statements := program.Arena.StatementRange(program.Statements)

	countID, ok := result.Declaration(statements[1])
	if !ok {
		t.Fatal("declaration following an error was not registered")
	}

	sayNode := program.Arena.Get(statements[2])
	referenceID := ast.NodeID(sayNode.Data)

	resolvedID, ok := result.Reference(referenceID)
	if !ok || resolvedID != countID {
		t.Errorf("later reference = (%d, %t), want (%d, true)",
			resolvedID, ok, countID)
	}
}

func TestResolveDuplicateDeclaration(t *testing.T) {
	sourceText := "let count be 10\nlet count be missing + 5\nsay count\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 2 {
		t.Fatalf("diagnostic count = %d, want 2: %+v", len(diagnostics), diagnostics)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	firstID, ok := result.Declaration(statements[0])
	if !ok {
		t.Fatal("first declaration was not bound")
	}

	if id, ok := result.Declaration(statements[1]); ok {
		t.Errorf("duplicate declaration unexpectedly bound to %d", id)
	}

	duplicate := diagnostics[0]

	if duplicate.Code != CodeDuplicateDeclaration {
		t.Errorf("duplicate code = %q, want %q",
			duplicate.Code, CodeDuplicateDeclaration)
	}
	if duplicate.Severity != diagnostic.SeverityError {
		t.Errorf("duplicate severity = %v, want error", duplicate.Severity)
	}
	if duplicate.Message != "duplicate declaration: count" {
		t.Errorf("duplicate message = %q", duplicate.Message)
	}

	if got := arena.SourceText(duplicate.Primary.Span); got != "count" {
		t.Errorf("duplicate highlights %q, want count", got)
	}

	if len(duplicate.Related) != 1 {
		t.Fatalf("related location count = %d, want 1", len(duplicate.Related))
	}

	original := duplicate.Related[0]

	if original.Message != "first declared here" {
		t.Errorf("related message = %q", original.Message)
	}
	if got := arena.SourceText(original.Location.Span); got != "count" {
		t.Errorf("original highlights %q, want count", got)
	}
	if original.Location.Span.Start >= duplicate.Primary.Span.Start {
		t.Error("original declaration should precede duplicate")
	}

	undefined := diagnostics[1]

	if undefined.Code != CodeUndefinedName {
		t.Errorf("second diagnostic code = %q, want %q",
			undefined.Code, CodeUndefinedName)
	}
	if got := arena.SourceText(undefined.Primary.Span); got != "missing" {
		t.Errorf("undefined diagnostic highlights %q, want missing", got)
	}

	sayNode := arena.Get(statements[2])
	sayRef := ast.NodeID(sayNode.Data)

	resolvedID, ok := result.Reference(sayRef)
	if !ok || resolvedID != firstID {
		t.Errorf("final count reference = (%d, %t), want (%d, true)",
			resolvedID, ok, firstID)
	}
}

func TestResolveDuplicateInitializerUsesOriginalSymbol(t *testing.T) {
	sourceText := "let count be 10\nlet count be count + 1\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %+v", len(diagnostics), diagnostics)
	}
	if diagnostics[0].Code != CodeDuplicateDeclaration {
		t.Fatalf("unexpected diagnostic code: %q", diagnostics[0].Code)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	originalID, ok := result.Declaration(statements[0])
	if !ok {
		t.Fatal("original declaration was not bound")
	}

	duplicateNode := arena.Get(statements[1])
	payload := arena.BindingPayload(duplicateNode.Data)
	expression := arena.Get(payload.Value)
	binary := arena.BinaryPayload(expression.Data)

	referenceID, ok := result.Reference(binary.Left)
	if !ok || referenceID != originalID {
		t.Errorf("initializer reference = (%d, %t), want (%d, true)",
			referenceID, ok, originalID)
	}
}

func TestResolveRepeatedDuplicatesPreserveFirstDeclaration(t *testing.T) {
	sourceText := "let x be 1\nlet x be 2\nlet x be 3\nsay x\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if len(diagnostics) != 2 {
		t.Fatalf("diagnostic count = %d, want 2: %+v", len(diagnostics), diagnostics)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	firstID, ok := result.Declaration(statements[0])
	if !ok {
		t.Fatal("first declaration was not bound")
	}

	firstNode := arena.Get(statements[0])
	firstPayload := arena.BindingPayload(firstNode.Data)

	for i := 1; i <= 2; i++ {
		if _, ok := result.Declaration(statements[i]); ok {
			t.Errorf("duplicate declaration %d unexpectedly bound", i)
		}

		diag := diagnostics[i-1]
		if diag.Code != CodeDuplicateDeclaration {
			t.Errorf("diagnostic %d code = %q", i-1, diag.Code)
		}
		if len(diag.Related) != 1 {
			t.Fatalf("diagnostic %d related count = %d, want 1",
				i-1, len(diag.Related))
		}
		if diag.Related[0].Location.Span != firstPayload.Name {
			t.Errorf("diagnostic %d does not reference first declaration", i-1)
		}
	}

	sayNode := arena.Get(statements[3])
	referenceID, ok := result.Reference(ast.NodeID(sayNode.Data))

	if !ok || referenceID != firstID {
		t.Errorf("final reference = (%d, %t), want (%d, true)",
			referenceID, ok, firstID)
	}
}

func TestResolveShadowedDeclaration(t *testing.T) {
	sourceText := "let value be 10\nlet value be value + 1\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	r := newResolver(arena)

	r.resolveStatement(statements[0])

	outerID, ok := r.result.Declaration(statements[0])
	if !ok {
		t.Fatal("outer declaration was not bound")
	}

	outerScope := r.currentScope
	r.currentScope = newScope(outerScope)

	r.resolveStatement(statements[1])

	if len(r.diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %+v",
			len(r.diagnostics), r.diagnostics)
	}

	diag := r.diagnostics[0]

	if diag.Code != CodeShadowedDeclaration {
		t.Errorf("code = %q, want %q", diag.Code, CodeShadowedDeclaration)
	}
	if diag.Severity != diagnostic.SeverityWarning {
		t.Errorf("severity = %v, want warning", diag.Severity)
	}
	if diag.Message != "declaration shadows outer variable: value" {
		t.Errorf("message = %q", diag.Message)
	}

	outerNode := arena.Get(statements[0])
	outerPayload := arena.BindingPayload(outerNode.Data)

	innerNode := arena.Get(statements[1])
	innerPayload := arena.BindingPayload(innerNode.Data)

	if diag.Primary.Span != innerPayload.Name {
		t.Errorf("primary span = %+v, want %+v",
			diag.Primary.Span, innerPayload.Name)
	}

	if len(diag.Related) != 1 {
		t.Fatalf("related count = %d, want 1", len(diag.Related))
	}

	if diag.Related[0].Location.Span != outerPayload.Name {
		t.Errorf("related span = %+v, want %+v",
			diag.Related[0].Location.Span, outerPayload.Name)
	}

	innerID, ok := r.result.Declaration(statements[1])
	if !ok {
		t.Fatal("shadowing declaration was not bound")
	}

	if innerID == outerID {
		t.Fatal("inner declaration reused outer symbol ID")
	}

	expression := arena.Get(innerPayload.Value)
	binary := arena.BinaryPayload(expression.Data)

	referenceID, ok := r.result.Reference(binary.Left)
	if !ok || referenceID != innerID {
		t.Errorf("initializer reference = (%d, %t), want (%d, true)",
			referenceID, ok, innerID)
	}

	if id, ok := outerScope.lookup("value"); !ok || id != outerID {
		t.Errorf("outer binding changed: (%d, %t)", id, ok)
	}
}

func TestResolveShadowingNearestAncestor(t *testing.T) {
	sourceText := "let value be 1\nlet value be 2\nlet value be 3\n"

	program, err := parser.Parse(sourceText)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	r := newResolver(arena)

	r.resolveStatement(statements[0])
	outerID, _ := r.result.Declaration(statements[0])

	r.currentScope = newScope(r.currentScope)
	r.resolveStatement(statements[1])

	middleID, ok := r.result.Declaration(statements[1])
	if !ok {
		t.Fatal("middle declaration was not bound")
	}

	r.currentScope = newScope(r.currentScope)
	r.resolveStatement(statements[2])

	innerID, ok := r.result.Declaration(statements[2])
	if !ok {
		t.Fatal("inner declaration was not bound")
	}

	if outerID == middleID || middleID == innerID || outerID == innerID {
		t.Fatal("distinct declarations share symbol IDs")
	}

	if len(r.diagnostics) != 2 {
		t.Fatalf("diagnostic count = %d, want 2: %+v",
			len(r.diagnostics), r.diagnostics)
	}

	for i, diag := range r.diagnostics {
		if diag.Code != CodeShadowedDeclaration {
			t.Errorf("diagnostic %d code = %q", i, diag.Code)
		}
	}

	middleNode := arena.Get(statements[1])
	middlePayload := arena.BindingPayload(middleNode.Data)

	innerWarning := r.diagnostics[1]

	if len(innerWarning.Related) != 1 {
		t.Fatalf("inner warning related count = %d, want 1",
			len(innerWarning.Related))
	}

	if innerWarning.Related[0].Location.Span != middlePayload.Name {
		t.Error("inner warning should reference nearest enclosing declaration")
	}
}

func TestResolveEmptyProgram(t *testing.T) {
	program, err := parser.Parse("")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	result, diagnostics := Resolve(program)

	if result == nil {
		t.Fatal("Resolve returned nil result")
	}
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}
	if len(result.symbols.symbols) != 0 {
		t.Errorf("symbol count = %d, want 0", len(result.symbols.symbols))
	}
	if len(result.bindings) != 0 {
		t.Errorf("binding count = %d, want 0", len(result.bindings))
	}
}

func TestResolveDeterministic(t *testing.T) {
	program, err := parser.Parse(
		"let x be 10\nlet y be x + 2\nsay y\nsay missing\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	first, firstDiagnostics := Resolve(program)
	second, secondDiagnostics := Resolve(program)

	if !reflect.DeepEqual(first, second) {
		t.Error("repeated resolution produced different results")
	}
	if !reflect.DeepEqual(firstDiagnostics, secondDiagnostics) {
		t.Error("repeated resolution produced different diagnostics")
	}
}

func TestResolveDoesNotMutateAST(t *testing.T) {
	program, err := parser.Parse(
		"let x be 10\nchange x to x + 1\nsay x\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	arena := program.Arena
	before := make([]ast.FlatNode, arena.Len())

	for i := range before {
		before[i] = *arena.Get(ast.NodeID(i))
	}

	statementIDs := append([]ast.NodeID(nil),
		arena.StatementRange(program.Statements)...)

	_, diagnostics := Resolve(program)
	if len(diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diagnostics)
	}

	if arena.Len() != len(before) {
		t.Fatalf("AST node count changed")
	}

	for i, original := range before {
		if current := *arena.Get(ast.NodeID(i)); current != original {
			t.Errorf("AST node %d was modified", i)
		}
	}

	if !reflect.DeepEqual(
		arena.StatementRange(program.Statements), statementIDs,
	) {
		t.Error("AST statement sequence was modified")
	}
}

func TestResolveIndependentResults(t *testing.T) {
	program, err := parser.Parse("let x be 10\nsay x\n")
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	first, firstDiagnostics := Resolve(program)
	second, secondDiagnostics := Resolve(program)

	if len(firstDiagnostics) != 0 || len(secondDiagnostics) != 0 {
		t.Fatalf("unexpected diagnostics")
	}

	statements := program.Arena.StatementRange(program.Statements)

	firstID, firstOK := first.Declaration(statements[0])
	secondID, secondOK := second.Declaration(statements[0])

	if !firstOK || !secondOK || firstID != secondID {
		t.Fatal("independent resolutions produced inconsistent IDs")
	}

	// Symbol IDs are local to each Resolution, even when their
	// numeric values happen to match.
	if &first.symbols.symbols[0] == &second.symbols.symbols[0] {
		t.Fatal("independent resolutions share symbol storage")
	}

	first.bindings[statements[0]] = InvalidSymbolID

	if _, ok := second.Declaration(statements[0]); !ok {
		t.Error("modifying first resolution affected second")
	}
}

func TestResolveIndependentScopesReuseName(t *testing.T) {
	program, err := parser.Parse(
		"let value be 1\nlet value be 2\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)
	r := newResolver(arena)

	root := r.currentScope

	r.currentScope = newScope(root)
	r.resolveStatement(statements[0])
	firstID, firstOK := r.result.Declaration(statements[0])

	r.currentScope = newScope(root)
	r.resolveStatement(statements[1])
	secondID, secondOK := r.result.Declaration(statements[1])

	if !firstOK || !secondOK {
		t.Fatal("declarations in independent scopes must both bind")
	}
	if firstID == secondID {
		t.Fatal("independent declarations must have distinct symbol IDs")
	}
	if len(r.diagnostics) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", r.diagnostics)
	}

	if _, exists := root.lookupLocal("value"); exists {
		t.Error("child declaration leaked into parent scope")
	}
}

func TestResolveScopeExitRestoresOuterBinding(t *testing.T) {
	program, err := parser.Parse(
		"let value be 1\nlet value be 2\nsay value\n",
	)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)
	r := newResolver(arena)

	r.resolveStatement(statements[0])
	outerID, ok := r.result.Declaration(statements[0])
	if !ok {
		t.Fatal("outer declaration was not bound")
	}

	outerScope := r.currentScope
	r.currentScope = newScope(outerScope)
	r.resolveStatement(statements[1])

	innerID, ok := r.result.Declaration(statements[1])
	if !ok {
		t.Fatal("inner declaration was not bound")
	}
	if innerID == outerID {
		t.Fatal("inner and outer declarations share a symbol ID")
	}

	// Simulate exiting the nested lexical scope.
	r.currentScope = outerScope
	r.resolveStatement(statements[2])

	sayNode := arena.Get(statements[2])
	refID := ast.NodeID(sayNode.Data)
	resolvedID, ok := r.result.Reference(refID)

	if !ok || resolvedID != outerID {
		t.Errorf("reference after scope exit = (%d, %t), want (%d, true)",
			resolvedID, ok, outerID)
	}

	if len(r.diagnostics) != 1 {
		t.Fatalf("diagnostic count = %d, want 1: %+v",
			len(r.diagnostics), r.diagnostics)
	}
	if r.diagnostics[0].Code != CodeShadowedDeclaration {
		t.Errorf("unexpected diagnostic: %+v", r.diagnostics[0])
	}
}

func TestNewResolverPreallocatedRootScope(t *testing.T) {
	arena := ast.NewArena("")
	r := newResolverWithCapacity(arena, 8)

	if r.currentScope == nil {
		t.Fatal("root scope was not initialized")
	}

	if r.currentScope.names == nil {
		t.Fatal("root scope map was not preallocated")
	}

	if len(r.currentScope.names) != 0 {
		t.Fatal("preallocated root scope contains declarations")
	}

	if len(r.result.symbols.symbols) != 0 {
		t.Fatal("preallocated symbol arena contains symbols")
	}

	if cap(r.result.symbols.symbols) != 8 {
		t.Fatalf("symbol capacity = %d, want 8",
			cap(r.result.symbols.symbols))
	}

	if !r.currentScope.declare("test", SymbolID(0)) {
		t.Fatal("declaration into preallocated scope failed")
	}

	if id, ok := r.currentScope.lookupLocal("test"); !ok || id != 0 {
		t.Fatalf("lookupLocal(test) = (%d, %v), want (0, true)",
			id, ok)
	}
}
