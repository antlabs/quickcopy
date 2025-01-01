package mytest

type copy1 struct {
	A int8
	B int16
	C int32
	D int64
}

type copy2 struct {
	A int8
	B int16
	C int32
	D int64
}

// :quickcopy
func QuickCopy(dst *copy1, src *copy2) {
	dst.A = src.A
	dst.B =
		src.B
	dst.
		C = src.
		C
	dst.D = src.D

}
