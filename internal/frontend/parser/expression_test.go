package parser

import (
	"slices"
	"strings"
	"testing"

	diagnosticpkg "github.com/manatunga/eloque/internal/diagnostic"
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
		name         string
		source       string
		wantMessage  string
		wantStart    uint32
		wantEnd      uint32
		wantExpected []string
		wantActual   string
	}{
		{
			name:        "missing right operand",
			source:      "2 +",
			wantMessage: "expected an expression",
			wantStart:   3,
			wantEnd:     3,
			wantExpected: []string{
				lexer.TokenInt.String(),
				lexer.TokenString.String(),
				lexer.TokenIdent.String(),
				lexer.TokenLeftParen.String(),
				lexer.TokenPlus.String(),
				lexer.TokenMinus.String(),
			},
			wantActual: lexer.TokenEOF.String(),
		},
		{
			name:        "empty grouping",
			source:      "()",
			wantMessage: "expected an expression",
			wantStart:   1,
			wantEnd:     2,
			wantExpected: []string{
				lexer.TokenInt.String(),
				lexer.TokenString.String(),
				lexer.TokenIdent.String(),
				lexer.TokenLeftParen.String(),
				lexer.TokenPlus.String(),
				lexer.TokenMinus.String(),
			},
			wantActual: lexer.TokenRightParen.String(),
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
			if diagnostic.Code != CodeUnexpectedToken {
				t.Errorf("code = %q, want %q",
					diagnostic.Code, CodeUnexpectedToken)
			}

			if diagnostic.Severity != diagnosticpkg.SeverityError {
				t.Errorf("severity = %v, want error",
					diagnostic.Severity)
			}

			if diagnostic.Primary.Span.Start != tt.wantStart ||
				diagnostic.Primary.Span.End != tt.wantEnd {
				t.Errorf("span = %+v, want [%d,%d)",
					diagnostic.Primary.Span, tt.wantStart, tt.wantEnd)
			}

			if !slices.Equal(
				diagnostic.TokenExpectation.Expected,
				tt.wantExpected,
			) {
				t.Errorf(
					"expected tokens = %v, want %v",
					diagnostic.TokenExpectation.Expected,
					tt.wantExpected,
				)
			}

			if diagnostic.TokenExpectation.Actual != tt.wantActual {
				t.Errorf("actual token = %q, want %q",
					diagnostic.TokenExpectation.Actual, tt.wantActual)
			}

			if len(diagnostic.TokenExpectation.Expected) == 0 {
				t.Error("expected token set is empty")
			}
		})
	}
}

func TestFlatExpressionBoundaries(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		wantKind      ast.NodeKind
		wantRemaining lexer.TokenKind
	}{
		{
			name:          "newline boundary",
			source:        "42\n",
			wantKind:      ast.NodeIntegerLiteral,
			wantRemaining: lexer.TokenNewline,
		},
		{
			name:          "EOF boundary",
			source:        "42",
			wantKind:      ast.NodeIntegerLiteral,
			wantRemaining: lexer.TokenEOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FlatParser{
				Parser: newStreamingParser(tt.source),
				arena:  ast.NewArena(tt.source),
			}

			result, diag := p.parseFlatExpression(0)
			if diag != nil {
				t.Fatalf("unexpected diagnostic: %+v", diag)
			}

			node := p.arena.Get(result.id)
			if node.Kind != tt.wantKind {
				t.Errorf(
					"root kind = %v, want %v",
					node.Kind,
					tt.wantKind,
				)
			}

			if got := p.peek().Kind; got != tt.wantRemaining {
				t.Errorf(
					"remaining token = %v, want %v",
					got,
					tt.wantRemaining,
				)
			}
		})
	}
}

func TestFlatExpressionMinimumBindingPower(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		minPower      int
		wantKind      ast.NodeKind
		wantRemaining lexer.TokenKind
	}{
		{
			name:          "addition below minimum",
			source:        "2 + 3",
			minPower:      bindingPowerAdditive + 1,
			wantKind:      ast.NodeIntegerLiteral,
			wantRemaining: lexer.TokenPlus,
		},
		{
			name:          "multiplication below minimum",
			source:        "2 * 3",
			minPower:      bindingPowerMultiplicative + 1,
			wantKind:      ast.NodeIntegerLiteral,
			wantRemaining: lexer.TokenStar,
		},
		{
			name:          "exponentiation below minimum",
			source:        "2 ** 3",
			minPower:      bindingPowerExponentiation + 1,
			wantKind:      ast.NodeIntegerLiteral,
			wantRemaining: lexer.TokenPower,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &FlatParser{
				Parser: newStreamingParser(tt.source),
				arena:  ast.NewArena(tt.source),
			}

			result, diag := p.parseFlatExpression(tt.minPower)
			if diag != nil {
				t.Fatalf("unexpected diagnostic: %+v", diag)
			}

			node := p.arena.Get(result.id)
			if node.Kind != tt.wantKind {
				t.Errorf("root kind = %v, want %v", node.Kind, tt.wantKind)
			}

			if got := p.peek().Kind; got != tt.wantRemaining {
				t.Errorf("remaining token = %v, want %v", got, tt.wantRemaining)
			}
		})
	}
}

