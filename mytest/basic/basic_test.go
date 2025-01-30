package basic

import (
	"fmt"
	"strconv"
)

// :quickcopy
func stringIntSlice(dst *[]string, src *[]int) {
	*dst =
		copySliceStringFromSliceInt(// :quickcopy
		*src)
}

func intStringSlice(dst *[]int, src *[]string) {
	*dst =
		copySliceIntFromSliceString(*src)
}

func copySliceStringFromSliceInt(src []int) []string {

	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	for i := range src {
		dst[i] = fmt.Sprint(src[i])

	}
	return dst
}

func copySliceIntFromSliceString(src []string) []int {

	if src == nil {
		return nil
	}
	dst := make([]int, len(src))
	for i := range src {
		dst[i] = func(s string) int {
			i,

				_ := strconv.Atoi(s)
			return i
		}(
			src[i])
	}
	return dst // :quickcopy
}
