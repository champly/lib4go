package buffer

import "testing"

func init() {
	RegistryBuffer(&demo{})
}

type demo struct {
	TempBufferCtx
}

func (d *demo) New() interface{} {
	return nil
}

func (d *demo) Reset(buf interface{}) {
	return
}

func TestRegistry(t *testing.T) {

}
