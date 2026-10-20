package lexer

import "testing"

func TestLexError(t *testing.T) {
	err := &LexError{
		Position: Position{Line: 2, Column: 8},
		Message:  "unexpected character",
	}
	expected := "line 2, column 8: unexpected character"
	result := err.Error()

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}
