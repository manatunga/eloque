package gobench

import (
	"fmt"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

type workload struct {
	name  string
	build func(int) string
}

var workloads = []workload{
	{"Declarations", buildDeclarations},
	{"Expressions", buildExpressions},
	{"Strings", buildStrings},
	{"Comments", buildComments},
	{"Mixed", buildMixed},
}

var parsedFileSink any

func wrapProgram(body string) string {
	return "package benchmark\n\nfunc workload() {\n" +
		body + "}\n"
}

func buildDeclarations(count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(&b, "value_%d := %d\n", i, i)
	}
	return wrapProgram(b.String())
}

func buildExpressions(count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(
			&b,
			"value_%d := ((%d + 3) * (7 - 2)) * ((%d + 3) * (7 - 2))\n",
			i, i, i,
		)
	}
	return wrapProgram(b.String())
}

func buildStrings(count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(
			&b,
			"fmt.Println(\"message %d: the quick brown fox jumps over the lazy dog\")\n",
			i,
		)
	}
	return wrapProgram(b.String())
}

func buildComments(count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(
			&b,
			"// This comment describes declaration number %d and its purpose.\n",
			i,
		)
		fmt.Fprintf(&b, "value_%d := %d\n", i, i)
	}
	return wrapProgram(b.String())
}

func buildMixed(count int) string {
	var b strings.Builder
	for i := range count {
		fmt.Fprintf(&b, "// Processing value %d\n", i)
		fmt.Fprintf(&b, "value_%d := (%d + 3) * 4\n", i, i)
		fmt.Fprintf(&b, "value_%d = value_%d * value_%d\n", i, i, i)
		fmt.Fprintf(&b, "fmt.Println(\"processed value %d\")\n", i)
		fmt.Fprintf(&b, "fmt.Println(value_%d)\n", i)
	}
	return wrapProgram(b.String())
}

func BenchmarkGoFrontend(b *testing.B) {
	for _, w := range workloads {
		for _, size := range []int{10, 100, 1000, 10000} {
			source := w.build(size)

			b.Run(fmt.Sprintf("%s/N%d", w.name, size), func(b *testing.B) {
				// Validate the source outside the timed loop.
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
}
