package parser

import (
	"fmt"

	"github.com/manatunga/eloque/internal/frontend/ast"
)

// parseLegacy constructs an abstract syntax tree from Eloque source text.
//
// Tokens are produced incrementally by the lexer rather than buffered
// into a complete token sequence before parsing.
//
// Parse returns the resulting program on success. If lexical analysis
// or parsing fails, it returns a nil program and the corresponding error.
// Lexical failures retain their original error type, while syntax
// diagnostics are returned as errors.
func parseLegacy(sourceText string) (*ast.Program, error) {
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

// Parse constructs an arena-backed abstract syntax tree from
// Eloque source text.
//
// Tokens are produced incrementally by the streaming lexer rather
// than buffered into a complete token sequence.
//
// Parse returns a FlatProgram on success. Lexical errors retain
// their original error type, while syntax diagnostics are returned
// as errors with the same formatting used by Parse.
//
// Lexical errors take precedence over syntax diagnostics when both
// are encountered during parsing.
func Parse(sourceText string) (*ast.FlatProgram, error) {
	parser := NewFlatParser(sourceText)
	program, diagnostic := parser.parseFlatProgram()
	lexErr := parser.tokenSource.lexicalError()
	if lexErr != nil {
		return nil, lexErr
	}
	if diagnostic != nil {
		return nil, fmt.Errorf("syntax error: %v", diagnostic)
	}
	return program, nil
}
