package parser

import (
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseFlatProgram parses a complete sequence of top-level Eloque
// statements into an arena-backed abstract syntax tree.
//
// Blank lines are skipped, while successfully parsed statements are
// appended to the arena's statement sequence in source order.
//
// Each statement must be followed by NEWLINE or EOF. Statement parsers
// deliberately leave these terminators unconsumed, allowing this method
// to validate boundaries consistently with the existing M1 parser.
//
// The returned FlatProgram references the arena and a half-open range
// of statement IDs rather than copying individual AST nodes.
//
// The range begins at the current statement-sequence position, so it
// identifies only statements appended during this invocation.
//
// Parsing stops at EOF or returns a positioned diagnostic for malformed
// statements or invalid statement boundaries. On failure, no FlatProgram
// is returned, although nodes already constructed in the arena remain.
//
// EOF is not consumed.
func (p *FlatParser) parseFlatProgram() (*ast.FlatProgram, *Diagnostic) {
	start := p.arena.StatementCount()

	for !p.check(lexer.TokenEOF) {
		if p.match(lexer.TokenNewline) {
			continue
		}

		statementID, diagnostic := p.parseFlatStatement()
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

		p.arena.AddStatement(statementID)
	}

	end := p.arena.StatementCount()

	var span source.ByteSpan

	if start == end {
		// Empty programs have a zero-width span at EOF,
		// including when the source contains only blank lines.
		eof := p.peek().Start

		span = source.ByteSpan{
			Start: eof,
			End:   eof,
		}
	} else {
		// Derive the program's extent from its first and last
		// statements, excluding surrounding blank lines.
		ids := p.arena.StatementRange(ast.StatementRange{
			Start: start,
			End:   end,
		})

		first := p.arena.Get(ids[0])
		last := p.arena.Get(ids[len(ids)-1])

		span = source.ByteSpan{
			Start: first.Span.Start,
			End:   last.Span.End,
		}
	}

	return &ast.FlatProgram{
		Arena: p.arena,
		Statements: ast.StatementRange{
			Start: start,
			End:   end,
		},
		Span: span,
	}, nil
}
