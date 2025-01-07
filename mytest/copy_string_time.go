package mytest

import "time"

type copy3 struct {
	T string
}

type copy4 struct {
	T time.Time
}

// :quickcopy
func QuickCopy3(dst *copy4, src *copy3) {
	dst.T = func(s string) time.Time {
		t, _ := time.
			Parse(time.RFC3339, s)
		return t
	}(src.
		T)
}

// :quickcopy

func QuickCopy4(dst *copy3, src *copy4) {
	dst.T = func(t time.Time) string {
		return t.Format(time.RFC3339)
	}(src.T)
}
