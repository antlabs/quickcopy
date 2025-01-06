package mytest

import (
	"testing"
)

func TestQuickCopyInt2(t *testing.T) {
	src := &copyint2Src{A: 1, B: 2, C: 3, D: 4}
	dst := &copyint2Dst{}

	quickcopyint2(dst, src)

	if dst.A != int(src.A) {
		t.Errorf("Expected A to be %d, got %d", src.A, dst.A)
	}
	if dst.B != int(src.B) {
		t.Errorf("Expected B to be %d, got %d", src.B, dst.B)
	}
	if dst.C != int(src.C) {
		t.Errorf("Expected C to be %d, got %d", src.C, dst.C)
	}
	if dst.D != int(src.D) {
		t.Errorf("Expected D to be %d, got %d", src.D, dst.D)
	}
}

func TestQuickCopyInt3(t *testing.T) {
	src := &copyint5src{A: 1, B: 2, C: 3, D: 4}
	dst := &copyint4Dst{}

	quickcopyint3(dst, src)

	if dst.A != int8(src.A) {
		t.Errorf("Expected A to be %d, got %d", src.A, dst.A)
	}
	if dst.B != int16(src.B) {
		t.Errorf("Expected B to be %d, got %d", src.B, dst.B)
	}
	if dst.C != int32(src.C) {
		t.Errorf("Expected C to be %d, got %d", src.C, dst.C)
	}
	if dst.D != int64(src.D) {
		t.Errorf("Expected D to be %d, got %d", src.D, dst.D)
	}
}

func TestQuickCopyInt4(t *testing.T) {
	src := &copyint6src{A: 1}
	dst := &copyint6dst{}

	quickcopyint4(dst, src)

	if dst.A != int16(src.A) {
		t.Errorf("Expected A to be %d, got %d", src.A, dst.A)
	}
}

func TestQuickCopyInt5(t *testing.T) {
	src := &copyint6dst{A: 1}
	dst := &copyint6src{}

	quickcopyint5(dst, src)

	if dst.A != int8(src.A) {
		t.Errorf("Expected A to be %d, got %d", src.A, dst.A)
	}
}
