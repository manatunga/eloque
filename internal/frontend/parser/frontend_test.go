package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestParseLexicalError(t *testing.T) {
	program, diag := Parse("say @")

	if program != nil {
		t.Fatal("expected nil program on lexical error")
	}
	if diag == nil {
		t.Fatal("expected lexical diagnostic")
	}
	if diag.Code != lexer.CodeUnexpectedCharacter {
		t.Errorf("code = %q, want %q",
			diag.Code, lexer.CodeUnexpectedCharacter)
	}
	if diag.Severity != diagnostic.SeverityError {
		t.Errorf("severity = %v, want error", diag.Severity)
	}
	if diag.Primary.Span != (source.ByteSpan{Start: 4, End: 5}) {
		t.Errorf("span = %+v, want [4,5)", diag.Primary.Span)
	}
}

func TestParseSyntaxError(t *testing.T) {
	program, diag := Parse("let age be")

	if program != nil {
		t.Fatal("expected nil program on syntax error")
	}
	if diag == nil {
		t.Fatal("expected syntax diagnostic")
	}
	if diag.Code != CodeUnexpectedToken {
		t.Errorf("code = %q, want %q",
			diag.Code, CodeUnexpectedToken)
	}
	if diag.Message != "expected an expression" {
		t.Errorf("message = %q, want expected expression", diag.Message)
	}
}

func TestParseEmptyProgram(t *testing.T) {
	program, diag := Parse("")
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}
	if program == nil {
		t.Fatal("expected non-nil program")
	}
	if program.Arena == nil {
		t.Fatal("expected non-nil arena")
	}
	if got := len(program.Arena.StatementRange(program.Statements)); got != 0 {
		t.Fatalf("expected zero statements, got %d", got)
	}
	if program.Span != (source.ByteSpan{Start: 0, End: 0}) {
		t.Errorf("empty program span = %+v, want [0,0)", program.Span)
	}
}

func TestParseTrailingNewline(t *testing.T) {
	program, diag := Parse("let age be 20\n")
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}
	if program == nil {
		t.Fatal("expected non-nil program")
	}
	if got := len(program.Arena.StatementRange(program.Statements)); got != 1 {
		t.Fatalf("expected one statement, got %d", got)
	}
}

func TestParseStatementOrder(t *testing.T) {
	const sourceText = "let age be 20\nsay age\nchange age to 21\n"

	program, diag := Parse(sourceText)
	if diag != nil {
		t.Fatalf("Parse() diagnostic = %+v", diag)
	}
	if program == nil || program.Arena == nil {
		t.Fatal("expected program with non-nil arena")
	}

	ids := program.Arena.StatementRange(program.Statements)
	wantKinds := []ast.NodeKind{
		ast.NodeLetStatement,
		ast.NodeSayStatement,
		ast.NodeChangeStatement,
	}

	if len(ids) != len(wantKinds) {
		t.Fatalf("statement count = %d, want %d", len(ids), len(wantKinds))
	}
	for i, id := range ids {
		if got := program.Arena.Get(id).Kind; got != wantKinds[i] {
			t.Errorf("statement %d kind = %v, want %v",
				i, got, wantKinds[i])
		}
	}
}

func TestParseFirstEncounteredError(t *testing.T) {
	_, diag := Parse("let age be\nsay @")

	if diag == nil {
		t.Fatal("expected diagnostic")
	}
	if diag.Code != CodeUnexpectedToken {
		t.Errorf("code = %q, want earlier syntax error %q",
			diag.Code, CodeUnexpectedToken)
	}
	if diag.Primary.Span != (source.ByteSpan{Start: 10, End: 11}) {
		t.Errorf("span = %+v, want [10,11)", diag.Primary.Span)
	}
}

func TestParseConversionExpressionsAcrossStatements(t *testing.T) {
	const sourceText = "" +
		"let narrow be convert value to Int16\n" +
		"change narrow to convert (narrow + 1) to Int32\n" +
		"say convert narrow to Float\n"

	program, diag := Parse(sourceText)
	if diag != nil {
		t.Fatalf("Parse() diagnostic = %+v", diag)
	}
	if program == nil || program.Arena == nil {
		t.Fatal("expected program with non-nil arena")
	}

	arena := program.Arena
	statements := arena.StatementRange(program.Statements)

	if len(statements) != 3 {
		t.Fatalf("statement count = %d, want 3", len(statements))
	}

	wantKinds := []ast.NodeKind{
		ast.NodeLetStatement,
		ast.NodeChangeStatement,
		ast.NodeSayStatement,
	}

	for i, id := range statements {
		if got := arena.Get(id).Kind; got != wantKinds[i] {
			t.Errorf("statement %d kind = %v, want %v",
				i, got, wantKinds[i])
		}
	}

	assertConversion := func(
		name string,
		id ast.NodeID,
		wantTarget string,
		wantOperandKind ast.NodeKind,
	) {
		t.Helper()

		node := arena.Get(id)
		if node.Kind != ast.NodeConversionExpression {
			t.Fatalf("%s kind = %v, want conversion expression",
				name, node.Kind)
		}

		payload := arena.ConversionPayload(node.Data)

		if got := arena.SourceText(payload.TargetType); got != wantTarget {
			t.Errorf("%s target = %q, want %q",
				name, got, wantTarget)
		}

		if got := arena.Get(payload.Operand).Kind; got != wantOperandKind {
			t.Errorf("%s operand kind = %v, want %v",
				name, got, wantOperandKind)
		}
	}

	// let narrow be convert value to Int16
	letPayload := arena.BindingPayload(arena.Get(statements[0]).Data)
	assertConversion(
		"let initializer",
		letPayload.Value,
		"Int16",
		ast.NodeIdentifier,
	)

	// change narrow to convert (narrow + 1) to Int32
	changePayload := arena.BindingPayload(arena.Get(statements[1]).Data)
	assertConversion(
		"change value",
		changePayload.Value,
		"Int32",
		ast.NodeBinaryExpression,
	)

	// say convert narrow to Float
	sayValue := ast.NodeID(arena.Get(statements[2]).Data)
	assertConversion(
		"say expression",
		sayValue,
		"Float",
		ast.NodeIdentifier,
	)
}

func TestParseRecoveringMalformedConversionRollsBack(t *testing.T) {
	result := ParseRecovering(
		"say convert value to\nsay 7\n",
		ParseOptions{},
	)

	if result.Status != ParseRecovered {
		t.Fatalf("status = %v, want ParseRecovered; diagnostics: %v",
			result.Status, result.Diagnostics)
	}

	if len(result.Diagnostics) != 1 {
		t.Fatalf("diagnostics = %d, want 1: %v",
			len(result.Diagnostics), result.Diagnostics)
	}

	if result.Program == nil || result.Program.Arena == nil {
		t.Fatal("expected recovered program with non-nil arena")
	}

	arena := result.Program.Arena

	if got := arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}

	statements := arena.StatementRange(result.Program.Statements)
	statement := arena.Get(statements[0])

	if statement.Kind != ast.NodeSayStatement {
		t.Fatalf("recovered statement kind = %v, want SAY",
			statement.Kind)
	}

	value := arena.Get(ast.NodeID(statement.Data))
	if value.Kind != ast.NodeIntegerLiteral {
		t.Fatalf("recovered expression kind = %v, want integer literal",
			value.Kind)
	}

	// Only the integer literal and its SAY statement should remain.
	if got := arena.Len(); got != 2 {
		t.Fatalf("arena node count = %d, want 2", got)
	}
}
