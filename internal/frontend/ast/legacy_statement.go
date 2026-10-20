package ast

import "github.com/manatunga/eloque/internal/source"

// Program represents the root of an Eloque abstract syntax tree.
// Statements contains the source file's top-level statements in their
// original order. The program is a Node but is neither an Expression
// nor a Statement, allowing it to serve as the compilation unit.
// Location spans the complete source program.
type Program struct {
	Statements []Statement
	Location   source.Span
}

// LetStatement represents a mutable variable declaration in Eloque.
// Name identifies the binding being declared, while Value contains
// the expression used to initialize it. The declaration introduces
// a new binding, whose name and initializer are subsequently checked
// by the resolver and type checker.
// Location spans the complete declaration, including its initializer.
type LetStatement struct {
	Name     string
	Value    Expression
	Location source.Span
}

// SayStatement represents an output statement in Eloque.
// Value contains the expression whose evaluated result will be
// produced as output during execution. The AST preserves the
// expression without evaluating it or performing any output.
// Location spans the complete statement, including its expression.
type SayStatement struct {
	Value    Expression
	Location source.Span
}

// ChangeStatement represents an assignment to an existing binding.
// Name identifies the binding being updated, while Value contains
// the expression whose result will replace its current value.
// The resolver verifies that the binding exists, and the type checker
// validates that the assignment is permitted and type-compatible.
// Location spans the complete assignment statement.
type ChangeStatement struct {
	Name     string
	Value    Expression
	Location source.Span
}

// Span returns the source range occupied by the complete program.
// The range covers the source file represented by this AST.
func (n *Program) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the variable declaration.
// The range includes the declaration syntax, binding name, and initializer.
func (n *LetStatement) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the output statement.
// The range includes the output keyword and its expression.
func (n *SayStatement) Span() source.Span {
	return n.Location
}

// Span returns the source range occupied by the complete assignment.
// The range includes the assignment keyword, binding name, and
// expression whose value will be assigned.
func (n *ChangeStatement) Span() source.Span {
	return n.Location
}

func (n *LetStatement) statementNode() {}

func (n *SayStatement) statementNode() {}

func (n *ChangeStatement) statementNode() {}
