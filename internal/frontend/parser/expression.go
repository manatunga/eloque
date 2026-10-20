package parser

import (
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseFlatPrefix parses a prefix unary expression or delegates to
// parseFlatPrimary when no unary operator is present.
//
// Unary plus and minus are right-nesting operators. Their operands
// are parsed at the unary binding power, allowing exponentiation to
// bind more tightly than unary operators.
//
// The resulting unary node is stored in the AST arena and references
// its operand through a NodeID. Its source span extends from the unary
// operator to the end of the operand's complete syntactic extent.
//
// Malformed operands produce positioned diagnostics without constructing
// the enclosing unary node.
func (p *FlatParser) parseFlatPrefix() (flatExpressionResult, *Diagnostic) {
	operator := p.peek()

	if operator.Kind != lexer.TokenPlus && operator.Kind != lexer.TokenMinus {
		return p.parseFlatPrimary()
	}

	p.advance()

	operand, diagnostic := p.parseFlatExpression(bindingPowerUnary)
	if diagnostic != nil {
		return flatExpressionResult{}, diagnostic
	}

	span := source.ByteSpan{
		Start: operator.Start,
		End:   operand.extent.End,
	}

	id := p.arena.AddUnary(span, operator.Kind, operand.id)

	return flatExpressionResult{
		id:     id,
		extent: span,
	}, nil
}

// parseFlatExpression parses an Eloque expression using Pratt parsing
// and constructs its nodes in the AST arena.
//
// minBindingPower determines the minimum precedence required for a
// binary operator to be consumed at the current parsing level.
//
// Parsing begins with a prefix or primary expression, then consumes
// binary operators according to the existing M1 binding-power rules.
//
// Addition, subtraction, multiplication, division, floor division,
// and modulo are left-associative. Exponentiation is right-associative
// and binds more tightly than prefix unary operators.
//
// Each binary node references its operands through NodeID values and
// stores its operator in a compact binary payload. The returned extent
// includes any grouping parentheses consumed by child expressions.
//
// Syntax errors return positioned diagnostics. Nodes constructed before
// an error may remain in the arena; the parser does not roll back
// partially constructed expressions.
func (p *FlatParser) parseFlatExpression(
	minBindingPower int,
) (flatExpressionResult, *Diagnostic) {
	left, diagnostic := p.parseFlatPrefix()
	if diagnostic != nil {
		return flatExpressionResult{}, diagnostic
	}

	for {
		bindingPower, isBinaryOperator := binaryBindingPower(p.peek().Kind)
		if !isBinaryOperator || bindingPower < minBindingPower {
			break
		}

		operator := p.advance()

		rightMinBindingPower := bindingPower + 1
		if operator.Kind == lexer.TokenPower {
			rightMinBindingPower = bindingPower
		}

		right, diagnostic := p.parseFlatExpression(rightMinBindingPower)
		if diagnostic != nil {
			return flatExpressionResult{}, diagnostic
		}

		span := source.ByteSpan{
			Start: left.extent.Start,
			End:   right.extent.End,
		}

		id := p.arena.AddBinary(
			span,
			left.id,
			operator.Kind,
			right.id,
		)

		left = flatExpressionResult{
			id:     id,
			extent: span,
		}
	}

	return left, nil
}

// parseFlatPrefix parses a prefix unary expression or delegates to
// parseFlatPrimary when no unary operator is present.
//
// Unary plus and minus are right-nesting operators. Their operands
// are parsed at the unary binding power, allowing exponentiation to
// bind more tightly than unary operators.
//
// The resulting unary node is stored in the AST arena and references
// its operand through a NodeID. Its source span extends from the unary
// operator to the end of the operand's complete syntactic extent.
//
// Malformed operands produce positioned diagnostics without constructing
// the enclosing unary node.
func (p *FlatParser) parseFlatPrimary() (flatExpressionResult, *Diagnostic) {
	token := p.peek()

	span := source.ByteSpan{
		Start: token.Start,
		End:   token.End,
	}

	switch token.Kind {
	case lexer.TokenInt:
		p.advance()

		return flatExpressionResult{
			id:     p.arena.AddInteger(span),
			extent: span,
		}, nil

	case lexer.TokenString:
		p.advance()

		value := decodeStringLiteral(p.lexeme(token))

		return flatExpressionResult{
			id:     p.arena.AddString(span, value),
			extent: span,
		}, nil

	case lexer.TokenIdent:
		p.advance()

		return flatExpressionResult{
			id:     p.arena.AddIdentifier(span),
			extent: span,
		}, nil

	case lexer.TokenLeftParen:
		opening := p.advance()

		inner, diagnostic := p.parseFlatExpression(0)
		if diagnostic != nil {
			return flatExpressionResult{}, diagnostic
		}

		closing, diagnostic := p.expect(
			lexer.TokenRightParen,
			"expected ')' to close parenthesized expression",
		)
		if diagnostic != nil {
			return flatExpressionResult{}, diagnostic
		}

		return flatExpressionResult{
			id: inner.id,
			extent: source.ByteSpan{
				Start: opening.Start,
				End:   closing.End,
			},
		}, nil

	default:
		return flatExpressionResult{}, p.unexpectedExpression(token)
	}
}
