package source

// Position identifies a location in Eloque source code.
// Line and column numbers are 1-based.
type Position struct {
	Line   int
	Column int
}

// Span identifies a half-open range in Eloque source code.
// Start is inclusive and End is exclusive.
type Span struct {
	Start Position
	End   Position
}

// ByteSpan identifies a half-open byte range in source text.
// Start is inclusive and End is exclusive.
type ByteSpan struct {
	Start uint32
	End   uint32
}