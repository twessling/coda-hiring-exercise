package ratelimit

import (
	"sync"
	"time"
)

const windowSize = 100

type speed int

const (
	fast speed = iota
	slow
)

// RateLimiter uses a circular buffer to keep track of the last N fast or slow counts.
// these will be used to re-calculate the percentage of slow counts (Score).
type RateLimiter struct {
	window          []speed
	position        int
	fastCount       int
	slowCount       int
	currentStage    waitTimeCalculator
	lock            sync.Mutex
	lastHandleTime  time.Time
	currentWaitTime time.Duration
	slowThreshold   time.Duration
}

func NewRateLimiter(slowThreshold time.Duration) *RateLimiter {
	return &RateLimiter{
		window:        make([]speed, windowSize),
		currentStage:  stage_ok,
		slowThreshold: slowThreshold,
		fastCount:     100, // start out as if it's fast all the way
	}
}

// add newDuration to the mix, recalculate & return the current weight (1-100) based on the new values.
func (w *RateLimiter) TrackNewDuration(newDuration time.Duration) {
	w.lock.Lock()
	defer w.lock.Unlock()
	newValue := fast
	if newDuration > w.slowThreshold {
		newValue = slow
	}
	oldValue := w.window[w.position]
	defer func() {
		w.position = (w.position + 1) % windowSize
	}()

	if oldValue == newValue {
		return
	}

	if oldValue == slow {
		w.slowCount--
		if w.slowCount < 0 {
			w.slowCount = 0
		}
	} else {
		w.fastCount--
		if w.fastCount < 0 {
			w.fastCount = 0
		}
	}

	if newValue == slow {
		w.window[w.position] = slow
		w.slowCount++
	} else {
		w.window[w.position] = fast
		w.fastCount++
	}
	w.lastHandleTime = time.Now()
	w.updateStage()
}

func (w *RateLimiter) updateStage() {
	newScore := w.score()
	oldStage := w.currentStage
	for _, s := range all_stages {
		if s.contains(newScore) {
			w.currentStage = s
			w.currentWaitTime = s.calculateNewWaitTime(oldStage, w.currentWaitTime, newScore, w.scoreLastN(10))
		}
	}
}

func (w *RateLimiter) score() float64 {
	if w.slowCount+w.fastCount == 0 {
		return 1
	}

	return float64(w.fastCount) / float64(w.slowCount+w.fastCount)
}

func (w *RateLimiter) scoreLastN(n int) float64 {
	if n > w.slowCount+w.fastCount {
		n = w.slowCount + w.fastCount
	}

	startPos := w.position - 1
	if startPos < 0 {
		startPos = len(w.window) - 1
	}

	slowCount, fastCount := 0, 0
	for m := 0; m < n; m++ {
		idx := startPos - m
		if idx < 0 {
			idx = len(w.window) + idx - 1
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

func (w *RateLimiter) CanHandleCall() bool {
	return time.Now().After(w.lastHandleTime.Add(w.currentWaitTime))
}
