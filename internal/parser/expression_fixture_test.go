package parser

import (
	"fmt"
	"strings"

	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// legacyExpressionToken preserves the original test fixture's explicit
// source-position expectations while adapting to offset-based tokens.
type legacyExpressionToken struct {
	Kind   lexer.TokenKind
	Lexeme string
	Span   source.Span
}

func newLegacyExpressionParser(fixtures []legacyExpressionToken) *Parser {
	var text strings.Builder
	tokens := make([]lexer.Token, 0, len(fixtures))
	line, column := 1, 1

	appendText := func(s string) {
		text.WriteString(s)
		for i := 0; i < len(s); i++ {
			if s[i] == '\n' {
				line++
				column = 1
			} else {
				column++
			}
		}
	}
	padTo := func(pos source.Position) {
		if pos.Line < line || (pos.Line == line && pos.Column < column) {
			panic(fmt.Sprintf("overlapping legacy token span: current %d:%d, target %+v", line, column, pos))
		}
		for line < pos.Line {
			appendText("\n")
		}
		if pos.Column > column {
			appendText(strings.Repeat(" ", pos.Column-column))
		}
	}
	for _, fixture := range fixtures {
		if fixture.Span.Start.Line > 0 && fixture.Span.Start.Column > 0 {
			padTo(fixture.Span.Start)
		}
		start := uint32(text.Len())
		lexeme := fixture.Lexeme
		if lexeme == "" && fixture.Kind != lexer.TokenEOF && fixture.Span.Start.Line > 0 {
			width := fixture.Span.End.Column - fixture.Span.Start.Column
			if width > 0 {
				lexeme = strings.Repeat("?", width)
			}
		}
		appendText(lexeme)
		end := uint32(text.Len())
		tokens = append(tokens, lexer.Token{Kind: fixture.Kind, Start: start, End: end})
	}
	return New(text.String(), tokens)
}
