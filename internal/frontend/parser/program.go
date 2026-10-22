package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// flatProgramFromRange constructs a program referencing statements
// committed to the arena within the specified half-open range.
//
// Empty programs use the current EOF position when available;
// otherwise they use the end of the source as a safe fallback.
func (p *FlatParser) flatProgramFromRange(
	start, end uint32,
) *ast.FlatProgram {
	var span source.ByteSpan

	if start == end {
		offset := uint32(len(p.source))

		if p.tokenSource.lexicalError() == nil &&
			p.check(lexer.TokenEOF) {
			offset = p.peek().Start
		}

		span = source.ByteSpan{
			Start: offset,
			End:   offset,
		}
	} else {
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
	}
}

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
func (p *FlatParser) parseFlatProgram() (*ast.FlatProgram, *diagnostic.Diagnostic) {
	start := p.arena.StatementCount()

	for !p.check(lexer.TokenEOF) {
		if p.match(lexer.TokenNewline) {
			continue
		}

		statementID, diag := p.parseFlatStatement()
		if diag != nil {
			return nil, diag
		}

		nextToken := p.peek()
		if nextToken.Kind != lexer.TokenNewline &&
			nextToken.Kind != lexer.TokenEOF {
			return nil, &diagnostic.Diagnostic{
				Code:     CodeUnexpectedToken,
				Severity: diagnostic.SeverityError,
				Message:  "expected a newline or end of file after statement",
				Primary: diagnostic.Location{
					File: 0,
					Span: source.ByteSpan{
						Start: nextToken.Start,
						End:   nextToken.End,
					},
				},
				TokenExpectation: &diagnostic.TokenExpectation{
					Expected: []string{
						lexer.TokenNewline.String(),
						lexer.TokenEOF.String(),
					},
					Actual: nextToken.Kind.String(),
				},
			}
		}

		p.arena.AddStatement(statementID)
	}

	end := p.arena.StatementCount()

	return p.flatProgramFromRange(start, end), nil
}
