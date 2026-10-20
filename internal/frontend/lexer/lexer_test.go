package lexer

import (
	"errors"
	"reflect"
	"testing"
)

func TestAdvance(t *testing.T) {
	lexer := New("abc")
	first := lexer.advance()
	second := lexer.advance()

	if first != 'a' {
		t.Errorf("expected %q, got %q", 'a', first)
	}
	if second != 'b' {
		t.Errorf("expected %q, got %q", 'b', second)
	}
	if lexer.current != 2 {
		t.Errorf("expected current %d, got %d", 2, lexer.current)
	}
	if lexer.start != 0 {
		t.Errorf("expected start 0, got %d", lexer.start)
	}
}

func TestAdvancePosition(t *testing.T) {
	testCases := []struct {
		expectedByte    byte
		expectedCurrent int
		expectedLine    int
		expectedColumn  int
	}{
		{'a', 1, 1, 2},
		{'\n', 2, 2, 1},
		{'b', 3, 2, 2},
		{'\n', 4, 3, 1},
		{'\n', 5, 4, 1},
		{'c', 6, 4, 2},
	}
	lexer := New("a\nb\n\nc")

	if lexer.current != 0 || lexer.line != 1 || lexer.column != 1 {
		t.Errorf("unexpected initial position: current=%d, line=%d, column=%d",
			lexer.current, lexer.line, lexer.column)
	}

	for _, tC := range testCases {
		result := lexer.advance()

		if result != tC.expectedByte {
			t.Errorf("expected byte %q, got %q", tC.expectedByte, result)
		}
		if lexer.current != tC.expectedCurrent {
			t.Errorf("expected current %d, got %d",
				tC.expectedCurrent, lexer.current)
		}
		if lexer.line != tC.expectedLine {
			t.Errorf("expected line %d, got %d",
				tC.expectedLine, lexer.line)
		}
		if lexer.column != tC.expectedColumn {
			t.Errorf("expected column %d, got %d",
				tC.expectedColumn, lexer.column)
		}
	}
}

func TestIsAtEnd(t *testing.T) {
	lexer := New("abc")
	if lexer.isAtEnd() {
		t.Error("expected lexer should not be at end right after initialization.")
	}

	lexer.advance()
	lexer.advance()
	lexer.advance()
	if !lexer.isAtEnd() {
		t.Error("expected isAtEnd to return true when cursor equals source length")
	}

	lexer.current = 4
	if !lexer.isAtEnd() {
		t.Error("expected isAtEnd to return true when cursor exceeds source length")
	}
}

func TestPeek(t *testing.T) {
	lexer := New("abc")
	afterInit := lexer.peek()
	if afterInit != 'a' || lexer.current != 0 {
		t.Errorf("expected return %q and current %d, got %q and %d", 'a', 0, afterInit, lexer.current)
	}

	lexer.advance()
	afterFirst := lexer.peek()
	if afterFirst != 'b' || lexer.current != 1 {
		t.Errorf("expected return %q and current %d, got %q and %d", 'b', 1, afterFirst, lexer.current)
	}

	afterFirstAgain := lexer.peek()
	if afterFirstAgain != 'b' || lexer.current != 1 {
		t.Errorf("expected return %q and current %d, got %q and %d", 'b', 1, afterFirstAgain, lexer.current)
	}

	lexer.advance()
	lexer.advance()
	afterLast := lexer.peek()
	if afterLast != 0 || lexer.current != 3 {
		t.Errorf("expected return %d and current %d, got %d and %d", 0, 3, afterLast, lexer.current)
	}

	if lexer.start != 0 {
		t.Errorf("expected start to remain 0, got %d", lexer.start)
	}
}

func TestIsIdentifierStart(t *testing.T) {
	testCases := []struct {
		input    byte
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'_', true},
		{'5', false},
		{' ', false},
		{'-', false},
	}
	for _, tC := range testCases {
		t.Run(string(tC.input), func(t *testing.T) {
			result := isIdentifierStart(tC.input)
			if result != tC.expected {
				t.Errorf("input %q: expected %t, got %t", tC.input, tC.expected, result)
			}
		})
	}
}

func TestIsIdentifierContinue(t *testing.T) {
	testCases := []struct {
		input    byte
		expected bool
	}{
		{'a', true},
		{'Z', true},
		{'_', true},
		{'0', true},
		{'9', true},
		{' ', false},
		{'-', false},
		{'.', false},
	}
	for _, tC := range testCases {
		t.Run(string(tC.input), func(t *testing.T) {
			result := isIdentifierContinue(tC.input)
			if result != tC.expected {
				t.Errorf("input %q: expected %t, got %t", tC.input, tC.expected, result)
			}
		})
	}
}

func TestScanIdentifier(t *testing.T) {
	testCases := []struct {
		input           string
		expected        string
		expectedCurrent int
	}{
		{"age", "age", 3},
		{"player_1 be 20", "player_1", 8},
		{"_count42", "_count42", 8},
		{"x", "x", 1},
		{"name-value", "name", 4},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)
			lexer.advance()
			result := lexer.scanIdentifier()
			if lexer.start != 0 {
				t.Errorf("expected start to remain 0, got %d", lexer.start)
			}
			if result != tC.expected || lexer.current != tC.expectedCurrent {
				t.Errorf("input %q: expected %q with current %d, got %q and %d",
					tC.input, tC.expected, tC.expectedCurrent, result, lexer.current)
			}
		})
	}
}

func TestScanIdentifierKind(t *testing.T) {
	testCases := []struct {
		input    string
		expected TokenKind
	}{
		{"let", TokenLet},
		{"be", TokenBe},
		{"say", TokenSay},
		{"player_1", TokenIdent},
		{"letter", TokenIdent},
		{"Let", TokenIdent},
		{"say hello", TokenSay},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)
			lexer.advance()
			result := lexer.scanIdentifierKind()
			if result != tC.expected {
				t.Errorf("input %q: expected %q, got %q", tC.input, tC.expected, result)
			}
		})
	}
}

func TestMarkTokenStart(t *testing.T) {
	lexer := New("let age be 20")
	lexer.advance()
	lexer.advance()
	lexer.advance()
	lexer.advance()

	lexer.markTokenStart()
	if lexer.start != 4 {
		t.Errorf("expected start %d, got %d", 4, lexer.start)
	}
	if lexer.current != 4 {
		t.Errorf("expected current %d, got %d", 4, lexer.current)
	}
	if lexer.line != 1 {
		t.Errorf("expected line %d, got %d", 1, lexer.line)
	}
	if lexer.column != 5 {
		t.Errorf("expected column %d, got %d", 5, lexer.column)
	}
	if lexer.startPosition.Line != 1 {
		t.Errorf("expected next start line %d, got %d", 1, lexer.startPosition.Line)
	}
	if lexer.startPosition.Column != 5 {
		t.Errorf("expected next start column %d, got %d", 5, lexer.startPosition.Column)
	}

	lexer.advance()
	if lexer.column != 6 {
		t.Errorf("expected column %d, got %d", 6, lexer.column)
	}
	if lexer.startPosition.Column != 5 {
		t.Errorf("expected next start column %d, got %d", 5, lexer.startPosition.Column)
	}
	if lexer.start != 4 {
		t.Errorf("expected start %d, got %d", 4, lexer.start)
	}
}

func TestMakeToken(t *testing.T) {
	source := "let age"
	lexer := New(source)
	lexer.markTokenStart()
	lexer.advance()
	lexer.advance()
	lexer.advance()

	token1 := lexer.makeToken(TokenLet)
	if token1.Kind != TokenLet {
		t.Errorf("expected kind %s, got %s", TokenLet, token1.Kind)
	}
	gotLexemeToken1 := source[token1.Start:token1.End]
	if gotLexemeToken1 != "let" {
		t.Errorf("expected lexeme %q, got %q", "let", gotLexemeToken1)
	}
	startToken1 := tokenPosition(source, token1.Start)
	if startToken1.Line != 1 {
		t.Errorf("expected start line %d, got %d", 1, startToken1.Line)
	}
	if startToken1.Column != 1 {
		t.Errorf("expected start column %d, got %d", 1, startToken1.Column)
	}

	if lexer.current != 3 {
		t.Errorf("expected current %d, got %d", 3, lexer.current)
	}
	if lexer.start != 0 {
		t.Errorf("expected source start %d, got %d", 0, lexer.start)
	}
	if lexer.column != 4 {
		t.Errorf("expected column %d, got %d", 4, lexer.column)
	}

	lexer.advance()
	lexer.markTokenStart()
	lexer.advance()
	lexer.advance()
	lexer.advance()

	token2 := lexer.makeToken(TokenIdent)
	if token2.Kind != TokenIdent {
		t.Errorf("expected kind %s, got %s", TokenIdent, token2.Kind)
	}
	gotLexemeToken2 := source[token2.Start:token2.End]
	if gotLexemeToken2 != "age" {
		t.Errorf("expected lexeme %q, got %q", "age", gotLexemeToken2)
	}
	if testTokenSpan(lexer.source, token2).Start.Line != 1 {
		t.Errorf("expected start line %d, got %d", 1, testTokenSpan(lexer.source, token2).Start.Line)
	}
	if testTokenSpan(lexer.source, token2).Start.Column != 5 {
		t.Errorf("expected start column %d, got %d", 5, testTokenSpan(lexer.source, token2).Start.Column)
	}
}

