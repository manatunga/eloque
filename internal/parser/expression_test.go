package parser

import (
	"fmt"
	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
	"reflect"
	"testing"
)

func TestParsePrimaryInteger(t *testing.T) {
	span := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 3},
	}
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenInt, Lexeme: "42", Span: span},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parsePrimary()
	if diagnostic != nil {
		t.Fatalf("parsePrimary() returned unexpected diagnostic: %+v", diagnostic)
	}
	expression, ok := result.expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.IntegerLiteral", result.expression)
	}
	if expression.Value != "42" {
		t.Errorf("Value = %q, want %q", expression.Value, "42")
	}
	if expression.Span() != span {
		t.Errorf("Span() = %+v, want %+v", expression.Span(), span)
	}
	if result.extent != span {
		t.Errorf("extent = %+v, want %+v", result.extent, span)
	}
	if p.current != 1 {
		t.Errorf("current = %d, want 1", p.current)
	}
}
func TestParsePrimaryIdentifier(t *testing.T) {
	span := source.Span{
		Start: source.Position{Line: 2, Column: 5},
		End:   source.Position{Line: 2, Column: 10},
	}
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenIdent, Lexeme: "score", Span: span},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parsePrimary()
	if diagnostic != nil {
		t.Fatalf("parsePrimary() returned unexpected diagnostic: %+v", diagnostic)
	}
	expression, ok := result.expression.(*ast.Identifier)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.Identifier", result.expression)
	}
	if expression.Name != "score" {
		t.Errorf("Name = %q, want %q", expression.Name, "score")
	}
	if expression.Span() != span {
		t.Errorf("Span() = %+v, want %+v", expression.Span(), span)
	}
	if result.extent != span {
		t.Errorf("extent = %+v, want %+v", result.extent, span)
	}
	if p.current != 1 {
		t.Errorf("current = %d, want 1", p.current)
	}
}
func TestParsePrimaryString(t *testing.T) {
	tests := []struct {
		name   string
		lexeme string
		want   string
	}{
		{
			name:   "double quoted",
			lexeme: `"hello"`,
			want:   "hello",
		},
		{
			name:   "single quoted",
			lexeme: `'hello'`,
			want:   "hello",
		},
		{
			name:   "newline escape",
			lexeme: `"hello\nworld"`,
			want:   "hello\nworld",
		},
		{
			name:   "tab escape",
			lexeme: `"hello\tworld"`,
			want:   "hello\tworld",
		},
		{
			name:   "carriage return escape",
			lexeme: `"hello\rworld"`,
			want:   "hello\rworld",
		},
		{
			name:   "backslash escape",
			lexeme: `"C:\\Users"`,
			want:   `C:\Users`,
		},
		{
			name:   "double quote escape",
			lexeme: `"say \"hello\""`,
			want:   `say "hello"`,
		},
		{
			name:   "single quote escape",
			lexeme: `'it\'s fine'`,
			want:   "it's fine",
		},
		{
			name:   "empty string",
			lexeme: `""`,
			want:   "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End: source.Position{
					Line:   1,
					Column: len(tt.lexeme) + 1,
				},
			}
			p := newLegacyExpressionParser([]legacyExpressionToken{
				{
					Kind:   lexer.TokenString,
					Lexeme: tt.lexeme,
					Span:   span,
				},
				{Kind: lexer.TokenEOF},
			})
			result, diagnostic := p.parsePrimary()
			if diagnostic != nil {
				t.Fatalf("parsePrimary() returned unexpected diagnostic: %+v", diagnostic)
			}
			expression, ok := result.expression.(*ast.StringLiteral)
			if !ok {
				t.Fatalf("expression type = %T, want *ast.StringLiteral", result.expression)
			}
			if expression.Value != tt.want {
				t.Errorf("Value = %q, want %q", expression.Value, tt.want)
			}
			if expression.Span() != span {
				t.Errorf("Span() = %+v, want %+v", expression.Span(), span)
			}
			if result.extent != span {
				t.Errorf("extent = %+v, want %+v", result.extent, span)
			}
			if p.current != 1 {
				t.Errorf("current = %d, want 1", p.current)
			}
		})
	}
}
func TestParsePrimarySpans(t *testing.T) {
	tests := []struct {
		name  string
		token legacyExpressionToken
	}{
		{
			name: "integer",
			token: legacyExpressionToken{
				Kind:   lexer.TokenInt,
				Lexeme: "123",
			},
		},
		{
			name: "string",
			token: legacyExpressionToken{
				Kind:   lexer.TokenString,
				Lexeme: `"abc"`,
			},
		},
		{
			name: "identifier",
			token: legacyExpressionToken{
				Kind:   lexer.TokenIdent,
				Lexeme: "value",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := source.Span{
				Start: source.Position{Line: 3, Column: 7},
				End: source.Position{
					Line:   3,
					Column: 7 + len(tt.token.Lexeme),
				},
			}
			tt.token.Span = span
			p := newLegacyExpressionParser([]legacyExpressionToken{
				tt.token,
				{Kind: lexer.TokenEOF},
			})
			result, diagnostic := p.parsePrimary()
			if diagnostic != nil {
				t.Fatalf("parsePrimary() returned unexpected diagnostic: %+v", diagnostic)
			}
			if result.expression.Span() != span {
				t.Errorf("AST span = %+v, want %+v", result.expression.Span(), span)
			}
			if result.extent != span {
				t.Errorf("extent = %+v, want %+v", result.extent, span)
			}
		})
	}
}
func TestParsePrimaryUnexpectedToken(t *testing.T) {
	tests := []struct {
		name string
		kind lexer.TokenKind
	}{
		{name: "right parenthesis", kind: lexer.TokenRightParen},
		{name: "newline", kind: lexer.TokenNewline},
		{name: "EOF", kind: lexer.TokenEOF},
		{name: "keyword", kind: lexer.TokenLet},
	}
	expected := []lexer.TokenKind{
		lexer.TokenInt,
		lexer.TokenString,
		lexer.TokenIdent,
		lexer.TokenLeftParen,
		lexer.TokenPlus,
		lexer.TokenMinus,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := source.Span{
				Start: source.Position{Line: 4, Column: 6},
				End:   source.Position{Line: 4, Column: 6},
			}
			tokens := []legacyExpressionToken{
				{Kind: tt.kind, Span: span},
			}
			if tt.kind != lexer.TokenEOF {
				tokens = append(tokens, legacyExpressionToken{Kind: lexer.TokenEOF})
			}
			p := newLegacyExpressionParser(tokens)
			result, diagnostic := p.parsePrimary()
			if diagnostic == nil {
				t.Fatal("parsePrimary() returned no diagnostic for invalid token")
			}
			if result.expression != nil {
				t.Errorf("expression = %T, want nil", result.expression)
			}
			if diagnostic.Message != "expected an expression" {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, "expected an expression")
			}
			if diagnostic.Location != span {
				t.Errorf("Location = %+v, want %+v", diagnostic.Location, span)
			}
			if !reflect.DeepEqual(diagnostic.Expected, expected) {
				t.Errorf("Expected = %v, want %v", diagnostic.Expected, expected)
			}
			if diagnostic.Actual == nil {
				t.Fatal("Actual = nil, want encountered token kind")
			}
			if *diagnostic.Actual != tt.kind {
				t.Errorf("Actual = %v, want %v", *diagnostic.Actual, tt.kind)
			}
			if p.current != 0 {
				t.Errorf("current = %d, want 0", p.current)
			}
		})
	}
}
func TestParsePrimaryGrouping(t *testing.T) {
	tests := []struct {
		name       string
		lexemes    []string
		innerStart int
		innerEnd   int
		outerEnd   int
	}{
		{
			name:       "single grouping",
			lexemes:    []string{"(", "42", ")"},
			innerStart: 2,
			innerEnd:   4,
			outerEnd:   5,
		},
		{
			name:       "nested grouping",
			lexemes:    []string{"(", "(", "42", ")", ")"},
			innerStart: 3,
			innerEnd:   5,
			outerEnd:   7,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := make([]legacyExpressionToken, 0, len(tt.lexemes)+1)
			column := 1
			for _, lexeme := range tt.lexemes {
				var kind lexer.TokenKind

				switch lexeme {
				case "(":
					kind = lexer.TokenLeftParen
				case ")":
					kind = lexer.TokenRightParen
				case "42":
					kind = lexer.TokenInt
				default:
					t.Fatalf("unexpected test lexeme: %q", lexeme)
				}

				tokens = append(tokens, legacyExpressionToken{
					Kind:   kind,
					Lexeme: lexeme,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: column},
						End:   source.Position{Line: 1, Column: column + len(lexeme)},
					},
				})
				column += len(lexeme)
			}
			tokens = append(tokens, legacyExpressionToken{
				Kind: lexer.TokenEOF,
				Span: source.Span{
					Start: source.Position{Line: 1, Column: column},
					End:   source.Position{Line: 1, Column: column},
				},
			})
			p := newLegacyExpressionParser(tokens)
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			integer, ok := result.expression.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("expression type = %T, want *ast.IntegerLiteral", result.expression)
			}
			if integer.Value != "42" {
				t.Errorf("Value = %q, want %q", integer.Value, "42")
			}
			wantInner := source.Span{
				Start: source.Position{Line: 1, Column: tt.innerStart},
				End:   source.Position{Line: 1, Column: tt.innerEnd},
			}
			wantOuter := source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: tt.outerEnd},
			}
			if integer.Span() != wantInner {
				t.Errorf("AST span = %+v, want %+v", integer.Span(), wantInner)
			}
			if result.extent != wantOuter {
				t.Errorf("extent = %+v, want %+v", result.extent, wantOuter)
			}
			if p.current != len(tokens)-1 {
				t.Errorf("current = %d, want %d", p.current, len(tokens)-1)
			}
		})
	}
}
func TestParsePrimaryGroupingErrors(t *testing.T) {
	tests := []struct {
		name         string
		kinds        []lexer.TokenKind
		lexemes      []string
		wantMessage  string
		wantActual   lexer.TokenKind
		wantExpected lexer.TokenKind
		wantColumn   int
	}{
		{
			name:         "empty grouping",
			kinds:        []lexer.TokenKind{lexer.TokenLeftParen, lexer.TokenRightParen, lexer.TokenEOF},
			lexemes:      []string{"(", ")", ""},
			wantMessage:  "expected an expression",
			wantActual:   lexer.TokenRightParen,
			wantExpected: lexer.TokenInt,
			wantColumn:   2,
		},
		{
			name:         "missing closing parenthesis",
			kinds:        []lexer.TokenKind{lexer.TokenLeftParen, lexer.TokenInt, lexer.TokenEOF},
			lexemes:      []string{"(", "42", ""},
			wantMessage:  "expected ')' to close parenthesized expression",
			wantActual:   lexer.TokenEOF,
			wantExpected: lexer.TokenRightParen,
			wantColumn:   4,
		},
		{
			name:         "missing outer closing parenthesis",
			kinds:        []lexer.TokenKind{lexer.TokenLeftParen, lexer.TokenLeftParen, lexer.TokenInt, lexer.TokenRightParen, lexer.TokenEOF},
			lexemes:      []string{"(", "(", "42", ")", ""},
			wantMessage:  "expected ')' to close parenthesized expression",
			wantActual:   lexer.TokenEOF,
			wantExpected: lexer.TokenRightParen,
			wantColumn:   6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := make([]legacyExpressionToken, 0, len(tt.kinds))
			column := 1
			for i, kind := range tt.kinds {
				lexeme := tt.lexemes[i]
				tokens = append(tokens, legacyExpressionToken{
					Kind:   kind,
					Lexeme: lexeme,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: column},
						End:   source.Position{Line: 1, Column: column + len(lexeme)},
					},
				})
				column += len(lexeme)
			}
			p := newLegacyExpressionParser(tokens)
			_, diagnostic := p.parseExpression(0)
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}
			if diagnostic.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", diagnostic.Message, tt.wantMessage)
			}
			if diagnostic.Location.Start.Column != tt.wantColumn {
				t.Errorf("error column = %d, want %d",
					diagnostic.Location.Start.Column, tt.wantColumn)
			}
			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v", diagnostic.Actual, tt.wantActual)
			}
			found := false
			for _, kind := range diagnostic.Expected {
				if kind == tt.wantExpected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected = %v, missing %v", diagnostic.Expected, tt.wantExpected)
			}
		})
	}
}
func TestParseBinaryArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		operator lexer.TokenKind
		lexeme   string
	}{
		{"addition", lexer.TokenPlus, "+"},
		{"subtraction", lexer.TokenMinus, "-"},
		{"multiplication", lexer.TokenStar, "*"},
		{"division", lexer.TokenSlash, "/"},
		{"floor division", lexer.TokenFloorDiv, "//"},
		{"modulo", lexer.TokenPercent, "%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: tt.operator, Lexeme: tt.lexeme},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenEOF},
			}
			p := newLegacyExpressionParser(tokens)
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			binary, ok := result.expression.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("expression type = %T, want *ast.BinaryExpression", result.expression)
			}
			if binary.Operator != tt.operator {
				t.Errorf("Operator = %v, want %v", binary.Operator, tt.operator)
			}
			left, ok := binary.Left.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("Left type = %T, want *ast.IntegerLiteral", binary.Left)
			}
			if left.Value != "10" {
				t.Errorf("Left.Value = %q, want %q", left.Value, "10")
			}
			right, ok := binary.Right.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("Right type = %T, want *ast.IntegerLiteral", binary.Right)
			}
			if right.Value != "3" {
				t.Errorf("Right.Value = %q, want %q", right.Value, "3")
			}
			if p.current != 3 {
				t.Errorf("current = %d, want 3", p.current)
			}
		})
	}
}
func TestParseBinaryPrecedence(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenInt, Lexeme: "10"},
		{Kind: lexer.TokenPlus, Lexeme: "+"},
		{Kind: lexer.TokenInt, Lexeme: "20"},
		{Kind: lexer.TokenStar, Lexeme: "*"},
		{Kind: lexer.TokenInt, Lexeme: "3"},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	root, ok := result.expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
	}
	if root.Operator != lexer.TokenPlus {
		t.Fatalf("root operator = %v, want PLUS", root.Operator)
	}
	left, ok := root.Left.(*ast.IntegerLiteral)
	if !ok || left.Value != "10" {
		t.Errorf("root.Left = %#v, want integer 10", root.Left)
	}
	right, ok := root.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root.Right type = %T, want *ast.BinaryExpression", root.Right)
	}
	if right.Operator != lexer.TokenStar {
		t.Errorf("right operator = %v, want STAR", right.Operator)
	}
	multiplierLeft, ok := right.Left.(*ast.IntegerLiteral)
	if !ok || multiplierLeft.Value != "20" {
		t.Errorf("multiplication left operand = %#v, want integer 20", right.Left)
	}
	multiplierRight, ok := right.Right.(*ast.IntegerLiteral)
	if !ok || multiplierRight.Value != "3" {
		t.Errorf("multiplication right operand = %#v, want integer 3", right.Right)
	}
}
func TestParseBinaryAssociativity(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenInt, Lexeme: "10"},
		{Kind: lexer.TokenMinus, Lexeme: "-"},
		{Kind: lexer.TokenInt, Lexeme: "5"},
		{Kind: lexer.TokenMinus, Lexeme: "-"},
		{Kind: lexer.TokenInt, Lexeme: "2"},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	root, ok := result.expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
	}
	if root.Operator != lexer.TokenMinus {
		t.Fatalf("root operator = %v, want MINUS", root.Operator)
	}
	left, ok := root.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root.Left type = %T, want *ast.BinaryExpression", root.Left)
	}
	if left.Operator != lexer.TokenMinus {
		t.Errorf("left operator = %v, want MINUS", left.Operator)
	}
	first, ok := left.Left.(*ast.IntegerLiteral)
	if !ok || first.Value != "10" {
		t.Errorf("first operand = %#v, want integer 10", left.Left)
	}
	second, ok := left.Right.(*ast.IntegerLiteral)
	if !ok || second.Value != "5" {
		t.Errorf("second operand = %#v, want integer 5", left.Right)
	}
	third, ok := root.Right.(*ast.IntegerLiteral)
	if !ok || third.Value != "2" {
		t.Errorf("third operand = %#v, want integer 2", root.Right)
	}
}
func TestParseBinaryGrouping(t *testing.T) {
	tokens := []legacyExpressionToken{
		{
			Kind: lexer.TokenLeftParen, Lexeme: "(",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: 2},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "10",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 2},
				End:   source.Position{Line: 1, Column: 4},
			},
		},
		{
			Kind: lexer.TokenPlus, Lexeme: "+",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 5},
				End:   source.Position{Line: 1, Column: 6},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "20",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 7},
				End:   source.Position{Line: 1, Column: 9},
			},
		},
		{
			Kind: lexer.TokenRightParen, Lexeme: ")",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 9},
				End:   source.Position{Line: 1, Column: 10},
			},
		},
		{
			Kind: lexer.TokenStar, Lexeme: "*",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 11},
				End:   source.Position{Line: 1, Column: 12},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "3",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 13},
				End:   source.Position{Line: 1, Column: 14},
			},
		},
		{Kind: lexer.TokenEOF},
	}
	p := newLegacyExpressionParser(tokens)
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	root, ok := result.expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
	}
	if root.Operator != lexer.TokenStar {
		t.Fatalf("root operator = %v, want STAR", root.Operator)
	}
	inner, ok := root.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root.Left type = %T, want *ast.BinaryExpression", root.Left)
	}
	if inner.Operator != lexer.TokenPlus {
		t.Errorf("inner operator = %v, want PLUS", inner.Operator)
	}
	wantInner := source.Span{
		Start: source.Position{Line: 1, Column: 2},
		End:   source.Position{Line: 1, Column: 9},
	}
	wantOuter := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 14},
	}
	if inner.Span() != wantInner {
		t.Errorf("inner span = %+v, want %+v", inner.Span(), wantInner)
	}
	if root.Span() != wantOuter {
		t.Errorf("outer span = %+v, want %+v", root.Span(), wantOuter)
	}
	if result.extent != wantOuter {
		t.Errorf("extent = %+v, want %+v", result.extent, wantOuter)
	}
}
func TestParseBinaryMissingOperand(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []legacyExpressionToken
		wantActual lexer.TokenKind
		wantColumn int
	}{
		{
			name: "missing operand at EOF",
			tokens: []legacyExpressionToken{
				{
					Kind: lexer.TokenInt, Lexeme: "10",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 1},
						End:   source.Position{Line: 1, Column: 3},
					},
				},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{
					Kind: lexer.TokenEOF,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 5},
						End:   source.Position{Line: 1, Column: 5},
					},
				},
			},
			wantActual: lexer.TokenEOF,
			wantColumn: 5,
		},
		{
			name: "missing operand before closing parenthesis",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{
					Kind: lexer.TokenRightParen, Lexeme: ")",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 6},
						End:   source.Position{Line: 1, Column: 7},
					},
				},
				{Kind: lexer.TokenEOF},
			},
			wantActual: lexer.TokenRightParen,
			wantColumn: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			_, diagnostic := p.parseExpression(0)
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}
			if diagnostic.Message != "expected an expression" {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, "expected an expression")
			}
			if diagnostic.Location.Start.Column != tt.wantColumn {
				t.Errorf("error column = %d, want %d",
					diagnostic.Location.Start.Column, tt.wantColumn)
			}
			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v", diagnostic.Actual, tt.wantActual)
			}
		})
	}
}
func TestParseBinaryMinimumBindingPower(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenInt, Lexeme: "20"},
		{Kind: lexer.TokenPlus, Lexeme: "+"},
		{Kind: lexer.TokenInt, Lexeme: "30"},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(bindingPowerMultiplicative + 1)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	integer, ok := result.expression.(*ast.IntegerLiteral)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.IntegerLiteral", result.expression)
	}
	if integer.Value != "20" {
		t.Errorf("Value = %q, want %q", integer.Value, "20")
	}
	if p.current != 1 {
		t.Errorf("current = %d, want 1", p.current)
	}
	if p.peek().Kind != lexer.TokenPlus {
		t.Errorf("next token = %v, want PLUS", p.peek().Kind)
	}
}
func TestParseUnaryArithmetic(t *testing.T) {
	tests := []struct {
		name     string
		operator lexer.TokenKind
		lexeme   string
		value    string
	}{
		{"unary minus", lexer.TokenMinus, "-", "42"},
		{"unary plus", lexer.TokenPlus, "+", "10"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser([]legacyExpressionToken{
				{Kind: tt.operator, Lexeme: tt.lexeme},
				{Kind: lexer.TokenInt, Lexeme: tt.value},
				{Kind: lexer.TokenEOF},
			})
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			unary, ok := result.expression.(*ast.UnaryExpression)
			if !ok {
				t.Fatalf("expression type = %T, want *ast.UnaryExpression", result.expression)
			}
			if unary.Operator != tt.operator {
				t.Errorf("Operator = %v, want %v", unary.Operator, tt.operator)
			}
			operand, ok := unary.Operand.(*ast.IntegerLiteral)
			if !ok {
				t.Fatalf("Operand type = %T, want *ast.IntegerLiteral", unary.Operand)
			}
			if operand.Value != tt.value {
				t.Errorf("Operand.Value = %q, want %q", operand.Value, tt.value)
			}
			if p.current != 2 {
				t.Errorf("current = %d, want 2", p.current)
			}
		})
	}
}
func TestParseNestedUnary(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenMinus, Lexeme: "-"},
		{Kind: lexer.TokenMinus, Lexeme: "-"},
		{Kind: lexer.TokenInt, Lexeme: "5"},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	outer, ok := result.expression.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("outer type = %T, want *ast.UnaryExpression", result.expression)
	}
	if outer.Operator != lexer.TokenMinus {
		t.Errorf("outer operator = %v, want MINUS", outer.Operator)
	}
	inner, ok := outer.Operand.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("inner type = %T, want *ast.UnaryExpression", outer.Operand)
	}
	if inner.Operator != lexer.TokenMinus {
		t.Errorf("inner operator = %v, want MINUS", inner.Operator)
	}
	operand, ok := inner.Operand.(*ast.IntegerLiteral)
	if !ok || operand.Value != "5" {
		t.Errorf("inner operand = %#v, want integer 5", inner.Operand)
	}
}
func TestParseUnaryBinaryPrecedence(t *testing.T) {
	tests := []struct {
		name           string
		tokens         []legacyExpressionToken
		binaryOperator lexer.TokenKind
		unaryOnLeft    bool
	}{
		{
			name: "unary before multiplication",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenEOF},
			},
			binaryOperator: lexer.TokenStar,
			unaryOnLeft:    true,
		},
		{
			name: "binary followed by unary",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenEOF},
			},
			binaryOperator: lexer.TokenMinus,
			unaryOnLeft:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			root, ok := result.expression.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
			}
			if root.Operator != tt.binaryOperator {
				t.Errorf("root operator = %v, want %v", root.Operator, tt.binaryOperator)
			}
			var unaryOperand ast.Expression
			if tt.unaryOnLeft {
				unaryOperand = root.Left
			} else {
				unaryOperand = root.Right
			}
			unary, ok := unaryOperand.(*ast.UnaryExpression)
			if !ok {
				t.Fatalf("unary operand type = %T, want *ast.UnaryExpression", unaryOperand)
			}
			if unary.Operator != lexer.TokenMinus {
				t.Errorf("unary operator = %v, want MINUS", unary.Operator)
			}
		})
	}
}
func TestParseUnaryGrouping(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{
			Kind: lexer.TokenMinus, Lexeme: "-",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: 2},
			},
		},
		{
			Kind: lexer.TokenLeftParen, Lexeme: "(",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 2},
				End:   source.Position{Line: 1, Column: 3},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "10",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 3},
				End:   source.Position{Line: 1, Column: 5},
			},
		},
		{
			Kind: lexer.TokenPlus, Lexeme: "+",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 6},
				End:   source.Position{Line: 1, Column: 7},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "20",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 8},
				End:   source.Position{Line: 1, Column: 10},
			},
		},
		{
			Kind: lexer.TokenRightParen, Lexeme: ")",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 10},
				End:   source.Position{Line: 1, Column: 11},
			},
		},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	outer, ok := result.expression.(*ast.UnaryExpression)
	if !ok {
		t.Fatalf("outer type = %T, want *ast.UnaryExpression", result.expression)
	}
	inner, ok := outer.Operand.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("inner type = %T, want *ast.BinaryExpression", outer.Operand)
	}
	if inner.Operator != lexer.TokenPlus {
		t.Errorf("inner operator = %v, want PLUS", inner.Operator)
	}
	wantInner := source.Span{
		Start: source.Position{Line: 1, Column: 3},
		End:   source.Position{Line: 1, Column: 10},
	}
	wantOuter := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 11},
	}
	if inner.Span() != wantInner {
		t.Errorf("inner span = %+v, want %+v", inner.Span(), wantInner)
	}
	if outer.Span() != wantOuter {
		t.Errorf("outer span = %+v, want %+v", outer.Span(), wantOuter)
	}
	if result.extent != wantOuter {
		t.Errorf("extent = %+v, want %+v", result.extent, wantOuter)
	}
}
func TestParseUnaryMissingOperand(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []legacyExpressionToken
		wantActual lexer.TokenKind
		wantColumn int
	}{
		{
			name: "missing operand at EOF",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{
					Kind: lexer.TokenEOF,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 2},
						End:   source.Position{Line: 1, Column: 2},
					},
				},
			},
			wantActual: lexer.TokenEOF,
			wantColumn: 2,
		},
		{
			name: "invalid operand before closing parenthesis",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{
					Kind: lexer.TokenRightParen, Lexeme: ")",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 2},
						End:   source.Position{Line: 1, Column: 3},
					},
				},
				{Kind: lexer.TokenEOF},
			},
			wantActual: lexer.TokenRightParen,
			wantColumn: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			_, diagnostic := p.parseExpression(0)
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}
			if diagnostic.Message != "expected an expression" {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, "expected an expression")
			}
			if diagnostic.Location.Start.Column != tt.wantColumn {
				t.Errorf("error column = %d, want %d",
					diagnostic.Location.Start.Column, tt.wantColumn)
			}
			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v", diagnostic.Actual, tt.wantActual)
			}
		})
	}
}
func TestParseExponentiation(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenInt, Lexeme: "2"},
		{Kind: lexer.TokenPower, Lexeme: "**"},
		{Kind: lexer.TokenInt, Lexeme: "3"},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	root, ok := result.expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("expression type = %T, want *ast.BinaryExpression", result.expression)
	}
	if root.Operator != lexer.TokenPower {
		t.Errorf("operator = %v, want TokenPower", root.Operator)
	}
	left, ok := root.Left.(*ast.IntegerLiteral)
	if !ok || left.Value != "2" {
		t.Errorf("left operand = %#v, want integer 2", root.Left)
	}
	right, ok := root.Right.(*ast.IntegerLiteral)
	if !ok || right.Value != "3" {
		t.Errorf("right operand = %#v, want integer 3", root.Right)
	}
	if p.current != 3 {
		t.Errorf("current = %d, want 3", p.current)
	}
}
func TestParseExponentiationAssociativity(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{Kind: lexer.TokenInt, Lexeme: "2"},
		{Kind: lexer.TokenPower, Lexeme: "**"},
		{Kind: lexer.TokenInt, Lexeme: "3"},
		{Kind: lexer.TokenPower, Lexeme: "**"},
		{Kind: lexer.TokenInt, Lexeme: "2"},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	root, ok := result.expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
	}
	if root.Operator != lexer.TokenPower {
		t.Fatalf("root operator = %v, want TokenPower", root.Operator)
	}
	left, ok := root.Left.(*ast.IntegerLiteral)
	if !ok || left.Value != "2" {
		t.Errorf("root.Left = %#v, want integer 2", root.Left)
	}
	right, ok := root.Right.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root.Right type = %T, want *ast.BinaryExpression", root.Right)
	}
	if right.Operator != lexer.TokenPower {
		t.Errorf("nested operator = %v, want TokenPower", right.Operator)
	}
	middle, ok := right.Left.(*ast.IntegerLiteral)
	if !ok || middle.Value != "3" {
		t.Errorf("nested left = %#v, want integer 3", right.Left)
	}
	final, ok := right.Right.(*ast.IntegerLiteral)
	if !ok || final.Value != "2" {
		t.Errorf("nested right = %#v, want integer 2", right.Right)
	}
}
func TestParseExponentiationGrouping(t *testing.T) {
	p := newLegacyExpressionParser([]legacyExpressionToken{
		{
			Kind: lexer.TokenLeftParen, Lexeme: "(",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: 2},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "2",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 2},
				End:   source.Position{Line: 1, Column: 3},
			},
		},
		{
			Kind: lexer.TokenPower, Lexeme: "**",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 4},
				End:   source.Position{Line: 1, Column: 6},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "3",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 7},
				End:   source.Position{Line: 1, Column: 8},
			},
		},
		{
			Kind: lexer.TokenRightParen, Lexeme: ")",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 8},
				End:   source.Position{Line: 1, Column: 9},
			},
		},
		{
			Kind: lexer.TokenPower, Lexeme: "**",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 10},
				End:   source.Position{Line: 1, Column: 12},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "2",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 13},
				End:   source.Position{Line: 1, Column: 14},
			},
		},
		{Kind: lexer.TokenEOF},
	})
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	root, ok := result.expression.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
	}
	if root.Operator != lexer.TokenPower {
		t.Errorf("root operator = %v, want TokenPower", root.Operator)
	}
	inner, ok := root.Left.(*ast.BinaryExpression)
	if !ok {
		t.Fatalf("root.Left type = %T, want *ast.BinaryExpression", root.Left)
	}
	if inner.Operator != lexer.TokenPower {
		t.Errorf("inner operator = %v, want TokenPower", inner.Operator)
	}
	wantInner := source.Span{
		Start: source.Position{Line: 1, Column: 2},
		End:   source.Position{Line: 1, Column: 8},
	}
	wantOuter := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 14},
	}
	if inner.Span() != wantInner {
		t.Errorf("inner span = %+v, want %+v", inner.Span(), wantInner)
	}
	if root.Span() != wantOuter {
		t.Errorf("outer span = %+v, want %+v", root.Span(), wantOuter)
	}
	if result.extent != wantOuter {
		t.Errorf("extent = %+v, want %+v", result.extent, wantOuter)
	}
}
func TestParseExponentiationUnaryPrecedence(t *testing.T) {
	tests := []struct {
		name   string
		tokens []legacyExpressionToken
		check  func(t *testing.T, expression ast.Expression)
	}{
		{
			name: "negative base expression",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenEOF},
			},
			check: func(t *testing.T, expression ast.Expression) {
				t.Helper()
				unary, ok := expression.(*ast.UnaryExpression)
				if !ok {
					t.Fatalf("root = %T, want *ast.UnaryExpression", expression)
				}
				if unary.Operator != lexer.TokenMinus {
					t.Errorf("unary operator = %v, want TokenMinus", unary.Operator)
				}
				power, ok := unary.Operand.(*ast.BinaryExpression)
				if !ok || power.Operator != lexer.TokenPower {
					t.Fatalf("unary operand = %#v, want power expression", unary.Operand)
				}
			},
		},
		{
			name: "negative exponent",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenEOF},
			},
			check: func(t *testing.T, expression ast.Expression) {
				t.Helper()
				power, ok := expression.(*ast.BinaryExpression)
				if !ok || power.Operator != lexer.TokenPower {
					t.Fatalf("root = %#v, want power expression", expression)
				}
				unary, ok := power.Right.(*ast.UnaryExpression)
				if !ok || unary.Operator != lexer.TokenMinus {
					t.Fatalf("right operand = %#v, want unary minus", power.Right)
				}
			},
		},
		{
			name: "negative base expression and exponent",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenEOF},
			},
			check: func(t *testing.T, expression ast.Expression) {
				t.Helper()
				outer, ok := expression.(*ast.UnaryExpression)
				if !ok || outer.Operator != lexer.TokenMinus {
					t.Fatalf("root = %#v, want unary minus", expression)
				}
				power, ok := outer.Operand.(*ast.BinaryExpression)
				if !ok || power.Operator != lexer.TokenPower {
					t.Fatalf("outer operand = %#v, want power expression", outer.Operand)
				}
				exponent, ok := power.Right.(*ast.UnaryExpression)
				if !ok || exponent.Operator != lexer.TokenMinus {
					t.Fatalf("exponent = %#v, want unary minus", power.Right)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			tt.check(t, result.expression)
		})
	}
}
func TestParseExponentiationMultiplicativePrecedence(t *testing.T) {
	tests := []struct {
		name         string
		tokens       []legacyExpressionToken
		powerOnRight bool
	}{
		{
			name: "multiplication then power",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "4"},
				{Kind: lexer.TokenEOF},
			},
			powerOnRight: true,
		},
		{
			name: "power then multiplication",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "4"},
				{Kind: lexer.TokenEOF},
			},
			powerOnRight: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			root, ok := result.expression.(*ast.BinaryExpression)
			if !ok {
				t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
			}
			if root.Operator != lexer.TokenStar {
				t.Fatalf("root operator = %v, want TokenStar", root.Operator)
			}
			var powerOperand ast.Expression
			if tt.powerOnRight {
				powerOperand = root.Right
			} else {
				powerOperand = root.Left
			}
			power, ok := powerOperand.(*ast.BinaryExpression)
			if !ok || power.Operator != lexer.TokenPower {
				t.Errorf("power operand = %#v, want power expression", powerOperand)
			}
		})
	}
}
func TestParseExponentiationMissingOperand(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []legacyExpressionToken
		wantActual lexer.TokenKind
		wantColumn int
	}{
		{
			name: "missing exponent at EOF",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{
					Kind: lexer.TokenEOF,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 5},
						End:   source.Position{Line: 1, Column: 5},
					},
				},
			},
			wantActual: lexer.TokenEOF,
			wantColumn: 5,
		},
		{
			name: "invalid exponent before closing parenthesis",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{
					Kind: lexer.TokenRightParen, Lexeme: ")",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 6},
						End:   source.Position{Line: 1, Column: 7},
					},
				},
				{Kind: lexer.TokenEOF},
			},
			wantActual: lexer.TokenRightParen,
			wantColumn: 6,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			_, diagnostic := p.parseExpression(0)
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}
			if diagnostic.Message != "expected an expression" {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, "expected an expression")
			}
			if diagnostic.Location.Start.Column != tt.wantColumn {
				t.Errorf("error column = %d, want %d",
					diagnostic.Location.Start.Column, tt.wantColumn)
			}
			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v", diagnostic.Actual, tt.wantActual)
			}
		})
	}
}
func TestParseExpressionBoundaries(t *testing.T) {
	tests := []struct {
		name            string
		tokens          []legacyExpressionToken
		minBindingPower int
		wantRootKind    lexer.TokenKind
		wantNextKind    lexer.TokenKind
		wantCurrent     int
	}{
		{
			name: "closing parenthesis boundary",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenRightParen, Lexeme: ")"},
				{Kind: lexer.TokenEOF},
			},
			wantRootKind: lexer.TokenPlus,
			wantNextKind: lexer.TokenRightParen,
			wantCurrent:  3,
		},
		{
			name: "newline boundary",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantRootKind: lexer.TokenPlus,
			wantNextKind: lexer.TokenNewline,
			wantCurrent:  3,
		},
		{
			name: "EOF boundary",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "42"},
				{Kind: lexer.TokenEOF},
			},
			wantRootKind: lexer.TokenInt,
			wantNextKind: lexer.TokenEOF,
			wantCurrent:  1,
		},
		{
			name: "nested operator precedence boundary",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenRightParen, Lexeme: ")"},
				{Kind: lexer.TokenEOF},
			},
			wantRootKind: lexer.TokenPlus,
			wantNextKind: lexer.TokenRightParen,
			wantCurrent:  5,
		},
		{
			name: "unexpected adjacent operand",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenInt, Lexeme: "30"},
				{Kind: lexer.TokenEOF},
			},
			wantRootKind: lexer.TokenPlus,
			wantNextKind: lexer.TokenInt,
			wantCurrent:  3,
		},
		{
			name: "minimum binding power boundary",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "30"},
				{Kind: lexer.TokenEOF},
			},
			minBindingPower: bindingPowerMultiplicative + 1,
			wantRootKind:    lexer.TokenInt,
			wantNextKind:    lexer.TokenStar,
			wantCurrent:     1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			result, diagnostic := p.parseExpression(tt.minBindingPower)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			switch tt.wantRootKind {
			case lexer.TokenInt:
				if _, ok := result.expression.(*ast.IntegerLiteral); !ok {
					t.Errorf("root type = %T, want *ast.IntegerLiteral", result.expression)
				}
			default:
				binary, ok := result.expression.(*ast.BinaryExpression)
				if !ok {
					t.Fatalf("root type = %T, want *ast.BinaryExpression", result.expression)
				}
				if binary.Operator != tt.wantRootKind {
					t.Errorf("root operator = %v, want %v",
						binary.Operator, tt.wantRootKind)
				}
			}
			if p.current != tt.wantCurrent {
				t.Errorf("current = %d, want %d", p.current, tt.wantCurrent)
			}
			if p.peek().Kind != tt.wantNextKind {
				t.Errorf("next token = %v, want %v",
					p.peek().Kind, tt.wantNextKind)
			}
		})
	}
}
func TestParseExpressionDiagnosticAccuracy(t *testing.T) {
	expressionStarters := []lexer.TokenKind{
		lexer.TokenInt,
		lexer.TokenString,
		lexer.TokenIdent,
		lexer.TokenLeftParen,
		lexer.TokenPlus,
		lexer.TokenMinus,
	}
	tests := []struct {
		name         string
		tokens       []legacyExpressionToken
		wantMessage  string
		wantLocation source.Span
		wantExpected []lexer.TokenKind
		wantActual   lexer.TokenKind
	}{
		{
			name: "missing addition operand",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{
					Kind: lexer.TokenEOF,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 5},
						End:   source.Position{Line: 1, Column: 5},
					},
				},
			},
			wantMessage: "expected an expression",
			wantLocation: source.Span{
				Start: source.Position{Line: 1, Column: 5},
				End:   source.Position{Line: 1, Column: 5},
			},
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenEOF,
		},
		{
			name: "unexpected closing parenthesis after multiplication",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{
					Kind: lexer.TokenRightParen, Lexeme: ")",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 6},
						End:   source.Position{Line: 1, Column: 7},
					},
				},
				{Kind: lexer.TokenEOF},
			},
			wantMessage: "expected an expression",
			wantLocation: source.Span{
				Start: source.Position{Line: 1, Column: 6},
				End:   source.Position{Line: 1, Column: 7},
			},
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenRightParen,
		},
		{
			name: "invalid exponent operand",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{
					Kind: lexer.TokenStar, Lexeme: "*",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 6},
						End:   source.Position{Line: 1, Column: 7},
					},
				},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage: "expected an expression",
			wantLocation: source.Span{
				Start: source.Position{Line: 1, Column: 6},
				End:   source.Position{Line: 1, Column: 7},
			},
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenStar,
		},
		{
			name: "empty parentheses",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenLeftParen, Lexeme: "("},
				{
					Kind: lexer.TokenRightParen, Lexeme: ")",
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 2},
						End:   source.Position{Line: 1, Column: 3},
					},
				},
				{Kind: lexer.TokenEOF},
			},
			wantMessage: "expected an expression",
			wantLocation: source.Span{
				Start: source.Position{Line: 1, Column: 2},
				End:   source.Position{Line: 1, Column: 3},
			},
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenRightParen,
		},
		{
			name: "missing closing parenthesis",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenLeftParen, Lexeme: "("},
				{Kind: lexer.TokenInt, Lexeme: "10"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{
					Kind: lexer.TokenEOF,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 9},
						End:   source.Position{Line: 1, Column: 9},
					},
				},
			},
			wantMessage: "expected ')' to close parenthesized expression",
			wantLocation: source.Span{
				Start: source.Position{Line: 1, Column: 9},
				End:   source.Position{Line: 1, Column: 9},
			},
			wantExpected: []lexer.TokenKind{lexer.TokenRightParen},
			wantActual:   lexer.TokenEOF,
		},
		{
			name: "missing nested unary operand",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{
					Kind: lexer.TokenEOF,
					Span: source.Span{
						Start: source.Position{Line: 1, Column: 3},
						End:   source.Position{Line: 1, Column: 3},
					},
				},
			},
			wantMessage: "expected an expression",
			wantLocation: source.Span{
				Start: source.Position{Line: 1, Column: 3},
				End:   source.Position{Line: 1, Column: 3},
			},
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenEOF,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			_, diagnostic := p.parseExpression(0)
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}
			if diagnostic.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, tt.wantMessage)
			}
			if diagnostic.Location != tt.wantLocation {
				t.Errorf("Location = %+v, want %+v",
					diagnostic.Location, tt.wantLocation)
			}
			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf("Expected = %v, want %v",
					diagnostic.Expected, tt.wantExpected)
			}
			if diagnostic.Actual == nil {
				t.Fatal("Actual = nil, want a token kind")
			}
			if *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v",
					*diagnostic.Actual, tt.wantActual)
			}
		})
	}
}

