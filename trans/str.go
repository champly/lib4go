package trans

import "unsafe"

// Str2Bytes string trans []byte
func Str2Bytes(s string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&s))
	b := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&b))
}

// Bytes2Str []byte trans string
func Bytes2Str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
