package queue

import (
	"sync"
	"time"

	"k8s.io/klog/v2"
)

type queueImpl struct {
	delay time.Duration
	tasks []Task
	cond  *sync.Cond

	closing bool
	mutext  sync.RWMutex
	closed  bool
}

func NewQueue(errorDelay time.Duration) Instance {
	return &queueImpl{
		delay:   errorDelay,
		tasks:   make([]Task, 0),
		closing: false,
		closed:  true,
		cond:    sync.NewCond(&sync.Mutex{}),
	}
}

func (q *queueImpl) Push(task Task) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	if !q.closing {
		q.tasks = append(q.tasks, task)
	}
	q.cond.Signal()
}

func (q *queueImpl) Run(stop <-chan struct{}) {
	q.mutext.Lock()
	if !q.closed {
		q.mutext.Unlock()
		panic("queue can not be run twice")
	}
	q.closed = false
	q.mutext.Unlock()

	q.cond.L.Lock()
	q.closing = false
	q.cond.L.Unlock()

	go func() {
		<-stop
		q.cond.L.Lock()
		q.cond.Signal()
		q.closing = true
		q.cond.L.Unlock()
	}()

	for {
		q.cond.L.Lock()
		for !q.closing && len(q.tasks) == 0 {
			q.cond.Wait()
		}

		if len(q.tasks) == 0 {
			q.cond.L.Unlock()
			return
		}

		var task Task
		task, q.tasks = q.tasks[0], q.tasks[1:]
		q.cond.L.Unlock()

		if err := task(); err != nil {
			klog.Infof("Work item handle failed (%v), retry after delay %v", err, q.delay)
			time.AfterFunc(q.delay, func() {
				q.Push(task)
			})
		}
	}
}
