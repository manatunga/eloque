package parser

import (
	"github.com/manatunga/eloque/internal/diagnostic"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// unexpectedExpression constructs a diagnostic for an invalid expression
// token without consuming it.
//
// The diagnostic identifies the offending token and lists the token
// kinds that can begin an expression.
//
// Diagnostic construction is isolated from successful parsing paths
// to avoid unnecessary heap allocations.
func (p *Parser) unexpectedExpression(
	token lexer.Token,
) *diagnostic.Diagnostic {
	return &diagnostic.Diagnostic{
		Code:     CodeUnexpectedToken,
		Severity: diagnostic.SeverityError,
		Message:  "expected an expression",
		Primary: diagnostic.Location{
			File: 0,
			Span: source.ByteSpan{
				Start: token.Start,
				End:   token.End,
			},
		},
		TokenExpectation: &diagnostic.TokenExpectation{
			Expected: []string{
				lexer.TokenInt.String(),
				lexer.TokenString.String(),
				lexer.TokenIdent.String(),
				lexer.TokenLeftParen.String(),
				lexer.TokenPlus.String(),
				lexer.TokenMinus.String(),
			},
			Actual: token.Kind.String(),
		},
	}
}

// unexpectedStatement constructs a diagnostic for an unsupported
// statement-starting token.
//
// The diagnostic records the offending token's source span and the
// supported statement starters: LET, CHANGE, and SAY.
//
// Diagnostic construction is isolated from successful parsing paths
// to avoid unnecessary heap allocations.
func (p *Parser) unexpectedStatement(
	token lexer.Token,
) *diagnostic.Diagnostic {
	return &diagnostic.Diagnostic{
		Code:     CodeUnexpectedToken,
		Severity: diagnostic.SeverityError,
		Message:  "expected a statement ('let', 'change' or 'say')",
		Primary: diagnostic.Location{
			File: 0,
			Span: source.ByteSpan{
				Start: token.Start,
				End:   token.End,
			},
		},
		TokenExpectation: &diagnostic.TokenExpectation{
			Expected: []string{
				lexer.TokenLet.String(),
				lexer.TokenChange.String(),
				lexer.TokenSay.String(),
			},
			Actual: token.Kind.String(),
		},
	}
}
