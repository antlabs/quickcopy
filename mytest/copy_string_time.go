package mytest

import "time"

type copy3 struct {
	T string
}

type copy4 struct {
	T time.Time
}

func QuickCopy3(dst *copy3, src *copy4) { // :quickcopy
	dst.T = func(t time.Time) string {
		return t.Format( // :quickcopy
		time.RFC3339,
		)
	}(src.T)
}

func QuickCopy4(dst *copy4, src *copy3) {
	dst.T = func(s string) time.Time {
		t, _ := time.Parse(time.RFC3339,
			s)
		return t
	}(src.T)
}

// :quickcopy
