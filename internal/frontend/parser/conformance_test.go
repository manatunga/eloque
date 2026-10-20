package parser

import (
	"reflect"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// conformanceParser tokenizes valid source text before constructing
// a parser. Lexical errors are test setup failures, not syntax errors.
func conformanceParser(t *testing.T, text string) *Parser {
	t.Helper()

	scanner := lexer.New(text)
	tokens := make([]lexer.Token, 0, 16)

	for {
		token, err := scanner.NextToken()
		if err != nil {
			t.Fatalf("unexpected lexical error: %v", err)
		}

		tokens = append(tokens, token)
		if token.Kind == lexer.TokenEOF {
			break
		}
	}

	return New(text, tokens)
}

func conformanceSpan(
	startLine, startColumn, endLine, endColumn int,
) source.Span {
	return source.Span{
		Start: source.Position{
			Line: startLine, Column: startColumn,
		},
		End: source.Position{
			Line: endLine, Column: endColumn,
		},
	}
}

func TestProgramConformanceValid(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		wantTypes []ast.Statement
		wantSpan  source.Span
	}{
		{
			name:     "empty source",
			source:   "",
			wantSpan: conformanceSpan(1, 1, 1, 1),
		},
		{
			name:     "blank lines",
			source:   "\n\n",
			wantSpan: conformanceSpan(3, 1, 3, 1),
		},
		{
			name:      "integer output",
			source:    "say 42",
			wantTypes: []ast.Statement{&ast.SayStatement{}},
			wantSpan:  conformanceSpan(1, 1, 1, 7),
		},
		{
			name:      "variable declaration",
			source:    "let age be 20",
			wantTypes: []ast.Statement{&ast.LetStatement{}},
			wantSpan:  conformanceSpan(1, 1, 1, 14),
		},
		{
			name:      "variable reassignment",
			source:    "change age to 21",
			wantTypes: []ast.Statement{&ast.ChangeStatement{}},
			wantSpan:  conformanceSpan(1, 1, 1, 17),
		},
		{
			name:      "grouped arithmetic",
			source:    "say (2 + 3) * 4",
			wantTypes: []ast.Statement{&ast.SayStatement{}},
			wantSpan:  conformanceSpan(1, 1, 1, 16),
		},
		{
			name:      "unary exponentiation",
			source:    "say -2 ** 3",
			wantTypes: []ast.Statement{&ast.SayStatement{}},
			wantSpan:  conformanceSpan(1, 1, 1, 12),
		},
		{
			name:   "multiline program",
			source: "let x be 1\nsay x",
			wantTypes: []ast.Statement{
				&ast.LetStatement{},
				&ast.SayStatement{},
			},
			wantSpan: conformanceSpan(1, 1, 2, 6),
		},
		{
			name:   "surrounding blank lines",
			source: "\nlet x be 1\n\nsay x\n",
			wantTypes: []ast.Statement{
				&ast.LetStatement{},
				&ast.SayStatement{},
			},
			wantSpan: conformanceSpan(2, 1, 4, 6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := conformanceParser(t, tt.source)
			program, diagnostic := p.parseProgram()

			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			if program == nil {
				t.Fatal("expected a program, got nil")
			}
			if len(program.Statements) != len(tt.wantTypes) {
				t.Fatalf(
					"statement count = %d, want %d",
					len(program.Statements), len(tt.wantTypes),
				)
			}

			for i, want := range tt.wantTypes {
				if reflect.TypeOf(program.Statements[i]) != reflect.TypeOf(want) {
					t.Errorf(
						"statement %d = %T, want %T",
						i, program.Statements[i], want,
					)
				}
			}

			if got := program.Span(); got != tt.wantSpan {
				t.Errorf("span = %+v, want %+v", got, tt.wantSpan)
			}
			if p.peek().Kind != lexer.TokenEOF {
				t.Errorf("remaining token = %v, want EOF", p.peek().Kind)
			}
		})
	}
}

