package tool

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"time"

	"github.com/champly/lib4go/security/md5"
)

// GetGUID get guid
func GetGUID() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return md5.Encrypt(base64.URLEncoding.EncodeToString(b))
}

func GoWithRecover(handler func(), recoverHandler func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "%s goroutine panic:%v\n%s\n", time.Now().Format("2006-01-02 15:04:05"), r, string(debug.Stack()))

				if recoverHandler != nil {
					go func() {
						defer func() {
							if p := recover(); p != nil {
								fmt.Fprintf(os.Stderr, "recover goroutine panic:%v\n%s\n", p, string(debug.Stack()))
							}
						}()

						recoverHandler(r)
					}()
				}
			}
		}()

		handler()
	}()
}
