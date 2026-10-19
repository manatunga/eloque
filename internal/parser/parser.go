package parser

import (
	"sort"

	"github.com/manatunga/eloque/internal/lexer"
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
	source       string
	tokenSource  tokenSource
	lineOffsets  []int
	positionLine int
}

// buildLineOffsets returns the starting byte offset of every source line.
//
// The first entry is always zero, representing line 1. Each subsequent
// entry identifies the byte immediately following a newline character.
// The resulting offsets are sorted in ascending order.
func buildLineOffsets(sourceText string) []int {
	offsets := []int{0}

	for i := 0; i < len(sourceText); i++ {
		if sourceText[i] == '\n' {
			offsets = append(offsets, i+1)
		}
	}

	return offsets
}

// New constructs a parser for the given source text and token sequence.
// Tokens must refer to byte offsets in sourceText and must end with EOF.

func New(sourceText string, tokens []lexer.Token) *Parser {
	return &Parser{
		source:      sourceText,
		tokenSource: newBufferedTokenSource(tokens),
		lineOffsets: buildLineOffsets(sourceText),
	}
}

// newStreamingParser constructs a parser whose tokens are
// produced lazily by the existing lexer.
func newStreamingParser(sourceText string) *Parser {
	return &Parser{
		source:      sourceText,
		tokenSource: newStreamingTokenSource(sourceText),
		lineOffsets: buildLineOffsets(sourceText),
	}
}

// offsetPosition converts a source byte offset into a 1-based line and
// column position.
//
// Forward lookups advance from the cached line index, avoiding repeated
// binary searches during sequential parsing. Backward lookups use binary
// search to recover the correct line.
//
// The cache is updated after every lookup. Columns are measured in bytes,
// consistent with the lexer's current source-position representation.
func (p *Parser) offsetPosition(offset uint32) source.Position {
	byteOffset := int(offset)
	lineIndex := p.positionLine

	if byteOffset < p.lineOffsets[lineIndex] {
		lineIndex = sort.Search(
			len(p.lineOffsets),
			func(i int) bool {
				return p.lineOffsets[i] > byteOffset
			},
		) - 1
	} else {
		for lineIndex+1 < len(p.lineOffsets) &&
			p.lineOffsets[lineIndex+1] <= byteOffset {
			lineIndex++
		}
	}

	p.positionLine = lineIndex

	return source.Position{
		Line:   lineIndex + 1,
		Column: byteOffset - p.lineOffsets[lineIndex] + 1,
	}
}

// lexeme returns the source substring represented by a token.
func (p *Parser) lexeme(token lexer.Token) string {
	return p.source[int(token.Start):int(token.End)]
}

// tokenSpan reconstructs a token's half-open source span.
//
// The token's start and end byte offsets are converted into 1-based
// line and column positions using the parser's cached line lookup.
func (p *Parser) tokenSpan(token lexer.Token) source.Span {
	return source.Span{
		Start: p.offsetPosition(token.Start),
		End:   p.offsetPosition(token.End),
	}
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

// unexpectedToken constructs a positioned diagnostic for a token mismatch.
//
// The diagnostic records the expected token kind, the actual token kind,
// and the offending token's source span. It does not consume any tokens.
//
// Keeping diagnostic construction on the failure path avoids unnecessary
// heap allocations during successful parsing.
func (p *Parser) unexpectedToken(
	token lexer.Token,
	expected lexer.TokenKind,
	message string,
) *Diagnostic {
	actual := token.Kind

	return &Diagnostic{
		Message:  message,
		Location: p.tokenSpan(token),
		Expected: []lexer.TokenKind{expected},
		Actual:   &actual,
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
) (lexer.Token, *Diagnostic) {
	token := p.peek()

	if token.Kind == kind {
		return p.advance(), nil
	}

	return lexer.Token{}, p.unexpectedToken(token, kind, message)
}
