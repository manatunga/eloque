package parser

import (
	"slices"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func newFlatStatementTestParser(sourceText string) *FlatParser {
	return NewFlatParser(sourceText)
}

func TestFlatStatementBindings(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantKind ast.NodeKind
		wantName string
	}{
		{
			name:     "let declaration",
			source:   "let age be 20",
			wantKind: ast.NodeLetStatement,
			wantName: "age",
		},
		{
			name:     "change assignment",
			source:   "change age to 21",
			wantKind: ast.NodeChangeStatement,
			wantName: "age",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFlatStatementTestParser(tt.source)

			id, diagnostic := p.parseFlatStatement()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			node := p.arena.Get(id)
			if node.Kind != tt.wantKind {
				t.Fatalf("node kind = %v, want %v",
					node.Kind, tt.wantKind)
			}

			binding := p.arena.BindingPayload(node.Data)

			if got := p.arena.SourceText(binding.Name); got != tt.wantName {
				t.Errorf("binding name = %q, want %q", got, tt.wantName)
			}

			value := p.arena.Get(binding.Value)
			if value.Kind != ast.NodeIntegerLiteral {
				t.Errorf("value kind = %v, want integer", value.Kind)
			}

			wantSpan := source.ByteSpan{
				Start: 0,
				End:   uint32(len(tt.source)),
			}
			if node.Span != wantSpan {
				t.Errorf("statement span = %+v, want %+v",
					node.Span, wantSpan)
			}

			if got := p.peek().Kind; got != lexer.TokenEOF {
				t.Errorf("remaining token = %v, want EOF", got)
			}
		})
	}
}

func TestFlatSayStatement(t *testing.T) {
	sourceText := "say 2 + 3 * 4"
	p := newFlatStatementTestParser(sourceText)

	id, diagnostic := p.parseFlatStatement()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	node := p.arena.Get(id)
	if node.Kind != ast.NodeSayStatement {
		t.Fatalf("node kind = %v, want say", node.Kind)
	}

	expression := p.arena.Get(ast.NodeID(node.Data))
	if expression.Kind != ast.NodeBinaryExpression {
		t.Fatalf("expression kind = %v, want binary", expression.Kind)
	}

	binary := p.arena.BinaryPayload(expression.Data)
	if binary.Operator != lexer.TokenPlus {
		t.Errorf("root operator = %v, want plus", binary.Operator)
	}

	right := p.arena.Get(binary.Right)
	if right.Kind != ast.NodeBinaryExpression {
		t.Fatalf("right expression kind = %v, want binary", right.Kind)
	}

	multiplication := p.arena.BinaryPayload(right.Data)
	if multiplication.Operator != lexer.TokenStar {
		t.Errorf("right operator = %v, want multiplication",
			multiplication.Operator)
	}

	if node.Span.Start != 0 ||
		node.Span.End != uint32(len(sourceText)) {
		t.Errorf("statement span = %+v, want entire source", node.Span)
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatStatementGroupingSpan(t *testing.T) {
	sourceText := "let result be (2 + 3)"
	p := newFlatStatementTestParser(sourceText)

	id, diagnostic := p.parseFlatStatement()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	node := p.arena.Get(id)
	binding := p.arena.BindingPayload(node.Data)
	value := p.arena.Get(binding.Value)

	if node.Span.End != uint32(len(sourceText)) {
		t.Errorf("statement end = %d, want %d",
			node.Span.End, len(sourceText))
	}

	// The binary expression begins after the opening parenthesis
	// and ends before the closing parenthesis.
	if got := p.arena.SourceText(value.Span); got != "2 + 3" {
		t.Errorf("expression source = %q, want %q", got, "2 + 3")
	}

	if got := p.peek().Kind; got != lexer.TokenEOF {
		t.Errorf("remaining token = %v, want EOF", got)
	}
}

func TestFlatStatementTerminators(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		wantNext lexer.TokenKind
	}{
		{"let newline", "let x be 1\n", lexer.TokenNewline},
		{"change newline", "change x to 2\n", lexer.TokenNewline},
		{"say newline", "say x\n", lexer.TokenNewline},
		{"let EOF", "let x be 1", lexer.TokenEOF},
		{"change EOF", "change x to 2", lexer.TokenEOF},
		{"say EOF", "say x", lexer.TokenEOF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFlatStatementTestParser(tt.source)

			_, diagnostic := p.parseFlatStatement()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			if got := p.peek().Kind; got != tt.wantNext {
				t.Errorf("remaining token = %v, want %v",
					got, tt.wantNext)
			}
		})
	}
}

