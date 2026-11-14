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

func TestLookupIdentifierConvert(t *testing.T) {
	tests := []struct {
		input string
		want  TokenKind
	}{
		{"convert", TokenConvert},
		{"convertible", TokenIdent},
		{"convert2", TokenIdent},
		{"my_convert", TokenIdent},
		{"Convert", TokenIdent},
		{"CONVERT", TokenIdent},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := lookupIdentifier(tt.input); got != tt.want {
				t.Errorf("lookupIdentifier(%q) = %s, want %s",
					tt.input, got, tt.want)
			}
		})
	}
}

func TestTokenConvertString(t *testing.T) {
	if got := TokenConvert.String(); got != "CONVERT" {
		t.Errorf("TokenConvert.String() = %q, want %q", got, "CONVERT")
	}
}

func TestTokenTypeAnnotationStrings(t *testing.T) {
	tests := []struct {
		kind TokenKind
		want string
	}{
		{TokenAs, "AS"},
		{TokenQuestion, "QUESTION"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.kind.String(); got != tt.want {
				t.Errorf("%s.String() = %q, want %q",
					tt.kind, got, tt.want)
			}
		})
	}
}
