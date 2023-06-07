package encoding

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func GBK2UTF8(s string) (string, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewDecoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func UTF82GBK(s string) (string, error) {
	reader := transform.NewReader(bytes.NewReader([]byte(s)), simplifiedchinese.GBK.NewEncoder())
	d, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(d), nil
}

func EncodeUCS2(s string) string {
	var r string
	for _, item := range s {
		r += fmt.Sprintf("%04x", int(item))
	}
	return r
}

func DecodeUCS2(s string) string {
	if len(s)%4 != 0 {
		return s
	}

	r := []string{}
	for i := 0; i < len(s); i += 4 {
		t, _ := strconv.ParseInt(s[i:i+4], 16, 32)
		r = append(r, fmt.Sprintf("%c", t))
	}
	return strings.Join(r, "")
}

func Decode7Bit(s string) string {
	var binStr string
	for i := len(s); i > 0; i -= 2 {
		ii, _ := strconv.ParseInt(s[i-2:i], 16, 32)
		binStr += int32toBinary(int(ii))
	}

	result := ""
	for i := len(binStr); i > 0; i -= 7 {
		if i > 7 {
			result += fmt.Sprintf("%c", binStrToInt(binStr[i-7:i]))
		} else {
			b := binStrToInt(binStr[:i])
			if b != 0 {
				result += fmt.Sprintf("%c", b)
			}
		}
	}
	return result
}

func Encode7Bit(s string) string {
	tStr := []string{}
	for _, v := range s {
		tStr = append(tStr, int32toBinary(int(v))[1:])
	}

	binStr := ""
	for i := len(tStr) - 1; i >= 0; i-- {
		binStr += tStr[i]
	}
	r := ""
	for i := len(binStr); i > 0; i -= 8 {
		if i > 8 {
			r += fmt.Sprintf("%02X", binStrToInt(binStr[i-8:i]))
		} else {
			r += fmt.Sprintf("%02X", binStrToInt(binStr[:i]))
		}
	}
	return r
}

func int32toBinary(orig int) string {
	r := ""
	i := 7
	for orig != 0 {
		t := orig % 2
		r = fmt.Sprintf("%d", t) + r
		i--
		orig /= 2
	}
	for i >= 0 {
		r = "0" + r
		i--
	}
	return r
}

func binStrToInt(orig string) int {
	r := 0
	for _, i := range orig {
		r *= 2
		if string(i) == "1" {
			r++
		}
	}
	return r
}
