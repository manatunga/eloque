package parser

import (
	"github.com/manatunga/eloque/internal/frontend/ast"
	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

// parseFlatLetStatement parses a mutable variable declaration into
// Eloque's arena-backed AST.
//
// The current token must be LET. The parser consumes the declaration
// keyword, an identifier, the BE keyword, and an initializer expression.
//
// The variable name is represented by its original source byte span,
// avoiding a separate string allocation for the identifier. The
// initializer is referenced through its arena-assigned NodeID.
//
// The resulting statement spans the complete declaration, including
// any parentheses surrounding the initializer.
//
// NEWLINE and EOF are left unconsumed for the program parser.
// Malformed declarations return the same diagnostics as M1.
func (p *FlatParser) parseFlatLetStatement() (ast.NodeID, *Diagnostic) {
	keyword := p.advance()

	identifier, diagnostic := p.expect(
		lexer.TokenIdent,
		"expected a variable name after 'let'",
	)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	_, diagnostic = p.expect(
		lexer.TokenBe,
		"expected 'be' after variable name to initialize a variable value",
	)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	value, diagnostic := p.parseFlatExpression(0)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	span := source.ByteSpan{
		Start: keyword.Start,
		End:   value.extent.End,
	}

	nameSpan := source.ByteSpan{
		Start: identifier.Start,
		End:   identifier.End,
	}

	return p.arena.AddLet(span, nameSpan, value.id), nil
}

// parseFlatChangeStatement parses reassignment to an existing variable
// into Eloque's arena-backed AST.
//
// The current token must be CHANGE. The parser consumes the keyword,
// an identifier, the TO keyword, and an assignment expression.
//
// The target name is stored as a source byte span, while the assigned
// expression is referenced by NodeID. The statement's span extends
// from CHANGE through the complete syntactic extent of the expression.
//
// Statement terminators remain unconsumed. Invalid assignments return
// positioned diagnostics consistent with the M1 parser.
func (p *FlatParser) parseFlatChangeStatement() (ast.NodeID, *Diagnostic) {
	keyword := p.advance()

	identifier, diagnostic := p.expect(
		lexer.TokenIdent,
		"expected a variable name after 'change'",
	)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	_, diagnostic = p.expect(
		lexer.TokenTo,
		"expected 'to' after variable name to assign a new variable value",
	)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	value, diagnostic := p.parseFlatExpression(0)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	span := source.ByteSpan{
		Start: keyword.Start,
		End:   value.extent.End,
	}

	nameSpan := source.ByteSpan{
		Start: identifier.Start,
		End:   identifier.End,
	}

	return p.arena.AddChange(span, nameSpan, value.id), nil
}

// parseFlatSayStatement parses an Eloque output statement into the
// arena-backed AST.
//
// The current token must be SAY. The following expression is parsed
// using the H2 Pratt parser and referenced by NodeID.
//
// The resulting statement spans from the SAY keyword through the
// expression's complete syntactic extent, including grouping parentheses.
//
// NEWLINE and EOF remain unconsumed. Missing or malformed expressions
// produce the same positioned diagnostics as M1.
func (p *FlatParser) parseFlatSayStatement() (ast.NodeID, *Diagnostic) {
	keyword := p.advance()

	value, diagnostic := p.parseFlatExpression(0)
	if diagnostic != nil {
		return ast.InvalidNodeID, diagnostic
	}

	span := source.ByteSpan{
		Start: keyword.Start,
		End:   value.extent.End,
	}

	return p.arena.AddSay(span, value.id), nil
}

// parseFlatStatement dispatches to the appropriate H2 statement parser.
//
// The current token determines the statement type. Supported forms
// are LET declarations, CHANGE assignments, and SAY output statements.
//
// Each successful parse returns the NodeID of a statement stored in
// the shared AST arena. Statement terminators remain unconsumed.
//
// Unsupported statement starters use the existing M1 diagnostic helper,
// preserving its expected-token list and source location behavior.
func (p *FlatParser) parseFlatStatement() (ast.NodeID, *Diagnostic) {
	switch p.peek().Kind {
	case lexer.TokenLet:
		return p.parseFlatLetStatement()

	case lexer.TokenChange:
		return p.parseFlatChangeStatement()

	case lexer.TokenSay:
		return p.parseFlatSayStatement()

	default:
		return ast.InvalidNodeID, p.unexpectedStatement(p.peek())
	}
}
