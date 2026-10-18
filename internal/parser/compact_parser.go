package parser

import (
	"sort"

	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// CompactParser parses compact tokens into the existing Eloque AST.
type CompactParser struct {
	source      string
	tokens      []lexer.CompactToken
	lineOffsets []int
	current     int

	// positionLine caches the zero-based line index of the most
	// recently resolved source position.
	positionLine int
}

// NewCompact constructs an experimental parser for compact tokens.
func NewCompact(sourceText string, tokens []lexer.CompactToken) *CompactParser {
	return &CompactParser{
		source:      sourceText,
		tokens:      tokens,
		lineOffsets: buildCompactLineOffsets(sourceText),
	}
}

// buildCompactLineOffsets records the byte offset of each line's start.
func buildCompactLineOffsets(sourceText string) []int {
	offsets := []int{0}

	for i := 0; i < len(sourceText); i++ {
		if sourceText[i] == '\n' {
			offsets = append(offsets, i+1)
		}
	}

	return offsets
}

// offsetPosition converts a byte offset into a source position.
// Forward lookups reuse the cached line; backward lookups use
// binary search to recover the correct line.
func (p *CompactParser) offsetPosition(offset uint32) source.Position {
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
func (p *CompactParser) lexeme(token lexer.CompactToken) string {
	return p.source[int(token.Start):int(token.End)]
}

// tokenSpan reconstructs the source span of a compact token.
func (p *CompactParser) tokenSpan(token lexer.CompactToken) source.Span {
	return source.Span{
		Start: p.offsetPosition(token.Start),
		End:   p.offsetPosition(token.End),
	}
}

// peek returns the current token without consuming it.
func (p *CompactParser) peek() lexer.CompactToken {
	return p.tokens[p.current]
}

// advance consumes the current token unless it is EOF.
func (p *CompactParser) advance() lexer.CompactToken {
	token := p.peek()

	if token.Kind != lexer.TokenEOF {
		p.current++
	}

	return token
}

// check reports whether the current token has the specified kind.
func (p *CompactParser) check(kind lexer.TokenKind) bool {
	return p.peek().Kind == kind
}

// match consumes the current token if it matches one of the kinds.
func (p *CompactParser) match(kinds ...lexer.TokenKind) bool {
	for _, kind := range kinds {
		if p.check(kind) {
			p.advance()
			return true
		}
	}

	return false
}

// expect consumes the expected token or returns a diagnostic.
func (p *CompactParser) expect(
	kind lexer.TokenKind,
	message string,
) (lexer.CompactToken, *Diagnostic) {
	token := p.peek()

	if token.Kind == kind {
		return p.advance(), nil
	}

	return lexer.CompactToken{}, &Diagnostic{
		Message:  message,
		Location: p.tokenSpan(token),
		Expected: []lexer.TokenKind{kind},
		Actual:   &token.Kind,
	}
}