func TestSkipHorizontalWhitespace(t *testing.T) {
	testCases := []struct {
		input           string
		expectedCurrent int
		expectedColumn  int
		expectedNext    byte
	}{
		{"   age", 3, 4, 'a'},
		{"age", 0, 1, 'a'},
		{"  \nage", 2, 3, '\n'},
		{"  \tage", 2, 3, '\t'},
		{"   ", 3, 4, 0},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)
			lexer.skipHorizontalWhitespace()

			if lexer.current != tC.expectedCurrent {
				t.Errorf("input %q: expected current %d, got %d", tC.input, tC.expectedCurrent, lexer.current)
			}
			if lexer.column != tC.expectedColumn {
				t.Errorf("input %q: expected column %d, got %d", tC.input, tC.expectedColumn, lexer.column)
			}
			if lexer.peek() != tC.expectedNext {
				t.Errorf("input %q: expected next %q, got %q", tC.input, tC.expectedNext, lexer.peek())
			}
			if lexer.line != 1 {
				t.Errorf("input %q: expected line %d, got %d", tC.input, 1, lexer.line)
			}
		})
	}
}

func TestNewlineToken(t *testing.T) {
	lexer := New("\nsay")
	lexer.markTokenStart()
	lexer.advance()

	token := lexer.makeToken(TokenNewline)
	if token.Kind != TokenNewline {
		t.Errorf("expected kind %s, got %s", TokenNewline, token.Kind)
	}
	if testTokenLexeme(lexer.source, token) != "\n" {
		t.Errorf("expected lexeme %q, got %q", "\n", testTokenLexeme(lexer.source, token))
	}
	if testTokenSpan(lexer.source, token).Start.Line != 1 {
		t.Errorf("expected start line %d, got %d", 1, testTokenSpan(lexer.source, token).Start.Line)
	}
	if testTokenSpan(lexer.source, token).Start.Column != 1 {
		t.Errorf("expected start column %d, got %d", 1, testTokenSpan(lexer.source, token).Start.Column)
	}

	if lexer.current != 1 {
		t.Errorf("expected current %d, got %d", 1, lexer.current)
	}
	if lexer.line != 2 {
		t.Errorf("expected line %d, got %d", 2, lexer.line)
	}
	if lexer.column != 1 {
		t.Errorf("expected column %d, got %d", 1, lexer.column)
	}
}

func TestNextToken(t *testing.T) {
	testCases := []struct {
		input          string
		expectedKind   TokenKind
		expectedLexeme string
		expectedLine   int
		expectedColumn int
	}{
		{"", TokenEOF, "", 1, 1},
		{"   ", TokenEOF, "", 1, 4},
		{"  \n", TokenNewline, "\n", 1, 3},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)
			token, err := lexer.NextToken()

			if err != nil {
				t.Fatalf("unexpected lexical error: %v", err)
			}
			if token.Kind != tC.expectedKind {
				t.Errorf("input %q: expected kind %s, got %s", tC.input, tC.expectedKind, token.Kind)
			}
			if testTokenLexeme(lexer.source, token) != tC.expectedLexeme {
				t.Errorf("input %q: expected lexeme %q, got %q", tC.input, tC.expectedLexeme, testTokenLexeme(lexer.source, token))
			}
			if testTokenSpan(lexer.source, token).Start.Line != tC.expectedLine {
				t.Errorf("input %q: expected start line %d, got %d", tC.input, tC.expectedLine, testTokenSpan(lexer.source, token).Start.Line)
			}
			if testTokenSpan(lexer.source, token).Start.Column != tC.expectedColumn {
				t.Errorf("input %q: expected start column %d, got %d", tC.input, tC.expectedColumn, testTokenSpan(lexer.source, token).Start.Column)
			}
		})
	}
}

func TestNextTokenSequence(t *testing.T) {
	testCases := []struct {
		expectedKind   TokenKind
		expectedLexeme string
		expectedLine   int
		expectedColumn int
	}{
		{TokenLet, "let", 1, 1},
		{TokenIdent, "player_1", 1, 5},
		{TokenBe, "be", 1, 14},
		{TokenNewline, "\n", 1, 16},
		{TokenSay, "say", 2, 1},
		{TokenIdent, "player_1", 2, 5},
		{TokenEOF, "", 2, 13},
	}
	lexer := New("let player_1 be\nsay player_1")
	for _, tC := range testCases {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("unexpected lexical error: %v", err)
		}
		if token.Kind != tC.expectedKind {
			t.Errorf("expected kind %s, got %s", tC.expectedKind, token.Kind)
		}
		if testTokenLexeme(lexer.source, token) != tC.expectedLexeme {
			t.Errorf("expected lexeme %q, got %q", tC.expectedLexeme, testTokenLexeme(lexer.source, token))
		}
		if testTokenSpan(lexer.source, token).Start.Line != tC.expectedLine {
			t.Errorf("expected start line %d, got %d", tC.expectedLine, testTokenSpan(lexer.source, token).Start.Line)
		}
		if testTokenSpan(lexer.source, token).Start.Column != tC.expectedColumn {
			t.Errorf("expected start column %d, got %d", tC.expectedColumn, testTokenSpan(lexer.source, token).Start.Column)
		}
	}
}

func TestIsDigit(t *testing.T) {
	testCases := []struct {
		input    byte
		expected bool
	}{
		{'0', true},
		{'5', true},
		{'9', true},
		{'a', false},
		{'_', false},
		{' ', false},
		{'-', false},
	}
	for _, tC := range testCases {
		t.Run(string(tC.input), func(t *testing.T) {
			result := isDigit(tC.input)
			if result != tC.expected {
				t.Errorf("input %q: expected %t, got %t", tC.input, tC.expected, result)
			}
		})
	}
}

func TestScanInteger(t *testing.T) {
	testCases := []struct {
		input           string
		expectedLexeme  string
		expectedCurrent int
	}{
		{"123", "123", 3},
		{"0", "0", 1},
		{"007", "007", 3},
		{"42 age", "42", 2},
		{"123abc", "123", 3},
		{"99\n", "99", 2},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)
			lexer.advance()
			result := lexer.scanInteger()
			if lexer.start != 0 {
				t.Errorf("expected start to remain 0, got %d", lexer.start)
			}
			if result != tC.expectedLexeme {
				t.Errorf("input %q: expected lexeme %q, got %q", tC.input, tC.expectedLexeme, result)
			}
			if lexer.current != tC.expectedCurrent {
				t.Errorf("input %q: expected current %d, got %d", tC.input, tC.expectedCurrent, lexer.current)
			}

			if tC.input == "99\n" && lexer.peek() != '\n' {
				t.Errorf("input %q: expected next character %q, got %q", tC.input, '\n', lexer.peek())
			}
		})
	}
}

func TestNextTokenIntegerSequence(t *testing.T) {
	testCases := []struct {
		expectedKind   TokenKind
		expectedLexeme string
		expectedLine   int
		expectedColumn int
	}{
		{TokenLet, "let", 1, 1},
		{TokenIdent, "age", 1, 5},
		{TokenBe, "be", 1, 9},
		{TokenInt, "20", 1, 12},
		{TokenEOF, "", 1, 14},
	}
	lexer := New("let age be 20")
	for _, tC := range testCases {
		token, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("unexpected lexical error: %v", err)
		}
		if token.Kind != tC.expectedKind {
			t.Errorf("expected kind %s, got %s", tC.expectedKind, token.Kind)
		}
		if testTokenLexeme(lexer.source, token) != tC.expectedLexeme {
			t.Errorf("expected lexeme %q, got %q", tC.expectedLexeme, testTokenLexeme(lexer.source, token))
		}
		if testTokenSpan(lexer.source, token).Start.Line != tC.expectedLine {
			t.Errorf("expected start line %d, got %d", tC.expectedLine, testTokenSpan(lexer.source, token).Start.Line)
		}
		if testTokenSpan(lexer.source, token).Start.Column != tC.expectedColumn {
			t.Errorf("expected start column %d, got %d", tC.expectedColumn, testTokenSpan(lexer.source, token).Start.Column)
		}
	}
}

