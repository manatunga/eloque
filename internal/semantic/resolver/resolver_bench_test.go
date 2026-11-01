package resolver

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/parser"
)

var benchmarkResolutionSink *Resolution
var benchmarkDiagnosticsSink int

func benchmarkProgram(b *testing.B, sourceText string) *ast.FlatProgram {
	b.Helper()

	program, err := parser.Parse(sourceText)
	if err != nil {
		b.Fatalf("Parse failed: %v", err)
	}
	return program
}

func benchmarkSource(kind string, count int) string {
	var source strings.Builder
	source.Grow(count * 48)

	switch kind {
	case "declarations":
		for i := 0; i < count; i++ {
			fmt.Fprintf(&source, "let value_%d be %d\n", i, i)
		}

	case "references":
		source.WriteString("let value be 1\n")
		for i := 0; i < count; i++ {
			source.WriteString("say value\n")
		}

	case "mixed":
		source.WriteString("let total be 0\n")
		for i := 0; i < count; i++ {
			fmt.Fprintf(&source, "let value_%d be total + %d\n", i, i)
			fmt.Fprintf(&source, "change total to total + value_%d\n", i)
			source.WriteString("say total\n")
		}

	case "undefined":
		for i := 0; i < count; i++ {
			fmt.Fprintf(&source, "say missing_%d\n", i)
		}

	case "sparse-undefined":
		source.WriteString("let value be 1\n")
		for i := 0; i < count; i++ {
			if i%100 == 0 {
				fmt.Fprintf(&source, "say missing_%d\n", i)
			} else {
				source.WriteString("say value\n")
			}
		}
	}

	return source.String()
}

func BenchmarkResolve(b *testing.B) {
	workloads := []string{
		"declarations",
		"references",
		"mixed",
		"undefined",
		"sparse-undefined",
	}
	sizes := []int{10, 100, 1000, 3000}

	for _, workload := range workloads {
		for _, size := range sizes {
			b.Run(fmt.Sprintf("%s/%d", workload, size), func(b *testing.B) {
				program := benchmarkProgram(
					b,
					benchmarkSource(workload, size),
				)

				b.ReportAllocs()
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					result, diagnostics := Resolve(program)
					benchmarkResolutionSink = result
					benchmarkDiagnosticsSink = len(diagnostics)
				}
			})
		}
	}
}