// formatExpressionTree converts a parsed expression into a compact
// prefix representation for testing operator precedence and grouping.
//
// This helper is intended only for parser tests and does not
// evaluate expressions or modify the AST.
func formatExpressionTree(expression ast.Expression) string {
	switch node := expression.(type) {
	case *ast.IntegerLiteral:
		return node.Value
	case *ast.StringLiteral:
		return fmt.Sprintf("%q", node.Value)
	case *ast.Identifier:
		return node.Name
	case *ast.UnaryExpression:
		return fmt.Sprintf("(%s %s)",
			node.Operator, formatExpressionTree(node.Operand))
	case *ast.BinaryExpression:
		return fmt.Sprintf("(%s %s %s)",
			node.Operator,
			formatExpressionTree(node.Left),
			formatExpressionTree(node.Right))
	default:
		return fmt.Sprintf("<unknown %T>", expression)
	}
}
func TestParseExpressionConformance(t *testing.T) {
	tests := []struct {
		name   string
		tokens []legacyExpressionToken
		want   string
	}{
		{
			name: "mixed arithmetic precedence",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "1"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "4"},
				{Kind: lexer.TokenEOF},
			},
			want: "(PLUS 1 (STAR 2 (POWER 3 4)))",
		},
		{
			name: "complex unary and arithmetic expression",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "4"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "5"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenLeftParen, Lexeme: "("},
				{Kind: lexer.TokenInt, Lexeme: "6"},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "1"},
				{Kind: lexer.TokenRightParen, Lexeme: ")"},
				{Kind: lexer.TokenEOF},
			},
			want: "(PLUS (MINUS (POWER 2 (POWER 3 (MINUS 4)))) (STAR 5 (MINUS 6 1)))",
		},
		{
			name: "floor division and modulo associativity",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenFloorDiv, Lexeme: "//"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenPercent, Lexeme: "%"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenEOF},
			},
			want: "(PERCENT (FLOOR_DIV 20 3) 2)",
		},
		{
			name: "explicit exponentiation grouping",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenLeftParen, Lexeme: "("},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenRightParen, Lexeme: ")"},
				{Kind: lexer.TokenEOF},
			},
			want: "(POWER 2 (POWER 3 2))",
		},
		{
			name: "parenthesized negative base",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenLeftParen, Lexeme: "("},
				{Kind: lexer.TokenMinus, Lexeme: "-"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenRightParen, Lexeme: ")"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenEOF},
			},
			want: "(POWER (MINUS 2) 2)",
		},
		{
			name: "long right associative exponentiation chain",
			tokens: []legacyExpressionToken{
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPower, Lexeme: "**"},
				{Kind: lexer.TokenInt, Lexeme: "1"},
				{Kind: lexer.TokenEOF},
			},
			want: "(POWER 2 (POWER 3 (POWER 2 1)))",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newLegacyExpressionParser(tt.tokens)
			result, diagnostic := p.parseExpression(0)
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			got := formatExpressionTree(result.expression)
			if got != tt.want {
				t.Errorf("expression tree = %q\nwant            = %q",
					got, tt.want)
			}
			if p.peek().Kind != lexer.TokenEOF {
				t.Errorf("next token = %v, want EOF", p.peek().Kind)
			}
		})
	}
}
func TestParseDeeplyNestedParentheses(t *testing.T) {
	const depth = 1000
	tokens := make([]legacyExpressionToken, 0, depth*2+2)
	for range depth {
		tokens = append(tokens, legacyExpressionToken{Kind: lexer.TokenLeftParen})
	}
	tokens = append(tokens, legacyExpressionToken{
		Kind:   lexer.TokenInt,
		Lexeme: "42",
	})
	for range depth {
		tokens = append(tokens, legacyExpressionToken{Kind: lexer.TokenRightParen})
	}
	tokens = append(tokens, legacyExpressionToken{Kind: lexer.TokenEOF})
	p := newLegacyExpressionParser(tokens)
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	if _, ok := result.expression.(*ast.IntegerLiteral); !ok {
		t.Errorf("expression type = %T, want *ast.IntegerLiteral", result.expression)
	}
	if p.peek().Kind != lexer.TokenEOF {
		t.Errorf("next token = %v, want EOF", p.peek().Kind)
	}
}
func TestParseLongExponentiationChain(t *testing.T) {
	const operators = 1000
	tokens := make([]legacyExpressionToken, 0, operators*2+2)
	for range operators {
		tokens = append(tokens,
			legacyExpressionToken{Kind: lexer.TokenInt, Lexeme: "2"},
			legacyExpressionToken{Kind: lexer.TokenPower, Lexeme: "**"},
		)
	}
	tokens = append(tokens,
		legacyExpressionToken{Kind: lexer.TokenInt, Lexeme: "2"},
		legacyExpressionToken{Kind: lexer.TokenEOF},
	)
	p := newLegacyExpressionParser(tokens)
	result, diagnostic := p.parseExpression(0)
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	expression := result.expression
	for i := 0; i < operators; i++ {
		binary, ok := expression.(*ast.BinaryExpression)
		if !ok {
			t.Fatalf("node %d = %T, want *ast.BinaryExpression", i, expression)
		}
		if binary.Operator != lexer.TokenPower {
			t.Fatalf("node %d operator = %v, want POWER", i, binary.Operator)
		}
		expression = binary.Right
	}
	if _, ok := expression.(*ast.IntegerLiteral); !ok {
		t.Errorf("final operand = %T, want *ast.IntegerLiteral", expression)
	}
	if p.peek().Kind != lexer.TokenEOF {
		t.Errorf("next token = %v, want EOF", p.peek().Kind)
	}
}
