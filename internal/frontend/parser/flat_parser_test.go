package parser

import "testing"

// TestNewFlatParser verifies that the H2 constructor initializes
// the streaming parser and an empty source-backed AST arena.
func TestNewFlatParser(t *testing.T) {
	const sourceText = "let value be 42\n"

	p := NewFlatParser(sourceText)

	if p == nil {
		t.Fatal("NewFlatParser returned nil")
	}
	if p.Parser == nil {
		t.Fatal("embedded parser is nil")
	}
	if p.arena == nil {
		t.Fatal("AST arena is nil")
	}
	if got := p.arena.Len(); got != 0 {
		t.Fatalf("initial node count = %d, want 0", got)
	}

	program, diagnostic := p.parseFlatProgram()
	if diagnostic != nil {
		t.Fatalf("parseFlatProgram returned diagnostic: %v", diagnostic)
	}
	if program == nil {
		t.Fatal("parseFlatProgram returned nil program")
	}
	if program.Arena != p.arena {
		t.Fatal("parsed program does not reference the parser's arena")
	}
	if got := p.arena.StatementCount(); got != 1 {
		t.Fatalf("statement count = %d, want 1", got)
	}
}
