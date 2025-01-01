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

func QuickCopy(dst *copy2, src *copy1) {
	dst.A = src.A
	dst.B =
		src.

			// :quickcopy
			B
	dst.C = src.
		C
	dst.D = src.D
}
