package trans

import "testing"

func TestStr2Bytes(t *testing.T) {
	t.Log(Str2Bytes("aaaa"))
}

func TestBytes2Str(t *testing.T) {
	t.Log(Bytes2Str([]byte{97, 97, 97, 97}))
}

func BenchmarkStr2Bytes(t *testing.B) {
	for i := 0; i < t.N; i++ {
		Str2Bytes("aaaa")
	}
}

func BenchmarkBytes2Str(t *testing.B) {
	for i := 0; i < t.N; i++ {
		Bytes2Str([]byte{1, 2, 3})
	}
}
