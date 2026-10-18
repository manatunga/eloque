package lexer

// Lexer maintains the state required to tokenize Eloque source code.
//
// The start and current fields track byte offsets into the source.
// Line and column positions are 1-based and are maintained for lexical
// error reporting.
//
// indentStack tracks active indentation levels, while pendingTokens
// queues structural tokens such as INDENT and DEDENT.
//
// parenDepth and parenPositions track unmatched opening parentheses.
// Newlines inside parentheses are ignored during tokenization.
type Lexer struct {
	source         string
	start          int
	current        int
	line           int
	column         int
	startPosition  Position
	indentStack    []int
	parenDepth     int
	atLineStart    bool
	pendingTokens  []Token
	pendingIndex   int
	parenPositions []Position
}

// New creates a lexer for the provided Eloque source text.
// Scanning begins at byte offset 0, line 1, column 1.
func New(source string) *Lexer {
	return &Lexer{
		source:        source,
		start:         0,
		current:       0,
		line:          1,
		column:        1,
		startPosition: Position{Line: 1, Column: 1},
		indentStack:   make([]int, 1, 4),
		atLineStart:   true,
	}
}

// advance consumes the next source byte and updates the cursor position.
// A newline moves the cursor to column 1 of the following line.
// The caller must ensure the lexer has not reached EOF.
func (l *Lexer) advance() byte {
	currentByte := l.peek()
	if currentByte == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	l.current++
	return currentByte
}

// isAtEnd reports whether the cursor has reached or passed
// the end of the source text.
func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

// peek returns the next unread byte without advancing the cursor.
// It returns 0 when the lexer has reached EOF.
func (l *Lexer) peek() byte {
	if l.isAtEnd() {
		return 0
	}
	return l.source[l.current]
}

// isIdentifierStart reports whether a byte can begin an Eloque identifier.
// Identifiers currently begin with an ASCII letter or underscore.
func isIdentifierStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_'
}

// isIdentifierContinue reports whether a byte can appear after the
// first character of an identifier, including ASCII digits.
func isIdentifierContinue(ch byte) bool {
	return isIdentifierStart(ch) ||
		(ch >= '0' && ch <= '9')
}

// scanIdentifier consumes the remaining characters of an identifier.
// The first character must already have been consumed by advance.
// It returns the complete identifier text without classifying it.
func (l *Lexer) scanIdentifier() string {
	for isIdentifierContinue(l.peek()) {
		l.advance()
	}
	return l.source[l.start:l.current]
}

// scanIdentifierKind scans an identifier and determines its token kind.
// Reserved keywords receive their dedicated kinds; other words are identifiers.
func (l *Lexer) scanIdentifierKind() TokenKind {
	word := l.scanIdentifier()
	return lookupIdentifier(word)
}

// markTokenStart records the byte offset and source position at which
// the next token begins. The saved position remains unchanged while
// the token is being scanned.
func (l *Lexer) markTokenStart() {
	l.start = l.current
	l.startPosition = Position{
		Line:   l.line,
		Column: l.column,
	}
}

// makeToken constructs a compact token using source byte offsets.
func (l *Lexer) makeToken(kind TokenKind) Token {
	return Token{
		Start: uint32(l.start),
		End:   uint32(l.current),
		Kind:  kind,
	}
}

// skipHorizontalWhitespace consumes consecutive ASCII space characters.
// It deliberately leaves newlines unread so they can become tokens.
// Tabs are not consumed by this helper.
func (l *Lexer) skipHorizontalWhitespace() {
	for l.peek() == ' ' {
		l.advance()
	}
}

// isDigit reports whether a byte is an ASCII decimal digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// scanInteger consumes consecutive decimal digits and returns their
// original source text. The first digit must already be consumed.
// Literal validation is performed separately by nextToken.
func (l *Lexer) scanInteger() string {
	for isDigit(l.peek()) {
		l.advance()
	}
	return l.source[l.start:l.current]
}

