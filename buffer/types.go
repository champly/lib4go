package buffer

type BufferPoolCtx interface {
	Index() int

	New() interface{}

	Reset(interface{})
}