package parser

import (
	"fmt"
	"strings"
	"testing"

	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
)

// benchmarkProgramResult prevents the benchmark's parsed result
// from being discarded by compiler optimisations.
var benchmarkProgramResult *ast.Program

// makeBenchmarkProgramTokens constructs source text and a token
// stream containing the requested number of statements, cycling
// through let, change, and say statements.
func makeBenchmarkProgramTokens(statementCount int) (string, []lexer.Token) {
	var sourceText strings.Builder
	tokens := make([]lexer.Token, 0, statementCount*7+1)

	// appendToken adds a token and its corresponding source text.
	// Token positions use zero-based byte offsets.
	appendToken := func(kind lexer.TokenKind, lexeme string) {
		start := uint32(sourceText.Len())
		sourceText.WriteString(lexeme)
		end := uint32(sourceText.Len())

		tokens = append(tokens, lexer.Token{
			Kind:  kind,
			Start: start,
			End:   end,
		})
	}

	// appendSpace separates adjacent tokens without producing
	// additional tokens.
	appendSpace := func() {
		sourceText.WriteByte(' ')
	}

	for i := 0; i < statementCount; i++ {
		switch i % 3 {
		case 0:
			appendToken(lexer.TokenLet, "let")
			appendSpace()
			appendToken(lexer.TokenIdent, "value")
			appendSpace()
			appendToken(lexer.TokenBe, "be")
			appendSpace()
			appendToken(lexer.TokenInt, "10")
			appendSpace()
			appendToken(lexer.TokenPlus, "+")
			appendSpace()
			appendToken(lexer.TokenInt, "20")

		case 1:
			appendToken(lexer.TokenChange, "change")
			appendSpace()
			appendToken(lexer.TokenIdent, "value")
			appendSpace()
			appendToken(lexer.TokenTo, "to")
			appendSpace()
			appendToken(lexer.TokenIdent, "value")
			appendSpace()
			appendToken(lexer.TokenStar, "*")
			appendSpace()
			appendToken(lexer.TokenInt, "2")

		case 2:
			appendToken(lexer.TokenSay, "say")
			appendSpace()
			appendToken(lexer.TokenIdent, "value")
		}

		if i < statementCount-1 {
			appendToken(lexer.TokenNewline, "\n")
		}
	}

	eof := uint32(sourceText.Len())
	tokens = append(tokens, lexer.Token{
		Kind:  lexer.TokenEOF,
		Start: eof,
		End:   eof,
	})

	return sourceText.String(), tokens
}

// benchmarkParseProgram measures program parsing independently
// of token construction and lexing.
func benchmarkParseProgram(b *testing.B, statementCount int) {
	sourceText, tokens := makeBenchmarkProgramTokens(statementCount)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p := New(sourceText, tokens)

		program, diagnostic := p.parseProgram()
		if diagnostic != nil {
			b.Fatalf("unexpected diagnostic: %+v", diagnostic)
		}
		if program == nil {
			b.Fatal("expected a non-nil program")
		}
		if len(program.Statements) != statementCount {
			b.Fatalf(
				"statement count = %d, want %d",
				len(program.Statements),
				statementCount,
			)
		}

		benchmarkProgramResult = program
	}
}

// BenchmarkParseProgram measures parsing performance across
// small, medium, and large programs.
func BenchmarkParseProgram(b *testing.B) {
	for _, statementCount := range []int{3, 300, 3000} {
		b.Run(fmt.Sprintf("%dStatements", statementCount), func(b *testing.B) {
			benchmarkParseProgram(b, statementCount)
		})
	}
}
