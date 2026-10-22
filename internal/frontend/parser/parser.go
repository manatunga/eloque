package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// Parser maintains the state required to construct an Eloque abstract
// syntax tree from a sequence of lexical tokens.
//
// Tokens retain their source byte offsets rather than storing lexemes
// or line/column spans. The parser keeps the original source text to
// reconstruct this information when constructing AST nodes and diagnostics.
//
// lineOffsets stores the byte offset at which each source line begins.
// positionLine caches the zero-based index of the most recently resolved
// line, allowing sequential source-position lookups to avoid repeated
// binary searches.
//
// current identifies the next token to consume. Tokens must appear in
// source order and end with a terminal EOF token.
type Parser struct {
	source      string
	tokenSource tokenSource
}

// New constructs a parser for the given source text and token sequence.
// Tokens must refer to byte offsets in sourceText and must end with EOF.

func New(sourceText string, tokens []lexer.Token) *Parser {
	return &Parser{
		source:      sourceText,
		tokenSource: newBufferedTokenSource(tokens),
	}
}

// newStreamingParser constructs a parser whose tokens are
// produced lazily by the existing lexer.
func newStreamingParser(sourceText string) *Parser {
	return &Parser{
		source:      sourceText,
		tokenSource: newStreamingTokenSource(sourceText),
	}
}

// newRecoveringParser constructs a parser configured for recoverable
// lexical scanning and parser-directed synchronization.
//
// It preserves the normal parser's source-position tracking while
// using a recovery-aware streaming token source.
func newRecoveringParser(sourceText string) *Parser {
	return &Parser{
		source:      sourceText,
		tokenSource: newRecoveringTokenSource(sourceText),
	}
}

// lexeme returns the source substring represented by a token.
func (p *Parser) lexeme(token lexer.Token) string {
	return p.source[int(token.Start):int(token.End)]
}

// peek returns the current token without consuming it.
//
// Token navigation is delegated to the underlying token source,
// regardless of whether tokens are buffered or streamed.
func (p *Parser) peek() lexer.Token {
	return p.tokenSource.peek()
}

// advance consumes and returns the current token.
// The token source owns cursor advancement and preserves EOF.
func (p *Parser) advance() lexer.Token {
	return p.tokenSource.advance()
}

// check reports whether the current token has the specified kind.
func (p *Parser) check(kind lexer.TokenKind) bool {
	return p.peek().Kind == kind
}

// match consumes the current token if it matches one of the kinds.
func (p *Parser) match(kinds ...lexer.TokenKind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}

	return false
}

// unexpectedToken constructs a structured diagnostic for a token mismatch.
//
// The diagnostic identifies the offending token by its half-open source
// byte span and preserves the expected and actual token kinds as stable
// string names for compiler tooling.
//
// It does not consume tokens. Diagnostic construction occurs only on
// the failure path, avoiding allocations during successful parsing.
func (p *Parser) unexpectedToken(
	token lexer.Token,
	expected lexer.TokenKind,
	message string,
) *diagnostic.Diagnostic {
	return &diagnostic.Diagnostic{
		Code:     CodeUnexpectedToken,
		Severity: diagnostic.SeverityError,
		Message:  message,
		Primary: diagnostic.Location{
			File: 0,
			Span: source.ByteSpan{
				Start: token.Start,
				End:   token.End,
			},
		},
		TokenExpectation: &diagnostic.TokenExpectation{
			Expected: []string{expected.String()},
			Actual:   token.Kind.String(),
		},
	}
}

// expect consumes and returns the current token if it matches kind.
//
// On mismatch, it returns a zero-value token and a positioned diagnostic
// without advancing the parser. The supplied message describes the
// expected syntax.
func (p *Parser) expect(
	kind lexer.TokenKind,
	message string,
) (lexer.Token, *diagnostic.Diagnostic) {
	token := p.peek()

	if token.Kind == kind {
		return p.advance(), nil
	}

	return lexer.Token{}, p.unexpectedToken(token, kind, message)
}
