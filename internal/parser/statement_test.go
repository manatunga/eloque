package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// statementLegacyToken preserves the original test fixtures during the
// compact-token migration. Production tokens are created with source offsets.
type statementLegacyToken struct {
	Kind   lexer.TokenKind
	Lexeme string
	Span   source.Span
}

func newStatementLegacyParser(old []statementLegacyToken) *Parser {
	var text strings.Builder
	tokens := make([]lexer.Token, 0, len(old))
	for _, item := range old {
		// Explicit spans in the original fixtures are 1-based line/column.
		if item.Span.Start.Line > 0 {
			for line := 1; line < item.Span.Start.Line; line++ {
				// Only advance to the requested line if necessary.
				if strings.Count(text.String(), "\n") < line {
					text.WriteByte('\n')
				}
			}
			lastNewline := strings.LastIndex(text.String(), "\n")
			column := text.Len() - lastNewline
			for column < item.Span.Start.Column {
				text.WriteByte(' ')
				column++
			}
		} else if text.Len() > 0 && item.Kind != lexer.TokenNewline && item.Kind != lexer.TokenEOF {
			text.WriteByte(' ')
		}
		start := uint32(text.Len())
		lexeme := item.Lexeme
		if lexeme == "" {
			switch item.Kind {
			case lexer.TokenIdent:
				lexeme = "x"
			case lexer.TokenInt:
				lexeme = "0"
			case lexer.TokenNewline:
				lexeme = "\n"
			case lexer.TokenEOF:
				lexeme = ""
			}
		}
		if item.Span.Start.Line > 0 && item.Kind == lexer.TokenEOF {
			// EOF has no source text; its position is represented by start.
			lexeme = ""
		}
		text.WriteString(lexeme)
		end := uint32(text.Len())
		if item.Span.Start.Line > 0 && item.Span.End.Line == item.Span.Start.Line && item.Span.End.Column > item.Span.Start.Column {
			width := uint32(item.Span.End.Column - item.Span.Start.Column)
			if end < start+width {
				text.WriteString(strings.Repeat(" ", int(start+width-end)))
				end = start + width
			}
		}
		tokens = append(tokens, lexer.Token{Kind: item.Kind, Start: start, End: end})
	}
	return New(text.String(), tokens)
}

func TestParseLetStatement(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []statementLegacyToken
		wantName   string
		wantValue  string
		wantNext   lexer.TokenKind
		wantCursor int
	}{
		{
			name: "integer initializer",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenEOF},
			},
			wantName:   "age",
			wantValue:  "20",
			wantNext:   lexer.TokenEOF,
			wantCursor: 4,
		},
		{
			name: "arithmetic initializer",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "result"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "4"},
				{Kind: lexer.TokenEOF},
			},
			wantName:   "result",
			wantValue:  "(PLUS 2 (STAR 3 4))",
			wantNext:   lexer.TokenEOF,
			wantCursor: 8,
		},
		{
			name: "newline remains unconsumed",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "count"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenInt, Lexeme: "5"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantName:   "count",
			wantValue:  "5",
			wantNext:   lexer.TokenNewline,
			wantCursor: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseLetStatement()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			declaration, ok := statement.(*ast.LetStatement)
			if !ok {
				t.Fatalf("statement type = %T, want *ast.LetStatement", statement)
			}

			if declaration.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", declaration.Name, tt.wantName)
			}

			if got := formatExpressionTree(declaration.Value); got != tt.wantValue {
				t.Errorf("initializer tree = %q, want %q", got, tt.wantValue)
			}

			if p.current != tt.wantCursor {
				t.Errorf("current = %d, want %d", p.current, tt.wantCursor)
			}

			if p.peek().Kind != tt.wantNext {
				t.Errorf("next token = %v, want %v", p.peek().Kind, tt.wantNext)
			}
		})
	}
}