func TestHasLeadingZero(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"0", false},
		{"7", false},
		{"42", false},
		{"1002", false},
		{"007", true},
		{"00", true},
		{"0123", true},
		{"", false},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			result := hasLeadingZero(tC.input)
			if result != tC.expected {
				t.Errorf("input %q: expected %t, got %t", tC.input, tC.expected, result)
			}
		})
	}
}

func TestNextTokenUnexpectedCharacter(t *testing.T) {
	lexer := New("let age be @")

	for i := 0; i < 3; i++ {
		_, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("unexpected error before invalid character: %v", err)
		}
	}

	token, err := lexer.NextToken()
	if err == nil {
		t.Fatal("expected lexical error, got nil")
	}

	lexErr, ok := err.(*LexError)
	if !ok {
		t.Fatalf("expected *LexError, got %T", err)
	}

	expectedMessage := "unexpected character"
	if lexErr.Position.Line != 1 {
		t.Errorf("expected error line %d, got %d", 1, lexErr.Position.Line)
	}
	if lexErr.Position.Column != 12 {
		t.Errorf("expected error column %d, got %d", 12, lexErr.Position.Column)
	}
	if lexErr.Message != expectedMessage {
		t.Errorf("expected message %q, got %q", expectedMessage, lexErr.Message)
	}

	if token.Kind != TokenInvalid {
		t.Errorf("expected kind %s, got %s", TokenInvalid, token.Kind)
	}
}

func TestNextTokenLeadingZero(t *testing.T) {
	lexer := New("let age be 007")

	for i := 0; i < 3; i++ {
		_, err := lexer.NextToken()
		if err != nil {
			t.Fatalf("unexpected error before invalid integer literal: %v", err)
		}
	}

	token, err := lexer.NextToken()
	if err == nil {
		t.Fatal("expected lexical error, got nil")
	}

	lexErr, ok := err.(*LexError)
	if !ok {
		t.Fatalf("expected *LexError, got %T", err)
	}

	expectedMessage := "integer literal cannot have leading zeros"
	if lexErr.Position.Line != 1 {
		t.Errorf("expected error line %d, got %d", 1, lexErr.Position.Line)
	}
	if lexErr.Position.Column != 12 {
		t.Errorf("expected error column %d, got %d", 12, lexErr.Position.Column)
	}
	if lexErr.Message != expectedMessage {
		t.Errorf("expected message %q, got %q", expectedMessage, lexErr.Message)
	}

	if token.Kind != TokenInvalid {
		t.Errorf("expected kind %s, got %s", TokenInvalid, token.Kind)
	}
	if lexer.current != 14 {
		t.Errorf("expected current %d, got %d", 14, lexer.current)
	}
}

func TestNextTokenInvalidIntegerSuffix(t *testing.T) {
	testCases := []struct {
		input               string
		expectedLine        int
		expectedColumn      int
		expectedMessage     string
		expectedValidTokens int
	}{
		{"123abc", 1, 1, "invalid integer literal", 0},
		{"42_name", 1, 1, "invalid integer literal", 0},
		{"let age be 123abc", 1, 12, "invalid integer literal", 3},
		{"007abc", 1, 1, "invalid integer literal", 0},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)

			for i := 0; i < tC.expectedValidTokens; i++ {
				_, err := lexer.NextToken()
				if err != nil {
					t.Fatalf("unexpected error before invalid integer literal: %v", err)
				}
			}

			token, err := lexer.NextToken()
			if err == nil {
				t.Fatal("expected lexical error, got nil")
			}

			lexErr, ok := err.(*LexError)
			if !ok {
				t.Fatalf("expected *LexError, got %T", err)
			}

			if lexErr.Position.Line != tC.expectedLine {
				t.Errorf("expected error line %d, got %d", tC.expectedLine, lexErr.Position.Line)
			}
			if lexErr.Position.Column != tC.expectedColumn {
				t.Errorf("expected error column %d, got %d", tC.expectedColumn, lexErr.Position.Column)
			}
			if lexErr.Message != tC.expectedMessage {
				t.Errorf("expected message %q, got %q", tC.expectedMessage, lexErr.Message)
			}

			if token.Kind != TokenInvalid {
				t.Errorf("expected kind %s, got %s", TokenInvalid, token.Kind)
			}
		})
	}
}

func TestNextTokenValidIntegers(t *testing.T) {
	testCases := []struct {
		input          string
		expectedLexeme string
	}{
		{"0", "0"},
		{"7", "7"},
		{"1002", "1002"},
		{"123 ", "123"},
		{"99\n", "99"},
	}

	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			lexer := New(tC.input)

			token, err := lexer.NextToken()
			if err != nil {
				t.Fatalf("unexpected lexical error: %v", err)
			}

			if token.Kind != TokenInt {
				t.Errorf("expected kind %s, got %s", TokenInt, token.Kind)
			}
			if testTokenLexeme(lexer.source, token) != tC.expectedLexeme {
				t.Errorf("expected lexeme %q, got %q", tC.expectedLexeme, testTokenLexeme(lexer.source, token))
			}
			if testTokenSpan(lexer.source, token).Start.Line != 1 {
				t.Errorf("expected start line %d, got %d", 1, testTokenSpan(lexer.source, token).Start.Line)
			}
			if testTokenSpan(lexer.source, token).Start.Column != 1 {
				t.Errorf("expected start column %d, got %d", 1, testTokenSpan(lexer.source, token).Start.Column)
			}
		})
	}
}

func TestPopPendingToken(t *testing.T) {
	l := New("")

	first := Token{
		Kind:  TokenDedent,
		Start: 0, End: 0,
	}
	second := Token{
		Kind:  TokenDedent,
		Start: 0, End: 0,
	}

	l.pendingTokens = []Token{first, second}

	for i := 0; i < 2; i++ {
		token, ok := l.popPendingToken()
		if !ok {
			t.Fatalf("pop %d: expected a pending token", i+1)
		}
		if token.Kind != TokenDedent {
			t.Errorf("pop %d: expected DEDENT, got %q", i+1, token.Kind)
		}
		if token.Start != first.Start || token.End != first.End {
			t.Errorf("pop %d: expected position %+v, got %+v",
				i+1, first, token)
		}
	}

	token, ok := l.popPendingToken()
	if ok {
		t.Errorf("expected empty queue, got token %+v", token)
	}
	if token != (Token{}) {
		t.Errorf("expected zero-value token, got %+v", token)
	}
}

func TestPopPendingTokenReusesStorage(t *testing.T) {
	l := New("")

	l.pendingTokens = append(l.pendingTokens, Token{Kind: TokenIndent})
	originalAddress := &l.pendingTokens[0]

	first, ok := l.popPendingToken()
	if !ok || first.Kind != TokenIndent {
		t.Fatalf("first pop = (%v, %v), want INDENT, true", first, ok)
	}

	if len(l.pendingTokens) != 0 || l.pendingIndex != 0 {
		t.Fatalf(
			"drained queue: length = %d, index = %d; want 0, 0",
			len(l.pendingTokens), l.pendingIndex,
		)
	}

	// Append another token and verify that the backing array is reused.
	l.pendingTokens = append(l.pendingTokens, Token{Kind: TokenDedent})

	if &l.pendingTokens[0] != originalAddress {
		t.Fatal("pending-token queue did not reuse its backing array")
	}

	second, ok := l.popPendingToken()
	if !ok || second.Kind != TokenDedent {
		t.Fatalf("second pop = (%v, %v), want DEDENT, true", second, ok)
	}
}

func TestNextTokenPendingTokens(t *testing.T) {
	source := "say"
	l := New(source)

	first := Token{Kind: TokenDedent, Start: 0, End: 0}
	second := Token{Kind: TokenIndent, Start: 0, End: 0}

	l.pendingTokens = []Token{first, second}

	expected := []expectedToken{
		{Kind: TokenDedent, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 1}}},
		{Kind: TokenIndent, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 1}}},
		{
			Kind:   TokenSay,
			Lexeme: "say",
			Span:   Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}},
		},
	}

	for i, want := range expected {
		got, err := l.NextToken()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %v", i+1, err)
		}

		assertToken(t, source, got, want)

		if i < 2 && l.current != 0 {
			t.Errorf(
				"token %d: expected source cursor to remain at 0, got %d",
				i+1,
				l.current,
			)
		}
	}

	if len(l.pendingTokens) != 0 {
		t.Errorf("expected empty pending-token queue, got %d tokens", len(l.pendingTokens))
	}
}

