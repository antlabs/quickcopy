package mytest

import (
	"testing"
	"time"
)

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
