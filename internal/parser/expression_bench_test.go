package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

// benchmarkExpression measures parsing an already-tokenized expression.
func benchmarkExpression(b *testing.B, sourceText string) {
	b.Helper()

	l := lexer.New(sourceText)
	var tokens []lexer.Token

	for {
		token, err := l.NextToken()
		if err != nil {
			b.Fatalf("lexing failed: %v", err)
		}

		tokens = append(tokens, token)
		if token.Kind == lexer.TokenEOF {
			break
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		p := New(sourceText, tokens)

		_, diagnostic := p.parseExpression(0)
		if diagnostic != nil {
			b.Fatalf("unexpected diagnostic: %+v", diagnostic)
		}

		if p.peek().Kind != lexer.TokenEOF {
			b.Fatalf("unexpected remaining token: %v", p.peek().Kind)
		}
	}
}

// benchmarkChainSource constructs repeated binary expressions.
func benchmarkChainSource(
	operators int,
	operator string,
) string {
	var builder strings.Builder
	builder.WriteString("2")

	for i := 0; i < operators; i++ {
		builder.WriteString(" ")
		builder.WriteString(operator)
		builder.WriteString(" 2")
	}

	return builder.String()
}

// benchmarkGroupingSource constructs nested parentheses.
func benchmarkGroupingSource(depth int) string {
	return strings.Repeat("(", depth) +
		"42" +
		strings.Repeat(")", depth)
}

func BenchmarkParseSimpleArithmetic(b *testing.B) {
	benchmarkExpression(b, benchmarkChainSource(2, "+"))
}

func BenchmarkParseMixedPrecedence(b *testing.B) {
	benchmarkExpression(b, "2 + 3 * 4 ** 5")
}

func BenchmarkParseAdditiveChain(b *testing.B) {
	for _, operators := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Operators%d", operators), func(b *testing.B) {
			benchmarkExpression(b, benchmarkChainSource(operators, "+"))
		})
	}
}

func BenchmarkParseExponentiationChain(b *testing.B) {
	for _, operators := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Operators%d", operators), func(b *testing.B) {
			benchmarkExpression(b, benchmarkChainSource(operators, "**"))
		})
	}
}

func BenchmarkParseNestedParentheses(b *testing.B) {
	for _, depth := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Depth%d", depth), func(b *testing.B) {
			benchmarkExpression(b, benchmarkGroupingSource(depth))
		})
	}
}
