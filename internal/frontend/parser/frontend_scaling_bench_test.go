package parser

import (
	"fmt"
	"testing"
)

// BenchmarkFrontendScaling measures the production Eloque frontend
// across increasingly large source programs.
//
// The N10000 workloads extend beyond the regular frontend benchmark
// suite to help identify nonlinear scaling and allocation regressions.
//
// Each iteration includes lexing, parsing, and AST construction.
func BenchmarkFrontendScaling(b *testing.B) {
	for _, workload := range frontendWorkloads {
		for _, size := range []int{10, 100, 1000, 10000} {
			sourceText := workload.build(size)

			b.Run(
				fmt.Sprintf("%s/N%d", workload.name, size),
				func(b *testing.B) {
					program, diag := Parse(sourceText)
					if diag != nil {
						b.Fatalf(
							"invalid benchmark source: %+v",
							diag,
						)
					}
					frontendFlatProgramSink = program

					b.ReportAllocs()
					b.SetBytes(int64(len(sourceText)))
					b.ResetTimer()

					for b.Loop() {
						program, diag := Parse(sourceText)
						if diag != nil {
							b.Fatalf(
								"frontend failed: %+v",
								diag,
							)
						}

						frontendFlatProgramSink = program
					}
				},
			)
		}
	}
}
