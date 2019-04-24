package bytes

import "testing"

func TestStringToSliceByte(t *testing.T) {
	input := "饕餮大餐！@#¥%……&*（）™ "
	result := StringToSliceByte(input)
	if string(result) != input {
		t.Errorf("test StringToSliceByte error:result:%s, input:%s", string(result), input)
	}
	t.Log(string(result))
}
