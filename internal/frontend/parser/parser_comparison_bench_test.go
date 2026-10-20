package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/frontend/ast"
)

// Benchmark results are retained in package-level variables to ensure
// that parsed programs remain observable after each benchmark iteration.
//
// benchmarkFlatCapacityProgram retains the parsed AST so the
// compiler cannot eliminate the benchmarked construction.
var (
	benchmarkM1Program *ast.Program
	benchmarkH2Program *ast.FlatProgram
)

// benchmarkProgramSource constructs a deterministic Eloque program
// containing repeated variable declarations, reassignments, output
// statements, and arithmetic expressions.
//
// The statement count controls workload size. Every repetition uses
// a distinct variable name so the generated source remains realistic.
//
// Source generation occurs outside the timed benchmark loop.
func benchmarkProgramSource(repetitions int) string {
	var builder strings.Builder

	for i := 0; i < repetitions; i++ {
		fmt.Fprintf(&builder, "let value_%d be 10 + 20 * 3\n", i)
		fmt.Fprintf(&builder, "change value_%d to value_%d + 1\n", i, i)
		fmt.Fprintf(&builder, "say value_%d\n", i)
	}

	return builder.String()
}

// benchmarkM1Parse measures the complete M1 streaming-parser path,
// including tokenization performed while parsing.
//
// Parser construction is included in the timed operation, while
// source generation and benchmark configuration are excluded.
func benchmarkM1Parse(b *testing.B, sourceText string) {
	b.Helper()
	b.ReportAllocs()
	b.SetBytes(int64(len(sourceText)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parser := newStreamingParser(sourceText)

		program, diagnostic := parser.parseProgram()
		if diagnostic != nil {
			b.Fatalf("M1 parse failed: %+v", diagnostic)
		}

		benchmarkM1Program = program
	}
}

// benchmarkH2Parse measures the complete H2 streaming-parser path,
// including tokenization performed while parsing.
//
// Parser construction is included in the timed operation, while
// source generation and benchmark configuration are excluded.
func benchmarkH2Parse(b *testing.B, sourceText string) {
	b.Helper()
	b.ReportAllocs()
	b.SetBytes(int64(len(sourceText)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		parser := newFlatStatementTestParser(sourceText)

		program, diagnostic := parser.parseFlatProgram()
		if diagnostic != nil {
			b.Fatalf("H2 parse failed: %+v", diagnostic)
		}

		benchmarkH2Program = program
	}
}

// BenchmarkParserM1H2 compares the M1 and H2 parser implementations
// using identical Eloque programs at three workload sizes.
//
// Small, medium, and large cases help reveal how parsing time and
// allocation behavior change as the AST grows.
//
// These benchmarks measure end-to-end parsing, including streaming
// lexing, rather than isolated AST node construction.
func BenchmarkParserM1H2(b *testing.B) {
	cases := []struct {
		name        string
		repetitions int
	}{
		{name: "Small", repetitions: 10},
		{name: "Medium", repetitions: 100},
		{name: "Large", repetitions: 1000},
	}

	for _, tc := range cases {
		sourceText := benchmarkProgramSource(tc.repetitions)

		b.Run(tc.name, func(b *testing.B) {
			b.Run("M1", func(b *testing.B) {
				benchmarkM1Parse(b, sourceText)
			})

			b.Run("H2", func(b *testing.B) {
				benchmarkH2Parse(b, sourceText)
			})
		})
	}
}
