package md5

import (
	"strings"
	"testing"
)

func TestEncrypt(t *testing.T) {
	data := map[string]string{
		"123456": "E10ADC3949BA59ABBE56E057F20F883E",
		"md5":    "1BC29B36F623BA82AAF6724FD3B16718",
		" ":      "7215ee9c7d9dc229d2921a40e899ec5f",
		"":       "d41d8cd98f00b204e9800998ecf8427e",
		"中文":     "A7BAC2239FCDCB3A067903D8077C4A07",
	}
	for input, expect := range data {
		actual := Encrypt(input)
		if !strings.EqualFold(actual, expect) {
			t.Errorf("test [security.md5.Encrypt] fail: input is:%s, acutal is:%s, but expect is:%s", input, actual, expect)
		}
	}
}
