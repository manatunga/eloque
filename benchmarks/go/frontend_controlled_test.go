package gobench

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func buildControlledGoArithmeticWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 64)

	for i := range count {
		fmt.Fprintf(
			&builder,
			"value_%d := ((%d + 3) * (7 - 2)) + (%d * 4)\n",
			i, i, i,
		)
	}

	return wrapProgram(builder.String())
}

func BenchmarkControlledGoFrontend(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 10000} {
		source := buildControlledGoArithmeticWorkload(size)

		b.Run(fmt.Sprintf("N%d", size), func(b *testing.B) {
			file, err := parser.ParseFile(
				token.NewFileSet(), "", source, 0,
			)
			if err != nil {
				b.Fatalf("invalid source: %v", err)
			}
			parsedFileSink = file

			b.ReportAllocs()
			b.SetBytes(int64(len(source)))
			b.ResetTimer()

			for b.Loop() {
				file, err := parser.ParseFile(
					token.NewFileSet(), "", source, 0,
				)
				if err != nil {
					b.Fatal(err)
				}
				parsedFileSink = file
			}
		})
	}
}
