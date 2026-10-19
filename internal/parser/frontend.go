package parser

import (
	"fmt"

	"github.com/manatunga/eloque/internal/ast"
)

// Parse constructs an abstract syntax tree from Eloque source text.
//
// Tokens are produced incrementally by the lexer rather than buffered
// into a complete token sequence before parsing.
//
// Parse returns the resulting program on success. If lexical analysis
// or parsing fails, it returns a nil program and the corresponding error.
// Lexical failures retain their original error type, while syntax
// diagnostics are returned as errors.
func Parse(sourceText string) (*ast.Program, error) {
	parser := newStreamingParser(sourceText)
	program, diagnostic := parser.parseProgram()
	lexErr := parser.tokenSource.lexicalError()
	if lexErr != nil {
		return nil, lexErr
	}
	if diagnostic != nil {
		return nil, fmt.Errorf("syntax error: %v", diagnostic)
	}
	return program, nil
}
