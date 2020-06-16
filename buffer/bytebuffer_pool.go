package buffer

import "sync"

const (
	minShift = 6
	maxShift = 18
	errSlot  = -1
)

var (
	bbPool *byteBufferPool
)

func init() {
	bbPool = newByteBufferPool()
}

type bufferSlot struct {
	defaultSize int
	pool        sync.Pool
}

type byteBufferPool struct {
	minShift int
	minSize  int
	maxSize  int

	pool []*bufferSlot
}

func newByteBufferPool() *byteBufferPool {
	p := &byteBufferPool{
		minShift: minShift,
		minSize:  1 << minShift,
		maxSize:  1 << maxShift,
	}

	for i := 0; i < maxShift-minShift; i++ {
		slab := &bufferSlot{
			defaultSize: 1 << (minShift + i),
		}
		p.pool = append(p.pool, slab)
	}

	return p
}

func (p *byteBufferPool) slot(size int) int {
	if size > p.maxSize {
		return errSlot
	}
	var slot, shift int

	if size > p.minShift {
		size--
		for size > 0 {
			size = size >> 1
			shift++
		}
		slot = shift - p.minShift
	}
	return slot
}

func newBytes(size int) []byte {
	return make([]byte, size)
}

func (p *byteBufferPool) take(size int) *[]byte {
	slot := p.slot(size)
	if slot == errSlot {
		b := newBytes(size)
		return &b
	}

	v := p.pool[slot].pool.Get()
	if v == nil {
		b := newBytes(p.pool[slot].defaultSize)
		b = b[:size]
		return &b
	}
	b := v.(*[]byte)
	*b = (*b)[:size]
	return b
}

func (p *byteBufferPool) give(buf *[]byte) {
	if buf == nil {
		return
	}
	size := cap(*buf)
	slot := p.slot(size)
	if slot == errSlot {
		return
	}
	if size != int(p.pool[slot].defaultSize) {
		return
	}
	p.pool[slot].pool.Put(buf)
}

type ByteBufferPoolContainer struct {
	bytes []*[]byte
	*byteBufferPool
}

func NewByteBufferPoolContainer() *ByteBufferPoolContainer {
	return &ByteBufferPoolContainer{
		byteBufferPool: bbPool,
	}
}

func (c *ByteBufferPoolContainer) Reset() {
	for _, buf := range c.bytes {
		c.give(buf)
	}
	c.bytes = c.bytes[:0]
}

func (c *ByteBufferPoolContainer) Take(size int) *[]byte {
	buf := c.take(size)
	c.bytes = append(c.bytes, buf)
	return buf
}

func GetBytes(size int) *[]byte {
	return bbPool.take(size)
}

func PutBytes(buf *[]byte) {
	bbPool.give(buf)
}
