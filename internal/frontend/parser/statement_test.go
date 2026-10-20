package parser

import (
	"reflect"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// newFlatStatementTestParser constructs an H2 parser using the
// production streaming lexer and arena initialization.
//
// Tests use real Eloque source text so token offsets, keyword
// classification, and statement boundaries are exercised together.
func newFlatStatementTestParser(sourceText string) *FlatParser {
	return NewFlatParser(sourceText)
}

// TestFlatStatementBindings verifies the statement node kind, variable
// name span, initializer reference, and source extent for both binding
// statement forms.
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

// TestFlatSayStatement verifies that output statements reference their
// expression through the arena without introducing a binding payload.
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

// TestFlatStatementGroupingSpan verifies that grouping parentheses
// contribute to the statement's syntactic span without becoming
// additional expression nodes.
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

// TestFlatStatementTerminators verifies that individual statement
// parsers leave NEWLINE and EOF handling to the program parser.
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

// TestFlatStatementDiagnostics verifies that malformed H2 statements
// preserve M1's diagnostic messages and expected-token information.
func TestFlatStatementDiagnostics(t *testing.T) {
	tests := []struct {
		name         string
		source       string
		wantMessage  string
		wantActual   lexer.TokenKind
		wantExpected []lexer.TokenKind
	}{
		{
			name:         "let missing name",
			source:       "let be 20",
			wantMessage:  "expected a variable name after 'let'",
			wantActual:   lexer.TokenBe,
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
		},
		{
			name:         "let missing be",
			source:       "let x 20",
			wantMessage:  "expected 'be' after variable name to initialize a variable value",
			wantActual:   lexer.TokenInt,
			wantExpected: []lexer.TokenKind{lexer.TokenBe},
		},
		{
			name:         "change missing name",
			source:       "change to 2",
			wantMessage:  "expected a variable name after 'change'",
			wantActual:   lexer.TokenTo,
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
		},
		{
			name:         "change missing to",
			source:       "change x 2",
			wantMessage:  "expected 'to' after variable name to assign a new variable value",
			wantActual:   lexer.TokenInt,
			wantExpected: []lexer.TokenKind{lexer.TokenTo},
		},
		{
			name:        "say missing expression",
			source:      "say",
			wantMessage: "expected an expression",
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
			name:        "unsupported statement",
			source:      "42",
			wantMessage: "expected a statement ('let', 'change' or 'say')",
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

			id, diagnostic := p.parseFlatStatement()
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if id != ast.InvalidNodeID {
				t.Errorf("node ID = %v, want InvalidNodeID", id)
			}

			if diagnostic.Message != tt.wantMessage {
				t.Errorf("message = %q, want %q",
					diagnostic.Message, tt.wantMessage)
			}

			if diagnostic.Actual == nil ||
				*diagnostic.Actual != tt.wantActual {
				t.Errorf("actual = %v, want %v",
					diagnostic.Actual, tt.wantActual)
			}

			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf("expected tokens = %v, want %v",
					diagnostic.Expected, tt.wantExpected)
			}
		})
	}
}
