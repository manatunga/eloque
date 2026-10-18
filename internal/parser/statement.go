package parser

import (
	"github.com/manatunga/eloque/internal/ast"
	"github.com/manatunga/eloque/internal/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseLetStatement parses a mutable variable declaration in Eloque.
//
// The current token must be TokenLet. A declaration consists of the
// "let" keyword, a variable identifier, the "be" keyword, and an
// initializer expression.
//
// The resulting LetStatement spans the complete declaration, from
// the opening keyword through the initializer's syntactic extent.
//
// Statement terminators (NEWLINE or EOF) are left unconsumed for the
// enclosing program parser to handle.
//
// Malformed declarations produce positioned syntax diagnostics.
func (p *Parser) parseLetStatement() (ast.Statement, *Diagnostic) {
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

	initializer, diagnostic := p.parseExpression(0)
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

// parseChangeStatement parses reassignment to an existing variable.
//
// The current token must be TokenChange. An assignment consists of the
// "change" keyword, a variable identifier, the "to" keyword, and an
// expression representing the new value.
//
// The resulting ChangeStatement spans the complete assignment, from
// the opening keyword through the expression's syntactic extent.
//
// Statement terminators (NEWLINE or EOF) are left unconsumed for the
// enclosing program parser to handle.
//
// Malformed assignments produce positioned syntax diagnostics.
func (p *Parser) parseChangeStatement() (ast.Statement, *Diagnostic) {
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

	assigned, diagnostic := p.parseExpression(0)
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

// parseSayStatement parses an Eloque output statement.
//
// The current token must be TokenSay. An output statement consists
// of the "say" keyword followed by an expression.
//
// The resulting SayStatement spans the complete statement, from the
// opening keyword through the expression's syntactic extent.
//
// Statement terminators (NEWLINE or EOF) are left unconsumed for the
// enclosing program parser to handle.
//
// Missing or malformed expressions produce positioned syntax diagnostics.
func (p *Parser) parseSayStatement() (ast.Statement, *Diagnostic) {
	sayToken := p.advance()

	value, diagnostic := p.parseExpression(0)
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

// parseStatement parses a single supported Eloque statement.
//
// The current token determines which statement parser is invoked.
// Currently supported statements are variable declarations,
// variable reassignments, and output statements.
//
// Statement terminators are not consumed. The enclosing program
// parser is responsible for validating statement boundaries.
//
// Unsupported statement starters produce positioned syntax diagnostics.
func (p *Parser) parseStatement() (ast.Statement, *Diagnostic) {
	currentToken := p.peek()

	switch currentToken.Kind {
	case lexer.TokenLet:
		return p.parseLetStatement()

	case lexer.TokenChange:
		return p.parseChangeStatement()

	case lexer.TokenSay:
		return p.parseSayStatement()

	default:
		return nil, p.unexpectedStatement(currentToken)
	}
}
