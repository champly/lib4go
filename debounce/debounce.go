package debounce

import (
	"time"

	"k8s.io/klog"
)

type Request interface {
	Merge(Request) Request
}

type Debounce struct {
	ch            chan Request
	debounceAfter time.Duration
	debounceMax   time.Duration
	pushFn        func(req Request)
}

func New(debounceAfter, debounceMax time.Duration, pushFn func(req Request)) *Debounce {
	d := &Debounce{
		ch:            make(chan Request, 0),
		debounceAfter: debounceAfter,
		debounceMax:   debounceMax,
		pushFn:        pushFn,
	}
	go d.start()
	return d
}

func (d *Debounce) start() {
	var timeChan <-chan time.Time
	var startDebounce time.Time
	var lastConfigUpdateTime time.Time

	pushCounter := 0
	debouncedEvents := 0

	var req Request

	free := true
	freeCh := make(chan struct{}, 1)

	push := func(req Request) {
		d.pushFn(req)
		freeCh <- struct{}{}
	}

	pushWorker := func() {
		eventDelay := time.Since(startDebounce)
		quietTime := time.Since(lastConfigUpdateTime)
		if eventDelay >= d.debounceMax || quietTime >= d.debounceAfter {
			if req != nil {
				pushCounter++
				klog.Infof("Push debounce stable[%d] %d: %v since last change, %v since last push",
					pushCounter, debouncedEvents,
					quietTime, eventDelay)

				free = false
				go push(req)
				req = nil
				debouncedEvents = 0
			}
		} else {
			timeChan = time.After(d.debounceAfter - quietTime)
		}
	}

	for {
		select {
		case <-freeCh:
			free = true
			pushWorker()
		case r, ok := <-d.ch:
			if !ok {
				return
			}

			lastConfigUpdateTime = time.Now()
			if debouncedEvents == 0 {
				timeChan = time.After(d.debounceAfter)
				startDebounce = lastConfigUpdateTime
			}
			debouncedEvents++

			if req == nil {
				req = r
				continue
			}
			req = req.Merge(r)
		case <-timeChan:
			if free {
				pushWorker()
			}
		}
	}
}

func (d *Debounce) Put(req Request) {
	d.ch <- req
}

func (d *Debounce) Close() {
	close(d.ch)
}