func TestFlatExpressionNestedUnary(t *testing.T) {
	sourceText := "--42"

	p := &FlatParser{
		Parser: newStreamingParser(sourceText),
		arena:  ast.NewArena(sourceText),
	}

	result, diag := p.parseFlatExpression(0)
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	outer := p.arena.Get(result.id)
	if outer.Kind != ast.NodeUnaryExpression {
		t.Fatalf("outer kind = %v, want unary", outer.Kind)
	}

	outerPayload := p.arena.UnaryPayload(outer.Data)
	if outerPayload.Operator != lexer.TokenMinus {
		t.Errorf("outer operator = %v, want minus", outerPayload.Operator)
	}

	inner := p.arena.Get(outerPayload.Operand)
	if inner.Kind != ast.NodeUnaryExpression {
		t.Fatalf("inner kind = %v, want unary", inner.Kind)
	}

	innerPayload := p.arena.UnaryPayload(inner.Data)
	if innerPayload.Operator != lexer.TokenMinus {
		t.Errorf("inner operator = %v, want minus", innerPayload.Operator)
	}

	operand := p.arena.Get(innerPayload.Operand)
	if operand.Kind != ast.NodeIntegerLiteral {
		t.Errorf("operand kind = %v, want integer", operand.Kind)
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatExpressionStringEscapeCases(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{"empty string", `""`, ""},
		{"single quoted", `'hello'`, "hello"},
		{"newline", `"a\nb"`, "a\nb"},
		{"tab", `"a\tb"`, "a\tb"},
		{"carriage return", `"a\rb"`, "a\rb"},
		{"backslash", `"a\\b"`, "a\\b"},
		{"double quote", `"a\"b"`, "a\"b"},
		{"single quote", `'a\'b'`, "a'b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewFlatParser(tt.source)

			result, diag := p.parseFlatExpression(0)
			if diag != nil {
				t.Fatalf("unexpected diagnostic: %+v", diag)
			}

			node := p.arena.Get(result.id)
			if node.Kind != ast.NodeStringLiteral {
				t.Fatalf("node kind = %v, want string literal", node.Kind)
			}

			if got := p.arena.StringValue(result.id); got != tt.want {
				t.Errorf("decoded value = %q, want %q", got, tt.want)
			}

			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("remaining token = %v, want EOF", got)
			}
		})
	}
}

func TestFlatExpressionFloorDivisionModuloAssociativity(t *testing.T) {
	p := NewFlatParser("20 // 3 % 2")

	result, diag := p.parseFlatExpression(0)
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	root := p.arena.Get(result.id)
	if root.Kind != ast.NodeBinaryExpression {
		t.Fatalf("root kind = %v, want binary", root.Kind)
	}

	modulo := p.arena.BinaryPayload(root.Data)
	if modulo.Operator != lexer.TokenPercent {
		t.Errorf("root operator = %v, want %%", modulo.Operator)
	}

	left := p.arena.Get(modulo.Left)
	if left.Kind != ast.NodeBinaryExpression {
		t.Fatalf("left kind = %v, want binary", left.Kind)
	}

	division := p.arena.BinaryPayload(left.Data)
	if division.Operator != lexer.TokenFloorDiv {
		t.Errorf("left operator = %v, want //", division.Operator)
	}
}

func TestFlatExpressionLongExponentiationChain(t *testing.T) {
	p := NewFlatParser("2 ** 3 ** 4 ** 5 ** 6")

	result, diag := p.parseFlatExpression(0)
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	current := result.id

	for i := 0; i < 4; i++ {
		node := p.arena.Get(current)
		if node.Kind != ast.NodeBinaryExpression {
			t.Fatalf("chain level %d: kind = %v, want binary", i, node.Kind)
		}

		payload := p.arena.BinaryPayload(node.Data)
		if payload.Operator != lexer.TokenPower {
			t.Fatalf("chain level %d: operator = %v, want power", i, payload.Operator)
		}

		left := p.arena.Get(payload.Left)
		if left.Kind != ast.NodeIntegerLiteral {
			t.Errorf("chain level %d: left kind = %v, want integer", i, left.Kind)
		}

		current = payload.Right
	}

	last := p.arena.Get(current)
	if last.Kind != ast.NodeIntegerLiteral {
		t.Errorf("final operand kind = %v, want integer", last.Kind)
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatExpressionDeeplyNestedParentheses(t *testing.T) {
	const depth = 100

	sourceText := strings.Repeat("(", depth) +
		"42" +
		strings.Repeat(")", depth)

	p := NewFlatParser(sourceText)

	result, diag := p.parseFlatExpression(0)
	if diag != nil {
		t.Fatalf("unexpected diagnostic: %+v", diag)
	}

	node := p.arena.Get(result.id)
	if node.Kind != ast.NodeIntegerLiteral {
		t.Fatalf("node kind = %v, want integer literal", node.Kind)
	}

	if node.Span.Start != uint32(depth) ||
		node.Span.End != uint32(depth+2) {
		t.Errorf(
			"literal span = %+v, want [%d,%d)",
			node.Span,
			depth,
			depth+2,
		)
	}

	if result.extent.Start != 0 ||
		result.extent.End != uint32(len(sourceText)) {
		t.Errorf(
			"expression extent = %+v, want [0,%d)",
			result.extent,
			len(sourceText),
		)
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}
