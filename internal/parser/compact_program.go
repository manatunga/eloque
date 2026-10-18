package parser

import (
	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseCompactProgram parses a complete sequence of top-level
// Eloque statements from compact tokens.
//
// It skips blank lines, enforces statement boundaries, and
// preserves statement order and source locations.
//
// Empty programs receive a zero-width span at EOF.
func (p *CompactParser) parseCompactProgram() (*ast.Program, *Diagnostic) {
	program := &ast.Program{
		Statements: []ast.Statement{},
	}

	for !p.check(lexer.TokenEOF) {
		if p.peek().Kind == lexer.TokenNewline {
			p.advance()
			continue
		}

		statement, diagnostic := p.parseCompactStatement()
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