func TestMeasureIndentation(t *testing.T) {
	testCases := []struct {
		name          string
		source        string
		expectedWidth int
		expectedNext  byte
	}{
		{"no indentation", "say age", 0, 's'},
		{"one level", "    say age", 4, 's'},
		{"two levels", "        say age", 8, 's'},
		{"blank line", "    \n", 4, '\n'},
		{"comment line", "    # comment", 4, '#'},
		{"tab after spaces", "    \tsay", 4, '\t'},
		{"empty source", "", 0, 0},
		{"spaces only", "    ", 4, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			l := New(tc.source)

			width := l.measureIndentation()

			if width != tc.expectedWidth {
				t.Errorf(
					"expected indentation width %d, got %d",
					tc.expectedWidth,
					width,
				)
			}

			if l.current != tc.expectedWidth {
				t.Errorf(
					"expected cursor offset %d, got %d",
					tc.expectedWidth,
					l.current,
				)
			}

			if l.column != tc.expectedWidth+1 {
				t.Errorf(
					"expected column %d, got %d",
					tc.expectedWidth+1,
					l.column,
				)
			}

			if next := l.peek(); next != tc.expectedNext {
				t.Errorf(
					"expected next byte %q, got %q",
					tc.expectedNext,
					next,
				)
			}

			if len(l.indentStack) != 1 || l.indentStack[0] != 0 {
				t.Errorf(
					"expected unchanged indentation stack [0], got %v",
					l.indentStack,
				)
			}

			if !l.atLineStart {
				t.Error("expected atLineStart to remain true")
			}
		})
	}
}

func TestIsValidIndentationWidth(t *testing.T) {
	testCases := []struct {
		name     string
		width    int
		expected bool
	}{
		{"zero indentation", 0, true},
		{"one level", 4, true},
		{"two levels", 8, true},
		{"three levels", 12, true},
		{"two spaces", 2, false},
		{"six spaces", 6, false},
		{"ten spaces", 10, false},
		{"negative multiple", -4, false},
		{"negative non-multiple", -2, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidIndentationWidth(tc.width)

			if got != tc.expected {
				t.Errorf(
					"isValidIndentationWidthTest(%d): expected %t, got %t",
					tc.width,
					tc.expected,
					got,
				)
			}
		})
	}
}
func TestProcessIndentation(t *testing.T) {
	tests := []struct {
		name           string
		initialStack   []int
		width          int
		expectedStack  []int
		expectedTokens []expectedToken
		wantError      bool
	}{
		{
			name:          "increase from base level",
			initialStack:  []int{0},
			width:         4,
			expectedStack: []int{0, 4},
			expectedTokens: []expectedToken{
				{Kind: TokenIndent, Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 5}}},
			},
		},
		{
			name:          "increase from existing level",
			initialStack:  []int{0, 4},
			width:         8,
			expectedStack: []int{0, 4, 8},
			expectedTokens: []expectedToken{
				{Kind: TokenIndent, Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 5}}},
			},
		},
		{
			name:          "unchanged indentation",
			initialStack:  []int{0, 4},
			width:         4,
			expectedStack: []int{0, 4},
		},
		{
			name:          "invalid indentation width",
			initialStack:  []int{0, 4},
			width:         6,
			expectedStack: []int{0, 4},
			wantError:     true,
		},
		{
			name:          "negative indentation width",
			initialStack:  []int{0},
			width:         -4,
			expectedStack: []int{0},
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New("")
			l.line = 2
			l.column = 5
			l.indentStack = append([]int(nil), tt.initialStack...)

			err := l.processIndentation(tt.width)

			if (err != nil) != tt.wantError {
				t.Fatalf("processIndentation(%d) error = %v, wantError = %t",
					tt.width, err, tt.wantError)
			}

			if !reflect.DeepEqual(l.indentStack, tt.expectedStack) {
				t.Errorf("indentStack = %v, want %v",
					l.indentStack, tt.expectedStack)
			}

			if len(l.pendingTokens) != len(tt.expectedTokens) {
				t.Fatalf("pendingTokens = %v, want %v",
					l.pendingTokens, tt.expectedTokens)
			}

			for i, expected := range tt.expectedTokens {
				actual := l.pendingTokens[i]
				if actual.Kind != expected.Kind {
					t.Errorf("pendingTokens[%d].Kind = %v, want %v", i, actual.Kind, expected.Kind)
				}
				if actual.Start != actual.End {
					t.Errorf("pendingTokens[%d] = %+v, want a zero-width indentation token", i, actual)
				}
			}
		})
	}
}

func TestProcessIndentationDedent(t *testing.T) {
	tests := []struct {
		name          string
		initialStack  []int
		width         int
		expectedStack []int
		dedentCount   int
		wantError     bool
	}{
		{
			name:          "single dedent",
			initialStack:  []int{0, 4, 8},
			width:         4,
			expectedStack: []int{0, 4},
			dedentCount:   1,
		},
		{
			name:          "multiple dedents",
			initialStack:  []int{0, 4, 8, 12},
			width:         4,
			expectedStack: []int{0, 4},
			dedentCount:   2,
		},
		{
			name:          "dedent to base",
			initialStack:  []int{0, 4, 8},
			width:         0,
			expectedStack: []int{0},
			dedentCount:   2,
		},
		{
			name:          "nonexistent indentation level",
			initialStack:  []int{0, 4, 12},
			width:         8,
			expectedStack: []int{0, 4, 12},
			wantError:     true,
		},
		{
			name:          "invalid indentation alignment",
			initialStack:  []int{0, 4, 8},
			width:         6,
			expectedStack: []int{0, 4, 8},
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := ""
			l := New(source)
			l.line = 3
			l.column = tt.width + 1
			l.indentStack = append([]int(nil), tt.initialStack...)

			err := l.processIndentation(tt.width)

			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, wantError = %t", err, tt.wantError)
			}

			if !reflect.DeepEqual(l.indentStack, tt.expectedStack) {
				t.Errorf("indentStack = %v, want %v",
					l.indentStack, tt.expectedStack)
			}

			if len(l.pendingTokens) != tt.dedentCount {
				t.Fatalf("pending token count = %d, want %d",
					len(l.pendingTokens), tt.dedentCount)
			}

			for i, token := range l.pendingTokens {
				if token.Kind != TokenDedent {
					t.Errorf("token %d kind = %s, want DEDENT", i, token.Kind)
				}

				gotLexeme := source[token.Start:token.End]
				if gotLexeme != "" {
					t.Errorf("token %d lexeme = %q, want empty", i, gotLexeme)
				}
			}
		})
	}
}

func TestNextTokenIndentation(t *testing.T) {
	source := "say 1\n    say 2\n        say 3\n    say 4\n"

	l := New(source)

	expected := []expectedToken{
		{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
		{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
		{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 2, Column: 1}}},

		{Kind: TokenIndent, Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 5}}},
		{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 8}}},
		{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 2, Column: 9}, End: Position{Line: 2, Column: 10}}},
		{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 2, Column: 10}, End: Position{Line: 3, Column: 1}}},

		{Kind: TokenIndent, Span: Span{Start: Position{Line: 3, Column: 9}, End: Position{Line: 3, Column: 9}}},
		{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 3, Column: 9}, End: Position{Line: 3, Column: 12}}},
		{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 3, Column: 13}, End: Position{Line: 3, Column: 14}}},
		{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 3, Column: 14}, End: Position{Line: 4, Column: 1}}},

		{Kind: TokenDedent, Span: Span{Start: Position{Line: 4, Column: 5}, End: Position{Line: 4, Column: 5}}},
		{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 4, Column: 5}, End: Position{Line: 4, Column: 8}}},
		{Kind: TokenInt, Lexeme: "4", Span: Span{Start: Position{Line: 4, Column: 9}, End: Position{Line: 4, Column: 10}}},
		{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 4, Column: 10}, End: Position{Line: 5, Column: 1}}},
	}

	for i, want := range expected {
		got, err := l.NextToken()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %v", i, err)
		}
		assertToken(t, source, got, want)
	}
}

func TestNextTokenBlankLinesPreserveIndentation(t *testing.T) {
	source := "say 1\n    say 2\n        \n\n    say 3\n"

	l := New(source)

	expectedKinds := []TokenKind{
		TokenSay, TokenInt, TokenNewline,
		TokenIndent, TokenSay, TokenInt, TokenNewline,
		TokenNewline,
		TokenNewline,
		TokenSay, TokenInt, TokenNewline,
	}

	for i, want := range expectedKinds {
		got, err := l.NextToken()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %v", i, err)
		}

		if got.Kind != want {
			t.Fatalf("token %d: got %s, want %s", i, got.Kind, want)
		}
	}

	expectedStack := []int{0, 4}
	if !reflect.DeepEqual(l.indentStack, expectedStack) {
		t.Errorf("indentStack = %v, want %v", l.indentStack, expectedStack)
	}

	if len(l.pendingTokens) != 0 {
		t.Errorf("expected empty pending-token queue, got %v", l.pendingTokens)
	}
}

