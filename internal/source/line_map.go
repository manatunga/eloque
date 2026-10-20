package source

import "sort"

type LineMap struct {
    lineStarts []uint32
}

func NewLineMap(sourceText string) LineMap {
	lineStarts := make([]uint32, 1, (len(sourceText)/50)+2)
	lineStarts[0] = 0
	for i := 0; i < len(sourceText); i++ {
		if sourceText[i] == '\n' {
			lineStarts = append(lineStarts, uint32(i+1))
		}
	}
	return LineMap{
		lineStarts: lineStarts,
	}
}

func (m *LineMap) Position(offset uint32) Position {
	if offset == 0 {
		return Position{Line: 1, Column: 1}
	}

	idx := sort.Search(len(m.lineStarts), func(i int) bool {
		return m.lineStarts[i] > offset
	})
	lineIdx := idx - 1

	if lineIdx < 0 {
		lineIdx = 0
	} else if lineIdx >= len(m.lineStarts) {
		lineIdx = len(m.lineStarts) - 1
	}

	lineStart := int(m.lineStarts[lineIdx])
	column := (int(offset) - lineStart) + 1
	return Position{
		Line: lineIdx + 1,
		Column: column,
	}
}

func (m *LineMap) Span(span ByteSpan) Span {
	return Span{
		Start: m.Position(span.Start),
		End: m.Position(span.End),
	}
}