package ping

import (
	"testing"
)

func TestSend(t *testing.T) {
	t.Log(Send("10.28.255.253"))
	// for i := 0; i < 10; i++ {
	// 	t.Log(Send("10.28.255.254"))
	// 	time.Sleep(time.Second)
	// }
}
