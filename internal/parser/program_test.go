package parser

import (
	"reflect"
	"testing"

	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// programTestParser lexes a real source fixture so token offsets and
// diagnostic locations stay consistent with the source text.
func programTestParser(text string) *Parser {
	l := lexer.New(text)
	tokens := make([]lexer.Token, 0, 16)
	for {
		tok, _ := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Kind == lexer.TokenEOF {
			break
		}
	}
	return New(text, tokens)
}

func TestParseProgramEmpty(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		wantCursor   int
		wantPosition source.Position
	}{
		{"immediate EOF", "", 0, source.Position{Line: 1, Column: 1}},
		{"blank lines only", "\n\n", 2, source.Position{Line: 3, Column: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := programTestParser(tt.text)
			program, diagnostic := p.parseProgram()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			if program == nil {
				t.Fatal("expected program, got nil")
			}
			if len(program.Statements) != 0 {
				t.Errorf("statement count = %d, want 0", len(program.Statements))
			}
			wantSpan := source.Span{Start: tt.wantPosition, End: tt.wantPosition}
			if program.Span() != wantSpan {
				t.Errorf("program span = %+v, want %+v", program.Span(), wantSpan)
			}
			if p.current != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", p.current, tt.wantCursor)
			}
			if p.peek().Kind != lexer.TokenEOF {
				t.Errorf("next token = %v, want EOF", p.peek().Kind)
			}
		})
	}
}

func TestParseProgramStatements(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantTypes  []ast.Statement
		wantSpan   source.Span
		wantCursor int
	}{
		{
			name:       "single statement at EOF",
			text:       "say 42",
			wantTypes:  []ast.Statement{&ast.SayStatement{}},
			wantSpan:   source.Span{Start: source.Position{Line: 1, Column: 1}, End: source.Position{Line: 1, Column: 7}},
			wantCursor: 2,
		},
		{
			name:       "single statement with final newline",
			text:       "say 42\n",
			wantTypes:  []ast.Statement{&ast.SayStatement{}},
			wantSpan:   source.Span{Start: source.Position{Line: 1, Column: 1}, End: source.Position{Line: 1, Column: 7}},
			wantCursor: 3,
		},
		{
			name:       "mixed statements with blank lines",
			text:       "\nlet age be 20\n\nchange age to 21\nsay age\n\n",
			wantTypes:  []ast.Statement{&ast.LetStatement{}, &ast.ChangeStatement{}, &ast.SayStatement{}},
			wantSpan:   source.Span{Start: source.Position{Line: 2, Column: 1}, End: source.Position{Line: 5, Column: 8}},
			wantCursor: 16,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := programTestParser(tt.text)
			program, diagnostic := p.parseProgram()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}
			if program == nil {
				t.Fatal("expected program, got nil")
			}
			if len(program.Statements) != len(tt.wantTypes) {
				t.Fatalf("statement count = %d, want %d", len(program.Statements), len(tt.wantTypes))
			}
			for i, want := range tt.wantTypes {
				if reflect.TypeOf(program.Statements[i]) != reflect.TypeOf(want) {
					t.Errorf("statement %d type = %T, want %T", i, program.Statements[i], want)
				}
			}
			if program.Span() != tt.wantSpan {
				t.Errorf("program span = %+v, want %+v", program.Span(), tt.wantSpan)
			}
			if p.current != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", p.current, tt.wantCursor)
			}
			if p.peek().Kind != lexer.TokenEOF {
				t.Errorf("next token = %v, want EOF", p.peek().Kind)
			}
		})
	}
}

func TestParseProgramGroupedExpressionSpan(t *testing.T) {
	p := programTestParser("say (2 + 3)")
	program, diagnostic := p.parseProgram()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}
	wantSpan := source.Span{Start: source.Position{Line: 1, Column: 1}, End: source.Position{Line: 1, Column: 12}}
	if program.Span() != wantSpan {
		t.Errorf("program span = %+v, want %+v", program.Span(), wantSpan)
	}
	if len(program.Statements) != 1 {
		t.Fatalf("statement count = %d, want 1", len(program.Statements))
	}
	if program.Statements[0].Span() != wantSpan {
		t.Errorf("statement span = %+v, want %+v", program.Statements[0].Span(), wantSpan)
	}
}

