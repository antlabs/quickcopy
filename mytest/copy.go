package mytest

type Copy struct {
	A int8
	B int16
	C int32
	D int64
}

type Copy2 struct {
	A int8
	B int16
	C int32
	D int64
}

// :quickcopy
func QuickCopy(dst *Copy, src *Copy2) {
	dst.A = src.A
	dst.B =
		src.B
	dst.
		C = src.
		C
	dst.D = src.D

}
