package parser

import (
	"fmt"
	"strings"
	"testing"
)

// frontendWorkload describes a valid Eloque source-generation scenario.
type frontendWorkload struct {
	name  string
	build func(int) string
}

// frontendWorkloads defines the programs used to compare both frontends.
var frontendWorkloads = []frontendWorkload{
	{
		name:  "Declarations",
		build: buildDeclarationWorkload,
	},
	{
		name:  "Expressions",
		build: buildExpressionWorkload,
	},
	{
		name:  "Strings",
		build: buildStringWorkload,
	},
	{
		name:  "Comments",
		build: buildCommentWorkload,
	},
	{
		name:  "Mixed",
		build: buildMixedWorkload,
	},
}

// buildDeclarationWorkload creates programs dominated by declarations.
func buildDeclarationWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 32)

	for i := range count {
		fmt.Fprintf(&builder, "let value_%d be %d\n", i, i)
	}

	return builder.String()
}

// buildExpressionWorkload creates arithmetic-heavy programs.
func buildExpressionWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 64)

	for i := range count {
		fmt.Fprintf(
			&builder,
			"let value_%d be ((%d + 3) * (7 - 2)) ** 2\n",
			i, i,
		)
	}

	return builder.String()
}

// buildStringWorkload creates programs containing string literals.
func buildStringWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 80)

	for i := range count {
		fmt.Fprintf(
			&builder,
			"say \"message %d: the quick brown fox jumps over the lazy dog\"\n",
			i,
		)
	}

	return builder.String()
}

// buildCommentWorkload creates programs with substantial comments.
func buildCommentWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 120)

	for i := range count {
		fmt.Fprintf(
			&builder,
			"# This comment describes declaration number %d and its purpose.\n",
			i,
		)
		fmt.Fprintf(&builder, "let value_%d be %d\n", i, i)
	}

	return builder.String()
}

// buildMixedWorkload creates declarations, reassignments,
// expressions, strings, and comments in the same program.
func buildMixedWorkload(count int) string {
	var builder strings.Builder
	builder.Grow(count * 150)

	for i := range count {
		fmt.Fprintf(&builder, "# Processing value %d\n", i)
		fmt.Fprintf(&builder, "let value_%d be (%d + 3) * 4\n", i, i)
		fmt.Fprintf(
			&builder,
			"change value_%d to value_%d ** 2\n",
			i, i,
		)
		fmt.Fprintf(&builder, "say \"processed value %d\"\n", i)
		fmt.Fprintf(&builder, "say value_%d\n", i)
	}

	return builder.String()
}

// BenchmarkFrontendWorkloads compares complete lexing and parsing
// across different source-code profiles and program sizes.
func BenchmarkFrontendWorkloads(b *testing.B) {
	sizes := []int{10, 100, 1000}

	for _, workload := range frontendWorkloads {
		for _, size := range sizes {
			text := workload.build(size)

			b.Run(
				fmt.Sprintf("%s/N%d/Full", workload.name, size),
				func(b *testing.B) {
					program, diagnostic, err := parseFullFrontend(text)
					if err != nil || diagnostic != nil {
						b.Fatalf(
							"invalid benchmark source: lexical=%v syntax=%v",
							err, diagnostic,
						)
					}
					frontendProgramSink = program

					b.ReportAllocs()
					b.SetBytes(int64(len(text)))
					b.ResetTimer()

					for b.Loop() {
						program, diagnostic, err := parseFullFrontend(text)
						if err != nil || diagnostic != nil {
							b.Fatalf(
								"frontend failed: lexical=%v syntax=%v",
								err, diagnostic,
							)
						}
						frontendProgramSink = program
					}
				},
			)

			b.Run(
				fmt.Sprintf("%s/N%d/Compact", workload.name, size),
				func(b *testing.B) {
					program, diagnostic, err := parseCompactFrontend(text)
					if err != nil || diagnostic != nil {
						b.Fatalf(
							"invalid benchmark source: lexical=%v syntax=%v",
							err, diagnostic,
						)
					}
					frontendProgramSink = program

					b.ReportAllocs()
					b.SetBytes(int64(len(text)))
					b.ResetTimer()

					for b.Loop() {
						program, diagnostic, err := parseCompactFrontend(text)
						if err != nil || diagnostic != nil {
							b.Fatalf(
								"frontend failed: lexical=%v syntax=%v",
								err, diagnostic,
							)
						}
						frontendProgramSink = program
					}
				},
			)
		}
	}
}
