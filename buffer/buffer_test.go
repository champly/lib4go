package buffer

import (
	"context"
	"testing"
)

var (
	ins = demo{}
)

func init() {
	RegistryBuffer(&ins)
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
	ctx := context.TODO()
	ctx = NewBufferPoolContext(ctx)

	bv := PoolContext(ctx)
	defer bv.Give()

	value := bv.Find(&ins, nil)
	t.Log(value)
}