func TestParseLetStatementGroupingSpan(t *testing.T) {
	p := newStatementLegacyParser([]statementLegacyToken{
		{
			Kind: lexer.TokenLet, Lexeme: "let",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: 4},
			},
		},
		{Kind: lexer.TokenIdent, Lexeme: "result"},
		{Kind: lexer.TokenBe, Lexeme: "be"},
		{
			Kind: lexer.TokenLeftParen, Lexeme: "(",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 15},
				End:   source.Position{Line: 1, Column: 16},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "2",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 16},
				End:   source.Position{Line: 1, Column: 17},
			},
		},
		{Kind: lexer.TokenPlus, Lexeme: "+"},
		{
			Kind: lexer.TokenInt, Lexeme: "3",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 20},
				End:   source.Position{Line: 1, Column: 21},
			},
		},
		{
			Kind: lexer.TokenRightParen, Lexeme: ")",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 21},
				End:   source.Position{Line: 1, Column: 22},
			},
		},
		{Kind: lexer.TokenEOF},
	})

	statement, diagnostic := p.parseLetStatement()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	declaration, ok := statement.(*ast.LetStatement)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.LetStatement", statement)
	}

	wantStatementSpan := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 22},
	}
	wantExpressionSpan := source.Span{
		Start: source.Position{Line: 1, Column: 16},
		End:   source.Position{Line: 1, Column: 21},
	}

	if declaration.Span() != wantStatementSpan {
		t.Errorf("statement span = %+v, want %+v",
			declaration.Span(), wantStatementSpan)
	}

	if declaration.Value.Span() != wantExpressionSpan {
		t.Errorf("initializer AST span = %+v, want %+v",
			declaration.Value.Span(), wantExpressionSpan)
	}

	if p.peek().Kind != lexer.TokenEOF {
		t.Errorf("next token = %v, want EOF", p.peek().Kind)
	}
}

func TestParseLetStatementDiagnostics(t *testing.T) {
	tests := []struct {
		name         string
		tokens       []statementLegacyToken
		wantMessage  string
		wantExpected []lexer.TokenKind
		wantActual   lexer.TokenKind
		wantCursor   int
	}{
		{
			name: "missing identifier",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage:  "expected a variable name after 'let'",
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
			wantActual:   lexer.TokenBe,
			wantCursor:   1,
		},
		{
			name: "missing be keyword",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage:  "expected 'be' after variable name to initialize a variable value",
			wantExpected: []lexer.TokenKind{lexer.TokenBe},
			wantActual:   lexer.TokenInt,
			wantCursor:   2,
		},
		{
			name: "missing initializer at EOF",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage: "expected an expression",
			wantExpected: []lexer.TokenKind{
				lexer.TokenInt, lexer.TokenString, lexer.TokenIdent,
				lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus,
			},
			wantActual: lexer.TokenEOF,
			wantCursor: 3,
		},
		{
			name: "missing initializer at newline",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage: "expected an expression",
			wantExpected: []lexer.TokenKind{
				lexer.TokenInt, lexer.TokenString, lexer.TokenIdent,
				lexer.TokenLeftParen, lexer.TokenPlus, lexer.TokenMinus,
			},
			wantActual: lexer.TokenNewline,
			wantCursor: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseLetStatement()
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if statement != nil {
				t.Errorf("statement = %T, want nil", statement)
			}

			if diagnostic.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, tt.wantMessage)
			}

			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf("Expected = %v, want %v",
					diagnostic.Expected, tt.wantExpected)
			}

			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v",
					diagnostic.Actual, tt.wantActual)
			}

			if p.current != tt.wantCursor {
				t.Errorf("current = %d, want %d", p.current, tt.wantCursor)
			}
		})
	}
}

func TestParseChangeStatement(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []statementLegacyToken
		wantName   string
		wantValue  string
		wantNext   lexer.TokenKind
		wantCursor int
	}{
		{
			name: "integer assignment",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenInt, Lexeme: "21"},
				{Kind: lexer.TokenEOF},
			},
			wantName:   "age",
			wantValue:  "21",
			wantNext:   lexer.TokenEOF,
			wantCursor: 4,
		},
		{
			name: "arithmetic reassignment",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "1"},
				{Kind: lexer.TokenEOF},
			},
			wantName:   "age",
			wantValue:  "(PLUS age 1)",
			wantNext:   lexer.TokenEOF,
			wantCursor: 6,
		},
		{
			name: "newline remains unconsumed",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "count"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenInt, Lexeme: "5"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantName:   "count",
			wantValue:  "5",
			wantNext:   lexer.TokenNewline,
			wantCursor: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseChangeStatement()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			assignment, ok := statement.(*ast.ChangeStatement)
			if !ok {
				t.Fatalf("statement type = %T, want *ast.ChangeStatement", statement)
			}

			if assignment.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", assignment.Name, tt.wantName)
			}

			if got := formatExpressionTree(assignment.Value); got != tt.wantValue {
				t.Errorf("assignment expression = %q, want %q", got, tt.wantValue)
			}

			if p.current != tt.wantCursor {
				t.Errorf("current = %d, want %d", p.current, tt.wantCursor)
			}

			if p.peek().Kind != tt.wantNext {
				t.Errorf("next token = %v, want %v", p.peek().Kind, tt.wantNext)
			}
		})
	}
}

