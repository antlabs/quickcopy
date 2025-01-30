package mytest

import (
	"testing"
)

type copyint2Src struct {
	A	int8
	B	int16
	C	int32
	D	int64
}

type copyint2Dst struct {
	A	int
	B	int
	C	int
	D	int
}
type copyint3Dst struct {
	A	int64
	B	int64
	C	int64
	D	int64
}

type copyint4Dst struct {
	A	int8
	B	int16
	C	int32
	D	int64
}

type copyint5src struct {
	A	int
	B	int
	C	int
	D	int
}

// :quickcopy
func quickcopyint2(dst *copyint2Dst, src *copyint2Src) {

	dst.
		A = int(src.A)

	dst.
		B = int(src.
		B)

	dst.C =
		int(src.C,
		)

	dst.
		D = int(src.D)
}

// :quickcopy
func quickcopyint22(dst *copyint3Dst, src *copyint2Src) {

	dst.A = int64(src.A)

	dst.B = int64(src.B)

	dst.
		C = int64(src.C)

	dst.
		D = src.
		D
}

// :quickcopy --allow-narrow
func quickcopyint3(dst *copyint4Dst, src *copyint5src) {

	dst.
		A = int8(src.A)

	dst.
		B = int16(
		src.
			B)

	dst.C = int32(
		src.C)

	dst.D = int64(src.D)
}

type copyint6src struct {
	A int8
}
type copyint6dst struct {
	A int16
}

// :quickcopy
func quickcopyint4(dst *copyint6dst, src *copyint6src) {

	dst.
		A = int16(src.A)
}

// :quickcopy --allow-narrow
func quickcopyint5(dst *copyint6src, src *copyint6dst) {

	dst.
		A = int8(src.A)
}

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
