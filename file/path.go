package file

import (
	"os"
	"path/filepath"
)

func Exists(fp string) bool {
	_, err := os.Stat(fp)
	return err == nil || os.IsExist(err)
}

func Abs(fp string) (string, error) {
	return filepath.Abs(fp)
}