func TestProgramConformanceInvalid(t *testing.T) {
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
		source       string
		wantMessage  string
		wantSpan     source.Span
		wantExpected []lexer.TokenKind
		wantActual   lexer.TokenKind
	}{
		{
			name:         "missing declaration name",
			source:       "let be 20",
			wantMessage:  "expected a variable name after 'let'",
			wantSpan:     conformanceSpan(1, 5, 1, 7),
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
			wantActual:   lexer.TokenBe,
		},
		{
			name:         "missing be keyword",
			source:       "let x 20",
			wantMessage:  "expected 'be' after variable name to initialize a variable value",
			wantSpan:     conformanceSpan(1, 7, 1, 9),
			wantExpected: []lexer.TokenKind{lexer.TokenBe},
			wantActual:   lexer.TokenInt,
		},
		{
			name:         "missing to keyword",
			source:       "change x 10",
			wantMessage:  "expected 'to' after variable name to assign a new variable value",
			wantSpan:     conformanceSpan(1, 10, 1, 12),
			wantExpected: []lexer.TokenKind{lexer.TokenTo},
			wantActual:   lexer.TokenInt,
		},
		{
			name:         "missing output expression",
			source:       "say",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(1, 4, 1, 4),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenEOF,
		},
		{
			name:         "missing grouped operand",
			source:       "say (1 + )",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(1, 10, 1, 11),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenRightParen,
		},
		{
			name:         "trailing operator",
			source:       "say 1 +",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(1, 8, 1, 8),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenEOF,
		},
		{
			name:        "missing statement separator",
			source:      "say 1 say 2",
			wantMessage: "expected a newline or end of file after statement",
			wantSpan:    conformanceSpan(1, 7, 1, 10),
			wantExpected: []lexer.TokenKind{
				lexer.TokenNewline, lexer.TokenEOF,
			},
			wantActual: lexer.TokenSay,
		},
		{
			name:         "invalid second statement",
			source:       "let x be 1\nsay (2 + )",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(2, 10, 2, 11),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenRightParen,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := conformanceParser(t, tt.source)
			program, diagnostic := p.parseProgram()

			if diagnostic == nil {
				t.Fatal("expected a syntax diagnostic")
			}
			if program != nil {
				t.Errorf("malformed program returned an AST: %+v", program)
			}
			if diagnostic.Message != tt.wantMessage {
				t.Errorf(
					"message = %q, want %q",
					diagnostic.Message, tt.wantMessage,
				)
			}
			if diagnostic.Location != tt.wantSpan {
				t.Errorf(
					"location = %+v, want %+v",
					diagnostic.Location, tt.wantSpan,
				)
			}
			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf(
					"expected tokens = %v, want %v",
					diagnostic.Expected, tt.wantExpected,
				)
			}
			if diagnostic.Actual == nil {
				t.Fatal("missing actual token metadata")
			}
			if *diagnostic.Actual != tt.wantActual {
				t.Errorf(
					"actual token = %v, want %v",
					*diagnostic.Actual, tt.wantActual,
				)
			}
		})
	}
}