// hasLeadingZero reports whether a decimal integer literal contains
// multiple digits and begins with zero. The literal "0" is permitted.
func hasLeadingZero(lexeme string) bool {
	return len(lexeme) > 1 && lexeme[0] == '0'
}

// popPendingToken retrieves the oldest unread queued token.
// When all pending tokens have been consumed, it resets the queue
// while retaining its backing array for subsequent reuse.
func (l *Lexer) popPendingToken() (Token, bool) {
	if l.pendingIndex >= len(l.pendingTokens) {
		return Token{}, false
	}

	token := l.pendingTokens[l.pendingIndex]
	l.pendingIndex++

	if l.pendingIndex == len(l.pendingTokens) {
		l.pendingTokens = l.pendingTokens[:0]
		l.pendingIndex = 0
	}

	return token, true
}

// measureIndentation consumes consecutive ASCII spaces at the current
// cursor position and returns their count. It leaves the first non-space
// byte unread and does not validate or modify indentation levels.
// Callers should invoke it only when processing the start of a line.
func (l *Lexer) measureIndentation() int {
	width := 0
	for l.peek() == ' ' {
		l.advance()
		width++
	}
	return width
}

// isValidIndentationWidth reports whether a measured indentation width
// is non-negative and aligned to Eloque's four-space indentation levels.
// It does not check whether the width matches an active indentation level.
func isValidIndentationWidth(width int) bool {
	return width >= 0 && width%4 == 0
}

// processIndentation validates and applies a line's indentation width.
// It queues structural tokens without consuming source characters.
// Invalid indentation returns a positioned lexical error without
// modifying the indentation stack or pending-token queue.
func (l *Lexer) processIndentation(width int) error {
	if !isValidIndentationWidth(width) {
		return &LexError{
			Position: Position{Line: l.line, Column: l.column},
			Message:  "indentation must be a multiple of four spaces",
		}
	}
	currentIndent := l.indentStack[len(l.indentStack)-1]
	if width > currentIndent {
		l.indentStack = append(l.indentStack, width)
		l.pendingTokens = append(l.pendingTokens, Token{
			Start: uint32(l.current),
			End:   uint32(l.current),
			Kind:  TokenIndent,
		})
	}
	if width < currentIndent {
		index := len(l.indentStack) - 1
		for index > 0 && l.indentStack[index] > width {
			index--
		}
		if l.indentStack[index] != width {
			return &LexError{
				Position: Position{Line: l.line, Column: l.column},
				Message:  "dedentation does not match an active indentation level",
			}
		}
		dedentCount := len(l.indentStack) - 1 - index
		l.indentStack = l.indentStack[:index+1]
		for range dedentCount {
			l.pendingTokens = append(l.pendingTokens, Token{
				Start: uint32(l.current),
				End:   uint32(l.current),
				Kind:  TokenDedent,
			})
		}
	}
	return nil
}

// skipComment consumes a comment from the current cursor position up to,
// but not including, the next newline or end of source. It assumes the
// cursor is positioned at the '#' character and leaves any newline unread
// so the lexer can emit it as a separate token.
func (l *Lexer) skipComment() {
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}
}

// scanString scans a string literal delimited by quote, assuming the opening
// quote has already been consumed. It preserves escape sequences in the token
// lexeme and reports invalid escapes, newlines, or EOF before the closing quote
// as lexical errors.
func (l *Lexer) scanString(quote byte) (Token, error) {
	isValidEscape := [256]bool{
		'"': true, '\'': true, '\\': true, 'n': true, 't': true, 'r': true,
	}

	for l.peek() != quote {
		if l.isAtEnd() || l.peek() == '\n' {
			return Token{}, &LexError{
				Position: l.startPosition,
				Message:  "unterminated string literal",
			}
		}
		if l.peek() == '\\' {
			escapePosition := Position{Line: l.line, Column: l.column}
			l.advance()
			if l.isAtEnd() || l.peek() == '\n' {
				return Token{}, &LexError{
					Position: escapePosition,
					Message:  "invalid escape sequence",
				}
			}

			if !isValidEscape[l.peek()] {
				return Token{}, &LexError{
					Position: escapePosition,
					Message:  "invalid escape sequence",
				}
			}
		}
		l.advance()
		continue
	}
	l.advance()
	return l.makeToken(TokenString), nil
}

