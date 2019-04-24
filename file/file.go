package file

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"io/ioutil"
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

func ToString(fp string) (string, error) {
	b, err := ioutil.ReadFile(fp)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func Md5(fp string) (string, error) {
	f, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	m := md5.New()
	io.Copy(m, f)
	return hex.EncodeToString(m.Sum(nil)), nil
}

func Crc32(fp string) (string, error) {
	f, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	ieee := crc32.NewIEEE()
	io.Copy(ieee, f)
	return fmt.Sprintf("%x", ieee.Sum32()), nil
}
