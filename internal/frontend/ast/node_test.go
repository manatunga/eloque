package ast

import (
	"testing"
	"unsafe"
)

func TestFlatNodeSize(t *testing.T) {
	const expectedSize = 16

	actualSize := unsafe.Sizeof(FlatNode{})

	if actualSize != expectedSize {
		t.Errorf(
			"FlatNode size = %d bytes, want %d bytes",
			actualSize,
			expectedSize,
		)
	}
}

func TestBindingPayloadSize(t *testing.T) {
	const expectedSize = 12

	if got := unsafe.Sizeof(BindingPayload{}); got != expectedSize {
		t.Errorf(
			"BindingPayload size = %d bytes, want %d",
			got,
			expectedSize,
		)
	}
}

func TestFlatRepresentationSizes(t *testing.T) {
	cases := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"BinaryPayload", unsafe.Sizeof(BinaryPayload{}), 12},
		{"UnaryPayload", unsafe.Sizeof(UnaryPayload{}), 8},
		{"NodeID", unsafe.Sizeof(NodeID(0)), 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Errorf(
					"%s size = %d bytes, want %d",
					tc.name,
					tc.got,
					tc.want,
				)
			}
		})
	}
}
