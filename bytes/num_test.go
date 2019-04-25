package bytes

import "testing"

type tesType struct {
	result string
	input  interface{}
}

func TestNum(t *testing.T) {
	t.Log(Int8ToSliceByte(int8(1)))
	t.Log(Int16ToSliceByte(int16(1)))
	t.Log(Int32ToSliceByte(int32(1)))
	t.Log(Int64ToSliceByte(int64(1)))
	t.Log(Uint16ToSliceByte(uint16(1)))
	t.Log(Uint32ToSliceByte(uint32(1)))
	t.Log(Uint64ToSliceByte(uint64(1)))
}