func TestParseProgramDiagnostics(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		wantMessage  string
		wantExpected []lexer.TokenKind
		wantActual   lexer.TokenKind
		wantLocation source.Span
		wantCursor   int
	}{
		{
			name:         "missing separator before say",
			text:         "let age be 20 say age",
			wantMessage:  "expected a newline or end of file after statement",
			wantExpected: []lexer.TokenKind{lexer.TokenNewline, lexer.TokenEOF},
			wantActual:   lexer.TokenSay,
			wantLocation: source.Span{Start: source.Position{Line: 1, Column: 15}, End: source.Position{Line: 1, Column: 18}},
			wantCursor:   4,
		},
		{
			name:         "unsupported statement starter",
			text:         "\n  age",
			wantMessage:  "expected a statement ('let', 'change' or 'say')",
			wantExpected: []lexer.TokenKind{lexer.TokenLet, lexer.TokenChange, lexer.TokenSay},
			wantActual:   lexer.TokenIdent,
			wantLocation: source.Span{Start: source.Position{Line: 2, Column: 3}, End: source.Position{Line: 2, Column: 6}},
			wantCursor:   0,
		},
		{
			name:         "missing let identifier",
			text:         "\n\nlet be 20",
			wantMessage:  "expected a variable name after 'let'",
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
			wantActual:   lexer.TokenBe,
			wantLocation: source.Span{Start: source.Position{Line: 3, Column: 5}, End: source.Position{Line: 3, Column: 7}},
			wantCursor:   1,
		},
		{
			name:         "missing change to keyword",
			text:         "\n\n\nchange age 21",
			wantMessage:  "expected 'to' after variable name to assign a new variable value",
			wantExpected: []lexer.TokenKind{lexer.TokenTo},
			wantActual:   lexer.TokenInt,
			wantLocation: source.Span{Start: source.Position{Line: 4, Column: 12}, End: source.Position{Line: 4, Column: 14}},
			wantCursor:   2,
		},
		{
			name:         "missing say expression",
			text:         "\n\n\n\nsay\n",
			wantMessage:  "expected an expression",
			wantExpected: []lexer.TokenKind{lexer.TokenInt, lexer.TokenString, lexer.TokenIdent, lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus},
			wantActual:   lexer.TokenNewline,
			wantLocation: source.Span{Start: source.Position{Line: 5, Column: 4}, End: source.Position{Line: 6, Column: 1}},
			wantCursor:   1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the same intentional token stream as the legacy test:
			// leading line breaks establish locations but are not tokens.
			text := tt.text
			var tokens []lexer.Token
			switch tt.name {
			case "unsupported statement starter":
				tokens = []lexer.Token{{Kind: lexer.TokenIdent, Start: 3, End: 6}, {Kind: lexer.TokenEOF, Start: 6, End: 6}}
			case "missing let identifier":
				tokens = []lexer.Token{{Kind: lexer.TokenLet, Start: 2, End: 5}, {Kind: lexer.TokenBe, Start: 6, End: 8}, {Kind: lexer.TokenInt, Start: 9, End: 11}, {Kind: lexer.TokenEOF, Start: 11, End: 11}}
			case "missing change to keyword":
				tokens = []lexer.Token{{Kind: lexer.TokenChange, Start: 3, End: 9}, {Kind: lexer.TokenIdent, Start: 10, End: 13}, {Kind: lexer.TokenInt, Start: 14, End: 16}, {Kind: lexer.TokenEOF, Start: 16, End: 16}}
			case "missing say expression":
				tokens = []lexer.Token{{Kind: lexer.TokenSay, Start: 4, End: 7}, {Kind: lexer.TokenNewline, Start: 7, End: 8}, {Kind: lexer.TokenEOF, Start: 8, End: 8}}
			default:
				l := lexer.New(text)
				for {
					tok, _ := l.NextToken()
					tokens = append(tokens, tok)
					if tok.Kind == lexer.TokenEOF {
						break
					}
				}
			}
			p := New(text, tokens)
			program, diagnostic := p.parseProgram()
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}
			if program != nil {
				t.Errorf("program = %+v, want nil", program)
			}
			if diagnostic.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q", diagnostic.Message, tt.wantMessage)
			}
			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf("Expected = %v, want %v", diagnostic.Expected, tt.wantExpected)
			}
			if diagnostic.Actual == nil {
				t.Error("Actual = nil, want a token kind")
			} else if *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v", *diagnostic.Actual, tt.wantActual)
			}
			if diagnostic.Location != tt.wantLocation {
				t.Errorf("Location = %+v, want %+v", diagnostic.Location, tt.wantLocation)
			}
			if p.current != tt.wantCursor {
				t.Errorf("cursor = %d, want %d", p.current, tt.wantCursor)
			}
			if p.peek().Kind != tt.wantActual {
				t.Errorf("next token = %v, want %v", p.peek().Kind, tt.wantActual)
			}
		})
	}
}

