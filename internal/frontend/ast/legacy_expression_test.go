package ast

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestExpressionInterfaces(t *testing.T) {
	var _ Expression = (*IntegerLiteral)(nil)
	var _ Expression = (*StringLiteral)(nil)
	var _ Expression = (*Identifier)(nil)
	var _ Expression = (*UnaryExpression)(nil)
	var _ Expression = (*BinaryExpression)(nil)
}

func TestExpressionSpans(t *testing.T) {
	span := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 8},
	}

	tests := []struct {
		name string
		expr Expression
	}{
		{"integer", &IntegerLiteral{Value: "42", Location: span}},
		{"string", &StringLiteral{Value: "hello", Location: span}},
		{"identifier", &Identifier{Name: "result", Location: span}},
		{"unary", &UnaryExpression{
			Operator: lexer.TokenMinus,
			Operand:  &IntegerLiteral{Value: "42"},
			Location: span,
		}},
		{"binary", &BinaryExpression{
			Left:     &IntegerLiteral{Value: "10"},
			Operator: lexer.TokenPlus,
			Right:    &IntegerLiteral{Value: "20"},
			Location: span,
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.expr.Span(); got != span {
				t.Errorf("Span() = %+v, want %+v", got, span)
			}
		})
	}
}

func TestBinaryExpressionChildren(t *testing.T) {
	left := &IntegerLiteral{Value: "10"}
	right := &Identifier{Name: "value"}

	expr := &BinaryExpression{
		Left:     left,
		Operator: lexer.TokenPlus,
		Right:    right,
	}

	if expr.Left != left {
		t.Error("left operand was not preserved")
	}

	if expr.Right != right {
		t.Error("right operand was not preserved")
	}

	if expr.Operator != lexer.TokenPlus {
		t.Errorf("operator = %v, want PLUS", expr.Operator)
	}
}

func TestUnaryExpressionOperand(t *testing.T) {
	operand := &IntegerLiteral{Value: "42"}

	expr := &UnaryExpression{
		Operator: lexer.TokenMinus,
		Operand:  operand,
	}

	if expr.Operand != operand {
		t.Error("unary operand was not preserved")
	}

	if expr.Operator != lexer.TokenMinus {
		t.Errorf("operator = %v, want MINUS", expr.Operator)
	}
}
