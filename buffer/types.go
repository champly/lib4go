package buffer

import "io"

type BufferPoolCtx interface {
	Index() int

	New() interface{}

	Reset(interface{})
}

type IoBuffer interface {
	Read(p []byte) (n int, err error)

	ReadeOnce(r io.Reader) (n int64, err error)

	ReadFrom(r io.Reader) (n int64, err error)

	Grow(n int) error

	Write(p []byte) (n int, err error)

	WriteString(s string) (n int, err error)

	WriteByte(p byte) error

	WriteUint16(p uint16) error

	WriteUint32(p uint32) error

	WriteUint64(p uint64) error

	WriteTo(w io.Writer) (n int64, err error)

	Peek(n int) []byte

	Bytes() []byte

	Drain(offset int)

	Len() int

	Cap() int

	Reset()

	Clone() IoBuffer

	String() string

	Alloc(int)

	Free()

	Count(int32) int32

	EOF() bool

	SetEOF(eof bool)

	Append(data []byte) error

	CloseWithError(err error)
}
