package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/ast"
)

// ParseStatus describes how a frontend parsing attempt terminated.
//
// A recovered or aborted parse may retain a partial AST, but neither
// outcome guarantees that the source program is valid for execution.
type ParseStatus uint8

const (
	// ParseComplete indicates that parsing reached EOF without errors.
	ParseComplete ParseStatus = iota

	// ParseRecovered indicates that parsing reached EOF after recovering
	// from one or more errors.
	ParseRecovered

	// ParseAborted indicates that parsing stopped before successful
	// recovery, such as when the error budget was exhausted.
	ParseAborted
)

// ParseOptions configures diagnostic recovery during frontend parsing.
//
// MaxErrors limits the number of error-severity diagnostics collected.
// Zero selects the default limit of 20; negative values disable the
// limit. Warnings and informational diagnostics do not count toward it.
type ParseOptions struct {
	MaxErrors int
}

// ParseResult contains the AST, diagnostics, and termination status
// produced by a recovery-capable frontend parsing attempt.
//
// Program may contain only successfully recovered statements when
// Status is ParseRecovered or ParseAborted. Callers must not treat
// such a program as valid for execution without further validation.
type ParseResult struct {
	Program     *ast.FlatProgram
	Diagnostics []*diagnostic.Diagnostic
	Status      ParseStatus
}

// recoverySession tracks diagnostics and the error budget for a single
// recovery-capable parsing attempt.
//
// Diagnostics are accumulated lazily so successful parsing does not
// allocate a diagnostic collection. The session does not own parser
// navigation, lexical scanning, or AST construction.
type recoverySession struct {
	diagnostics []*diagnostic.Diagnostic
	maxErrors   int
	errorCount  int
}

// newRecoverySession initializes diagnostic accumulation for a single
// recovery-capable parsing attempt.
//
// A zero MaxErrors value selects the default error budget of 20.
// Negative values disable the error limit. The diagnostic slice remains
// nil until the first diagnostic is recorded.
func newRecoverySession(options ParseOptions) recoverySession {
	if options.MaxErrors == 0 {
		options.MaxErrors = 20
	}
	return recoverySession{maxErrors: options.MaxErrors}
}

// limitReached reports whether the session has exhausted its configured
// error budget.
//
// Negative budgets represent unlimited recovery and never reach a limit.
// Only error-severity diagnostics contribute to the error count.
func (s *recoverySession) limitReached() bool {
	return s.maxErrors > 0 && s.errorCount >= s.maxErrors
}

// add records a diagnostic in encounter order and updates the session's
// error count when its severity is SeverityError.
//
// Nil diagnostics are ignored. Additional errors are rejected once the
// configured budget is exhausted, while warnings and informational
// diagnostics remain acceptable.
//
// The return value indicates whether the error budget is exhausted
// after processing the diagnostic.
func (s *recoverySession) add(diag *diagnostic.Diagnostic) bool {
	if diag == nil {
		return s.limitReached()
	}
	if diag.Severity == diagnostic.SeverityError && s.limitReached() {
		return true
	}

	s.diagnostics = append(s.diagnostics, diag)
	if diag.Severity == diagnostic.SeverityError {
		s.errorCount++
	}

	return s.limitReached()
}

// ParseRecovering parses Eloque source while collecting multiple
// diagnostics and preserving successfully completed statements.
//
// The returned program may be partial when errors are encountered.
// Callers must inspect Status before treating it as executable.
func ParseRecovering(sourceText string, options ParseOptions) ParseResult {
	parser := newRecoveringFlatParser(sourceText)
	session := newRecoverySession(options)

	program, status := parser.parseRecoveringProgram(&session)

	return ParseResult{
		Program:     program,
		Diagnostics: session.diagnostics,
		Status:      status,
	}
}