func TestParseProgramNestedExpressionDiagnostic(t *testing.T) {
	offendingSpan := source.Span{Start: source.Position{Line: 1, Column: 10}, End: source.Position{Line: 1, Column: 11}}
	p := programTestParser("say (2 + )")
	program, diagnostic := p.parseProgram()
	if diagnostic == nil {
		t.Fatal("expected diagnostic, got nil")
	}
	if program != nil {
		t.Errorf("program = %+v, want nil", program)
	}
	if diagnostic.Message != "expected an expression" {
		t.Errorf("Message = %q, want %q", diagnostic.Message, "expected an expression")
	}
	wantExpected := []lexer.TokenKind{lexer.TokenInt, lexer.TokenString, lexer.TokenIdent, lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus}
	if !reflect.DeepEqual(diagnostic.Expected, wantExpected) {
		t.Errorf("Expected = %v, want %v", diagnostic.Expected, wantExpected)
	}
	if diagnostic.Actual == nil || *diagnostic.Actual != lexer.TokenRightParen {
		t.Errorf("Actual = %v, want %v", diagnostic.Actual, lexer.TokenRightParen)
	}
	if diagnostic.Location != offendingSpan {
		t.Errorf("Location = %+v, want %+v", diagnostic.Location, offendingSpan)
	}
	if p.current != 4 {
		t.Errorf("cursor = %d, want 4", p.current)
	}
	if p.peek().Kind != lexer.TokenRightParen {
		t.Errorf("next token = %v, want RIGHT_PAREN", p.peek().Kind)
	}
}

func TestParseProgramDiagnosticAfterValidStatements(t *testing.T) {
	// Preserve the original EOF location at line 3, column 14.
	text := "let age be 20\nsay age\nchange age to"
	eof := uint32(len(text))
	l := lexer.New(text)
	tokens := make([]lexer.Token, 0, 12)
	for {
		tok, _ := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Kind == lexer.TokenEOF {
			break
		}
	}
	tokens[len(tokens)-1] = lexer.Token{Kind: lexer.TokenEOF, Start: eof, End: eof}
	p := New(text, tokens)
	program, diagnostic := p.parseProgram()
	if diagnostic == nil {
		t.Fatal("expected diagnostic, got nil")
	}
	if program != nil {
		t.Errorf("program = %+v, want nil", program)
	}
	if diagnostic.Message != "expected an expression" {
		t.Errorf("Message = %q, want %q", diagnostic.Message, "expected an expression")
	}
	wantExpected := []lexer.TokenKind{lexer.TokenInt, lexer.TokenString, lexer.TokenIdent, lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus}
	if !reflect.DeepEqual(diagnostic.Expected, wantExpected) {
		t.Errorf("Expected = %v, want %v", diagnostic.Expected, wantExpected)
	}
	if diagnostic.Actual == nil || *diagnostic.Actual != lexer.TokenEOF {
		t.Errorf("Actual = %v, want EOF", diagnostic.Actual)
	}
	eofSpan := source.Span{Start: source.Position{Line: 3, Column: 14}, End: source.Position{Line: 3, Column: 14}}
	if diagnostic.Location != eofSpan {
		t.Errorf("Location = %+v, want %+v", diagnostic.Location, eofSpan)
	}
	if p.current != 11 {
		t.Errorf("cursor = %d, want 11", p.current)
	}
	if p.peek().Kind != lexer.TokenEOF {
		t.Errorf("next token = %v, want EOF", p.peek().Kind)
	}
}
