package parser

import (
	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseCompactLetStatement parses a mutable variable declaration
// while preserving the production parser's AST and diagnostics.
func (p *CompactParser) parseCompactLetStatement() (ast.Statement, *Diagnostic) {
	letToken := p.advance()

	identifier, diagnostic := p.expect(
		lexer.TokenIdent,
		"expected a variable name after 'let'",
	)
	if diagnostic != nil {
		return nil, diagnostic
	}

	_, diagnostic = p.expect(
		lexer.TokenBe,
		"expected 'be' after variable name to initialize a variable value",
	)
	if diagnostic != nil {
		return nil, diagnostic
	}

	initializer, diagnostic := p.parseCompactExpression(0)
	if diagnostic != nil {
		return nil, diagnostic
	}

	statement := &ast.LetStatement{
		Name:  p.lexeme(identifier),
		Value: initializer.expression,
		Location: source.Span{
			Start: p.tokenSpan(letToken).Start,
			End:   initializer.extent.End,
		},
	}

	return statement, nil
}

// parseCompactChangeStatement parses reassignment to an existing
// binding while retaining the original grammar and diagnostic messages.
func (p *CompactParser) parseCompactChangeStatement() (ast.Statement, *Diagnostic) {
	changeToken := p.advance()

	identifier, diagnostic := p.expect(
		lexer.TokenIdent,
		"expected a variable name after 'change'",
	)
	if diagnostic != nil {
		return nil, diagnostic
	}

	_, diagnostic = p.expect(
		lexer.TokenTo,
		"expected 'to' after variable name to assign a new variable value",
	)
	if diagnostic != nil {
		return nil, diagnostic
	}

	assigned, diagnostic := p.parseCompactExpression(0)
	if diagnostic != nil {
		return nil, diagnostic
	}

	statement := &ast.ChangeStatement{
		Name:  p.lexeme(identifier),
		Value: assigned.expression,
		Location: source.Span{
			Start: p.tokenSpan(changeToken).Start,
			End:   assigned.extent.End,
		},
	}

	return statement, nil
}

// parseCompactSayStatement parses an output statement containing
// an expression and preserves its complete source location.
func (p *CompactParser) parseCompactSayStatement() (ast.Statement, *Diagnostic) {
	sayToken := p.advance()

	value, diagnostic := p.parseCompactExpression(0)
	if diagnostic != nil {
		return nil, diagnostic
	}

	statement := &ast.SayStatement{
		Value: value.expression,
		Location: source.Span{
			Start: p.tokenSpan(sayToken).Start,
			End:   value.extent.End,
		},
	}

	return statement, nil
}

// unexpectedStatement constructs a diagnostic for an invalid statement token.
func (p *CompactParser) unexpectedStatement(
    token lexer.CompactToken,
) *Diagnostic {
    actual := token.Kind

    return &Diagnostic{
        Message:  "expected a statement ('let', 'change' or 'say')",
        Location: p.tokenSpan(token),
        Expected: []lexer.TokenKind{
			lexer.TokenLet,
			lexer.TokenChange,
			lexer.TokenSay,
        },
        Actual: &actual,
    }
}

// parseCompactStatement dispatches a compact-token statement to
// its corresponding parser or reports a positioned syntax error.
func (p *CompactParser) parseCompactStatement() (ast.Statement, *Diagnostic) {
	currentToken := p.peek()

	switch currentToken.Kind {
	case lexer.TokenLet:
		return p.parseCompactLetStatement()

	case lexer.TokenChange:
		return p.parseCompactChangeStatement()

	case lexer.TokenSay:
		return p.parseCompactSayStatement()

	default:
		return nil, p.unexpectedStatement(currentToken)
	}
}
