package types

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
	if len(def) > 0 {
		return def[0]
	}
	return 0
}
