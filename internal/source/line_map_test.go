package source

import (
	"testing"
)

func TestLineMapPosition(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		offset   uint32
		expected Position
	}{
		{
			name:     "empty source",
			source:   "",
			offset:   0,
			expected: Position{Line: 1, Column: 1},
		},
		{
			name:     "single line start",
			source:   "hello",
			offset:   0,
			expected: Position{Line: 1, Column: 1},
		},
		{
			name:     "single line middle",
			source:   "hello",
			offset:   3,
			expected: Position{Line: 1, Column: 4},
		},
		{
			name:     "single line end",
			source:   "hello",
			offset:   5,
			expected: Position{Line: 1, Column: 6},
		},
		{
			name:     "newline byte",
			source:   "hello\nworld",
			offset:   5,
			expected: Position{Line: 1, Column: 6},
		},
		{
			name:     "second line start",
			source:   "hello\nworld",
			offset:   6,
			expected: Position{Line: 2, Column: 1},
		},
		{
			name:     "second line middle",
			source:   "hello\nworld",
			offset:   8,
			expected: Position{Line: 2, Column: 3},
		},
		{
			name:     "trailing newline",
			source:   "hello\n",
			offset:   6,
			expected: Position{Line: 2, Column: 1},
		},
		{
			name:     "empty middle line",
			source:   "a\n\nb",
			offset:   2,
			expected: Position{Line: 2, Column: 1},
		},
		{
			name:     "line after empty line",
			source:   "a\n\nb",
			offset:   3,
			expected: Position{Line: 3, Column: 1},
		},
		{
			name:     "non ASCII byte column",
			source:   "éx\nz",
			offset:   2,
			expected: Position{Line: 1, Column: 3},
		},
		{
			name:     "non ASCII next line",
			source:   "éx\nz",
			offset:   4,
			expected: Position{Line: 2, Column: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineMap := NewLineMap(tt.source)

			got := lineMap.Position(tt.offset)
			if got != tt.expected {
				t.Errorf(
					"Position(%d) = %+v, want %+v",
					tt.offset,
					got,
					tt.expected,
				)
			}
		})
	}
}

func TestLineMapSpan(t *testing.T) {
	lineMap := NewLineMap("let age be 20\nsay age")

	input := ByteSpan{
		Start: 4,
		End:   18,
	}

	expected := Span{
		Start: Position{Line: 1, Column: 5},
		End:   Position{Line: 2, Column: 5},
	}

	got := lineMap.Span(input)
	if got != expected {
		t.Errorf("Span(%+v) = %+v, want %+v", input, got, expected)
	}
}

func TestLineMapLineStarts(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		expected []uint32
	}{
		{
			name:     "empty source",
			source:   "",
			expected: []uint32{0},
		},
		{
			name:     "single line",
			source:   "hello",
			expected: []uint32{0},
		},
		{
			name:     "multiple lines",
			source:   "a\nb\nc",
			expected: []uint32{0, 2, 4},
		},
		{
			name:     "empty lines",
			source:   "a\n\nb",
			expected: []uint32{0, 2, 3},
		},
		{
			name:     "trailing newline",
			source:   "a\n",
			expected: []uint32{0, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lineMap := NewLineMap(tt.source)

			if len(lineMap.lineStarts) != len(tt.expected) {
				t.Fatalf(
					"lineStarts length = %d, want %d",
					len(lineMap.lineStarts),
					len(tt.expected),
				)
			}

			for i, want := range tt.expected {
				if got := lineMap.lineStarts[i]; got != want {
					t.Errorf(
						"lineStarts[%d] = %d, want %d",
						i,
						got,
						want,
					)
				}
			}
		})
	}
}
