package parser

import (
	"reflect"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// normalizedNode provides a representation-independent AST structure
// for comparing the M1 pointer-based and H2 arena-backed parsers.
//
// Kind identifies the syntax construct. Value stores literal contents
// or binding names, Operator identifies unary and binary operations,
// and Children preserves expression-tree and statement ordering.
//
// Span records the construct's source byte range.
type normalizedNode struct {
	Kind     string
	Value    string
	Operator lexer.TokenKind
	Span     source.ByteSpan
	Children []normalizedNode
}

// normalizedSpan converts an M1 line/column span to H2's byte-offset
// representation using the original source text.
//
// M1 positions are 1-based and columns are currently byte-oriented.
func normalizedSpan(text string, span source.Span) source.ByteSpan {
	offset := func(position source.Position) uint32 {
		line := 1
		start := 0

		for i := 0; i < len(text) && line < position.Line; i++ {
			if text[i] == '\n' {
				line++
				start = i + 1
			}
		}

		return uint32(start + position.Column - 1)
	}

	return source.ByteSpan{
		Start: offset(span.Start),
		End:   offset(span.End),
	}
}

// normalizeM1Expression recursively converts an M1 expression into
// the common representation used by differential tests.
func normalizeM1Expression(
	text string,
	expression ast.Expression,
) normalizedNode {
	switch node := expression.(type) {
	case *ast.IntegerLiteral:
		return normalizedNode{
			Kind:  "integer",
			Value: node.Value,
			Span:  normalizedSpan(text, node.Span()),
		}

	case *ast.StringLiteral:
		return normalizedNode{
			Kind:  "string",
			Value: node.Value,
			Span:  normalizedSpan(text, node.Span()),
		}

	case *ast.Identifier:
		return normalizedNode{
			Kind:  "identifier",
			Value: node.Name,
			Span:  normalizedSpan(text, node.Span()),
		}

	case *ast.UnaryExpression:
		return normalizedNode{
			Kind:     "unary",
			Operator: node.Operator,
			Span:     normalizedSpan(text, node.Span()),
			Children: []normalizedNode{
				normalizeM1Expression(text, node.Operand),
			},
		}

	case *ast.BinaryExpression:
		return normalizedNode{
			Kind:     "binary",
			Operator: node.Operator,
			Span:     normalizedSpan(text, node.Span()),
			Children: []normalizedNode{
				normalizeM1Expression(text, node.Left),
				normalizeM1Expression(text, node.Right),
			},
		}

	default:
		panic("unrecognized M1 expression type")
	}
}

// normalizeM1Statement converts an M1 statement and its expression
// tree into the common representation.
func normalizeM1Statement(
	text string,
	statement ast.Statement,
) normalizedNode {
	switch node := statement.(type) {
	case *ast.LetStatement:
		return normalizedNode{
			Kind:  "let",
			Value: node.Name,
			Span:  normalizedSpan(text, node.Span()),
			Children: []normalizedNode{
				normalizeM1Expression(text, node.Value),
			},
		}

	case *ast.ChangeStatement:
		return normalizedNode{
			Kind:  "change",
			Value: node.Name,
			Span:  normalizedSpan(text, node.Span()),
			Children: []normalizedNode{
				normalizeM1Expression(text, node.Value),
			},
		}

	case *ast.SayStatement:
		return normalizedNode{
			Kind: "say",
			Span: normalizedSpan(text, node.Span()),
			Children: []normalizedNode{
				normalizeM1Expression(text, node.Value),
			},
		}

	default:
		panic("unrecognized M1 statement type")
	}
}

// normalizeH2Expression recursively converts an arena-backed
// expression into the common representation.
//
// Node IDs are resolved through the supplied arena. Binary and
// unary payload indices are read from the node's Data field.
func normalizeH2Expression(
	arena *ast.Arena,
	id ast.NodeID,
) normalizedNode {
	node := arena.Get(id)

	switch node.Kind {
	case ast.NodeIntegerLiteral:
		return normalizedNode{
			Kind:  "integer",
			Value: arena.SourceText(node.Span),
			Span:  node.Span,
		}

	case ast.NodeStringLiteral:
		return normalizedNode{
			Kind:  "string",
			Value: arena.StringValue(id),
			Span:  node.Span,
		}

	case ast.NodeIdentifier:
		return normalizedNode{
			Kind:  "identifier",
			Value: arena.SourceText(node.Span),
			Span:  node.Span,
		}

	case ast.NodeUnaryExpression:
		payload := arena.UnaryPayload(node.Data)
		return normalizedNode{
			Kind:     "unary",
			Operator: payload.Operator,
			Span:     node.Span,
			Children: []normalizedNode{
				normalizeH2Expression(arena, payload.Operand),
			},
		}

	case ast.NodeBinaryExpression:
		payload := arena.BinaryPayload(node.Data)
		return normalizedNode{
			Kind:     "binary",
			Operator: payload.Operator,
			Span:     node.Span,
			Children: []normalizedNode{
				normalizeH2Expression(arena, payload.Left),
				normalizeH2Expression(arena, payload.Right),
			},
		}

	default:
		panic("unrecognized H2 expression kind")
	}
}

// normalizeH2Statement converts an arena-backed statement into
// the common representation.
//
// Binding statements retrieve their names from source spans.
// SAY statements store their expression NodeID directly in Data.
func normalizeH2Statement(
	arena *ast.Arena,
	id ast.NodeID,
) normalizedNode {
	node := arena.Get(id)

	switch node.Kind {
	case ast.NodeLetStatement, ast.NodeChangeStatement:
		kind := "let"
		if node.Kind == ast.NodeChangeStatement {
			kind = "change"
		}

		payload := arena.BindingPayload(node.Data)
		return normalizedNode{
			Kind:  kind,
			Value: arena.SourceText(payload.Name),
			Span:  node.Span,
			Children: []normalizedNode{
				normalizeH2Expression(arena, payload.Value),
			},
		}

	case ast.NodeSayStatement:
		return normalizedNode{
			Kind: "say",
			Span: node.Span,
			Children: []normalizedNode{
				normalizeH2Expression(arena, ast.NodeID(node.Data)),
			},
		}

	default:
		panic("unrecognized H2 statement kind")
	}
}

// TestFlatParserDifferential compares successful programs parsed
// independently by M1 and H2.
//
// Both parsers consume the same source text through their existing
// streaming lexer. The resulting ASTs are normalized and compared
// for structure, values, operators, statement order, and byte spans.
func TestFlatParserDifferential(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"empty program", ""},
		{"blank lines", "\n\n\n"},
		{"integer declaration", "let age be 20"},
		{"string declaration", "let message be \"hello\\nworld\""},
		{"assignment", "change age to 21"},
		{"output", "say age"},
		{"operator precedence", "say 2 + 3 * 4"},
		{"left associativity", "say 10 - 3 - 2"},
		{"exponent associativity", "say 2 ** 3 ** 2"},
		{"unary precedence", "say -2 ** 3"},
		{"grouping", "say (2 + 3) * 4"},
		{"nested grouping", "say ((2 + 3))"},
		{"unary grouping", "say -(2 + 3)"},
		{
			"multiple statements",
			"let age be 20\nchange age to age + 1\nsay age",
		},
		{
			"blank lines between statements",
			"\nlet x be 1\n\nsay x\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m1 := newStreamingParser(tt.source)
			m1Program, m1Diagnostic := m1.parseProgram()

			h2 := newFlatStatementTestParser(tt.source)
			h2Program, h2Diagnostic := h2.parseFlatProgram()

			if m1Diagnostic != nil {
				t.Fatalf("M1 rejected fixture: %+v", m1Diagnostic)
			}
			if h2Diagnostic != nil {
				t.Fatalf("H2 rejected fixture: %+v", h2Diagnostic)
			}
			if m1Program == nil || h2Program == nil {
				t.Fatal("expected non-nil programs from both parsers")
			}

			wantProgramSpan := normalizedSpan(tt.source, m1Program.Span())

			if h2Program.Span != wantProgramSpan {
				t.Errorf(
					"program span mismatch: M1=%+v, H2=%+v",
					wantProgramSpan,
					h2Program.Span,
				)
			}

			var m1Nodes []normalizedNode
			for _, statement := range m1Program.Statements {
				m1Nodes = append(
					m1Nodes,
					normalizeM1Statement(tt.source, statement),
				)
			}

			var h2Nodes []normalizedNode
			for _, id := range h2Program.Arena.StatementRange(
				h2Program.Statements,
			) {
				h2Nodes = append(
					h2Nodes,
					normalizeH2Statement(h2Program.Arena, id),
				)
			}

			if !reflect.DeepEqual(m1Nodes, h2Nodes) {
				t.Errorf(
					"AST mismatch\nM1: %#v\nH2: %#v",
					m1Nodes,
					h2Nodes,
				)
			}

			if m1.peek().Kind != h2.peek().Kind {
				t.Errorf(
					"remaining token mismatch: M1=%v, H2=%v",
					m1.peek().Kind,
					h2.peek().Kind,
				)
			}
		})
	}
}

