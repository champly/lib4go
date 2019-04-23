package file

import (
	"os"
	"path"
	"path/filepath"
)

func CurrentPath() string {
	path, _ := filepath.Abs(os.Args[0])
	return path
}

func Ext(f string) string {
	return path.Ext(f)
}