func TestParseChangeStatementGroupingSpan(t *testing.T) {
	p := newStatementLegacyParser([]statementLegacyToken{
		{
			Kind: lexer.TokenChange, Lexeme: "change",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: 7},
			},
		},
		{Kind: lexer.TokenIdent, Lexeme: "result"},
		{Kind: lexer.TokenTo, Lexeme: "to"},
		{
			Kind: lexer.TokenLeftParen, Lexeme: "(",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 18},
				End:   source.Position{Line: 1, Column: 19},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "2",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 19},
				End:   source.Position{Line: 1, Column: 20},
			},
		},
		{Kind: lexer.TokenPlus, Lexeme: "+"},
		{
			Kind: lexer.TokenInt, Lexeme: "3",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 23},
				End:   source.Position{Line: 1, Column: 24},
			},
		},
		{
			Kind: lexer.TokenRightParen, Lexeme: ")",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 24},
				End:   source.Position{Line: 1, Column: 25},
			},
		},
		{Kind: lexer.TokenEOF},
	})

	statement, diagnostic := p.parseChangeStatement()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	assignment, ok := statement.(*ast.ChangeStatement)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.ChangeStatement", statement)
	}

	wantStatementSpan := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 25},
	}
	wantExpressionSpan := source.Span{
		Start: source.Position{Line: 1, Column: 19},
		End:   source.Position{Line: 1, Column: 24},
	}

	if assignment.Span() != wantStatementSpan {
		t.Errorf("statement span = %+v, want %+v",
			assignment.Span(), wantStatementSpan)
	}

	if assignment.Value.Span() != wantExpressionSpan {
		t.Errorf("expression span = %+v, want %+v",
			assignment.Value.Span(), wantExpressionSpan)
	}

	if p.peek().Kind != lexer.TokenEOF {
		t.Errorf("next token = %v, want EOF", p.peek().Kind)
	}
}

func TestParseChangeStatementDiagnostics(t *testing.T) {
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
		tokens       []statementLegacyToken
		wantMessage  string
		wantExpected []lexer.TokenKind
		wantActual   lexer.TokenKind
		wantCursor   int
	}{
		{
			name: "missing identifier",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage:  "expected a variable name after 'change'",
			wantExpected: []lexer.TokenKind{lexer.TokenIdent},
			wantActual:   lexer.TokenTo,
			wantCursor:   1,
		},
		{
			name: "missing to keyword",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage:  "expected 'to' after variable name to assign a new variable value",
			wantExpected: []lexer.TokenKind{lexer.TokenTo},
			wantActual:   lexer.TokenInt,
			wantCursor:   2,
		},
		{
			name: "missing expression at EOF",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage:  "expected an expression",
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenEOF,
			wantCursor:   3,
		},
		{
			name: "missing expression at newline",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantMessage:  "expected an expression",
			wantExpected: expressionStarters,
			wantActual:   lexer.TokenNewline,
			wantCursor:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseChangeStatement()
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if statement != nil {
				t.Errorf("statement = %T, want nil", statement)
			}

			if diagnostic.Message != tt.wantMessage {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, tt.wantMessage)
			}

			if !reflect.DeepEqual(diagnostic.Expected, tt.wantExpected) {
				t.Errorf("Expected = %v, want %v",
					diagnostic.Expected, tt.wantExpected)
			}

			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v",
					diagnostic.Actual, tt.wantActual)
			}

			if p.current != tt.wantCursor {
				t.Errorf("current = %d, want %d", p.current, tt.wantCursor)
			}
		})
	}
}