// TestFlatParserDifferentialDiagnostics compares syntax diagnostics
// emitted by M1 and H2 for the same malformed source programs.
//
// Messages, expected and actual tokens, and source locations must
// match exactly. Both parsers must reject the fixture.
func TestFlatParserDifferentialDiagnostics(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{"missing declaration name", "let be 20"},
		{"missing be", "let age 20"},
		{"missing initializer", "let age be"},
		{"missing assignment name", "change to 2"},
		{"missing to", "change age 2"},
		{"missing assigned value", "change age to"},
		{"missing say expression", "say"},
		{"invalid expression", "say (2 + )"},
		{"missing closing parenthesis", "say (2 + 3"},
		{"missing statement separator", "let age be 20 say age"},
		{"unsupported statement", "42"},
		{
			"failure after valid statements",
			"let age be 20\nsay age\nchange age to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m1 := newStreamingParser(tt.source)
			m1Program, m1Diagnostic := m1.parseProgram()

			h2 := newFlatStatementTestParser(tt.source)
			h2Program, h2Diagnostic := h2.parseFlatProgram()

			if m1Diagnostic == nil || h2Diagnostic == nil {
				t.Fatalf(
					"expected diagnostics from both parsers: M1=%+v, H2=%+v",
					m1Diagnostic,
					h2Diagnostic,
				)
			}

			if m1Program != nil || h2Program != nil {
				t.Fatal("expected nil programs on syntax failure")
			}

			if m1Diagnostic.Message != h2Diagnostic.Message {
				t.Errorf(
					"message mismatch: M1=%q, H2=%q",
					m1Diagnostic.Message,
					h2Diagnostic.Message,
				)
			}

			if !reflect.DeepEqual(
				m1Diagnostic.Expected,
				h2Diagnostic.Expected,
			) {
				t.Errorf(
					"expected-token mismatch: M1=%v, H2=%v",
					m1Diagnostic.Expected,
					h2Diagnostic.Expected,
				)
			}

			if !reflect.DeepEqual(
				m1Diagnostic.Actual,
				h2Diagnostic.Actual,
			) {
				t.Errorf(
					"actual-token mismatch: M1=%v, H2=%v",
					m1Diagnostic.Actual,
					h2Diagnostic.Actual,
				)
			}

			if m1Diagnostic.Location != h2Diagnostic.Location {
				t.Errorf(
					"location mismatch: M1=%+v, H2=%+v",
					m1Diagnostic.Location,
					h2Diagnostic.Location,
				)
			}

			if m1.peek().Kind != h2.peek().Kind {
				t.Errorf(
					"remaining token mismatch: M1=%v, H2=%v",
					m1.peek().Kind,
					h2.peek().Kind,
				)
			}
		})
	}
}
