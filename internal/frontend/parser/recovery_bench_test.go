package parser

import (
	"strconv"
	"strings"
	"testing"
)

func benchmarkSource(statements int) string {
	var b strings.Builder
	b.Grow(statements * 24)

	for i := range statements {
		b.WriteString("say ")
		b.WriteString(strconv.Itoa(i))
		b.WriteString(" + 42\n")
	}

	return b.String()
}

func benchmarkMalformedSource(statements int) string {
	var b strings.Builder
	b.Grow(statements * 16)

	for i := range statements {
		if i%5 == 0 {
			b.WriteString("say +\n")
		} else {
			b.WriteString("say 42\n")
		}
	}

	return b.String()
}

func BenchmarkParseFailFast(b *testing.B) {
	for _, size := range []int{10, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			src := benchmarkSource(size)
			b.ReportAllocs()
			b.SetBytes(int64(len(src)))
			b.ResetTimer()

			for b.Loop() {
				program, diag := Parse(src)
				if diag != nil || program == nil {
					b.Fatal("unexpected parse failure")
				}
			}
		})
	}
}

func BenchmarkParseRecoveringValid(b *testing.B) {
	for _, size := range []int{10, 1000} {
		b.Run(strconv.Itoa(size), func(b *testing.B) {
			src := benchmarkSource(size)
			b.ReportAllocs()
			b.SetBytes(int64(len(src)))
			b.ResetTimer()

			for b.Loop() {
				result := ParseRecovering(src, ParseOptions{})
				if result.Status != ParseComplete {
					b.Fatal("unexpected recovery status")
				}
			}
		})
	}
}

func BenchmarkParseRecoveringMalformed(b *testing.B) {
	src := benchmarkMalformedSource(1000)
	b.ReportAllocs()
	b.SetBytes(int64(len(src)))
	b.ResetTimer()

	for b.Loop() {
		result := ParseRecovering(
			src,
			ParseOptions{MaxErrors: -1},
		)
		if result.Status != ParseRecovered {
			b.Fatal("unexpected recovery status")
		}
	}
}
