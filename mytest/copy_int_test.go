package mytest

import (
	"testing"
)

type copy1 struct {
	A	int8
	B	int16
	C	int32
	D	int64
}

type copy2 struct {
	A	int8
	B	int16
	C	int32
	D	int64
}

// :quickcopy
func QuickCopy(dst *copy1, src *copy2) {

	dst.A = src.A

	dst.B = src.B

	dst.C = src.C

	dst.
		D = src.D
}
func TestQuickCopy(t *testing.T) {
	src := &copy2{A: 1, B: 2, C: 3, D: 4}
	dst := &copy1{}

	QuickCopy(dst, src)

	if dst.A != src.A {
		t.Errorf("Expected A to be %d, got %d", src.A, dst.A)
	}
	if dst.B != src.B {
		t.Errorf("Expected B to be %d, got %d", src.B, dst.B)
	}
	if dst.C != src.C {
		t.Errorf("Expected C to be %d, got %d", src.C, dst.C)
	}
	if dst.D != src.D {
		t.Errorf("Expected D to be %d, got %d", src.D, dst.D)
	}
}
