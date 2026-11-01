package resolver

import (
	"fmt"
	"testing"

	"github.com/manatunga/eloque/internal/diagnostic"
)

var diagnosticBenchmarkSink []diagnostic.Diagnostic

// benchmarkDiagnosticStorage measures diagnostic slice construction using
// either Go's default append growth or the resolver's geometric growth policy.
func benchmarkDiagnosticStorage(b *testing.B, count int, geometric bool) {
	b.Helper()

	sample := diagnostic.Diagnostic{
		Code:     CodeUndefinedName,
		Severity: diagnostic.SeverityError,
		Message:  "undefined variable: missing",
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var diagnostics []diagnostic.Diagnostic

		if geometric {
			r := resolver{}

			for j := 0; j < count; j++ {
				r.addDiagnostic(sample)
			}

			diagnostics = r.diagnostics
		} else {
			for j := 0; j < count; j++ {
				diagnostics = append(diagnostics, sample)
			}
		}

		diagnosticBenchmarkSink = diagnostics
	}
}

// BenchmarkDiagnosticStorage compares diagnostic storage policies across
// small, sparse, and error-heavy diagnostic counts.
func BenchmarkDiagnosticStorage(b *testing.B) {
	counts := []int{1, 8, 9, 16, 17, 30, 32, 33, 100, 1000, 3000}

	for _, count := range counts {
		b.Run(fmt.Sprintf("default/%d", count), func(b *testing.B) {
			benchmarkDiagnosticStorage(b, count, false)
		})

		b.Run(fmt.Sprintf("geometric/%d", count), func(b *testing.B) {
			benchmarkDiagnosticStorage(b, count, true)
		})
	}
}
