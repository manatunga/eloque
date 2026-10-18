package lexer

import "testing"

func TestTokenFields(t *testing.T) {
	token := Token{Kind: TokenLet, Start: 0, End: 3}

	if token.Kind != TokenLet {
		t.Errorf("expected kind %q, got %q", TokenLet, token.Kind)
	}
	if token.Start != 0 {
		t.Errorf("expected start offset 0, got %d", token.Start)
	}
	if token.End != 3 {
		t.Errorf("expected end offset 3, got %d", token.End)
	}
	if got := testTokenLexeme("let age", token); got != "let" {
		t.Errorf("expected lexeme %q, got %q", "let", got)
	}
	span := testTokenSpan("let age", token)
	if span.Start != (Position{Line: 1, Column: 1}) || span.End != (Position{Line: 1, Column: 4}) {
		t.Errorf("expected span 1:1–1:4, got %+v", span)
	}
}

func TestLookupIdentifier(t *testing.T) {
	testCases := []struct {
		input    string
		expected TokenKind
	}{
		{"let", TokenLet},
		{"be", TokenBe},
		{"say", TokenSay},
		{"age", TokenIdent},
		{"letter", TokenIdent},
		{"Let", TokenIdent},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := lookupIdentifier(tc.input)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}
