package diagnostic

import (
	"testing"

	"github.com/manatunga/eloque/internal/source"
)

func TestSeverityValues(t *testing.T) {
	if SeverityError != 0 {
		t.Errorf("SeverityError = %d, want 0", SeverityError)
	}
	if SeverityWarning != 1 {
		t.Errorf("SeverityWarning = %d, want 1", SeverityWarning)
	}
	if SeverityInfo != 2 {
		t.Errorf("SeverityInfo = %d, want 2", SeverityInfo)
	}
}

func TestDiagnosticLocation(t *testing.T) {
	location := Location{
		File: 1,
		Span: source.ByteSpan{
			Start: 4,
			End:   7,
		},
	}

	got := Diagnostic{
		Code:     "RESOLVE_UNDEFINED_NAME",
		Severity: SeverityError,
		Message:  "undefined name",
		Primary:  location,
	}

	if got.Primary != location {
		t.Errorf("Primary = %+v, want %+v", got.Primary, location)
	}

	if got.Primary.Span.End-got.Primary.Span.Start != 3 {
		t.Errorf("span length = %d, want 3",
			got.Primary.Span.End-got.Primary.Span.Start)
	}
}