func TestNextTokenEOFDedentation(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedKinds []TokenKind
	}{
		{
			name:   "single indentation level without final newline",
			source: "say 1\n    say 2",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt,
				TokenDedent, TokenEOF,
			},
		},
		{
			name:   "multiple indentation levels without final newline",
			source: "say 1\n    say 2\n        say 3",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt,
				TokenDedent, TokenDedent, TokenEOF,
			},
		},
		{
			name:   "multiple indentation levels with final newline",
			source: "say 1\n    say 2\n        say 3\n",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenDedent, TokenDedent, TokenEOF,
			},
		},
		{
			name:   "trailing blank lines",
			source: "say 1\n    say 2\n    \n\n",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenNewline, TokenNewline,
				TokenDedent, TokenEOF,
			},
		},
		{
			name:   "no indentation",
			source: "say 1",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expectedKinds {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				if got.Kind != want {
					t.Fatalf("token %d: got %s, want %s", i, got.Kind, want)
				}
			}

			if len(l.indentStack) != 1 || l.indentStack[0] != 0 {
				t.Errorf("indentStack = %v, want [0]", l.indentStack)
			}

			if len(l.pendingTokens) != 0 {
				t.Errorf("pendingTokens = %v, want empty queue", l.pendingTokens)
			}
		})
	}
}

func TestNextTokenRejectsTabs(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedError Position
	}{
		{
			name:          "tab at beginning of file",
			source:        "\tsay 1",
			expectedError: Position{Line: 1, Column: 1},
		},
		{
			name:          "tab after leading spaces",
			source:        "    \tsay 1",
			expectedError: Position{Line: 1, Column: 5},
		},
		{
			name:          "tab between tokens",
			source:        "say\t1",
			expectedError: Position{Line: 1, Column: 4},
		},
		{
			name:          "tab on subsequent line",
			source:        "say 1\n\t say 2",
			expectedError: Position{Line: 2, Column: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for {
				token, err := l.NextToken()
				if err != nil {
					lexErr, ok := err.(*LexError)
					if !ok {
						t.Fatalf("expected *LexError, got %T: %v", err, err)
					}

					if lexErr.Position != tt.expectedError {
						t.Errorf("error position = %+v, want %+v",
							lexErr.Position, tt.expectedError)
					}
					return
				}

				if token.Kind == TokenEOF {
					t.Fatal("expected lexical error for tab, got EOF")
				}
			}
		})
	}
}

func TestSkipComment(t *testing.T) {
	tests := []struct {
		name           string
		source         string
		expectedCursor int
		expectedLine   int
		expectedColumn int
		expectedPeek   byte
	}{
		{
			name:           "comment followed by newline",
			source:         "# hello world\nsay 42",
			expectedCursor: 13,
			expectedLine:   1,
			expectedColumn: 14,
			expectedPeek:   '\n',
		},
		{
			name:           "comment ending at EOF",
			source:         "# hello world",
			expectedCursor: 13,
			expectedLine:   1,
			expectedColumn: 14,
			expectedPeek:   0,
		},
		{
			name:           "empty comment followed by newline",
			source:         "#\nsay 42",
			expectedCursor: 1,
			expectedLine:   1,
			expectedColumn: 2,
			expectedPeek:   '\n',
		},
		{
			name:           "empty comment at EOF",
			source:         "#",
			expectedCursor: 1,
			expectedLine:   1,
			expectedColumn: 2,
			expectedPeek:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			l.skipComment()

			if l.current != tt.expectedCursor {
				t.Errorf("cursor = %d, want %d",
					l.current, tt.expectedCursor)
			}

			if l.line != tt.expectedLine {
				t.Errorf("line = %d, want %d",
					l.line, tt.expectedLine)
			}

			if l.column != tt.expectedColumn {
				t.Errorf("column = %d, want %d",
					l.column, tt.expectedColumn)
			}

			if got := l.peek(); got != tt.expectedPeek {
				t.Errorf("peek = %q, want %q",
					got, tt.expectedPeek)
			}

			if !l.atLineStart {
				t.Error("skipComment unexpectedly changed atLineStart")
			}
		})
	}
}

func TestNextTokenComments(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedKinds []TokenKind
	}{
		{
			name:   "inline comment",
			source: "say 42 # explanation\nsay 10",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenSay, TokenInt, TokenEOF,
			},
		},
		{
			name:   "comment-only line at deeper indentation",
			source: "say 1\n    say 2\n        # comment\n    say 3",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenNewline,
				TokenSay, TokenInt,
				TokenDedent, TokenEOF,
			},
		},
		{
			name:   "comment-only line at lower indentation",
			source: "say 1\n    say 2\n# comment\n    say 3",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenNewline,
				TokenSay, TokenInt,
				TokenDedent, TokenEOF,
			},
		},
		{
			name:   "comment ending at EOF",
			source: "say 42 # final comment",
			expectedKinds: []TokenKind{
				TokenSay, TokenInt, TokenEOF,
			},
		},
		{
			name:   "comment-only file",
			source: "# comment",
			expectedKinds: []TokenKind{
				TokenEOF,
			},
		},
		{
			name:   "multiple comment lines",
			source: "# first\n# second\nsay 1",
			expectedKinds: []TokenKind{
				TokenNewline, TokenNewline,
				TokenSay, TokenInt, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expectedKinds {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				if got.Kind != want {
					t.Fatalf("token %d: got %s, want %s", i, got.Kind, want)
				}
			}

			if len(l.indentStack) != 1 || l.indentStack[0] != 0 {
				t.Errorf("indentStack = %v, want [0]", l.indentStack)
			}

			if len(l.pendingTokens) != 0 {
				t.Errorf("pendingTokens = %v, want empty queue", l.pendingTokens)
			}
		})
	}
}

func TestNextTokenIgnoredLinesAndIndentationBoundaries(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []TokenKind
	}{
		{
			name:   "comment before first indented statement",
			source: "say 1\n    # comment\n    say 2",
			expected: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenNewline,
				TokenIndent, TokenSay, TokenInt,
				TokenDedent, TokenEOF,
			},
		},
		{
			name:   "blank line between indentation levels",
			source: "say 1\n    say 2\n\n        say 3",
			expected: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenNewline,
				TokenIndent, TokenSay, TokenInt,
				TokenDedent, TokenDedent, TokenEOF,
			},
		},
		{
			name:   "comment between nested block and dedent",
			source: "say 1\n    say 2\n        say 3\n    # comment\nsay 4",
			expected: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenNewline,
				TokenDedent, TokenDedent,
				TokenSay, TokenInt, TokenEOF,
			},
		},
		{
			name:   "trailing comment at EOF inside block",
			source: "say 1\n    say 2\n    # final comment",
			expected: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenDedent, TokenEOF,
			},
		},
		{
			name:   "whitespace-only line before dedent",
			source: "say 1\n    say 2\n        \nsay 3",
			expected: []TokenKind{
				TokenSay, TokenInt, TokenNewline,
				TokenIndent, TokenSay, TokenInt, TokenNewline,
				TokenNewline,
				TokenDedent, TokenSay, TokenInt, TokenEOF,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				if got.Kind != want {
					t.Fatalf(
						"token %d: got %s (%+v), want %s",
						i, got.Kind, got, want,
					)
				}
			}

			if len(l.indentStack) != 1 || l.indentStack[0] != 0 {
				t.Errorf("indentStack = %v, want [0]", l.indentStack)
			}

			if l.parenDepth != 0 {
				t.Errorf("parenDepth = %d, want 0", l.parenDepth)
			}

			if len(l.pendingTokens) != 0 {
				t.Errorf("pendingTokens = %v, want empty", l.pendingTokens)
			}

			got, err := l.NextToken()
			if err != nil {
				t.Fatalf("repeated EOF: unexpected error: %v", err)
			}
			if got.Kind != TokenEOF {
				t.Errorf("repeated EOF: got %s, want EOF", got.Kind)
			}
		})
	}
}

