package ast

import (
	"testing"

	"github.com/manatunga/eloque/internal/frontend/lexer"
	"github.com/manatunga/eloque/internal/source"
)

func TestArenaZeroValue(t *testing.T) {
	var arena Arena

	if got := arena.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestArenaAddAndGet(t *testing.T) {
	var arena Arena

	nodes := []FlatNode{
		{
			Kind: NodeIntegerLiteral,
			Span: source.ByteSpan{Start: 0, End: 2},
			Data: 42,
		},
		{
			Kind: NodeIdentifier,
			Span: source.ByteSpan{Start: 3, End: 6},
			Data: 7,
		},
		{
			Kind: NodeSayStatement,
			Span: source.ByteSpan{Start: 7, End: 13},
			Data: 0,
		},
	}

	for i, node := range nodes {
		id := arena.Add(node)

		if id != NodeID(i) {
			t.Fatalf("Add() returned ID %d, want %d", id, i)
		}

		if id == InvalidNodeID {
			t.Fatal("Add() returned reserved InvalidNodeID")
		}
	}

	if got := arena.Len(); got != len(nodes) {
		t.Fatalf("Len() = %d, want %d", got, len(nodes))
	}

	for i, want := range nodes {
		got := arena.Get(NodeID(i))

		if *got != want {
			t.Errorf("Get(%d) = %+v, want %+v", i, *got, want)
		}
	}
}

func TestArenaGrowthPreservesNodes(t *testing.T) {
	var arena Arena

	const nodeCount = 10_000

	for i := 0; i < nodeCount; i++ {
		id := arena.Add(FlatNode{
			Kind: NodeIntegerLiteral,
			Data: uint32(i),
		})

		if id != NodeID(i) {
			t.Fatalf("Add() returned ID %d, want %d", id, i)
		}
	}

	if got := arena.Len(); got != nodeCount {
		t.Fatalf("Len() = %d, want %d", got, nodeCount)
	}

	for i := 0; i < nodeCount; i++ {
		got := arena.Get(NodeID(i))

		if got.Kind != NodeIntegerLiteral || got.Data != uint32(i) {
			t.Fatalf("Get(%d) = %+v, unexpected node", i, *got)
		}
	}
}

func TestArenaGetAllowsMutation(t *testing.T) {
	var arena Arena

	id := arena.Add(FlatNode{
		Kind: NodeIntegerLiteral,
		Data: 10,
	})

	arena.Get(id).Data = 20

	if got := arena.Get(id).Data; got != 20 {
		t.Errorf("updated Data = %d, want 20", got)
	}
}

func TestArenaAddBinary(t *testing.T) {
	arena := NewArena("10 + 20")

	left := arena.Add(FlatNode{
		Kind: NodeIntegerLiteral,
		Data: 10,
	})
	right := arena.Add(FlatNode{
		Kind: NodeIntegerLiteral,
		Data: 20,
	})

	span := source.ByteSpan{Start: 0, End: 7}
	id := arena.AddBinary(span, left, lexer.TokenPlus, right)

	if id != 2 {
		t.Fatalf("binary node ID = %d, want 2", id)
	}

	node := arena.Get(id)

	if node.Kind != NodeBinaryExpression {
		t.Errorf("node kind = %d, want %d", node.Kind, NodeBinaryExpression)
	}
	if node.Span != span {
		t.Errorf("node span = %+v, want %+v", node.Span, span)
	}

	payload := arena.BinaryPayload(node.Data)

	if payload.Left != left ||
		payload.Right != right ||
		payload.Operator != lexer.TokenPlus {
		t.Errorf("binary payload = %+v, unexpected contents", payload)
	}
}

func TestArenaAddUnary(t *testing.T) {
	arena := NewArena("-10")

	operand := arena.Add(FlatNode{
		Kind: NodeIntegerLiteral,
		Data: 10,
	})

	span := source.ByteSpan{Start: 0, End: 3}
	id := arena.AddUnary(span, lexer.TokenMinus, operand)

	if id != 1 {
		t.Fatalf("unary node ID = %d, want 1", id)
	}

	node := arena.Get(id)

	if node.Kind != NodeUnaryExpression {
		t.Errorf("node kind = %d, want %d", node.Kind, NodeUnaryExpression)
	}

	if node.Span != span {
		t.Errorf("node span = %+v, want %+v", node.Span, span)
	}

	payload := arena.UnaryPayload(node.Data)

	if payload.Operand != operand ||
		payload.Operator != lexer.TokenMinus {
		t.Errorf("unary payload = %+v, unexpected contents", payload)
	}
}

func TestArenaNestedExpressions(t *testing.T) {
	arena := NewArena("-(10 + 20)")

	// Represents -(10 + 20).
	left := arena.Add(FlatNode{Kind: NodeIntegerLiteral})
	right := arena.Add(FlatNode{Kind: NodeIntegerLiteral})

	binaryID := arena.AddBinary(
		source.ByteSpan{Start: 2, End: 9},
		left,
		lexer.TokenPlus,
		right,
	)

	unaryID := arena.AddUnary(
		source.ByteSpan{Start: 0, End: 10},
		lexer.TokenMinus,
		binaryID,
	)

	unaryNode := arena.Get(unaryID)

	if unaryNode.Kind != NodeUnaryExpression {
		t.Fatalf(
			"unary node kind = %d, want %d",
			unaryNode.Kind,
			NodeUnaryExpression,
		)
	}

	if unaryNode.Span != (source.ByteSpan{Start: 0, End: 10}) {
		t.Errorf("unexpected unary expression span: %+v", unaryNode.Span)
	}

	unary := arena.UnaryPayload(unaryNode.Data)

	if unary.Operand != binaryID {
		t.Fatalf("unary operand = %d, want %d", unary.Operand, binaryID)
	}

	binaryNode := arena.Get(unary.Operand)

	if binaryNode.Kind != NodeBinaryExpression {
		t.Fatalf(
			"binary node kind = %d, want %d",
			binaryNode.Kind,
			NodeBinaryExpression,
		)
	}

	if binaryNode.Span != (source.ByteSpan{Start: 2, End: 9}) {
		t.Errorf("unexpected binary expression span: %+v", binaryNode.Span)
	}

	binary := arena.BinaryPayload(binaryNode.Data)

	if binary.Left != left || binary.Right != right {
		t.Errorf("nested binary payload = %+v, unexpected operands", binary)
	}
}

func TestArenaAddBinaryInvalidOperand(t *testing.T) {
	var arena Arena
	valid := arena.Add(FlatNode{Kind: NodeIntegerLiteral})

	beforeNodes := arena.Len()
	beforePayloads := len(arena.binaries)

	defer func() {
		if recover() == nil {
			t.Error("AddBinary did not panic for invalid operand")
		}
		if arena.Len() != beforeNodes ||
			len(arena.binaries) != beforePayloads {
			t.Error("arena was modified after invalid operand")
		}
	}()

	arena.AddBinary(
		source.ByteSpan{},
		valid,
		lexer.TokenPlus,
		InvalidNodeID,
	)
}

func TestArenaAddUnaryInvalidOperand(t *testing.T) {
	var arena Arena

	defer func() {
		if recover() == nil {
			t.Error("AddUnary did not panic for invalid operand")
		}
		if arena.Len() != 0 || len(arena.unaries) != 0 {
			t.Error("arena was modified after invalid operand")
		}
	}()

	arena.AddUnary(
		source.ByteSpan{},
		lexer.TokenMinus,
		InvalidNodeID,
	)
}
func TestArenaSourceText(t *testing.T) {
	arena := NewArena("let age be 20")

	tests := []struct {
		span source.ByteSpan
		want string
	}{
		{source.ByteSpan{Start: 4, End: 7}, "age"},
		{source.ByteSpan{Start: 11, End: 13}, "20"},
		{source.ByteSpan{Start: 0, End: 0}, ""},
	}

	for _, tt := range tests {
		if got := arena.SourceText(tt.span); got != tt.want {
			t.Errorf("SourceText(%+v) = %q, want %q", tt.span, got, tt.want)
		}
	}
}

func TestArenaSourceTextInvalidSpan(t *testing.T) {
	arena := NewArena("hello")

	defer func() {
		if recover() == nil {
			t.Error("expected panic for invalid source span")
		}
	}()

	arena.SourceText(source.ByteSpan{Start: 2, End: 10})
}

func TestArenaAddInteger(t *testing.T) {
	arena := NewArena("12345")

	id := arena.AddInteger(source.ByteSpan{
		Start: 0,
		End:   5,
	})

	node := arena.Get(id)

	if node.Kind != NodeIntegerLiteral {
		t.Fatalf("kind = %d, want %d", node.Kind, NodeIntegerLiteral)
	}

	if got := arena.SourceText(node.Span); got != "12345" {
		t.Errorf("integer text = %q, want %q", got, "12345")
	}
}

func TestArenaAddIdentifier(t *testing.T) {
	arena := NewArena("let age be 20")

	id := arena.AddIdentifier(source.ByteSpan{
		Start: 4,
		End:   7,
	})

	node := arena.Get(id)

	if node.Kind != NodeIdentifier {
		t.Fatalf("kind = %d, want %d", node.Kind, NodeIdentifier)
	}

	if got := arena.SourceText(node.Span); got != "age" {
		t.Errorf("identifier = %q, want %q", got, "age")
	}
}

func TestArenaAddPrimaryInvalidSpan(t *testing.T) {
	tests := []struct {
		name string
		add  func(*Arena)
	}{
		{
			name: "integer",
			add: func(a *Arena) {
				a.AddInteger(source.ByteSpan{Start: 0, End: 10})
			},
		},
		{
			name: "identifier",
			add: func(a *Arena) {
				a.AddIdentifier(source.ByteSpan{Start: 5, End: 2})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arena := NewArena("hello")

			defer func() {
				if recover() == nil {
					t.Error("expected panic for invalid span")
				}
				if arena.Len() != 0 {
					t.Error("arena modified after invalid span")
				}
			}()

			tt.add(arena)
		})
	}
}

func TestArenaAddString(t *testing.T) {
	arena := NewArena(`"hello"`)

	id := arena.AddString(
		source.ByteSpan{Start: 0, End: 7},
		"hello",
	)

	node := arena.Get(id)

	if node.Kind != NodeStringLiteral {
		t.Fatalf("kind = %d, want %d", node.Kind, NodeStringLiteral)
	}

	if got := arena.StringValue(id); got != "hello" {
		t.Errorf("StringValue() = %q, want %q", got, "hello")
	}
}

func TestArenaDecodedString(t *testing.T) {
	arena := NewArena(`"hello\nworld"`)

	id := arena.AddString(
		source.ByteSpan{Start: 0, End: 14},
		"hello\nworld",
	)

	if got := arena.StringValue(id); got != "hello\nworld" {
		t.Errorf("decoded value = %q, want %q", got, "hello\nworld")
	}
}

func TestArenaStringStorageGrowth(t *testing.T) {
	arena := NewArena(`"x"`)

	const count = 10_000

	for i := 0; i < count; i++ {
		arena.AddString(
			source.ByteSpan{Start: 0, End: 3},
			"x",
		)
	}

	if len(arena.strings) != count {
		t.Fatalf("string count = %d, want %d", len(arena.strings), count)
	}

	for i := 0; i < count; i++ {
		if got := arena.StringValue(NodeID(i)); got != "x" {
			t.Fatalf("string %d = %q, want x", i, got)
		}
	}
}

func TestArenaBindingStatements(t *testing.T) {
	tests := []struct {
		name string
		kind NodeKind
		add  func(*Arena, NodeID) NodeID
	}{
		{
			name: "let",
			kind: NodeLetStatement,
			add: func(a *Arena, value NodeID) NodeID {
				return a.AddLet(
					source.ByteSpan{Start: 0, End: 13},
					source.ByteSpan{Start: 4, End: 7},
					value,
				)
			},
		},
		{
			name: "change",
			kind: NodeChangeStatement,
			add: func(a *Arena, value NodeID) NodeID {
				return a.AddChange(
					source.ByteSpan{Start: 14, End: 30},
					source.ByteSpan{Start: 21, End: 24},
					value,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arena := NewArena("let age be 20\nchange age to 21")
			value := arena.AddInteger(source.ByteSpan{Start: 11, End: 13})

			id := tt.add(arena, value)
			node := arena.Get(id)

			if node.Kind != tt.kind {
				t.Fatalf("kind = %d, want %d", node.Kind, tt.kind)
			}

			payload := arena.BindingPayload(node.Data)

			if payload.Value != value {
				t.Errorf("value = %d, want %d", payload.Value, value)
			}

			if got := arena.SourceText(payload.Name); got != "age" {
				t.Errorf("binding name = %q, want age", got)
			}
		})
	}
}

func TestArenaAddSay(t *testing.T) {
	arena := NewArena("say age")

	value := arena.AddIdentifier(source.ByteSpan{Start: 4, End: 7})
	id := arena.AddSay(source.ByteSpan{Start: 0, End: 7}, value)

	node := arena.Get(id)

	if node.Kind != NodeSayStatement {
		t.Fatalf("kind = %d, want %d", node.Kind, NodeSayStatement)
	}

	if NodeID(node.Data) != value {
		t.Errorf("say value = %d, want %d", node.Data, value)
	}
}

func TestArenaBindingPayloadGrowth(t *testing.T) {
	arena := NewArena("let x be 1")
	value := arena.AddInteger(source.ByteSpan{Start: 9, End: 10})

	const count = 10_000

	for i := 0; i < count; i++ {
		id := arena.AddLet(
			source.ByteSpan{Start: 0, End: 10},
			source.ByteSpan{Start: 4, End: 5},
			value,
		)

		if id != NodeID(i+1) {
			t.Fatalf("node ID = %d, want %d", id, i+1)
		}
	}

	if len(arena.bindings) != count {
		t.Fatalf("binding count = %d, want %d", len(arena.bindings), count)
	}

	for i := 0; i < count; i++ {
		payload := arena.BindingPayload(uint32(i))

		if payload.Value != value {
			t.Fatalf("binding %d has incorrect value", i)
		}
	}
}

func TestArenaInvalidStatementReference(t *testing.T) {
	tests := []struct {
		name string
		add  func(*Arena)
	}{
		{
			name: "let",
			add: func(a *Arena) {
				a.AddLet(
					source.ByteSpan{Start: 0, End: 7},
					source.ByteSpan{Start: 4, End: 7},
					InvalidNodeID,
				)
			},
		},
		{
			name: "change",
			add: func(a *Arena) {
				a.AddChange(
					source.ByteSpan{Start: 0, End: 7},
					source.ByteSpan{Start: 4, End: 7},
					InvalidNodeID,
				)
			},
		},
		{
			name: "say",
			add: func(a *Arena) {
				a.AddSay(
					source.ByteSpan{Start: 0, End: 7},
					InvalidNodeID,
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arena := NewArena("let age")
			beforeNodes := arena.Len()
			beforeBindings := len(arena.bindings)

			defer func() {
				if recover() == nil {
					t.Error("expected panic for invalid reference")
				}

				if arena.Len() != beforeNodes ||
					len(arena.bindings) != beforeBindings {
					t.Error("arena modified after invalid reference")
				}
			}()

			tt.add(arena)
		})
	}
}

func TestArenaRejectsStatementAsExpression(t *testing.T) {
	tests := []struct {
		name string
		add  func(*Arena, NodeID)
	}{
		{
			name: "unary operand",
			add: func(a *Arena, id NodeID) {
				a.AddUnary(source.ByteSpan{}, lexer.TokenMinus, id)
			},
		},
		{
			name: "binary left operand",
			add: func(a *Arena, id NodeID) {
				a.AddBinary(
					source.ByteSpan{},
					id,
					lexer.TokenPlus,
					0,
				)
			},
		},
		{
			name: "binary right operand",
			add: func(a *Arena, id NodeID) {
				a.AddBinary(
					source.ByteSpan{},
					0,
					lexer.TokenPlus,
					id,
				)
			},
		},
		{
			name: "let value",
			add: func(a *Arena, id NodeID) {
				a.AddLet(source.ByteSpan{}, source.ByteSpan{}, id)
			},
		},
		{
			name: "change value",
			add: func(a *Arena, id NodeID) {
				a.AddChange(source.ByteSpan{}, source.ByteSpan{}, id)
			},
		},
		{
			name: "say value",
			add: func(a *Arena, id NodeID) {
				a.AddSay(source.ByteSpan{}, id)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			arena := NewArena("say 1")

			value := arena.AddInteger(
				source.ByteSpan{Start: 4, End: 5},
			)
			statement := arena.AddSay(
				source.ByteSpan{Start: 0, End: 5},
				value,
			)

			beforeNodes := arena.Len()
			beforeBinaries := len(arena.binaries)
			beforeUnaries := len(arena.unaries)
			beforeBindings := len(arena.bindings)

			defer func() {
				if recover() == nil {
					t.Error("expected panic for statement used as expression")
				}

				if arena.Len() != beforeNodes ||
					len(arena.binaries) != beforeBinaries ||
					len(arena.unaries) != beforeUnaries ||
					len(arena.bindings) != beforeBindings {
					t.Error("arena modified after rejected expression")
				}
			}()

			tt.add(arena, statement)
		})
	}
}

func TestArenaRollbackRestoresAllBuffers(t *testing.T) {
	a := NewArena("1 2")

	one := a.AddInteger(source.ByteSpan{Start: 0, End: 1})
	a.AddStatement(a.AddSay(
		source.ByteSpan{Start: 0, End: 1},
		one,
	))

	checkpoint := a.Checkpoint()

	two := a.AddInteger(source.ByteSpan{Start: 2, End: 3})

	a.AddString(source.ByteSpan{Start: 2, End: 3}, "two")

	a.AddBinaryPayload(BinaryPayload{})
	a.AddUnaryPayload(UnaryPayload{})
	a.AddBindingPayload(BindingPayload{})

	a.AddStatement(a.AddSay(
		source.ByteSpan{Start: 2, End: 3},
		two,
	))

	a.Rollback(checkpoint)

	if got := len(a.nodes); got != checkpoint.nodes {
		t.Errorf("nodes = %d, want %d", got, checkpoint.nodes)
	}
	if got := len(a.strings); got != checkpoint.strings {
		t.Errorf("strings = %d, want %d", got, checkpoint.strings)
	}
	if got := len(a.binaries); got != checkpoint.binaries {
		t.Errorf("binaries = %d, want %d", got, checkpoint.binaries)
	}
	if got := len(a.unaries); got != checkpoint.unaries {
		t.Errorf("unaries = %d, want %d", got, checkpoint.unaries)
	}
	if got := len(a.bindings); got != checkpoint.bindings {
		t.Errorf("bindings = %d, want %d", got, checkpoint.bindings)
	}
	if got := len(a.statements); got != checkpoint.statements {
		t.Errorf("statements = %d, want %d", got, checkpoint.statements)
	}

	if got := a.StatementCount(); got != 1 {
		t.Errorf("preserved statement count = %d, want 1", got)
	}
}