// NextToken returns the next token in source order, emitting queued
// structural tokens before scanning additional source text. At the start
// of a non-blank line, it measures and processes indentation before
// scanning ordinary tokens. It skips horizontal spaces, preserves
// newlines, and validates integer literals. Invalid input returns a
// zero-value token and a positioned lexical error; callers should stop
// tokenization when an error occurs.
func (l *Lexer) NextToken() (Token, error) {
	if token, ok := l.popPendingToken(); ok {
		return token, nil
	}

	if l.atLineStart && l.parenDepth == 0 {
		width := l.measureIndentation()

		if l.peek() != '\n' && l.peek() != '#' && !l.isAtEnd() {
			if err := l.processIndentation(width); err != nil {
				return Token{}, err
			}

			l.atLineStart = false
			if token, ok := l.popPendingToken(); ok {
				return token, nil
			}
		}
	}

	for {
		l.skipHorizontalWhitespace()
		l.markTokenStart()

		if l.peek() == '#' {
			l.skipComment()
			continue
		}

		if l.peek() == '\n' && l.parenDepth > 0 {
			l.advance()
			l.atLineStart = true
			continue
		}

		break
	}

	if l.isAtEnd() {
		if l.parenDepth > 0 {
			return Token{}, &LexError{
				Position: l.parenPositions[len(l.parenPositions)-1],
				Message:  "unclosed opening parenthesis",
			}
		}
		if len(l.indentStack) > 1 {
			if err := l.processIndentation(0); err != nil {
				return Token{}, err
			}
			if token, ok := l.popPendingToken(); ok {
				return token, nil
			}
		}
		return l.makeToken(TokenEOF), nil
	}

	char := l.advance()

	if char != '\n' {
		l.atLineStart = false
	}
	if char == '\n' {
		l.atLineStart = true
		return l.makeToken(TokenNewline), nil
	}

	if char == ':' {
		return l.makeToken(TokenColon), nil
	}

	if char == '+' {
		return l.makeToken(TokenPlus), nil
	}
	if char == '-' {
		return l.makeToken(TokenMinus), nil
	}
	if char == '*' {
		if l.peek() == '*' {
			l.advance()
			return l.makeToken(TokenPower), nil
		}
		return l.makeToken(TokenStar), nil
	}
	if char == '/' {
		if l.peek() == '/' {
			l.advance()
			return l.makeToken(TokenFloorDiv), nil
		}
		return l.makeToken(TokenSlash), nil
	}
	if char == '%' {
		return l.makeToken(TokenPercent), nil
	}

	if char == '"' || char == '\'' {
		return l.scanString(char)
	}

	if char == '(' {
		l.parenDepth++
		l.parenPositions = append(l.parenPositions, l.startPosition)
		return l.makeToken(TokenLeftParen), nil
	}
	if char == ')' {
		if l.parenDepth == 0 {
			return Token{}, &LexError{
				Position: l.startPosition,
				Message:  "unmatched closing parenthesis",
			}
		}
		l.parenDepth--
		l.parenPositions = l.parenPositions[:len(l.parenPositions)-1]
		return l.makeToken(TokenRightParen), nil
	}

	if isIdentifierStart(char) {
		kind := l.scanIdentifierKind()
		return l.makeToken(kind), nil
	}
	if isDigit(char) {
		lexeme := l.scanInteger()
		if isIdentifierStart(l.peek()) {
			return Token{}, &LexError{
				Position: l.startPosition,
				Message:  "invalid integer literal",
			}
		}
		if hasLeadingZero(lexeme) {
			return Token{}, &LexError{
				Position: l.startPosition,
				Message:  "integer literal cannot have leading zeros",
			}
		}
		return l.makeToken(TokenInt), nil
	}

	return Token{}, &LexError{
		Position: l.startPosition,
		Message:  "unexpected character",
	}
}