func TestTokenKindString(t *testing.T) {
	tests := []struct {
		kind TokenKind
		want string
	}{
		{TokenInvalid, "INVALID"},
		{TokenLet, "LET"},
		{TokenBe, "BE"},
		{TokenSay, "SAY"},
		{TokenIdent, "IDENT"},
		{TokenInt, "INT"},
		{TokenString, "STRING"},
		{TokenNewline, "NEWLINE"},
		{TokenEOF, "EOF"},
		{TokenIndent, "INDENT"},
		{TokenDedent, "DEDENT"},
		{TokenColon, "COLON"},
		{TokenLeftParen, "LEFT_PAREN"},
		{TokenRightParen, "RIGHT_PAREN"},
		{TokenPlus, "PLUS"},
		{TokenMinus, "MINUS"},
		{TokenStar, "STAR"},
		{TokenPower, "POWER"},
		{TokenSlash, "SLASH"},
		{TokenFloorDiv, "FLOOR_DIV"},
		{TokenPercent, "PERCENT"},
		{TokenChange, "CHANGE"},
		{TokenTo, "TO"},
		{TokenWith, "WITH"},
		{TokenGiving, "GIVING"},
		{TokenKind(255), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.kind.String(); got != tt.want {
			t.Errorf("TokenKind(%d).String() = %q, want %q",
				tt.kind, got, tt.want)
		}
	}
}

func TestNextTokenColon(t *testing.T) {
	source := "say:"
	l := New(source)

	expected := []expectedToken{
		{
			Kind:   TokenSay,
			Lexeme: "say",
			Span:   Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}},
		},
		{
			Kind:   TokenColon,
			Lexeme: ":",
			Span:   Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}},
		},
		{
			Kind:   TokenEOF,
			Lexeme: "",
			Span:   Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}},
		},
	}

	for i, want := range expected {
		got, err := l.NextToken()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %v", i, err)
		}

		assertToken(t, source, got, want)
	}
}

func TestNextTokenParentheses(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "parentheses around integer",
			source: "say (42)",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 9}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 9}, End: Position{Line: 1, Column: 9}}},
			},
		},
		{
			name:   "empty parentheses",
			source: "()",
			expected: []expectedToken{
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 3}}},
			},
		},
		{
			name:   "nested parentheses",
			source: "((7))",
			expected: []expectedToken{
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 6}}},
			},
		},
		{
			name:   "parentheses on separate lines",
			source: "(\n)",
			expected: []expectedToken{
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 2, Column: 1}, End: Position{Line: 2, Column: 2}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 2, Column: 2}, End: Position{Line: 2, Column: 2}}},
			},
		},
		{
			name:   "multiline parenthesized expression",
			source: "say (\n    42\n)",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 7}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 3, Column: 1}, End: Position{Line: 3, Column: 2}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 3, Column: 2}, End: Position{Line: 3, Column: 2}}},
			},
		},
		{
			name:   "nested multiline parentheses",
			source: "say (\n    (42)\n)",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 6}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 2, Column: 6}, End: Position{Line: 2, Column: 8}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 2, Column: 8}, End: Position{Line: 2, Column: 9}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 3, Column: 1}, End: Position{Line: 3, Column: 2}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 3, Column: 2}, End: Position{Line: 3, Column: 2}}},
			},
		},
		{
			name:   "comments inside parentheses",
			source: "say (\n    # explanation\n    42\n)",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 3, Column: 5}, End: Position{Line: 3, Column: 7}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 4, Column: 1}, End: Position{Line: 4, Column: 2}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 4, Column: 2}, End: Position{Line: 4, Column: 2}}},
			},
		},
		{
			name:   "restore indentation after closing parenthesis",
			source: "say:\n    say (\n        42\n    )\n    say 7\nsay 8",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenColon, Lexeme: ":", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 2, Column: 1}}},
				{Kind: TokenIndent, Lexeme: "", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 5}}},
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 8}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 2, Column: 9}, End: Position{Line: 2, Column: 10}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 3, Column: 9}, End: Position{Line: 3, Column: 11}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 4, Column: 5}, End: Position{Line: 4, Column: 6}}},
				{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 4, Column: 6}, End: Position{Line: 5, Column: 1}}},
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 5, Column: 5}, End: Position{Line: 5, Column: 8}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 5, Column: 9}, End: Position{Line: 5, Column: 10}}},
				{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 5, Column: 10}, End: Position{Line: 6, Column: 1}}},
				{Kind: TokenDedent, Lexeme: "", Span: Span{Start: Position{Line: 6, Column: 1}, End: Position{Line: 6, Column: 1}}},
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 6, Column: 1}, End: Position{Line: 6, Column: 4}}},
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 6, Column: 5}, End: Position{Line: 6, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 6, Column: 6}, End: Position{Line: 6, Column: 6}}},
			},
		},
		{
			name:   "closing parenthesis followed by normal newline",
			source: "say (\n    42)\nsay 7",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 7}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 2, Column: 7}, End: Position{Line: 2, Column: 8}}},
				{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 2, Column: 8}, End: Position{Line: 3, Column: 1}}},
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 3, Column: 1}, End: Position{Line: 3, Column: 4}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 3, Column: 5}, End: Position{Line: 3, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 3, Column: 6}, End: Position{Line: 3, Column: 6}}},
			},
		},
		{
			name:   "closing parenthesis followed by tokens on same line",
			source: "say (\n    42) say 7",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenLeftParen, Lexeme: "(", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "42", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 7}}},
				{Kind: TokenRightParen, Lexeme: ")", Span: Span{Start: Position{Line: 2, Column: 7}, End: Position{Line: 2, Column: 8}}},
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 2, Column: 9}, End: Position{Line: 2, Column: 12}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 2, Column: 13}, End: Position{Line: 2, Column: 14}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 2, Column: 14}, End: Position{Line: 2, Column: 14}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenInlineCommentsInsideParentheses(t *testing.T) {
	source := "say (\n    10 + # first operand\n    20 # second operand\n)"

	expected := []TokenKind{
		TokenSay,
		TokenLeftParen,
		TokenInt,
		TokenPlus,
		TokenInt,
		TokenRightParen,
		TokenEOF,
	}

	l := New(source)

	for i, want := range expected {
		got, err := l.NextToken()
		if err != nil {
			t.Fatalf("token %d: unexpected error: %v", i, err)
		}

		if got.Kind != want {
			t.Fatalf("token %d: got %s, want %s", i, got.Kind, want)
		}
	}

	if l.parenDepth != 0 {
		t.Errorf("parenDepth = %d, want 0", l.parenDepth)
	}

	if len(l.parenPositions) != 0 {
		t.Errorf("parenPositions = %v, want empty", l.parenPositions)
	}
}

func TestNextTokenParenDepth(t *testing.T) {
	tests := []struct {
		name          string
		source        string
		expectedDepth []int
	}{
		{
			name:          "single pair",
			source:        "()",
			expectedDepth: []int{1, 0, 0},
		},
		{
			name:          "nested pairs",
			source:        "((7))",
			expectedDepth: []int{1, 2, 2, 1, 0, 0},
		},
		{
			name:          "sequential pairs",
			source:        "()()",
			expectedDepth: []int{1, 0, 1, 0, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expectedDepth {
				token, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				if l.parenDepth != want {
					t.Errorf(
						"token %d (%s): parenDepth = %d, want %d",
						i,
						token.Kind,
						l.parenDepth,
						want,
					)
				}
			}
		})
	}
}

func TestNextTokenUnclosedOpeningParenthesis(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		position Position
	}{
		{
			name:     "standalone opening parenthesis",
			source:   "(",
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "opening parenthesis after expression",
			source:   "say (10 + 5",
			position: Position{Line: 1, Column: 5},
		},
		{
			name:     "nested unclosed parentheses",
			source:   "say ((10",
			position: Position{Line: 1, Column: 6},
		},
		{
			name:     "outer parenthesis remains unclosed",
			source:   "say ((10)",
			position: Position{Line: 1, Column: 5},
		},
		{
			name:     "multiline unclosed parenthesis",
			source:   "say (\n    42",
			position: Position{Line: 1, Column: 5},
		},
		{
			name:     "nested multiline unclosed parentheses",
			source:   "say (\n    (42",
			position: Position{Line: 2, Column: 5},
		},
		{
			name:     "unclosed parenthesis after matched pair",
			source:   "say () + (7",
			position: Position{Line: 1, Column: 10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for {
				token, err := l.NextToken()
				if err != nil {
					lexErr, ok := err.(*LexError)
					if !ok {
						t.Fatalf("expected *LexError, got %T: %v", err, err)
					}

					if lexErr.Position != tt.position {
						t.Errorf(
							"error position = %+v, want %+v",
							lexErr.Position,
							tt.position,
						)
					}

					if lexErr.Message != "unclosed opening parenthesis" {
						t.Errorf(
							"error message = %q, want %q",
							lexErr.Message,
							"unclosed opening parenthesis",
						)
					}

					if token != (Token{}) {
						t.Errorf(
							"error token = %+v, want zero-value Token",
							token,
						)
					}

					if l.parenDepth == 0 {
						t.Error("parenDepth = 0, want positive depth")
					}

					if len(l.parenPositions) != l.parenDepth {
						t.Errorf(
							"parenPositions length = %d, want %d",
							len(l.parenPositions),
							l.parenDepth,
						)
					}

					return
				}

				if token.Kind == TokenEOF {
					t.Fatal("expected unclosed opening parenthesis error, got EOF")
				}
			}
		})
	}
}