func TestFlatStatementDiagnostics(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		wantMessage  string
		wantStart    uint32
		wantEnd      uint32
		wantActual   lexer.TokenKind
		wantExpected []lexer.TokenKind
	}{
		{
			name:         "let missing name",
			source:       "let be 20",
			wantMessage:  "expected a variable name after 'let'",
			wantStart:    4,
			wantEnd:      6,
			wantActual:   lexer.TokenBe,
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
		},
		{
			name:         "let missing be",
			source:       "let x 20",
			wantMessage:  "expected 'be' after variable name to initialize a variable value",
			wantStart:    6,
			wantEnd:      8,
			wantActual:   lexer.TokenInt,
			wantExpected: []lexer.TokenKind{lexer.TokenBe},
		},
		{
			name:         "change missing name",
			source:       "change to 2",
			wantMessage:  "expected a variable name after 'change'",
			wantStart:    7,
			wantEnd:      9,
			wantActual:   lexer.TokenTo,
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
		},
		{
			name:         "change missing to",
			source:       "change x 2",
			wantMessage:  "expected 'to' after variable name to assign a new variable value",
			wantStart:    9,
			wantEnd:      10,
			wantActual:   lexer.TokenInt,
			wantExpected: []lexer.TokenKind{lexer.TokenTo},
		},
		{
			name:        "say missing expression at EOF",
			source:      "say",
			wantMessage: "expected an expression",
			wantStart:   3,
			wantEnd:     3,
			wantActual:  lexer.TokenEOF,
			wantExpected: []lexer.TokenKind{
				lexer.TokenInt,
				lexer.TokenString,
				lexer.TokenIdent,
				lexer.TokenLeftParen,
				lexer.TokenPlus,
				lexer.TokenMinus,
			},
		},
		{
			name:        "say missing expression at newline",
			source:      "say\n",
			wantMessage: "expected an expression",
			wantStart:   3,
			wantEnd:     4,
			wantActual:  lexer.TokenNewline,
			wantExpected: []lexer.TokenKind{
				lexer.TokenInt,
				lexer.TokenString,
				lexer.TokenIdent,
				lexer.TokenLeftParen,
				lexer.TokenPlus,
				lexer.TokenMinus,
			},
		},
		{
			name:        "unsupported statement",
			source:      "42",
			wantMessage: "expected a statement ('let', 'change' or 'say')",
			wantStart:   0,
			wantEnd:     2,
			wantActual:  lexer.TokenInt,
			wantExpected: []lexer.TokenKind{
				lexer.TokenLet,
				lexer.TokenChange,
				lexer.TokenSay,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newFlatStatementTestParser(tt.source)

			id, diag := p.parseFlatStatement()
			if diag == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if id != ast.InvalidNodeID {
				t.Errorf("node ID = %v, want InvalidNodeID", id)
			}

			if diag.Code != CodeUnexpectedToken {
				t.Errorf("code = %q, want %q", diag.Code, CodeUnexpectedToken)
			}

			if diag.Severity != diagnostic.SeverityError {
				t.Errorf("severity = %v, want error", diag.Severity)
			}

			if diag.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q", diag.Message, tt.wantMessage)
			}

			if diag.Primary.File != 0 {
				t.Errorf("file ID = %d, want 0", diag.Primary.File)
			}

			if diag.Primary.Span != (source.ByteSpan{
				Start: tt.wantStart,
				End:   tt.wantEnd,
			}) {
				t.Errorf("span = %+v, want [%d,%d)",
					diag.Primary.Span, tt.wantStart, tt.wantEnd)
			}

			if diag.TokenExpectation == nil {
				t.Fatal("missing token expectation")
			}

			if got := diag.TokenExpectation.Actual; got != tt.wantActual.String() {
				t.Errorf("actual = %q, want %q", got, tt.wantActual.String())
			}

			expected := make([]string, len(tt.wantExpected))
			for i, kind := range tt.wantExpected {
				expected[i] = kind.String()
			}

			if !slices.Equal(diag.TokenExpectation.Expected, expected) {
				t.Errorf("expected = %v, want %v",
					diag.TokenExpectation.Expected, expected)
			}
		})
	}
}
