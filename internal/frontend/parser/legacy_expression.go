package parser

import (
	"strings"

	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// expressionResult represents the result of parsing an Eloque expression.
//
// expression contains the semantic AST node, while extent records the
// complete syntactic source range consumed during parsing.
//
// These spans may differ for parenthesized expressions because grouping
// parentheses contribute to the syntactic extent without becoming
// separate AST nodes.
//
// Keeping the two values separate allows enclosing expressions to
// preserve accurate source locations without modifying child nodes.
type expressionResult struct {
	expression ast.Expression
	extent     source.Span
}

// Expression binding powers define the precedence of Eloque operators.
// Higher values indicate tighter binding. These constants follow the
// precedence rules established in the language specification and are
// used by the Pratt parser to determine expression grouping.
const (
	bindingPowerAdditive       = 50
	bindingPowerMultiplicative = 60
	bindingPowerUnary          = 70
	bindingPowerExponentiation = 80
)

// binaryBindingPower returns the binding power of a supported
// binary arithmetic operator.
//
// The boolean result indicates whether the token kind represents
// an operator supported by the current expression parser.
//
// Operators with higher binding powers bind more tightly.
// Addition, subtraction, multiplication, division, floor division,
// and modulo are left-associative. Exponentiation is right-associative.
func binaryBindingPower(kind lexer.TokenKind) (int, bool) {
	switch kind {
	case lexer.TokenPlus, lexer.TokenMinus:
		return bindingPowerAdditive, true
	case lexer.TokenStar, lexer.TokenSlash, lexer.TokenFloorDiv, lexer.TokenPercent:
		return bindingPowerMultiplicative, true
	case lexer.TokenPower:
		return bindingPowerExponentiation, true
	default:
		return 0, false
	}
}

// parsePrefix parses a prefix unary expression or a primary expression.
//
// Unary plus and minus are right-nesting prefix operators. Their operands
// are parsed using the unary binding power, allowing higher-precedence
// operators such as exponentiation to bind within the operand.
//
// When the current token is not a supported unary operator, parsing
// delegates to parsePrimary.
//
// The returned result contains both the semantic AST and its complete
// syntactic source extent. Syntax errors produce positioned diagnostics.
func (p *Parser) parsePrefix() (expressionResult, *Diagnostic) {
	operator := p.peek()

	if operator.Kind != lexer.TokenPlus && operator.Kind != lexer.TokenMinus {
		return p.parsePrimary()
	}

	p.advance()

	operand, diagnostic := p.parseExpression(bindingPowerUnary)
	if diagnostic != nil {
		return expressionResult{}, diagnostic
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

	return expressionResult{
		expression: expression,
		extent:     span,
	}, nil
}

// decodeStringLiteral converts a validated Eloque string token's
// lexeme into its decoded string value.
//
// The input includes the original surrounding quotation marks.
// Escape sequences are converted to their corresponding characters.
// The lexer is responsible for rejecting malformed string literals
// before this function is called.
func decodeStringLiteral(lexeme string) string {
	innerString := lexeme[1 : len(lexeme)-1]
	var result strings.Builder
	result.Grow(len(innerString))

	for i := 0; i < len(innerString); i++ {
		char := innerString[i]
		if char == '\\' && i+1 < len(innerString) {
			nextChar := innerString[i+1]
			switch nextChar {
			case 'n':
				result.WriteByte('\n')
			case 't':
				result.WriteByte('\t')
			case 'r':
				result.WriteByte('\r')
			case '\\':
				result.WriteByte('\\')
			case '\'':
				result.WriteByte('\'')
			case '"':
				result.WriteByte('"')
			default:
				result.WriteByte('\\')
				result.WriteByte(nextChar)
			}
			i++
		} else {
			result.WriteByte(char)
		}
	}
	return result.String()
}

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
) *Diagnostic {
	actual := token.Kind

	return &Diagnostic{
		Message:  "expected an expression",
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

// parsePrimary parses an integer literal, string literal, identifier,
// or parenthesized Eloque expression.
//
// Literal and identifier values are obtained directly from the original
// source text using the token's byte offsets. String literals are decoded
// after lexical validation.
//
// Parenthesized expressions reuse the inner semantic AST node rather
// than introducing a separate grouping node. Their syntactic extent
// includes both parentheses, while the inner AST node retains its
// original source location.
//
// Unexpected tokens produce a positioned diagnostic without consuming
// the offending token.
func (p *Parser) parsePrimary() (expressionResult, *Diagnostic) {
	currentToken := p.peek()

	switch currentToken.Kind {
	case lexer.TokenInt:
		p.advance()

		span := p.tokenSpan(currentToken)

		expression := &ast.IntegerLiteral{
			Value:    p.lexeme(currentToken),
			Location: span,
		}

		return expressionResult{
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

		return expressionResult{
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

		return expressionResult{
			expression: expression,
			extent:     span,
		}, nil

	case lexer.TokenLeftParen:
		openingToken := p.advance()

		innerResult, diagnostic := p.parseExpression(0)
		if diagnostic != nil {
			return expressionResult{}, diagnostic
		}

		closingToken, diagnostic := p.expect(
			lexer.TokenRightParen,
			"expected ')' to close parenthesized expression",
		)
		if diagnostic != nil {
			return expressionResult{}, diagnostic
		}

		return expressionResult{
			expression: innerResult.expression,
			extent: source.Span{
				Start: p.tokenSpan(openingToken).Start,
				End:   p.tokenSpan(closingToken).End,
			},
		}, nil

	default:
		return expressionResult{}, p.unexpectedExpression(currentToken)
	}
}

// parseExpression parses an Eloque expression using Pratt parsing.
//
// minBindingPower specifies the minimum binding power required for a
// binary operator to be consumed at the current parsing level.
//
// Parsing begins with a prefix unary or primary expression, followed
// by any binary operators whose binding powers satisfy the minimum.
//
// Addition, subtraction, multiplication, division, floor division,
// and modulo are left-associative. Exponentiation is right-associative.
//
// Exponentiation binds more tightly than prefix unary operators,
// allowing expressions such as -2 ** 2 and 2 ** -3 to follow Eloque's
// established precedence rules.
//
// The returned result preserves the semantic AST separately from its
// complete syntactic source extent. Syntax errors produce positioned
// diagnostics.
func (p *Parser) parseExpression(
	minBindingPower int,
) (expressionResult, *Diagnostic) {
	left, diagnostic := p.parsePrefix()
	if diagnostic != nil {
		return expressionResult{}, diagnostic
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

		right, diagnostic := p.parseExpression(rightMinBindingPower)
		if diagnostic != nil {
			return expressionResult{}, diagnostic
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

		left = expressionResult{
			expression: expression,
			extent:     span,
		}
	}

	return left, nil
}
