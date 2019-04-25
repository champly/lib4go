package file

import "testing"

func TestExists(t *testing.T) {
	data := map[string]bool{
		"/Users":       true,
		"/setst":       false,
		"path.go":      true,
		"../README.md": true,
		"../":          true,
		".":            true,
		"!!!":          false,
	}

	for path, expect := range data {
		actual := Exists(path)
		if expect != actual {
			t.Errorf("test [file.Exists] fail: input is:%s, acutal is:%t, but expect is:%t", path, actual, expect)
		}
	}
}

func TestAbs(t *testing.T) {
	data := map[string]string{
		"/root":        "/root",
		"/setst":       "/setst",
		"path.go":      "/Users/champly/go/src/github.com/champly/lib4go/file/path.go",
		"../README.md": "/Users/champly/go/src/github.com/champly/lib4go/README.md",
		"../":          "/Users/champly/go/src/github.com/champly/lib4go",
		".":            "/Users/champly/go/src/github.com/champly/lib4go/file",
		"!!!":          "/Users/champly/go/src/github.com/champly/lib4go/file/!!!",
	}

	for path, expect := range data {
		actual, err := Abs(path)
		if err != nil {
			t.Errorf("test [file.Abs] fail: input is:%s, expect is:%s, err is:%s", path, expect, err.Error())
			continue
		}
		if expect != actual {
			t.Errorf("test [file.Abs] fail: input is:%s, acutal is:%s, but expect is:%s", path, actual, expect)
		}
	}
}

func TestMd5(t *testing.T) {
	t.Log(Md5("file.go"))
}

func TestCrc32(t *testing.T) {
	t.Log(Crc32("file.go"))
}
