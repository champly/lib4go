package gctuner

import (
	"runtime"
	"runtime/debug"
	"time"
)

type finalizer struct {
	ch  chan time.Time
	ref *finalizerRef
}

type finalizerRef struct {
	parent *finalizer
}

// default GOGC = 100
// FIXME, if use sets the GOGC at startup
// how to get this value correctly?
var previousGOGC = 100

// don't trigger err log on every failure
var failCounter = -1

func getCurrentPercentAndChangeGOGC() {
	memPercent, err := getUsage()
	if err != nil {
		failCounter++
		if failCounter%10 == 0 {
			println("failed to adjust GC", err.Error())
		}
		return
	}

	// hard_target = live_dataset + live_dataset (GOGC / 100).
	//		hard_target = memoryLimitInPercent
	//		live_dataset = memPercent
	// so gogc = (hard_target - live_dataset) / live_dataset * 100
	newgogc := (memoryLimitInPercent - memPercent) / memPercent * 100.0

	// if newgogc < 0, we have to use the previous gogc to determine the next
	if newgogc < 0 {
		newgogc = float64(previousGOGC) * memoryLimitInPercent / memoryLimitInPercent
	}
	previousGOGC = debug.SetGCPercent(int(newgogc))
}

func finalizerHandler(f *finalizerRef) {
	select {
	case f.parent.ch <- time.Time{}:
	default:
	}

	getCurrentPercentAndChangeGOGC()
	runtime.SetFinalizer(f, finalizerHandler)
}

func NewTuner(useCgroup bool, percent float64) *finalizer {
	if useCgroup {
		getUsage = getUsageCGroup
	} else {
		getUsage = getUsageNormal
	}

	memoryLimitInPercent = percent

	f := &finalizer{
		ch: make(chan time.Time, 1),
	}

	f.ref = &finalizerRef{parent: f}
	runtime.SetFinalizer(f.ref, finalizerHandler)
	f.ref = nil
	return f
}
