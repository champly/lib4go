package types

import "testing"

func TestGetString(t *testing.T) {
	t.Log(GetString(1))
	t.Log(GetString(nil))
	t.Log(GetString(""))

	t.Log(GetString(1, "a"))
	t.Log(GetString(nil, "a"))
	t.Log(GetString("", "a"))
}
