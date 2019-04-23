package file

import "testing"

func TestExt(t *testing.T) {
	data := map[string]string{
		"/root":        "",
		"/setst":       "",
		"path.go":      ".go",
		"../README.md": ".md",
		"../":          "",
		".":            ".",
		"!!!":          "",
	}

	for path, expect := range data {
		actual := Ext(path)
		if expect != actual {
			t.Errorf("test [file.Ext] fail: input is:%s, acutal is:%s, but expect is:%s", path, actual, expect)
		}
	}
}

func TestCurrentPath(t *testing.T) {
	t.Log(CurrentPath())
}
