package parser

import (
	"fmt"
	"testing"
)

// BenchmarkFrontendScaling compares buffered and production streaming
// frontends across increasingly large source programs.
//
// The N10000 workloads extend beyond the regular frontend benchmark
// suite to help identify scaling regressions.
func BenchmarkFrontendScaling(b *testing.B) {
	for _, workload := range frontendWorkloads {
		for _, size := range []int{10, 100, 1000, 10000} {
			source := workload.build(size)

			b.Run(fmt.Sprintf("%s/N%d", workload.name, size), func(b *testing.B) {
				b.Run("Buffered", func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(source)))

					for b.Loop() {
						program, err := parseFrontend(source)
						if err != nil {
							b.Fatal(err)
						}

						frontendProgramSink = program
					}
				})

				b.Run("Streaming", func(b *testing.B) {
					b.ReportAllocs()
					b.SetBytes(int64(len(source)))

					for b.Loop() {
						program, err := Parse(source)
						if err != nil {
							b.Fatal(err)
						}

						frontendProgramSink = program
					}
				})
			})
		}
	}
}
