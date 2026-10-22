package lexer

import "github.com/manatunga/eloque/internal/diagnostic"

// RecoveryStatus describes whether lexical scanning can continue after
// producing a token or diagnostic.
//
// Recoverable failures leave the lexer positioned for a subsequent
// scanning attempt. Terminal failures require the caller to stop.
type RecoveryStatus uint8

const (
	// RecoveryReady indicates that scanning may continue normally.
	RecoveryReady RecoveryStatus = iota

	// RecoveryResynchronized indicates that a lexical error was handled
	// and the lexer is positioned to continue scanning.
	RecoveryResynchronized

	// RecoveryTerminal indicates that scanning cannot safely continue
	// without violating lexer state invariants.
	RecoveryTerminal
)

// RecoveryResult describes the outcome of a recovery-aware lexical
// scanning operation.
//
// A successful scan produces a token and no diagnostic. A failed scan
// produces a diagnostic and indicates whether scanning may continue.
// Callers must not interpret a zero-value token as a valid token.
type RecoveryResult struct {
	Token      Token
	Diagnostic *diagnostic.Diagnostic
	Status     RecoveryStatus
}

// recoverInvalidInteger consumes the remaining identifier characters
// belonging to a malformed integer literal.
//
// NextToken has already consumed the numeric prefix. This method
// discards the trailing identifier suffix so subsequent scanning
// cannot misinterpret it as a separate token.
//
// The lexer remains positioned immediately after the malformed literal.
func (l *Lexer) recoverInvalidInteger() {
	for isIdentifierContinue(l.peek()) {
		l.advance()
	}
}

// recoverInvalidString consumes the unread remainder of a malformed
// string literal after an invalid escape.
//
// The method recognizes escaped quote characters and stops after the
// matching closing quote, or immediately before a physical newline
// or EOF. Previously consumed characters are never rescanned.
//
// The lexer remains positioned after the malformed string or at its
// terminating source boundary.
func (l *Lexer) recoverInvalidString(quote byte) {
	for !l.isAtEnd() && l.peek() != '\n' {
		if l.peek() == quote {
			l.advance()
			return
		}
		if l.peek() == '\\' {
			l.advance()
			if !l.isAtEnd() && l.peek() != '\n' {
				l.advance()
			}
			continue
		}
		l.advance()
	}
}

// ResetRecoveryGrouping clears unmatched opening-parenthesis state
// during parser-directed recovery.
//
// It does not consume source characters, modify indentation levels,
// or affect queued structural tokens.
func (l *Lexer) ResetRecoveryGrouping() {
	l.parenDepth = 0
	l.parenPositions = l.parenPositions[:0]
}

// RecoverToLineBoundary discards unread source characters until the
// next physical newline or end of file.
//
// The newline is left unread so normal tokenization can emit it.
// When resetGrouping is true, unmatched opening-parenthesis state
// is cleared so subsequent newlines regain their structural meaning.
//
// The method does not modify the indentation stack or synthesize
// structural tokens. It returns true if a physical newline was found,
// or false if scanning reached EOF.
func (l *Lexer) RecoverToLineBoundary(resetGrouping bool) bool {
	for !l.isAtEnd() && l.peek() != '\n' {
		l.advance()
	}

	if resetGrouping {
		l.ResetRecoveryGrouping()
	}

	l.atLineStart = false

	return !l.isAtEnd()
}

// RecoveryCrossedNewline reports whether the most recent recoverable
// scan crossed a physical newline suppressed by parenthesis grouping.
//
// The flag is reset at the beginning of each NextRecoverable call.
func (l *Lexer) RecoveryCrossedNewline() bool {
	return l.recoveryCrossedNewline
}

// NextRecoverable scans the next token while supporting recovery from
// malformed lexical input.
//
// Successful scans return RecoveryReady and a valid token. Recoverable
// lexical failures return RecoveryResynchronized and a diagnostic after
// establishing a safe continuation state.
//
// RecoveryTerminal indicates that scanning cannot safely continue.
// Once a terminal condition is encountered, subsequent calls return
// RecoveryTerminal without a diagnostic or further source consumption.
//
// This method does not perform parser-level statement synchronization.
// Terminal state applies only to recovery-aware scanning; NextToken
// retains its existing behavior.
//
// A successful scan produces a token and no diagnostic. A recoverable
// failure produces a diagnostic and permits subsequent scanning.
// A terminal failure produces a diagnostic on its first occurrence;
// subsequent terminal results contain no diagnostic.
//
// Callers must not interpret a zero-value token as a valid token.
func (l *Lexer) NextRecoverable() RecoveryResult {
	if l.recoveryTerminated {
		return RecoveryResult{
			Status: RecoveryTerminal,
		}
	}

	l.recoveryCrossedNewline = false
	l.recoveryScanning = true

	token, diag := l.NextToken()

	l.recoveryScanning = false

	if diag == nil {
		return RecoveryResult{
			Token:  token,
			Status: RecoveryReady,
		}
	}

	switch diag.Code {
	case CodeUnexpectedCharacter,
		CodeUnmatchedClosingParen,
		CodeLeadingZero:
		return RecoveryResult{
			Diagnostic: diag,
			Status:     RecoveryResynchronized,
		}
	case CodeInvalidInteger:
		l.recoverInvalidInteger()
		diag.Primary.Span.End = uint32(l.current)
		return RecoveryResult{
			Diagnostic: diag,
			Status:     RecoveryResynchronized,
		}
	case CodeInvalidEscape:
		quote := l.source[l.start]
		l.recoverInvalidString(quote)

		return RecoveryResult{
			Diagnostic: diag,
			Status:     RecoveryResynchronized,
		}
	case CodeUnterminatedString:
		if !l.isAtEnd() && l.peek() == '\n' {
			return RecoveryResult{
				Diagnostic: diag,
				Status:     RecoveryResynchronized,
			}
		}

		l.recoveryTerminated = true

		return RecoveryResult{
			Diagnostic: diag,
			Status:     RecoveryTerminal,
		}
	case CodeInvalidIndentation, CodeInvalidDedent:
		l.RecoverToLineBoundary(false)

		return RecoveryResult{
			Diagnostic: diag,
			Status:     RecoveryResynchronized,
		}
	default:
		l.recoveryTerminated = true
		return RecoveryResult{
			Diagnostic: diag,
			Status:     RecoveryTerminal,
		}
	}
}