func TestParseSayStatement(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []statementLegacyToken
		wantValue  string
		wantNext   lexer.TokenKind
		wantCursor int
	}{
		{
			name: "string literal",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenString, Lexeme: `"Hello, world!"`},
				{Kind: lexer.TokenEOF},
			},
			wantValue:  `"Hello, world!"`,
			wantNext:   lexer.TokenEOF,
			wantCursor: 2,
		},
		{
			name: "identifier",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenEOF},
			},
			wantValue:  "age",
			wantNext:   lexer.TokenEOF,
			wantCursor: 2,
		},
		{
			name: "arithmetic precedence",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenInt, Lexeme: "2"},
				{Kind: lexer.TokenPlus, Lexeme: "+"},
				{Kind: lexer.TokenInt, Lexeme: "3"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "4"},
				{Kind: lexer.TokenEOF},
			},
			wantValue:  "(PLUS 2 (STAR 3 4))",
			wantNext:   lexer.TokenEOF,
			wantCursor: 6,
		},
		{
			name: "newline remains unconsumed",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenIdent, Lexeme: "result"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantValue:  "result",
			wantNext:   lexer.TokenNewline,
			wantCursor: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseSayStatement()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			output, ok := statement.(*ast.SayStatement)
			if !ok {
				t.Fatalf("statement type = %T, want *ast.SayStatement", statement)
			}

			if got := formatExpressionTree(output.Value); got != tt.wantValue {
				t.Errorf("expression = %q, want %q", got, tt.wantValue)
			}

			if p.current != tt.wantCursor {
				t.Errorf("current = %d, want %d", p.current, tt.wantCursor)
			}

			if p.peek().Kind != tt.wantNext {
				t.Errorf("next token = %v, want %v", p.peek().Kind, tt.wantNext)
			}
		})
	}
}

func TestParseSayStatementGroupingSpan(t *testing.T) {
	p := newStatementLegacyParser([]statementLegacyToken{
		{
			Kind: lexer.TokenSay, Lexeme: "say",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 1},
				End:   source.Position{Line: 1, Column: 4},
			},
		},
		{
			Kind: lexer.TokenLeftParen, Lexeme: "(",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 5},
				End:   source.Position{Line: 1, Column: 6},
			},
		},
		{
			Kind: lexer.TokenInt, Lexeme: "2",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 6},
				End:   source.Position{Line: 1, Column: 7},
			},
		},
		{Kind: lexer.TokenPlus, Lexeme: "+"},
		{
			Kind: lexer.TokenInt, Lexeme: "3",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 10},
				End:   source.Position{Line: 1, Column: 11},
			},
		},
		{
			Kind: lexer.TokenRightParen, Lexeme: ")",
			Span: source.Span{
				Start: source.Position{Line: 1, Column: 11},
				End:   source.Position{Line: 1, Column: 12},
			},
		},
		{Kind: lexer.TokenEOF},
	})

	statement, diagnostic := p.parseSayStatement()
	if diagnostic != nil {
		t.Fatalf("unexpected diagnostic: %+v", diagnostic)
	}

	output, ok := statement.(*ast.SayStatement)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.SayStatement", statement)
	}

	wantStatementSpan := source.Span{
		Start: source.Position{Line: 1, Column: 1},
		End:   source.Position{Line: 1, Column: 12},
	}
	wantExpressionSpan := source.Span{
		Start: source.Position{Line: 1, Column: 6},
		End:   source.Position{Line: 1, Column: 11},
	}

	if output.Span() != wantStatementSpan {
		t.Errorf("statement span = %+v, want %+v",
			output.Span(), wantStatementSpan)
	}

	if output.Value.Span() != wantExpressionSpan {
		t.Errorf("expression span = %+v, want %+v",
			output.Value.Span(), wantExpressionSpan)
	}

	if p.peek().Kind != lexer.TokenEOF {
		t.Errorf("next token = %v, want EOF", p.peek().Kind)
	}
}

