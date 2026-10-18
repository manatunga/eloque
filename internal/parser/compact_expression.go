package parser

import (
	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// compactExpressionResult contains a parsed expression and its complete
// syntactic extent, including grouping parentheses excluded from the AST.
type compactExpressionResult struct {
	expression ast.Expression
	extent     source.Span
}

// parseCompactPrefix parses a unary expression or primary expression
// while preserving the original Pratt parser's precedence rules.
func (p *CompactParser) parseCompactPrefix() (compactExpressionResult, *Diagnostic) {
	operator := p.peek()

	if operator.Kind != lexer.TokenPlus && operator.Kind != lexer.TokenMinus {
		return p.parseCompactPrimary()
	}

	p.advance()

	operand, diagnostic := p.parseCompactExpression(bindingPowerUnary)
	if diagnostic != nil {
		return compactExpressionResult{}, diagnostic
	}

	span := source.Span{
		Start: p.tokenSpan(operator).Start,
		End:   operand.extent.End,
	}

	expression := &ast.UnaryExpression{
		Operator: operator.Kind,
		Operand:  operand.expression,
		Location: span,
	}

	return compactExpressionResult{
		expression: expression,
		extent:     span,
	}, nil
}

// unexpectedExpression constructs a diagnostic for an invalid expression token.
func (p *CompactParser) unexpectedExpression(
    token lexer.CompactToken,
) *Diagnostic {
    actual := token.Kind

    return &Diagnostic{
        Message: "expected an expression",
        Location: p.tokenSpan(token),
        Expected: []lexer.TokenKind{
            lexer.TokenInt,
            lexer.TokenString,
            lexer.TokenIdent,
            lexer.TokenLeftParen,
			lexer.TokenPlus,
			lexer.TokenMinus,
        },
        Actual: &actual,
    }
}

// parseCompactPrimary parses integer literals, string literals,
// identifiers and parenthesized expressions from compact tokens.
func (p *CompactParser) parseCompactPrimary() (compactExpressionResult, *Diagnostic) {
	currentToken := p.peek()

	switch currentToken.Kind {
	case lexer.TokenInt:
		p.advance()

		span := p.tokenSpan(currentToken)

		expression := &ast.IntegerLiteral{
			Value:    p.lexeme(currentToken),
			Location: span,
		}

		return compactExpressionResult{
			expression: expression,
			extent:     span,
		}, nil

	case lexer.TokenString:
		p.advance()

		span := p.tokenSpan(currentToken)

		expression := &ast.StringLiteral{
			Value:    decodeStringLiteral(p.lexeme(currentToken)),
			Location: span,
		}

		return compactExpressionResult{
			expression: expression,
			extent:     span,
		}, nil

	case lexer.TokenIdent:
		p.advance()

		span := p.tokenSpan(currentToken)

		expression := &ast.Identifier{
			Name:     p.lexeme(currentToken),
			Location: span,
		}

		return compactExpressionResult{
			expression: expression,
			extent:     span,
		}, nil

	case lexer.TokenLeftParen:
		openingToken := p.advance()

		innerResult, diagnostic := p.parseCompactExpression(0)
		if diagnostic != nil {
			return compactExpressionResult{}, diagnostic
		}

		closingToken, diagnostic := p.expect(
			lexer.TokenRightParen,
			"expected ')' to close parenthesized expression",
		)
		if diagnostic != nil {
			return compactExpressionResult{}, diagnostic
		}

		return compactExpressionResult{
			expression: innerResult.expression,
			extent: source.Span{
				Start: p.tokenSpan(openingToken).Start,
				End:   p.tokenSpan(closingToken).End,
			},
		}, nil

	default:
		return compactExpressionResult{}, p.unexpectedExpression(currentToken)
	}
}

// parseCompactExpression parses an expression using Pratt parsing.
// Binary operators retain the production parser's precedence and
// associativity, including right-associative exponentiation.
func (p *CompactParser) parseCompactExpression(
	minBindingPower int,
) (compactExpressionResult, *Diagnostic) {
	left, diagnostic := p.parseCompactPrefix()
	if diagnostic != nil {
		return compactExpressionResult{}, diagnostic
	}

	for {
		bindingPower, isBinaryOperator := binaryBindingPower(p.peek().Kind)
		if !isBinaryOperator || bindingPower < minBindingPower {
			break
		}

		operatorToken := p.advance()

		rightMinBindingPower := bindingPower + 1
		if operatorToken.Kind == lexer.TokenPower {
			rightMinBindingPower = bindingPower
		}

		right, diagnostic := p.parseCompactExpression(rightMinBindingPower)
		if diagnostic != nil {
			return compactExpressionResult{}, diagnostic
		}

		span := source.Span{
			Start: left.extent.Start,
			End:   right.extent.End,
		}

		expression := &ast.BinaryExpression{
			Left:     left.expression,
			Operator: operatorToken.Kind,
			Right:    right.expression,
			Location: span,
		}

		left = compactExpressionResult{
			expression: expression,
			extent:     span,
		}
	}

	return left, nil
}
