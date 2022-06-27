package queue

import (
	"sync/atomic"
	"unsafe"
)

// LKQueue is a lock-free unbounded queue.
type LKQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

type node struct {
	value interface{}
	next  unsafe.Pointer
}

func NewLKQueue() *LKQueue {
	n := unsafe.Pointer(&node{})
	return &LKQueue{head: n, tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *LKQueue) Enqueue(v interface{}) {
	n := &node{value: v}
	for {
		tail := load(&q.tail)
		next := load(&tail.next)
		// are tail and next consistent?
		if tail == load(&q.tail) {
			if next == nil {
				if cas(&tail.next, next, n) {
					// Enqueue is done.
					cas(&q.tail, tail, n)
					return
				}
			} else {
				// tail was not pointing to the last node
				// try to swing tail to the next node.
				cas(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *LKQueue) Dequeue() interface{} {
	for {
		head := load(&q.head)
		tail := load(&q.tail)
		next := load(&head.next)
		// are head, tail and next consistent?
		if head == load(&q.head) {
			// is queue empty or tail falling behind
			if head == tail {
				// queue is empty
				if next == nil {
					return nil
				}
				// tail is falling behind. try to advance it
				cas(&q.tail, tail, next)
			} else {
				// read value before CAS, otherwise another dequeue might free the next node
				v := next.value
				if cas(&q.head, head, next) {
					// Dequeue is done. return
					return v
				}
			}
		}
	}
}

func load(p *unsafe.Pointer) (n *node) {
	return (*node)(atomic.LoadPointer(p))
}

func cas(p *unsafe.Pointer, old, new *node) (ok bool) {
	return atomic.CompareAndSwapPointer(p, unsafe.Pointer(old), unsafe.Pointer(new))
}
