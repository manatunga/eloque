package lexer

import "github.com/manatunga/eloque/internal/diagnostic"

// Lexical diagnostic codes identify errors encountered during tokenization.
//
// Each code represents a distinct category of malformed source input.
// Codes remain stable independently of diagnostic messages, allowing
// compiler tooling and tests to identify errors reliably.
const (
	// CodeUnexpectedCharacter identifies a source character that cannot
	// begin any recognized Eloque token.
	CodeUnexpectedCharacter diagnostic.Code = "LEX_UNEXPECTED_CHARACTER"

	// CodeInvalidInteger identifies a malformed decimal integer literal,
	// including invalid characters immediately following its digits.
	CodeInvalidInteger diagnostic.Code = "LEX_INVALID_INTEGER"

	// CodeLeadingZero identifies a multi-digit decimal integer literal
	// beginning with zero, which Eloque does not permit.
	CodeLeadingZero diagnostic.Code = "LEX_LEADING_ZERO"

	// CodeUnterminatedString identifies a string literal that reaches
	// a newline or end of source without a matching closing quote.
	CodeUnterminatedString diagnostic.Code = "LEX_UNTERMINATED_STRING"

	// CodeInvalidEscape identifies an unsupported or incomplete escape
	// sequence within a string literal.
	CodeInvalidEscape diagnostic.Code = "LEX_INVALID_ESCAPE"

	// CodeInvalidIndentation identifies indentation whose width does not
	// satisfy Eloque's four-space indentation requirement.
	CodeInvalidIndentation diagnostic.Code = "LEX_INVALID_INDENTATION"

	// CodeInvalidDedent identifies dedentation that does not correspond
	// to an active indentation level.
	CodeInvalidDedent diagnostic.Code = "LEX_INVALID_DEDENT"

	// CodeUnmatchedClosingParen identifies a closing parenthesis without
	// a corresponding opening parenthesis.
	CodeUnmatchedClosingParen diagnostic.Code = "LEX_UNMATCHED_CLOSING_PAREN"

	// CodeUnclosedOpeningParen identifies an opening parenthesis that
	// remains unmatched when the lexer reaches end of source.
	CodeUnclosedOpeningParen diagnostic.Code = "LEX_UNCLOSED_OPENING_PAREN"
)
