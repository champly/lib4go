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

var (
	nullByte []byte

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

		if p.err != nil {
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
		err = io.EOF
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.c.L == nil {
		p.c.L = &p.mu
	}
	p.err = err
	defer p.c.Signal()
}

func NewPipeBuffer(capacity int) IoBuffer {
	p := &pipe{
		IoBuffer: newIoBuffer(capacity),
	}
	p.c.L = &p.mu
	return p
}

func newIoBuffer(capacity int) IoBuffer {
	return nil
}
