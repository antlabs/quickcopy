package mytest

import (
	"testing"
	"time"
)

type copy3 struct {
	T string
}

type copy4 struct {
	T time.Time
}

// :quickcopy
func QuickCopy3(dst *copy4, src *copy3) {

	dst.T = func(s string) time.Time {
		t, _ := time.Parse(time.RFC3339, s)
		return t
	}(src.T)
}

// :quickcopy

func QuickCopy4(dst *copy3, src *copy4) {
	dst.T = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}(src.T)
}
func TestQuickCopy3(t *testing.T) {
	src := &copy3{T: "2025-01-06T21:28:46+08:00"}
	dst := &copy4{}

	QuickCopy3(dst, src)

	expectedTime, _ := time.Parse(time.RFC3339, src.T)
	if !dst.T.Equal(expectedTime) {
		t.Errorf("Expected T to be %v, got %v", expectedTime, dst.T)
	}
}

func TestQuickCopy4(t *testing.T) {
	expectedTime := "2025-01-06T21:28:46+08:00"
	srcTime, _ := time.Parse(time.RFC3339, expectedTime)
	src := &copy4{T: srcTime}
	dst := &copy3{}

	QuickCopy4(dst, src)

	if dst.T != expectedTime {
		t.Errorf("Expected T to be %s, got %s", expectedTime, dst.T)
	}
}
