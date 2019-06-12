package types

import (
	"fmt"
	"strconv"
)

func GetString(input interface{}, def ...string) string {
	r, ok := input.(string)
	if ok {
		return r
	}
	if len(def) > 0 {
		return def[0]
	}
	return ""
}

func GetInt(input interface{}, def ...int) int {
	r, ok := input.(int)
	if ok {
		return r
	}
	b, err := strconv.Atoi(fmt.Sprintf("%v", input))
	if err == nil {
		return b
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
