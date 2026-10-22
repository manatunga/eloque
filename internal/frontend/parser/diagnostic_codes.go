package parser

import "github.com/manatunga/eloque/internal/diagnostic"

const (
	// CodeUnexpectedToken identifies a token that violates the grammar.
	CodeUnexpectedToken diagnostic.Code = "PAR_UNEXPECTED_TOKEN"

	// CodeExpressionTooDeep identifies an expression that exceeds
	// the parser's maximum supported recursion depth.
	CodeExpressionTooDeep diagnostic.Code = "PAR_EXPRESSION_TOO_DEEP"
)
