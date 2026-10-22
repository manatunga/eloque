package diagnostic

import "github.com/manatunga/eloque/internal/source"

// Severity classifies the importance of a compiler diagnostic.
//
// Errors prevent normal compilation, while warnings and informational
// diagnostics are nonblocking by default.
type Severity uint8

const (
	// SeverityError identifies a blocking compilation error.
	SeverityError Severity = iota

	// SeverityWarning identifies a nonblocking issue that may require attention.
	SeverityWarning

	// SeverityInfo identifies supplementary diagnostic information.
	SeverityInfo
)

// Code is a stable, machine-readable identifier for a diagnostic.
//
// Codes are independent of human-readable messages and are grouped by
// compiler stage, such as LEX_*, PARSE_*, RESOLVE_*, and TYPE_*.
type Code string

// FileID identifies a source file within a single compilation snapshot.
//
// File identities are local to a compilation and are not guaranteed to
// remain stable across separate compilations or recompilations.
type FileID uint32

// Location identifies a source range associated with a diagnostic.
//
// File identifies the source containing the issue. Span stores a half-open
// byte range, with an inclusive start and exclusive end. Line and column
// positions are derived when diagnostics are presented.
type Location struct {
	File FileID
	Span source.ByteSpan
}

// RelatedLocation identifies another source range relevant to a diagnostic.
//
// Related locations can identify earlier declarations, conflicting symbols,
// or other source constructs that help explain the primary diagnostic.
// Message describes the relationship to the primary location.
type RelatedLocation struct {
	Location Location
	Message  string
}

// TokenExpectation describes a syntax error involving an unexpected token.
//
// Expected contains the stable names of acceptable token kinds.
// Actual identifies the token kind encountered by the parser.
// Token names are strings so the diagnostic package remains independent
// of lexer-specific types and can be consumed by external tooling.
type TokenExpectation struct {
	Expected []string
	Actual   string
}

// Diagnostic represents an issue reported by a compiler stage.
//
// Code provides a stable identifier for tooling, while Message provides
// a human-readable explanation. Primary identifies the main source range
// associated with the issue.
//
// Related and Notes provide optional supplementary context. Their nil
// values indicate that no additional information is available.
//
// Diagnostics are collected by individual compiler stages and later
// coordinated by the compilation driver. This type does not own source
// text, perform recovery, or determine diagnostic ordering.
type Diagnostic struct {
	Code             Code
	Severity         Severity
	Message          string
	Primary          Location
	Related          []RelatedLocation
	Notes            []string
	TokenExpectation *TokenExpectation
}
