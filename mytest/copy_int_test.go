package mytest

import (
	"testing"
)

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