func TestNextTokenUnmatchedClosingParenthesis(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		position Position
	}{
		{
			name:     "standalone closing parenthesis",
			source:   ")",
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "closing parenthesis after expression",
			source:   "say 10)",
			position: Position{Line: 1, Column: 7},
		},
		{
			name:     "extra closing parenthesis",
			source:   "say (10))",
			position: Position{Line: 1, Column: 9},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for {
				token, err := l.NextToken()
				if err != nil {
					lexErr, ok := err.(*LexError)
					if !ok {
						t.Fatalf("expected *LexError, got %T: %v", err, err)
					}

					if lexErr.Position != tt.position {
						t.Errorf("error position = %+v, want %+v",
							lexErr.Position, tt.position)
					}

					if lexErr.Message != "unmatched closing parenthesis" {
						t.Errorf("error message = %q, want %q",
							lexErr.Message, "unmatched closing parenthesis")
					}

					if token != (Token{}) {
						t.Errorf("error token = %+v, want zero-value Token", token)
					}

					if l.parenDepth != 0 {
						t.Errorf("parenDepth = %d, want 0", l.parenDepth)
					}

					return
				}

				if token.Kind == TokenEOF {
					t.Fatal("expected unmatched closing parenthesis error, got EOF")
				}
			}
		})
	}
}

func TestNextTokenPlus(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "addition with spaces",
			source: "say 10 + 5",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "10", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 9}}},
				{Kind: TokenInt, Lexeme: "5", Span: Span{Start: Position{Line: 1, Column: 10}, End: Position{Line: 1, Column: 11}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 11}, End: Position{Line: 1, Column: 11}}},
			},
		},
		{
			name:   "addition without spaces",
			source: "1+2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 4}}},
			},
		},
		{
			name:   "consecutive plus operators",
			source: "1++2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "addition across lines",
			source: "say 1\nsay 2+3",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenNewline, Lexeme: "\n", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 2, Column: 1}}},
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 2, Column: 1}, End: Position{Line: 2, Column: 4}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 6}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 2, Column: 6}, End: Position{Line: 2, Column: 7}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 2, Column: 7}, End: Position{Line: 2, Column: 8}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 2, Column: 8}, End: Position{Line: 2, Column: 8}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenMinus(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "subtraction with spaces",
			source: "say 10 - 3",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "10", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 9}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 10}, End: Position{Line: 1, Column: 11}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 11}, End: Position{Line: 1, Column: 11}}},
			},
		},
		{
			name:   "subtraction without spaces",
			source: "10-3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "10", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "unary negation",
			source: "say -7",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 7}}},
			},
		},
		{
			name:   "consecutive minus operators",
			source: "10--3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "10", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 6}}},
			},
		},
		{
			name:   "addition and subtraction",
			source: "1+2-3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 6}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenStar(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "multiplication with spaces",
			source: "say 4 * 5",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "4", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenStar, Lexeme: "*", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenInt, Lexeme: "5", Span: Span{Start: Position{Line: 1, Column: 9}, End: Position{Line: 1, Column: 10}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 10}, End: Position{Line: 1, Column: 10}}},
			},
		},
		{
			name:   "multiplication without spaces",
			source: "2*3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenStar, Lexeme: "*", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 4}}},
			},
		},
		{
			name:   "multiplication with unary minus",
			source: "2*-3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenStar, Lexeme: "*", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "exponentiation operator",
			source: "2**3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPower, Lexeme: "**", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "mixed arithmetic operators",
			source: "1+2*3-4",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenStar, Lexeme: "*", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenInt, Lexeme: "4", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 8}}},
			},
		},
		{
			name:   "power followed by multiplication",
			source: "2***3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPower, Lexeme: "**", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenStar, Lexeme: "*", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 6}}},
			},
		},
		{
			name:   "consecutive power operators",
			source: "2****3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPower, Lexeme: "**", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenPower, Lexeme: "**", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 7}}},
			},
		},
		{
			name:   "power with unary minus",
			source: "2**-3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPower, Lexeme: "**", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 6}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenDivision(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "ordinary division",
			source: "8/2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenSlash, Lexeme: "/", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 4}}},
			},
		},
		{
			name:   "floor division",
			source: "8//2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "division with spaces",
			source: "say 12 / 3",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "12", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenSlash, Lexeme: "/", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 9}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 10}, End: Position{Line: 1, Column: 11}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 11}, End: Position{Line: 1, Column: 11}}},
			},
		},
		{
			name:   "floor division with spaces",
			source: "say 12 // 3",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "12", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 10}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 11}, End: Position{Line: 1, Column: 12}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 12}, End: Position{Line: 1, Column: 12}}},
			},
		},
		{
			name:   "three consecutive slashes",
			source: "8///2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenSlash, Lexeme: "/", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 6}}},
			},
		},
		{
			name:   "four consecutive slashes",
			source: "8////2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 7}}},
			},
		},
		{
			name:   "division with unary minus",
			source: "8/-2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenSlash, Lexeme: "/", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "mixed division and multiplication",
			source: "8//2*3/4",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenStar, Lexeme: "*", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenSlash, Lexeme: "/", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenInt, Lexeme: "4", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 9}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 9}, End: Position{Line: 1, Column: 9}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenPercent(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "remainder without spaces",
			source: "7%3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 4}}},
			},
		},
		{
			name:   "remainder with spaces",
			source: "say 7 % 3",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 9}, End: Position{Line: 1, Column: 10}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 10}, End: Position{Line: 1, Column: 10}}},
			},
		},
		{
			name:   "negative dividend",
			source: "-7%3",
			expected: []expectedToken{
				{Kind: TokenMinus, Lexeme: "-", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "consecutive percent characters",
			source: "7%%3",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "7", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 5}}},
			},
		},
		{
			name:   "floor division and remainder",
			source: "8//3%2",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "8", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenFloorDiv, Lexeme: "//", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 7}}},
			},
		},
		{
			name:   "mixed arithmetic operators",
			source: "2**3%5+1",
			expected: []expectedToken{
				{Kind: TokenInt, Lexeme: "2", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 2}}},
				{Kind: TokenPower, Lexeme: "**", Span: Span{Start: Position{Line: 1, Column: 2}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenInt, Lexeme: "3", Span: Span{Start: Position{Line: 1, Column: 4}, End: Position{Line: 1, Column: 5}}},
				{Kind: TokenPercent, Lexeme: "%", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 6}}},
				{Kind: TokenInt, Lexeme: "5", Span: Span{Start: Position{Line: 1, Column: 6}, End: Position{Line: 1, Column: 7}}},
				{Kind: TokenPlus, Lexeme: "+", Span: Span{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenInt, Lexeme: "1", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 9}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 9}, End: Position{Line: 1, Column: 9}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}

				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []expectedToken
	}{
		{
			name:   "double-quoted string",
			source: `say "Hello"`,
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenString, Lexeme: `"Hello"`, Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 12}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 12}, End: Position{Line: 1, Column: 12}}},
			},
		},
		{
			name:   "single-quoted string",
			source: "say 'Hello'",
			expected: []expectedToken{
				{Kind: TokenSay, Lexeme: "say", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}}},
				{Kind: TokenString, Lexeme: "'Hello'", Span: Span{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 12}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 12}, End: Position{Line: 1, Column: 12}}},
			},
		},
		{
			name:   "empty double-quoted string",
			source: `""`,
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: `""`, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 3}}},
			},
		},
		{
			name:   "empty single-quoted string",
			source: "''",
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: "''", Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 3}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 3}, End: Position{Line: 1, Column: 3}}},
			},
		},
		{
			name:   "single quote inside double quotes",
			source: `"It's fine"`,
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: `"It's fine"`, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 12}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 12}, End: Position{Line: 1, Column: 12}}},
			},
		},
		{
			name:   "double quote inside single quotes",
			source: `'She said "Hi"'`,
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: `'She said "Hi"'`, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 16}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 16}, End: Position{Line: 1, Column: 16}}},
			},
		},
		{
			name:   "comment marker inside string",
			source: `"hello # world"`,
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: `"hello # world"`, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 16}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 16}, End: Position{Line: 1, Column: 16}}},
			},
		},
		{
			name:   "operators inside string",
			source: `"a + b * c / d"`,
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: `"a + b * c / d"`, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 16}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 16}, End: Position{Line: 1, Column: 16}}},
			},
		},
		{
			name:   "adjacent strings",
			source: `"hello"'world'`,
			expected: []expectedToken{
				{Kind: TokenString, Lexeme: `"hello"`, Span: Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 8}}},
				{Kind: TokenString, Lexeme: "'world'", Span: Span{Start: Position{Line: 1, Column: 8}, End: Position{Line: 1, Column: 15}}},
				{Kind: TokenEOF, Lexeme: "", Span: Span{Start: Position{Line: 1, Column: 15}, End: Position{Line: 1, Column: 15}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			for i, want := range tt.expected {
				got, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}
				assertToken(t, tt.source, got, want)
			}
		})
	}
}

