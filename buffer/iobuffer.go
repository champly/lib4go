package buffer

import (
	"errors"
	"io"
	"sync"
	"time"
)

const (
	AutoExpand      = -1
	MinRead         = 1 << 9
	MaxRead         = 1 << 17
	ResetOffMark    = -1
	DefaultSize     = 1 << 4
	MaxBufferLength = 1 << 20
	MaxThreshold    = 1 << 22
)

var nullByte []byte

var (
	EOF                  = errors.New("EOF")
	ErrTooLarge          = errors.New("io buffer: too large")
	ErrNegativeCount     = errors.New("io buffer: negative count")
	ErrInvalidWriteCount = errors.New("io buffer: invalid write count")
	ErrClosedPipeWrite   = errors.New("write on closed buffer")
	ConnReadTimeout      = 15 * time.Second
)

type pipe struct {
	IoBuffer

	mu sync.Mutex
	c  sync.Cond

	err error
}

func (p *pipe) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.IoBuffer == nil {
		return 0
	}
	return p.IoBuffer.Len()
}

func (p *pipe) Read(d []byte) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.c.L == nil {
		p.c.L = &p.mu
	}
	for {
		if p.IoBuffer != nil && p.IoBuffer.Len() > 0 {
			return p.IoBuffer.Read(d)
		}
		if err != nil {
			return 0, p.err
		}
		p.c.Wait()
	}
}

func (p *pipe) Write(d []byte) (n int, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.c.L == nil {
		p.c.L = &p.mu
	}

	defer p.c.Signal()
	if p.err != nil {
		return 0, ErrClosedPipeWrite
	}
	return len(d), p.IoBuffer.Append(d)
}

func (p *pipe) CloseWithError(err error) {
	if err == nil {
		p.err = io.EOF
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.c.L == nil {
		p.c.L = &p.mu
	}
	p.err = err
	p.c.Signal()
}

func NewPipeBuffer(capacity int) IoBuffer {
	return &pipe{
		IoBuffer: newIoBuffer(capacity),
	}
}

type ioBuffer struct {
	buf     []byte
	off     int
	offMark int
	count   int32
	eof     bool

	b *[]byte
}

func newIoBuffer(capacity int) IoBuffer {
	return nil
}

func (b *ioBuffer) Read(p []byte) (n int, err error) {
	if b.off >= len(b.buf) {
		b.Reset()

		if len(p) == 0 {
			return
		}
		return 0, io.EOF
	}

	n = copy(p, b.buf[b.off:])
	b.off += n
	return
}

func (b *ioBuffer) Grow(n int) error {
	_, ok := b.tryGrowByReslice(n)
	if !ok {
		b.grow(n)
	}

	return nil
}

func (b *ioBuffer) ReadOnce(r io.Reader) (n int64, err error) {
	var m int

	if b.off > 0 && b.off >= len(b.buf) {
		b.Reset()
	}

	if b.off >= (cap(b.buf) - len(b.buf)) {
		b.copy(0)
	}

	if b.off == len(b.buf) && cap(b.buf) > MaxBufferLength {
		b.Free()
		b.Alloc(MaxRead)
	}

	l := cap(b.buf) - len(b.buf)
	m, err = r.Read(b.buf[len(b.buf):cap(b.buf)])
	b.buf = b.buf[0 : len(b.buf)+m]
	n = int64(m)

	if l == m {
		b.copy(AutoExpand)
	}
	return n, err
}

func (b *ioBuffer) ReadFrom(r io.Reader) (n int64, err error) {
	if b.off >= len(b.buf) {
		b.Reset()
	}

	for {
		if free := cap(b.buf) - len(b.buf); free < MinRead {
			if b.off+free < MinRead {
				b.copy(MinRead)
			} else {
				b.copy(0)
			}
		}

		m, e := r.Read(b.buf[len(b.buf):cap(b.buf)])

		b.buf = b.buf[0 : len(b.buf)+m]
		n += int64(m)

		if e == io.EOF {
			break
		}

		if m == 0 {
			break
		}

		if e != nil {
			return n, e
		}
	}
	return
}

func (b *ioBuffer) Reset() {
	b.buf = b.buf[:0]
	b.off = 0
	b.offMark = ResetOffMark
	b.eof = false
}

func (b *ioBuffer) Alloc(size int) {
	if b.buf != nil {
		b.Free()
	}
	if size <= 0 {
		size = DefaultSize
	}
	b.b = b.makeSlice(size)
	b.buf = *b.b
	b.buf = b.buf[:0]
}

func (b *ioBuffer) Free() {
	b.Reset()
	b.giveSlice()
}

func (b *ioBuffer) Len() int {
	return len(b.buf) - b.off
}

func (b *ioBuffer) tryGrowByReslice(n int) (int, bool) {
	if l := len(b.buf); l+n <= cap(b.buf) {
		b.buf = b.buf[:l+n]

		return l, true
	}
	return 0, false
}

func (b *ioBuffer) grow(n int) int {
	m := b.Len()

	if m == 0 && b.off != 0 {
		b.Reset()
	}

	if i, ok := b.tryGrowByReslice(n); ok {
		return i
	}
	if m+n <= cap(b.buf)/2 {
		b.copy(0)
	} else {
		b.copy(n)
	}

	b.off = 0
	b.buf = b.buf[:m+n]
	return m
}

func (b *ioBuffer) copy(expand int) {
	var newBuf []byte
	var bufp *[]byte

	if expand > 0 || expand == AutoExpand {
		cap := cap(b.buf)
		if cap < 2*MinRead {
			cap = 2 * MinRead
		} else if cap < MaxThreshold {
			cap = 2 * cap
		} else {
			cap = cap + cap/4
		}

		if expand == AutoExpand {
			expand = 0
		}

		bufp = b.makeSlice(cap + expand)
		newBuf = *bufp
		copy(newBuf, b.buf[b.off:])
		PutBytes(b.b)
		b.b = bufp
	} else {
		newBuf = b.buf
		copy(newBuf, b.buf[b.off:])
	}
	b.buf = newBuf[:len(b.buf)-b.off]
	b.off = 0
}

func (b *ioBuffer) makeSlice(n int) *[]byte {
	return GetBytes(n)
}

func (b *ioBuffer) giveSlice() {
	if b.b != nil {
		PutBytes(b.b)
		b.b = nil
		b.buf = nullByte
	}
}

func (b *ioBuffer) CloseWithError(err error) {}
