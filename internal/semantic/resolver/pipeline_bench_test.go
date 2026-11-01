package resolver

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/parser"
)

var pipelineResolutionSink *Resolution
var pipelineDiagnosticCountSink int

// pipelineSource produces the same declaration/reference topology as m2_native.py.
func pipelineSource(workload string, n int) string {
	var b strings.Builder
	b.Grow(n * 48)
	declaration := func(name string, value int) {
		b.WriteString("let ")
		b.WriteString(name)
		b.WriteString(" be ")
		b.WriteString(strconv.Itoa(value))
		b.WriteByte('\n')
	}
	reference := func(name string) { b.WriteString("say "); b.WriteString(name); b.WriteByte('\n') }
	switch workload {
	case "declarations":
		for i := 0; i < n; i++ {
			declaration(fmt.Sprintf("value_%d", i), i)
		}
		for i := 0; i < n; i++ {
			reference(fmt.Sprintf("value_%d", i))
		}
	case "references":
		declaration("seed", 7)
		for i := 0; i < n; i++ {
			reference("seed")
		}
	case "mixed":
		for i := 0; i < n; i++ {
			name := fmt.Sprintf("value_%d", i)
			declaration(name, i)
			reference(name)
		}
	case "sparse-undefined":
		declaration("seed", 7)
		for i := 0; i < n; i++ {
			if i%100 == 0 {
				reference(fmt.Sprintf("missing_%d", i))
			} else {
				reference("seed")
			}
		}
	case "undefined":
		for i := 0; i < n; i++ {
			reference(fmt.Sprintf("missing_%d", i))
		}
	default:
		panic("unknown workload: " + workload)
	}
	return b.String()
}

func expectedPipelineErrors(workload string, n int) int {
	switch workload {
	case "undefined":
		return n
	case "sparse-undefined":
		return (n + 99) / 100
	default:
		return 0
	}
}

func benchmarkPipeline(b *testing.B, recovering bool) {
	workloads := []string{"declarations", "references", "mixed", "sparse-undefined", "undefined"}
	sizes := []int{10, 100, 1000, 3000}
	for _, workload := range workloads {
		for _, n := range sizes {
			b.Run(fmt.Sprintf("%s/N%d", workload, n), func(b *testing.B) {
				source := pipelineSource(workload, n)
				expected := expectedPipelineErrors(workload, n)
				// Preflight checks are not timed.
				var preflightCount int
				if recovering {
					parsed := parser.ParseRecovering(source, parser.ParseOptions{})
					if parsed.Status != parser.ParseComplete || parsed.Program == nil || len(parsed.Diagnostics) != 0 {
						b.Fatal("recovery parser rejected generated valid syntax")
					}
					resolution, diagnostics := Resolve(parsed.Program)
					preflightCount = len(diagnostics)
					if resolution == nil {
						b.Fatal("nil resolution")
					}
				} else {
					parsed, diagnostic := parser.Parse(source)
					if diagnostic != nil || parsed == nil {
						b.Fatalf("fail-fast parser rejected generated valid syntax: %v", diagnostic)
					}
					resolution, diagnostics := Resolve(parsed)
					preflightCount = len(diagnostics)
					if resolution == nil {
						b.Fatal("nil resolution")
					}
				}
				if preflightCount != expected {
					b.Fatalf("resolver diagnostics: got %d, want %d", preflightCount, expected)
				}
				b.ReportAllocs()
				b.SetBytes(int64(len(source)))
				b.ResetTimer()
				for b.Loop() {
					if recovering {
						parsed := parser.ParseRecovering(source, parser.ParseOptions{})
						if parsed.Status != parser.ParseComplete || parsed.Program == nil {
							b.Fatal("unexpected recovery parse failure")
						}
						resolution, diagnostics := Resolve(parsed.Program)
						pipelineResolutionSink = resolution
						pipelineDiagnosticCountSink = len(diagnostics)
					} else {
						parsed, diagnostic := parser.Parse(source)
						if diagnostic != nil || parsed == nil {
							b.Fatal("unexpected fail-fast parse failure")
						}
						resolution, diagnostics := Resolve(parsed)
						pipelineResolutionSink = resolution
						pipelineDiagnosticCountSink = len(diagnostics)
					}
				}
			})
		}
	}
}

func BenchmarkPipelineRecovering(b *testing.B) { benchmarkPipeline(b, true) }
func BenchmarkPipelineFailFast(b *testing.B)   { benchmarkPipeline(b, false) }
