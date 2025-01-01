package mytest

import "time"

type copy3 struct {
	T string
}

type copy4 struct {
	T time.Time
}

// :quickcopy
func QuickCopy3(dst *copy3, src *copy4) {
}

// :quickcopy
func QuickCopy4(dst *copy4, src *copy3) {
}
