package lexer

import "fmt"

// LexError represents an error encountered while scanning Eloque source.
// Position identifies the beginning of the offending token or character,
// and Message describes the lexical problem.
type LexError struct {
	Position Position
	Message  string
}

// Error formats a lexical diagnostic with its source line, column,
// and descriptive message. It allows *LexError to satisfy Go's
// built-in error interface.
func (e *LexError) Error() string {
	return fmt.Sprintf("line %d, column %d: %s",
		e.Position.Line, e.Position.Column, e.Message)
}
