package parser

import (
	"fmt"
	"strings"
	"testing"
)

func buildControlledArithmeticWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 64)

	for i := range count {
		fmt.Fprintf(
			&builder,
			"let value_%d be ((%d + 3) * (7 - 2)) + (%d * 4)\n",
			i, i, i,
		)
	}

	return builder.String()
}

func BenchmarkControlledEloqueFrontend(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		source := buildControlledArithmeticWorkload(size)

		b.Run(fmt.Sprintf("N%d", size), func(b *testing.B) {
			program, err := Parse(source)
			if err != nil {
				b.Fatalf("invalid source: %v", err)
			}
			frontendProgramSink = program

			b.ReportAllocs()
			b.SetBytes(int64(len(source)))
			b.ResetTimer()

			for b.Loop() {
				program, err := Parse(source)
				if err != nil {
					b.Fatal(err)
				}
				frontendProgramSink = program
			}
		})
	}
}
