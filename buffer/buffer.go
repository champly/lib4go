package buffer

import (
	"context"
	"sync"
	"sync/atomic"
	"unsafe"
)

const (
	maxBufferPool = 16
	ctxStoreKey   = "BUFFER_CACHE_POOL"
)

var (
	index int32
	bPool = bufferPoolArray[:]
	vPool = new(valuePool)

	bufferPoolArray [maxBufferPool]bufferPool
	nullBufferValue [maxBufferPool]interface{}
)

type TempBufferCtx struct {
	index int
}

func (t *TempBufferCtx) Index() int {
	return t.index
}

func (t *TempBufferCtx) New() interface{} {
	return nil
}

func (t *TempBufferCtx) Reset(interface{}) {
}

type bufferPool struct {
	ctx BufferPoolCtx
	sync.Pool
}

func (p *bufferPool) take() (value interface{}) {
	value = p.Get()
	if value == nil {
		value = p.ctx.New()
	}
	return
}

func (p *bufferPool) give(value interface{}) {
	p.ctx.Reset(value)
	p.Put(value)
}

type ifaceWords struct {
	typ  unsafe.Pointer
	data unsafe.Pointer
}

func setIndex(poolCtx BufferPoolCtx, index int) {
	p := (*ifaceWords)(unsafe.Pointer(&poolCtx))
	temp := (*TempBufferCtx)(p.data)
	temp.index = index
}

func RegistryBuffer(poolCtx BufferPoolCtx) {
	i := atomic.AddInt32(&index, 1)
	if i < 0 || i > maxBufferPool {
		panic("buffersize over full")
	}
	bPool[i].ctx = poolCtx
	setIndex(poolCtx, int(i))
}

type valuePool struct {
	sync.Pool
}

type bufferValue struct {
	value    [maxBufferPool]interface{}
	transmit [maxBufferPool]interface{}
}

func NewBufferPoolContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxStoreKey, newBufferValue())
}

func newBufferValue() (value *bufferValue) {
	v := vPool.Get()
	if v == nil {
		value = new(bufferValue)
	} else {
		value = v.(*bufferValue)
	}
	return
}

func (bv *bufferValue) Find(poolCtx BufferPoolCtx, x interface{}) interface{} {
	i := poolCtx.Index()
	if i <= 0 || i > int(index) {
		panic("buffer should call buffer.RegistryBuffer()")
	}
	if bv.value[i] != nil {
		return bv.value[i]
	}
	return bv.Take(poolCtx)
}

func (bv *bufferValue) Take(poolCtx BufferPoolCtx) (value interface{}) {
	i := poolCtx.Index()
	value = bPool[i].take()
	bv.value[i] = value
	return
}

func (bv *bufferValue) Give() {
	if index <= 0 {
		return
	}
	// first index is 1
	for i := 1; i < int(index); i++ {
		value := bv.value[i]
		if value != nil {
			bPool[i].give(value)
		}
		value = bv.transmit[i]
		if value != nil {
			bPool[i].give(value)
		}
	}
	bv.value = nullBufferValue
	bv.transmit = nullBufferValue

	// Give bufferValue to Pool
	vPool.Put(bv)
}

func PoolContext(ctx context.Context) *bufferValue {
	if ctx != nil {
		if val := ctx.Value(ctxStoreKey); val != nil {
			return val.(*bufferValue)
		}
	}
	return newBufferValue()
}