func TestProgramConformanceDiagnosticEdges(t *testing.T) {
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
		source       string
		wantMessage  string
		wantSpan     source.Span
		wantExpected []lexer.TokenKind
		wantActual   lexer.TokenKind
	}{
		{
			name:         "missing let identifier at EOF",
			source:       "let",
			wantMessage:  "expected a variable name after 'let'",
			wantSpan:     conformanceSpan(1, 4, 1, 4),
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
			wantActual:   lexer.TokenEOF,
		},
		{
			name:         "missing be at EOF",
			source:       "let value",
			wantMessage:  "expected 'be' after variable name to initialize a variable value",
			wantSpan:     conformanceSpan(1, 10, 1, 10),
			wantExpected: []lexer.TokenKind{lexer.TokenBe},
			wantActual:   lexer.TokenEOF,
		},
		{
			name:         "missing change identifier at EOF",
			source:       "change",
			wantMessage:  "expected a variable name after 'change'",
			wantSpan:     conformanceSpan(1, 7, 1, 7),
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
			wantActual:   lexer.TokenEOF,
		},
		{
			name:         "missing to at EOF",
			source:       "change value",
			wantMessage:  "expected 'to' after variable name to assign a new variable value",
			wantSpan:     conformanceSpan(1, 13, 1, 13),
			wantExpected: []lexer.TokenKind{lexer.TokenTo},
			wantActual:   lexer.TokenEOF,
		},
		{
			name:         "missing initializer on second line",
			source:       "say 1\nlet value be",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(2, 13, 2, 13),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenEOF,
		},
		{
			name:         "missing operand before newline",
			source:       "say 1 +\nsay 2",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(1, 8, 2, 1),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenNewline,
		},
		{
			name:         "nested grouping error on third line",
			source:       "say 1\nsay 2\nsay (3 * (4 + ))",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(3, 15, 3, 16),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenRightParen,
		},
		{
			name:         "first of two syntax errors",
			source:       "say 1 +\nlet be 2",
			wantMessage:  "expected an expression",
			wantSpan:     conformanceSpan(1, 8, 2, 1),
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenNewline,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := conformanceParser(t, tt.source)
			program, diagnostic := p.parseProgram()

			if diagnostic == nil {
				t.Fatal("expected a syntax diagnostic")
			}
			if program != nil {
				t.Errorf("malformed program returned an AST: %+v", program)
			}
			if diagnostic.Message != tt.wantMessage {
				t.Errorf(
					"message = %q, want %q",
					diagnostic.Message, tt.wantMessage,
				)
			}
			if diagnostic.Location != tt.wantSpan {
				t.Errorf(
					"location = %+v, want %+v",
					diagnostic.Location, tt.wantSpan,
				)
			}
			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf(
					"expected tokens = %v, want %v",
					diagnostic.Expected, tt.wantExpected,
				)
			}
			if diagnostic.Actual == nil {
				t.Fatal("missing actual token metadata")
			}
			if *diagnostic.Actual != tt.wantActual {
				t.Errorf(
					"actual token = %v, want %v",
					*diagnostic.Actual, tt.wantActual,
				)
			}
		})
	}
}

func TestProgramConformanceLexicalBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantMessage string
		wantLine    int
		wantColumn  int
	}{
		{
			name:        "unclosed parenthesis",
			source:      "say (1 + 2",
			wantMessage: "unclosed opening parenthesis",
			wantLine:    1,
			wantColumn:  5,
		},
		{
			name:        "unmatched closing parenthesis",
			source:      "say 1)",
			wantMessage: "unmatched closing parenthesis",
			wantLine:    1,
			wantColumn:  6,
		},
		{
			name:        "invalid integer suffix",
			source:      "say 123abc",
			wantMessage: "invalid integer literal",
			wantLine:    1,
			wantColumn:  5,
		},
		{
			name:        "leading zero",
			source:      "say 007",
			wantMessage: "integer literal cannot have leading zeros",
			wantLine:    1,
			wantColumn:  5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := lexer.New(tt.source)
			var lexicalError *lexer.LexError

			for {
				token, err := scanner.NextToken()
				if err != nil {
					var ok bool
					lexicalError, ok = err.(*lexer.LexError)
					if !ok {
						t.Fatalf("unexpected error type: %T", err)
					}
					break
				}
				if token.Kind == lexer.TokenEOF {
					break
				}
			}

			if lexicalError == nil {
				t.Fatal("expected lexical error")
			}
			if lexicalError.Message != tt.wantMessage {
				t.Errorf(
					"message = %q, want %q",
					lexicalError.Message, tt.wantMessage,
				)
			}
			if lexicalError.Position.Line != tt.wantLine ||
				lexicalError.Position.Column != tt.wantColumn {
				t.Errorf(
					"position = %d:%d, want %d:%d",
					lexicalError.Position.Line,
					lexicalError.Position.Column,
					tt.wantLine,
					tt.wantColumn,
				)
			}
		})
	}
}
