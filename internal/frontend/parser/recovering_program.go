package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseRecoveringProgram parses a sequence of statements while collecting
// diagnostics and retaining successfully completed statements.
//
// Lexical diagnostics take precedence over syntax diagnostics caused by
// invalid lexer tokens. Malformed statements are abandoned, and parsing
// resumes at a recoverable statement boundary when possible.
//
// The method returns a partial program even when recovery aborts.
func (p *FlatParser) parseRecoveringProgram(
	session *recoverySession,
) (*ast.FlatProgram, ParseStatus) {
	start := p.arena.StatementCount()
	status := ParseComplete

	for {
		if session.limitReached() {
			status = ParseAborted
			break
		}

		if p.tokenSource.lexicalError() != nil {
			session.add(p.tokenSource.takeDiagnostic())
			status = ParseRecovered

			if session.limitReached() || p.tokenSource.terminal {
				status = ParseAborted
				break
			}

			p.tokenSource.synchronizeLine(true)
			p.tokenSource.finishLineSynchronization()
			continue
		}

		if p.tokenSource.terminal {
			status = ParseAborted
			break
		}

		if p.check(lexer.TokenEOF) {
			break
		}

		if p.match(lexer.TokenNewline) {
			continue
		}

		checkpoint := p.arena.Checkpoint()
		statementID, diag := p.parseFlatStatement()

		// A lexical failure may have interrupted expression parsing.
		// In that case, the lexical diagnostic is authoritative.
		if p.tokenSource.lexicalError() != nil {
			p.arena.Rollback(checkpoint)
			session.add(p.tokenSource.takeDiagnostic())
			status = ParseRecovered

			if session.limitReached() || p.tokenSource.terminal {
				status = ParseAborted
				break
			}

			p.tokenSource.synchronizeLine(true)
			p.tokenSource.finishLineSynchronization()
			continue
		}

		if diag != nil {
			p.arena.Rollback(checkpoint)
			session.add(diag)
			status = ParseRecovered

			if session.limitReached() {
				status = ParseAborted
				break
			}

			p.tokenSource.synchronizeLine(true)
			p.tokenSource.finishLineSynchronization()
			continue
		}

		next := p.peek()
		if p.tokenSource.lexicalError() != nil {
			p.arena.Rollback(checkpoint)
			session.add(p.tokenSource.takeDiagnostic())
			status = ParseRecovered

			if session.limitReached() || p.tokenSource.terminal {
				status = ParseAborted
				break
			}

			p.tokenSource.synchronizeLine(true)
			p.tokenSource.finishLineSynchronization()
			continue
		}

		if next.Kind != lexer.TokenNewline &&
			next.Kind != lexer.TokenEOF {
			p.arena.Rollback(checkpoint)
			session.add(&diagnostic.Diagnostic{
				Code:     CodeUnexpectedToken,
				Severity: diagnostic.SeverityError,
				Message:  "expected a newline or end of file after statement",
				Primary: diagnostic.Location{
					File: 0,
					Span: source.ByteSpan{
						Start: next.Start,
						End:   next.End,
					},
				},
				TokenExpectation: &diagnostic.TokenExpectation{
					Expected: []string{
						lexer.TokenNewline.String(),
						lexer.TokenEOF.String(),
					},
					Actual: next.Kind.String(),
				},
			})

			status = ParseRecovered

			if session.limitReached() {
				status = ParseAborted
				break
			}

			p.tokenSource.synchronizeLine(true)
			p.tokenSource.finishLineSynchronization()
			continue
		}

		p.arena.AddStatement(statementID)
	}

	end := p.arena.StatementCount()
	return p.flatProgramFromRange(start, end), status
}
