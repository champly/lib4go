package bytes

import (
	"fmt"
	"reflect"
	"unsafe"
)

// SliceByteToString []byte to string
func SliceByteToString(b []byte) string {
	p := unsafe.Pointer((*reflect.StringHeader)(unsafe.Pointer(&b)).Data)

	var s string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&s))
	hdr.Data = uintptr(p)
	hdr.Len = len(s)
	return s
}

// SliceByteConcat concat interface{} to []byte
func SliceByteConcat(dst *[]byte, res interface{}) (err error) {
	switch v := res.(type) {
	case []byte:
		*dst = append(*dst, v...)
	case string:
		*dst = append(*dst, StringToSliceByte(v)...)
	case uint8:
		*dst = append(*dst, v)
	case uint16:
		for _, item := range Uint16ToSliceByte(v) {
			*dst = append(*dst, item)
		}
	case uint32:
		for _, item := range Uint32ToSliceByte(v) {
			*dst = append(*dst, item)
		}
	case uint64:
		for _, item := range Uint64ToSliceByte(v) {
			*dst = append(*dst, item)
		}
	case int8:
		*dst = append(*dst, Int8ToSliceByte(v))
	case int16:
		for _, item := range Int16ToSliceByte(v) {
			*dst = append(*dst, item)
		}
	case int32:
		for _, item := range Int32ToSliceByte(v) {
			*dst = append(*dst, item)
		}
	case int64:
		for _, item := range Int64ToSliceByte(v) {
			*dst = append(*dst, item)
		}
	default:
		err = fmt.Errorf("error type")
	}

	return
}
