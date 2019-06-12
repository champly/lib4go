package types

import (
	"fmt"
	"strconv"
)

func GetString(input interface{}, def ...string) string {
	if input != nil {
		if r := fmt.Sprintf("%v", input); r != "" {
			return r
		}
	}
	if len(def) > 0 {
		return def[0]
	}

	return ""
}

func GetInt(input interface{}, def ...int) int {
	b, err := strconv.Atoi(fmt.Sprintf("%v", input))
	if err == nil {
		return b
	}
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
