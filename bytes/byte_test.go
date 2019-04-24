package bytes

import "testing"

func TestSliceByteToString(t *testing.T) {
	str := SliceByteToString([]byte{65, 1, 3, 2, 123, 59})
	t.Log(str)
}

func TestSliceByteConcat(t *testing.T) {
	content := []byte{}
	SliceByteConcat(&content, []byte{1, 2, 3})
	t.Log(content)

	SliceByteConcat(&content, "123")
	t.Log(content)

	SliceByteConcat(&content, uint8(2))
	t.Log(content)

	SliceByteConcat(&content, uint16(2))
	t.Log(content)

	SliceByteConcat(&content, uint32(2))
	t.Log(content)

	SliceByteConcat(&content, uint64(2))
	t.Log(content)

	SliceByteConcat(&content, int8(2))
	t.Log(content)

	SliceByteConcat(&content, int16(2))
	t.Log(content)

	SliceByteConcat(&content, int32(2))
	t.Log(content)

	SliceByteConcat(&content, int64(2))
	t.Log(content)

	err := SliceByteConcat(&content, struct {
		A int
	}{A: 1})
	if err == nil || err.Error() != "error type" {
		t.Errorf("default type error:%+v", err)
	}
}
