package pool

import (
	"sync"
	"time"
)

const windowSize = 100
const slow bool = true
const fast bool = false

var slowAvgThreshold = time.Millisecond * 500

// movingWindow uses a Circular buffer to keep track of the last N fast or slow counts.
// these will be used to re-calculate the percentage of slow counts.
type movingWindow struct {
	window    []bool
	position  int
	fastCount int
	slowCount int
	lock      sync.Mutex
}

func newMovingWindow() *movingWindow {
	return &movingWindow{
		window: make([]bool, windowSize),
	}
}

// add newDuration to the mix, recalculate & return the current weight (1-100) based on the new values.
func (w *movingWindow) trackNewDuration(newDuration time.Duration) {
	w.lock.Lock()
	defer w.lock.Unlock()
	newValue := newDuration > slowAvgThreshold
	oldValue := w.window[w.position]
	w.position = (w.position + 1) % windowSize

	if oldValue == newValue {
		return
	}

	if oldValue == slow {
		w.slowCount--
	} else {
		w.fastCount--
	}

	if newValue == slow {
		w.slowCount++
	} else {
		w.fastCount++
	}
}

func (w *movingWindow) getScore() float64 {
	return float64(w.slowCount) / float64(w.slowCount+w.fastCount)
}
