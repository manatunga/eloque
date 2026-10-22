package parser

import (
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
)

func TestNewRecoverySession(t *testing.T) {
	tests := []struct {
		name     string
		max      int
		expected int
	}{
		{"default budget", 0, 20},
		{"custom budget", 5, 5},
		{"unlimited budget", -1, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := newRecoverySession(ParseOptions{
				MaxErrors: tt.max,
			})

			if session.maxErrors != tt.expected {
				t.Errorf("maxErrors = %d, want %d",
					session.maxErrors, tt.expected)
			}

			if session.diagnostics != nil {
				t.Error("expected nil diagnostic slice before first error")
			}

			if session.errorCount != 0 {
				t.Errorf("errorCount = %d, want 0", session.errorCount)
			}
		})
	}
}

func TestRecoverySessionErrorBudget(t *testing.T) {
	session := newRecoverySession(ParseOptions{MaxErrors: 2})

	first := &diagnostic.Diagnostic{
		Code:     "TEST_FIRST",
		Severity: diagnostic.SeverityError,
	}
	second := &diagnostic.Diagnostic{
		Code:     "TEST_SECOND",
		Severity: diagnostic.SeverityError,
	}
	third := &diagnostic.Diagnostic{
		Code:     "TEST_THIRD",
		Severity: diagnostic.SeverityError,
	}

	if session.add(first) {
		t.Fatal("budget exhausted after first error")
	}

	if !session.add(second) {
		t.Fatal("expected budget exhaustion after second error")
	}

	if !session.add(third) {
		t.Fatal("expected budget to remain exhausted")
	}

	if session.errorCount != 2 {
		t.Errorf("errorCount = %d, want 2", session.errorCount)
	}

	if len(session.diagnostics) != 2 {
		t.Errorf("diagnostic count = %d, want 2",
			len(session.diagnostics))
	}

	if session.diagnostics[0] != first ||
		session.diagnostics[1] != second {
		t.Error("diagnostics were not retained in encounter order")
	}
}

func TestRecoverySessionNonErrorDiagnostics(t *testing.T) {
	session := newRecoverySession(ParseOptions{MaxErrors: 1})

	errDiag := &diagnostic.Diagnostic{
		Severity: diagnostic.SeverityError,
	}
	warning := &diagnostic.Diagnostic{
		Severity: diagnostic.SeverityWarning,
	}
	info := &diagnostic.Diagnostic{
		Severity: diagnostic.SeverityInfo,
	}

	if !session.add(errDiag) {
		t.Fatal("expected error budget to be exhausted")
	}

	session.add(warning)
	session.add(info)

	if session.errorCount != 1 {
		t.Errorf("errorCount = %d, want 1", session.errorCount)
	}

	if len(session.diagnostics) != 3 {
		t.Errorf("diagnostic count = %d, want 3",
			len(session.diagnostics))
	}

	if session.diagnostics[1] != warning ||
		session.diagnostics[2] != info {
		t.Error("non-error diagnostics were not retained")
	}
}

func TestRecoverySessionUnlimitedBudget(t *testing.T) {
	session := newRecoverySession(ParseOptions{MaxErrors: -1})

	for i := 0; i < 100; i++ {
		if session.add(&diagnostic.Diagnostic{
			Severity: diagnostic.SeverityError,
		}) {
			t.Fatalf("unlimited budget exhausted at error %d", i+1)
		}
	}

	if session.limitReached() {
		t.Fatal("unlimited budget should never be exhausted")
	}

	if session.errorCount != 100 {
		t.Errorf("errorCount = %d, want 100", session.errorCount)
	}

	if len(session.diagnostics) != 100 {
		t.Errorf("diagnostic count = %d, want 100",
			len(session.diagnostics))
	}
}

func TestRecoverySessionNilDiagnostic(t *testing.T) {
	session := newRecoverySession(ParseOptions{MaxErrors: 1})

	if session.add(nil) {
		t.Fatal("nil diagnostic should not exhaust budget")
	}

	if session.diagnostics != nil {
		t.Error("nil diagnostic should not allocate a slice")
	}

	if session.errorCount != 0 {
		t.Errorf("errorCount = %d, want 0", session.errorCount)
	}
}
