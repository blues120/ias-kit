package string_helper

import (
	"strconv"
	"strings"
)

// StringToUIntSlice 字符串转换为uint切片
func StringToUIntSlice(s string) []uint64 {
	var result []uint64
	for _, v := range strings.Split(s, ",") {
		i, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			continue
		}
		result = append(result, i)
	}
	return result
}
