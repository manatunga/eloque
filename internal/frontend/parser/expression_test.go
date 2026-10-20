package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
)

func TestFlatExpressionPrecedence(t *testing.T) {
	sourceText := "2 + 3 * 4"

	parser := &FlatParser{
		Parser: newStreamingParser(sourceText),
		arena:  ast.NewArena(sourceText),
	}

	result, diagnostic := parser.parseFlatExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected parser diagnostic: %v", diagnostic)
	}

	root := parser.arena.Get(result.id)

	if root.Kind != ast.NodeBinaryExpression {
		t.Fatalf("root kind = %v, want binary expression", root.Kind)
	}

	addition := parser.arena.BinaryPayload(root.Data)

	if addition.Operator != lexer.TokenPlus {
		t.Fatalf("root operator = %v, want +", addition.Operator)
	}

	right := parser.arena.Get(addition.Right)

	if right.Kind != ast.NodeBinaryExpression {
		t.Fatalf("right kind = %v, want binary expression", right.Kind)
	}

	multiplication := parser.arena.BinaryPayload(right.Data)

	if multiplication.Operator != lexer.TokenStar {
		t.Errorf("right operator = %v, want *", multiplication.Operator)
	}

	if result.extent.Start != 0 || result.extent.End != uint32(len(sourceText)) {
		t.Errorf("expression extent = %+v, want [0, %d)", result.extent, len(sourceText))
	}
}

func TestFlatExpressionAssociativity(t *testing.T) {
	tests := []struct {
		name       string
		source     string
		operator   lexer.TokenKind
		nestedLeft bool
	}{
		{
			name:       "exponentiation is right associative",
			source:     "2 ** 3 ** 2",
			operator:   lexer.TokenPower,
			nestedLeft: false,
		},
		{
			name:       "subtraction is left associative",
			source:     "10 - 5 - 2",
			operator:   lexer.TokenMinus,
			nestedLeft: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FlatParser{
				Parser: newStreamingParser(tt.source),
				arena:  ast.NewArena(tt.source),
			}

			result, diagnostic := p.parseFlatExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			root := p.arena.Get(result.id)
			if root.Kind != ast.NodeBinaryExpression {
				t.Fatalf("root kind = %v, want binary", root.Kind)
			}

			payload := p.arena.BinaryPayload(root.Data)
			if payload.Operator != tt.operator {
				t.Errorf("root operator = %v, want %v",
					payload.Operator, tt.operator)
			}

			nestedID := payload.Right
			if tt.nestedLeft {
				nestedID = payload.Left
			}

			nested := p.arena.Get(nestedID)
			if nested.Kind != ast.NodeBinaryExpression {
				t.Fatalf("nested kind = %v, want binary", nested.Kind)
			}

			nestedPayload := p.arena.BinaryPayload(nested.Data)
			if nestedPayload.Operator != tt.operator {
				t.Errorf("nested operator = %v, want %v",
					nestedPayload.Operator, tt.operator)
			}

			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("remaining token = %v, want EOF", got)
			}
		})
	}
}

func TestFlatExpressionUnaryPrecedence(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		rootKind      ast.NodeKind
		binaryOnRight bool
	}{
		{
			name:     "exponentiation binds tighter than unary",
			source:   "-2 ** 2",
			rootKind: ast.NodeUnaryExpression,
		},
		{
			name:          "negative exponent",
			source:        "2 ** -3",
			rootKind:      ast.NodeBinaryExpression,
			binaryOnRight: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FlatParser{
				Parser: newStreamingParser(tt.source),
				arena:  ast.NewArena(tt.source),
			}

			result, diagnostic := p.parseFlatExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			root := p.arena.Get(result.id)
			if root.Kind != tt.rootKind {
				t.Fatalf("root kind = %v, want %v",
					root.Kind, tt.rootKind)
			}

			if tt.binaryOnRight {
				binary := p.arena.BinaryPayload(root.Data)

				if binary.Operator != lexer.TokenPower {
					t.Errorf("operator = %v, want power",
						binary.Operator)
				}

				right := p.arena.Get(binary.Right)
				if right.Kind != ast.NodeUnaryExpression {
					t.Fatalf("right kind = %v, want unary", right.Kind)
				}

				unary := p.arena.UnaryPayload(right.Data)
				if unary.Operator != lexer.TokenMinus {
					t.Errorf("unary operator = %v, want minus",
						unary.Operator)
				}
			} else {
				unary := p.arena.UnaryPayload(root.Data)
				if unary.Operator != lexer.TokenMinus {
					t.Errorf("unary operator = %v, want minus",
						unary.Operator)
				}

				operand := p.arena.Get(unary.Operand)
				if operand.Kind != ast.NodeBinaryExpression {
					t.Fatalf("operand kind = %v, want binary",
						operand.Kind)
				}

				binary := p.arena.BinaryPayload(operand.Data)
				if binary.Operator != lexer.TokenPower {
					t.Errorf("operand operator = %v, want power",
						binary.Operator)
				}
			}

			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("remaining token = %v, want EOF", got)
			}
		})
	}
}

func TestFlatExpressionGrouping(t *testing.T) {
	sourceText := "((42))"

	p := &FlatParser{
		Parser: newStreamingParser(sourceText),
		arena:  ast.NewArena(sourceText),
	}

	result, diagnostic := p.parseFlatExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	node := p.arena.Get(result.id)
	if node.Kind != ast.NodeIntegerLiteral {
		t.Fatalf("node kind = %v, want integer", node.Kind)
	}

	if p.arena.Len() != 1 {
		t.Errorf("arena length = %d, want 1", p.arena.Len())
	}

	if node.Span.Start != 2 || node.Span.End != 4 {
		t.Errorf("node span = %+v, want [2,4)", node.Span)
	}

	if result.extent.Start != 0 ||
		result.extent.End != uint32(len(sourceText)) {
		t.Errorf("extent = %+v, want entire source", result.extent)
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatExpressionStringDecoding(t *testing.T) {
	sourceText := `"hello\nworld"`

	p := &FlatParser{
		Parser: newStreamingParser(sourceText),
		arena:  ast.NewArena(sourceText),
	}

	result, diagnostic := p.parseFlatExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	node := p.arena.Get(result.id)
	if node.Kind != ast.NodeStringLiteral {
		t.Fatalf("node kind = %v, want string", node.Kind)
	}

	if got := p.arena.StringValue(result.id); got != "hello\nworld" {
		t.Errorf("decoded value = %q, want %q",
			got, "hello\nworld")
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatExpressionInvalidSyntax(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantMessage string
	}{
		{
			name:        "missing right operand",
			source:      "2 +",
			wantMessage: "expected an expression",
		},
		{
			name:        "empty grouping",
			source:      "()",
			wantMessage: "expected an expression",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FlatParser{
				Parser: newStreamingParser(tt.source),
				arena:  ast.NewArena(tt.source),
			}

			_, diagnostic := p.parseFlatExpression(0)
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if diagnostic.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q",
					diagnostic.Message, tt.wantMessage)
			}
		})
	}
}
