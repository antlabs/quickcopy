package mytest

type copyint2Src struct {
	A int8
	B int16
	C int32
	D int64
}

type copyint2Dst struct {
	A int
	B int
	C int
	D int
}
type copyint3Dst struct {
	A int64
	B int64
	C int64
	D int64
}

type copyint4Dst struct {
	A int8
	B int16
	C int32
	D int64
}

type copyint5src struct {
	A int
	B int
	C int
	D int
}

// :quickcopy
func quickcopyint2(dst *copyint2Dst, src *copyint2Src) {
	dst.
		A =
		int(src.A)
	dst.B = int(src.
		B)
	dst.C = int(src.
		C)
	dst.
		D = int(
		src.
			D)
}

// :quickcopy
func quickcopyint22(dst *copyint3Dst, src *copyint2Src) {
	dst.A =
		int64(src.A)
	dst.B = int64(src.B)
	dst.C =
		int64(src.
			C)
	dst.
		D =
		src.D
}

// :quickcopy --allow-narrow
func quickcopyint3(dst *copyint4Dst, src *copyint5src) {
	dst.
		A =
		int8(src.A)
	dst.B = int16(src.
		B)
	dst.
		C = int32(src.C)
	dst.D =
		int64(src.D)
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
		A =
		int16(src.A)
}

// :quickcopy --allow-narrow
func quickcopyint5(dst *copyint6src, src *copyint6dst) {
	dst.
		A =
		int8(src.A)
}
