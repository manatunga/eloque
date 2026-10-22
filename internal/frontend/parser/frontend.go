package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
)

// Parse constructs an arena-backed abstract syntax tree from Eloque
// source text.
//
// Tokens are produced incrementally by the streaming lexer rather
// than buffered into a complete token sequence.
//
// On success, Parse returns a FlatProgram and a nil diagnostic.
// On lexical or syntax failure, it returns a nil program and the
// corresponding structured diagnostic.
//
// Lexical diagnostics take precedence over syntax diagnostics when
// both are encountered during parsing.
//
// Parsing currently stops at the first error. Multi-error recovery
// will be introduced separately.
func Parse(sourceText string) (*ast.FlatProgram, *diagnostic.Diagnostic) {
	parser := NewFlatParser(sourceText)
	program, diag := parser.parseFlatProgram()

	if lexDiag := parser.tokenSource.lexicalError(); lexDiag != nil {
		return nil, lexDiag
	}
	if diag != nil {
		return nil, diag
	}

	return program, nil
}
