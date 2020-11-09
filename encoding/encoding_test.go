package encoding

import (
	"testing"
)

var (
	utf8Str = "Utf8 转 GBK测试"
	gbkStr  = ""
)

func TestUtf8ToGbk(t *testing.T) {
	r, err := UTF82GBK(utf8Str)
	if err != nil {
		t.Error("test [encoding.Utf8ToGbk] fail:", err)
	}
	gbkStr = r
}

func TestGbkToUtf8(t *testing.T) {
	r, err := GBK2UTF8(gbkStr)
	if err != nil {
		t.Error("test [encoding.GbkToUtf8] fail:", err)
	}
	if r != utf8Str {
		t.Errorf("test [encoding.Utf8ToGbk] fail, gbk:%s, utf8:%s", gbkStr, r)
	}
}

func TestEncodeUCS2(t *testing.T) {
	input := "1234"
	except := "0031003200330034"
	actual := EncodeUCS2(input)

	if except != actual {
		t.Errorf("test [encoding.EncodeUCS2] fail, input:%s, except:%s, actual:%s", input, except, actual)
	}
}

func TestDecodeUCS2(t *testing.T) {
	input := "0031003200330034"
	except := "1234"
	actual := DecodeUCS2(input)

	if except != actual {
		t.Errorf("test [encoding.DecodeUCS2] fail, input:%s, except:%s, actual:%s", input, except, actual)
	}
}

func TestEncode7Bit(t *testing.T) {
	input := "1234"
	except := "31D98C06"
	actual := Encode7Bit(input)

	if except != actual {
		t.Errorf("test [encoding.Encode7Bit] fail, input:%s, except:%s, actual:%s", input, except, actual)
	}
}

func TestDecode7Bit(t *testing.T) {
	input := "31D98C06"
	except := "1234"
	actual := Decode7Bit(input)

	if except != actual {
		t.Errorf("test [encoding.Decode7Bit] fail, input:%s, except:%s, actual:%s", input, except, actual)
	}
}
