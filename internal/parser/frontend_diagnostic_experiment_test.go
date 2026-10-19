package parser

import (
	"errors"
	"reflect"
	"testing"

	"github.com/manatunga/eloque/internal/lexer"
)

type frontendDiagnosticCase struct {
	name          string
	source        string
	wantErrorKind string
	wantLine      int
	wantColumn    int
}

func classifyFrontendError(err error) string {
	if err == nil {
		return "none"
	}

	var lexErr *lexer.LexError
	if errors.As(err, &lexErr) {
		return "lexical"
	}

	return "syntax"
}

func TestFrontendDiagnosticExperiment(t *testing.T) {
	cases := []frontendDiagnosticCase{
		{
			name:          "unexpected character",
			source:        "say @\n",
			wantErrorKind: "lexical",
			wantLine:      1,
			wantColumn:    5,
		},
		{
			name:          "invalid integer",
			source:        "say 12abc\n",
			wantErrorKind: "lexical",
			wantLine:      1,
			wantColumn:    5,
		},
		{
			name:          "leading zero",
			source:        "say 012\n",
			wantErrorKind: "lexical",
			wantLine:      1,
			wantColumn:    5,
		},
		{
			name:          "unterminated string",
			source:        "say \"hello\n",
			wantErrorKind: "lexical",
			wantLine:      1,
			wantColumn:    5,
		},
		{
			name:          "unmatched closing parenthesis",
			source:        "say 1)\n",
			wantErrorKind: "lexical",
			wantLine:      1,
			wantColumn:    6,
		},
		{
			name:          "missing expression",
			source:        "say\n",
			wantErrorKind: "syntax",
		},
		{
			name:          "missing declaration initializer",
			source:        "let value be\n",
			wantErrorKind: "syntax",
		},
		{
			name:          "syntax before lexical error",
			source:        "say\nsay @\n",
			wantErrorKind: "syntax",
		},
		{
			name:          "lexical before syntax error",
			source:        "say @\nsay\n",
			wantErrorKind: "lexical",
			wantLine:      1,
			wantColumn:    5,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Parse(tc.source)
			if err == nil {
				t.Fatal("expected an error")
			}

			kind := classifyFrontendError(err)
			if kind != tc.wantErrorKind {
				t.Fatalf(
					"error kind = %s, want %s: %v",
					kind, tc.wantErrorKind, err,
				)
			}

			if kind == "lexical" {
				var lexErr *lexer.LexError
				if !errors.As(err, &lexErr) {
					t.Fatal("expected *lexer.LexError")
				}

				if lexErr.Position.Line != tc.wantLine ||
					lexErr.Position.Column != tc.wantColumn {
					t.Errorf(
						"position = %d:%d, want %d:%d",
						lexErr.Position.Line,
						lexErr.Position.Column,
						tc.wantLine,
						tc.wantColumn,
					)
				}
			}
		})
	}
}

func TestFrontendDiagnosticParity(t *testing.T) {
	sources := []string{
		"say @\n",
		"say 12abc\n",
		"say 012\n",
		"say \"hello\n",
		"say 1)\n",
		"say\n",
		"let value be\n",
		"say @\nsay\n",
	}

	for _, source := range sources {
		t.Run(source, func(t *testing.T) {
			_, bufferedErr := parseFrontend(source)
			_, streamingErr := Parse(source)

			if classifyFrontendError(bufferedErr) !=
				classifyFrontendError(streamingErr) {
				t.Fatalf(
					"error category mismatch: buffered=%v streaming=%v",
					bufferedErr, streamingErr,
				)
			}

			var bufferedLex, streamingLex *lexer.LexError

			if errors.As(bufferedErr, &bufferedLex) &&
				errors.As(streamingErr, &streamingLex) {
				if !reflect.DeepEqual(bufferedLex, streamingLex) {
					t.Errorf(
						"lexical diagnostics differ: buffered=%+v streaming=%+v",
						bufferedLex, streamingLex,
					)
				}
			}
		})
	}
}
