package float

import (
	"strconv"
	"testing"
)

type FloatSource struct {
	F32	float32
	F64	float64
	Str	string
}

type FloatDest struct {
	F32	float64	// float32 -> float64
	F64	float32	// float64 -> float32
	Str	float64	// string -> float64
}

// :quickcopy --allow-narrow
func CopyFloat(dst *FloatDest, src *FloatSource) {

	dst.F32 = float64(src.F32)

	dst.F64 = float32(src.F64)

	dst.Str = func(s string) float64 {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}(src.
		Str,
	)
}

func TestFloatCopy(t *testing.T) {
	src := &FloatSource{
		F32:	3.14159,
		F64:	2.71828,
		Str:	"123.456",
	}

	dst := &FloatDest{}
	CopyFloat(dst, src)

	// 验证float32到float64的转换
	if float64(src.F32) != dst.F32 {
		t.Errorf("F32 copy failed, got %v, want %v", dst.F32, src.F32)
	}

	// 验证float64到float32的转换
	if float32(src.F64) != dst.F64 {
		t.Errorf("F64 copy failed, got %v, want %v", dst.F64, src.F64)
	}

	// 验证string到float64的转换
	expected := 123.456
	if dst.Str != expected {
		t.Errorf("Str to float64 copy failed, got %v, want %v", dst.Str, expected)
	}
}
