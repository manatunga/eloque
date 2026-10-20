package parser

import (
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseProgram parses a complete sequence of top-level Eloque statements.
//
// Blank lines are skipped, while statements are parsed and preserved in
// source order. Each statement must be followed by NEWLINE or EOF.
//
// The resulting Program spans from the beginning of its first statement
// to the end of its last statement. Empty programs receive a zero-width
// span at the EOF position.
//
// Parsing stops at EOF or returns a positioned syntax diagnostic when
// a statement is malformed or its boundary is invalid.
func (p *Parser) parseProgram() (*ast.Program, *Diagnostic) {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !p.check(lexer.TokenEOF) {
		if p.peek().Kind == lexer.TokenNewline {
			p.advance()
			continue
		}

		statement, diagnostic := p.parseStatement()
		if diagnostic != nil {
			return nil, diagnostic
		}

		nextToken := p.peek()
		if nextToken.Kind != lexer.TokenNewline &&
			nextToken.Kind != lexer.TokenEOF {
			return nil, &Diagnostic{
				Message:  "expected a newline or end of file after statement",
				Location: p.tokenSpan(nextToken),
				Expected: []lexer.TokenKind{
					lexer.TokenNewline,
					lexer.TokenEOF,
				},
				Actual: &nextToken.Kind,
			}
		}

		program.Statements = append(program.Statements, statement)
	}

	var span source.Span
	eofPosition := p.tokenSpan(p.peek()).Start

	if len(program.Statements) == 0 {
		span = source.Span{
			Start: eofPosition,
			End:   eofPosition,
		}
	} else {
		span = source.Span{
			Start: program.Statements[0].Span().Start,
			End:   program.Statements[len(program.Statements)-1].Span().End,
		}
	}

	program.Location = span
	return program, nil
}
