package lexer

import "strings"

// testTokenLexeme reconstructs a token's lexeme from its source byte range.
func testTokenLexeme(source string, token Token) string {
	return source[token.Start:token.End]
}

// testTokenSpan reconstructs the one-based source span for a compact token.
// This helper is test-only and does not add fields to production tokens.
func testTokenSpan(source string, token Token) Span {
	position := func(offset uint32) Position {
		prefix := source[:offset]
		line := strings.Count(prefix, "\n") + 1
		lastNewline := strings.LastIndexByte(prefix, '\n')
		return Position{Line: line, Column: len(prefix) - lastNewline}
	}
	return Span{Start: position(token.Start), End: position(token.End)}
}