func TestParseSayStatementDiagnostics(t *testing.T) {
	tests := []struct {
		name       string
		tokens     []statementLegacyToken
		wantActual lexer.TokenKind
		wantCursor int
	}{
		{
			name: "missing expression at EOF",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenEOF},
			},
			wantActual: lexer.TokenEOF,
			wantCursor: 1,
		},
		{
			name: "missing expression at newline",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenNewline, Lexeme: "\n"},
				{Kind: lexer.TokenEOF},
			},
			wantActual: lexer.TokenNewline,
			wantCursor: 1,
		},
		{
			name: "invalid expression starter",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenStar, Lexeme: "*"},
				{Kind: lexer.TokenInt, Lexeme: "5"},
				{Kind: lexer.TokenEOF},
			},
			wantActual: lexer.TokenStar,
			wantCursor: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseSayStatement()
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if statement != nil {
				t.Errorf("statement = %T, want nil", statement)
			}

			if diagnostic.Message != "expected an expression" {
				t.Errorf("Message = %q, want %q",
					diagnostic.Message, "expected an expression")
			}

			if diagnostic.Actual == nil || *diagnostic.Actual != tt.wantActual {
				t.Errorf("Actual = %v, want %v",
					diagnostic.Actual, tt.wantActual)
			}

			if p.current != tt.wantCursor {
				t.Errorf("current = %d, want %d", p.current, tt.wantCursor)
			}
		})
	}
}

func TestParseStatementDispatch(t *testing.T) {
	tests := []struct {
		name     string
		tokens   []statementLegacyToken
		wantType ast.Statement
	}{
		{
			name: "let declaration",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenLet, Lexeme: "let"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenBe, Lexeme: "be"},
				{Kind: lexer.TokenInt, Lexeme: "20"},
				{Kind: lexer.TokenEOF},
			},
			wantType: &ast.LetStatement{},
		},
		{
			name: "change assignment",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenChange, Lexeme: "change"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenTo, Lexeme: "to"},
				{Kind: lexer.TokenInt, Lexeme: "21"},
				{Kind: lexer.TokenEOF},
			},
			wantType: &ast.ChangeStatement{},
		},
		{
			name: "say output",
			tokens: []statementLegacyToken{
				{Kind: lexer.TokenSay, Lexeme: "say"},
				{Kind: lexer.TokenIdent, Lexeme: "age"},
				{Kind: lexer.TokenEOF},
			},
			wantType: &ast.SayStatement{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := newStatementLegacyParser(tt.tokens)

			statement, diagnostic := p.parseStatement()
			if diagnostic != nil {
				t.Fatalf("unexpected diagnostic: %+v", diagnostic)
			}

			if reflect.TypeOf(statement) != reflect.TypeOf(tt.wantType) {
				t.Errorf("statement type = %T, want %T",
					statement, tt.wantType)
			}

			if p.peek().Kind != lexer.TokenEOF {
				t.Errorf("next token = %v, want EOF", p.peek().Kind)
			}
		})
	}
}

func TestParseStatementUnsupportedStarter(t *testing.T) {
	tests := []struct {
		name string
		kind lexer.TokenKind
	}{
		{name: "identifier", kind: lexer.TokenIdent},
		{name: "integer", kind: lexer.TokenInt},
		{name: "newline", kind: lexer.TokenNewline},
		{name: "EOF", kind: lexer.TokenEOF},
	}

	wantExpected := []lexer.TokenKind{
		lexer.TokenLet,
		lexer.TokenChange,
		lexer.TokenSay,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			span := source.Span{
				Start: source.Position{Line: 2, Column: 3},
				End:   source.Position{Line: 2, Column: 4},
			}

			if tt.kind == lexer.TokenNewline {
				span.End = source.Position{Line: 3, Column: 1}
			}

			p := newStatementLegacyParser([]statementLegacyToken{
				{Kind: tt.kind, Span: span},
				{Kind: lexer.TokenEOF},
			})

			statement, diagnostic := p.parseStatement()
			if diagnostic == nil {
				t.Fatal("expected diagnostic, got nil")
			}

			if statement != nil {
				t.Errorf("statement = %T, want nil", statement)
			}

			if diagnostic.Message != "expected a statement ('let', 'change' or 'say')" {
				t.Errorf("Message = %q", diagnostic.Message)
			}

			if diagnostic.Location != span {
				t.Errorf("Location = %+v, want %+v", diagnostic.Location, span)
			}

			if !reflect.DeepEqual(diagnostic.Expected, wantExpected) {
				t.Errorf("Expected = %v, want %v",
					diagnostic.Expected, wantExpected)
			}

			if diagnostic.Actual == nil || *diagnostic.Actual != tt.kind {
				t.Errorf("Actual = %v, want %v",
					diagnostic.Actual, tt.kind)
			}

			if p.current != 0 {
				t.Errorf("current = %d, want 0", p.current)
			}
		})
	}
}