func TestNextTokenUnterminatedString(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		position Position
	}{
		{
			name:     "double quote reaches EOF",
			source:   `"hello`,
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "single quote reaches EOF",
			source:   "'hello",
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "double quote reaches newline",
			source:   "\"hello\nworld\"",
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "single quote reaches newline",
			source:   "'hello\nworld'",
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "mismatched double opening quote",
			source:   `"hello'`,
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "mismatched single opening quote",
			source:   `'hello"`,
			position: Position{Line: 1, Column: 1},
		},
		{
			name:     "unterminated string after keyword",
			source:   `say "hello`,
			position: Position{Line: 1, Column: 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			var err error
			for {
				token, scanErr := l.NextToken()
				if scanErr != nil {
					err = scanErr
					break
				}
				if token.Kind == TokenEOF {
					t.Fatal("expected lexical error, got EOF")
				}
			}

			var lexErr *LexError
			if !errors.As(err, &lexErr) {
				t.Fatalf("expected *LexError, got %T: %v", err, err)
			}

			if lexErr.Position != tt.position {
				t.Errorf("position: got %+v, want %+v", lexErr.Position, tt.position)
			}
			if lexErr.Message != "unterminated string literal" {
				t.Errorf("message: got %q, want %q", lexErr.Message, "unterminated string literal")
			}
		})
	}
}

func TestNextTokenStringEscapes(t *testing.T) {
	tests := []struct {
		name   string
		source string
	}{
		{
			name:   "escaped double quote",
			source: `"She said \"Hello\""`,
		},
		{
			name:   "escaped single quote",
			source: `'It\'s working'`,
		},
		{
			name:   "escaped backslash",
			source: `"C:\\Users\\Documents"`,
		},
		{
			name:   "newline escape",
			source: `"First\nSecond"`,
		},
		{
			name:   "tab escape",
			source: `"First\tSecond"`,
		},
		{
			name:   "carriage return escape",
			source: `"First\rSecond"`,
		},
		{
			name:   "single quote escaped inside double quotes",
			source: `"It\'s fine"`,
		},
		{
			name:   "double quote escaped inside single quotes",
			source: `'She said \"Hi\"'`,
		},
		{
			name:   "multiple escapes",
			source: `"a\nb\tc\\d\"e"`,
		},
		{
			name:   "escaped quote before closing quote",
			source: `"hello\""`,
		},
		{
			name:   "escaped backslash before closing quote",
			source: `"hello\\"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			got, err := l.NextToken()
			if err != nil {
				t.Fatalf("unexpected lexical error: %v", err)
			}

			want := expectedToken{
				Kind:   TokenString,
				Lexeme: tt.source,
				Span:   Span{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: len(tt.source) + 1}},
			}
			assertToken(t, tt.source, got, want)

			eof, err := l.NextToken()
			if err != nil {
				t.Fatalf("unexpected error after string: %v", err)
			}
			if eof.Kind != TokenEOF {
				t.Errorf("expected EOF after string, got %s", eof.Kind)
			}
			if testTokenSpan(l.source, eof).Start != (Position{Line: 1, Column: len(tt.source) + 1}) {
				t.Errorf("EOF position: got %+v, want column %d", testTokenSpan(l.source, eof).Start, len(tt.source)+1)
			}
		})
	}
}

func TestNextTokenInvalidStringEscapes(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		position Position
		message  string
	}{
		{
			name:     "unknown escape in double quotes",
			source:   `"hello\q"`,
			position: Position{Line: 1, Column: 7},
			message:  "invalid escape sequence",
		},
		{
			name:     "unknown escape in single quotes",
			source:   `'hello\x'`,
			position: Position{Line: 1, Column: 7},
			message:  "invalid escape sequence",
		},
		{
			name:     "numeric escape not yet supported",
			source:   `"hello\0"`,
			position: Position{Line: 1, Column: 7},
			message:  "invalid escape sequence",
		},
		{
			name:     "trailing backslash before EOF",
			source:   `"hello\`,
			position: Position{Line: 1, Column: 7},
			message:  "invalid escape sequence",
		},
		{
			name:     "backslash followed by actual newline",
			source:   "\"hello\\\nworld\"",
			position: Position{Line: 1, Column: 7},
			message:  "invalid escape sequence",
		},
		{
			name:     "unknown escape on second line",
			source:   "say 1\nsay \"bad\\q\"",
			position: Position{Line: 2, Column: 9},
			message:  "invalid escape sequence",
		},
		{
			name:     "unterminated string without escape",
			source:   `"hello`,
			position: Position{Line: 1, Column: 1},
			message:  "unterminated string literal",
		},
		{
			name:     "unterminated string at newline",
			source:   "\"hello\nworld\"",
			position: Position{Line: 1, Column: 1},
			message:  "unterminated string literal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)

			var err error
			for {
				token, scanErr := l.NextToken()
				if scanErr != nil {
					err = scanErr
					break
				}
				if token.Kind == TokenEOF {
					t.Fatal("expected lexical error, got EOF")
				}
			}

			var lexErr *LexError
			if !errors.As(err, &lexErr) {
				t.Fatalf("expected *LexError, got %T: %v", err, err)
			}

			if lexErr.Position != tt.position {
				t.Errorf("position: got %+v, want %+v", lexErr.Position, tt.position)
			}
			if lexErr.Message != tt.message {
				t.Errorf("message: got %q, want %q", lexErr.Message, tt.message)
			}
		})
	}
}

// TestNextTokenSpans verifies half-open source spans, including synthetic tokens.
func TestNextTokenSpans(t *testing.T) {
	tests := []struct {
		name   string
		source string
		spans  []Span
	}{
		{"empty input", "", []Span{{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 1}}}},
		{"ordinary tokens and EOF", "say 42", []Span{
			{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}},
			{Start: Position{Line: 1, Column: 5}, End: Position{Line: 1, Column: 7}},
			{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 7}},
		}},
		{"newline crosses lines", "say\n", []Span{
			{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}},
			{Start: Position{Line: 1, Column: 4}, End: Position{Line: 2, Column: 1}},
			{Start: Position{Line: 2, Column: 1}, End: Position{Line: 2, Column: 1}},
		}},
		{"indent and EOF dedent", "say\n    7", []Span{
			{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 4}},
			{Start: Position{Line: 1, Column: 4}, End: Position{Line: 2, Column: 1}},
			{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 5}},
			{Start: Position{Line: 2, Column: 5}, End: Position{Line: 2, Column: 6}},
			{Start: Position{Line: 2, Column: 6}, End: Position{Line: 2, Column: 6}},
			{Start: Position{Line: 2, Column: 6}, End: Position{Line: 2, Column: 6}},
		}},
		{"quoted string with escape", `"a\nb"`, []Span{
			{Start: Position{Line: 1, Column: 1}, End: Position{Line: 1, Column: 7}},
			{Start: Position{Line: 1, Column: 7}, End: Position{Line: 1, Column: 7}},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.source)
			for i, want := range tt.spans {
				token, err := l.NextToken()
				if err != nil {
					t.Fatalf("token %d: unexpected error: %v", i, err)
				}
				if testTokenSpan(l.source, token) != want {
					t.Errorf("token %d (%s): span = %+v, want %+v", i, token.Kind, testTokenSpan(l.source, token), want)
				}

			}
		})
	}
}

func TestLookupIdentifierConversationalKeywords(t *testing.T) {
	tests := []struct {
		word string
		want TokenKind
	}{
		{"change", TokenChange},
		{"to", TokenTo},
		{"with", TokenWith},
		{"giving", TokenGiving},
		{"Change", TokenIdent},
		{"changing", TokenIdent},
		{"together", TokenIdent},
		{"without", TokenIdent},
		{"give", TokenIdent},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			if got := lookupIdentifier(tt.word); got != tt.want {
				t.Errorf(
					"lookupIdentifier(%q) = %v, want %v",
					tt.word,
					got,
					tt.want,
				)
			}
		})
	}
}
