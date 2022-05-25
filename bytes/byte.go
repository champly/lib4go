package bytes

import (
	"fmt"
	"reflect"
	"unsafe"
)

// SliceByteToString []byte to string
func SliceByteToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{
		Data: bh.Data,
		Len:  bh.Len,
	}
	return *(*string)(unsafe.Pointer(&sh))
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
