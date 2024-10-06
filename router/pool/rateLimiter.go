package pool

import (
	"sync"
	"time"
)

const windowSize = 100

type speed int

const (
	slow speed = iota
	fast
)

var slowAvgThreshold = time.Millisecond * 200

// rateLimiter uses a circular buffer to keep track of the last N fast or slow counts.
// these will be used to re-calculate the percentage of slow counts (Score).
type rateLimiter struct {
	window          []speed
	position        int
	fastCount       int
	slowCount       int
	currentStage    waitTimeCalculator
	lock            sync.Mutex
	lastHandleTime  time.Time
	currentWaitTime time.Duration
}

func newRateLimiter() *rateLimiter {
	return &rateLimiter{
		window:       make([]speed, windowSize),
		currentStage: stage_ok,
	}
}

// add newDuration to the mix, recalculate & return the current weight (1-100) based on the new values.
func (w *rateLimiter) trackNewDuration(newDuration time.Duration) {
	w.lock.Lock()
	defer w.lock.Unlock()
	newValue := fast
	if newDuration > slowAvgThreshold {
		newValue = slow
	}
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
	w.lastHandleTime = time.Now()
	w.updateStage()
}

func (w *rateLimiter) updateStage() {
	newScore := w.score()
	oldStage := w.currentStage
	for _, s := range all_stages {
		if s.contains(newScore) {
			w.currentStage = s
			w.currentWaitTime = s.calculateNewWaitTime(oldStage, newScore, w.scoreLastN(10))
		}
	}
}

func (w *rateLimiter) score() float64 {
	if w.slowCount+w.fastCount == 0 {
		return 1
	}

	return float64(w.fastCount) / float64(w.slowCount+w.fastCount)
}

func (w *rateLimiter) scoreLastN(n int) float64 {
	if n > w.slowCount+w.fastCount {
		n = w.slowCount + w.fastCount
	}

	startPos := w.position - 1
	if startPos < 0 {
		startPos = len(w.window)
	}

	slowCount, fastCount := 0, 0
	for m := 0; m < n; m++ {
		idx := startPos - m
		if idx < 0 {
			idx = len(w.window) + idx
		}

		if w.window[idx] == slow {
			slowCount++
		} else {
			fastCount++
		}
	}

	if slowCount+fastCount == 0 {
		return 1
	}

	return float64(fastCount) / float64(slowCount+fastCount)
}

func (w *rateLimiter) canHandleCall() bool {
	return time.Now().After(w.lastHandleTime.Add(w.currentWaitTime))
}
