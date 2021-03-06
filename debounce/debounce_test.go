package debounce

import (
	"testing"
	"time"
)

type R struct {
	t time.Time
	i int
}

func (r *R) Merge(req Request) Request {
	if r == nil {
		return req
	}
	if req == nil {
		return r
	}

	return req
}

// func TestNew(t *testing.T) {
// 	d := New(time.Second*3, time.Second*10, func(req Request) {
// 		fmt.Println(time.Now())
// 		r := req.(*R)
// 		fmt.Println(r.t, r.i)
// 	})

// 	for i := 0; i < 15; i++ {
// 		d.Put(&R{
// 			t: time.Now(),
// 			i: i,
// 		})
// 		time.Sleep(time.Millisecond * 900)
// 	}

// 	time.Sleep(20 * time.Second)
// 	d.Close()
// }

func TestHighCurrent(t *testing.T) {
	d := New(time.Millisecond*100, time.Second*10, func(req Request) {
	})

	r := &R{
		t: time.Now(),
		i: 0,
	}
	for {
		for i := 0; i < 10000; i++ {
			d.Put(r)
		}
		// time.Sleep(time.Second)
	}
}
