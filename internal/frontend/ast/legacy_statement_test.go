package ast

import (
	"testing"

	"github.com/manatunga/eloque/internal/source"
)

func TestStatementInterfaces(t *testing.T) {
	var _ Node = (*Program)(nil)
	var _ Statement = (*LetStatement)(nil)
	var _ Statement = (*SayStatement)(nil)
	var _ Statement = (*ChangeStatement)(nil)
}

func TestStatementSpans(t *testing.T) {
	span := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 20},
	}

	tests := []struct {
		name string
		node Node
	}{
		{
			name: "program",
			node: &Program{Location: span},
		},
		{
			name: "let statement",
			node: &LetStatement{Location: span},
		},
		{
			name: "say statement",
			node: &SayStatement{Location: span},
		},
		{
			name: "change statement",
			node: &ChangeStatement{Location: span},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.node.Span(); got != span {
				t.Errorf("Span() = %+v, want %+v", got, span)
			}
		})
	}
}

func TestProgramStatementOrder(t *testing.T) {
	first := &LetStatement{Name: "result"}
	second := &SayStatement{
		Value: &Identifier{Name: "result"},
	}

	program := &Program{
		Statements: []Statement{first, second},
	}

	if len(program.Statements) != 2 {
		t.Fatalf("statement count = %d, want 2", len(program.Statements))
	}

	if program.Statements[0] != first {
		t.Error("first statement was not preserved")
	}

	if program.Statements[1] != second {
		t.Error("second statement was not preserved")
	}
}

func TestStatementExpressionReferences(t *testing.T) {
	value := &BinaryExpression{
		Left:  &IntegerLiteral{Value: "10"},
		Right: &IntegerLiteral{Value: "20"},
	}

	letStmt := &LetStatement{
		Name:  "result",
		Value: value,
	}

	sayStmt := &SayStatement{
		Value: value,
	}

	if letStmt.Name != "result" {
		t.Errorf("declaration name = %q, want %q", letStmt.Name, "result")
	}

	if letStmt.Value != value {
		t.Error("declaration initializer was not preserved")
	}

	if sayStmt.Value != value {
		t.Error("output expression was not preserved")
	}
}

func TestChangeStatement(t *testing.T) {
	span := source.Span{
		Start: source.Position{Line: 2, Column: 1},
		End:   source.Position{Line: 2, Column: 19},
	}

	value := &IntegerLiteral{
		Value: "20",
		Location: source.Span{
			Start: source.Position{Line: 2, Column: 17},
			End:   source.Position{Line: 2, Column: 19},
		},
	}

	stmt := &ChangeStatement{
		Name:     "score",
		Value:    value,
		Location: span,
	}

	if got := stmt.Span(); got != span {
		t.Errorf("Span() = %+v, want %+v", got, span)
	}

	if stmt.Name != "score" {
		t.Errorf("Name = %q, want %q", stmt.Name, "score")
	}

	if stmt.Value != value {
		t.Error("assignment expression was not preserved")
	}
}
