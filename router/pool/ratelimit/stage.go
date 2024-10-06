package ratelimit

import (
	"mrbarrel/lib/interval"
	"time"
)

type waitTimeCalculator interface {
	calculateNewWaitTime(oldStage waitTimeCalculator, oldWaitTime time.Duration, newScore, newScoreLast10 float64) time.Duration
	contains(f float64) bool
}
type stage struct {
	interval        *interval.Interval
	defaultWaittime time.Duration
}

// represents real bad, as good as dead performance.
type deadStage struct {
	stage
}

// represents good performance
type okStage struct {
	stage
}

// represents a large range of mediocre performance
type slowStage struct {
	stage
}

const no_wait time.Duration = 0
const ok_threshold float64 = 0.99
const dead_threshold float64 = 0.1

var (
	stage_ok   = &okStage{stage: stage{interval: interval.New(ok_threshold, true, 1.0, true), defaultWaittime: no_wait}}
	stage_slow = &slowStage{stage: stage{interval: interval.New(dead_threshold, false, ok_threshold, false), defaultWaittime: 100 * time.Millisecond}}
	stage_dead = &deadStage{stage: stage{interval: interval.New(0, true, dead_threshold, true), defaultWaittime: 10 * time.Second}}
	all_stages = []waitTimeCalculator{stage_ok, stage_slow, stage_dead}
)

func (s *deadStage) contains(score float64) bool {
	return s.interval.Contains(score)
}
func (s *deadStage) calculateNewWaitTime(_ waitTimeCalculator, _ time.Duration, _, _ float64) time.Duration {
	return s.defaultWaittime
}

func (s *okStage) contains(score float64) bool {
	return s.interval.Contains(score)
}
func (s *okStage) calculateNewWaitTime(_ waitTimeCalculator, _ time.Duration, _, _ float64) time.Duration {
	return s.defaultWaittime
}

func (s *slowStage) contains(score float64) bool {
	return s.interval.Contains(score)
}
func (s *slowStage) calculateNewWaitTime(oldStage waitTimeCalculator, oldWaitTime time.Duration, newScore, newScoreLast10 float64) time.Duration {

	// transitions & slow stage handling are the complex bits
	switch oldStage {
	case stage_ok:
		// we degraded, go slower a bit
		return s.defaultWaittime
	case stage_dead:
		// we coming back
		return time.Second
	case stage_slow:
		// better or worse, still not great, recalc wait time based on overall & last N new scores
		return s.calculateSlowStage(oldWaitTime, newScore, newScoreLast10)
	}
	return s.defaultWaittime
}

func (s *slowStage) calculateSlowStage(oldWaitTime time.Duration, newScore, newScoreLast10 float64) time.Duration {
	// just entered from OK == 100 ms wait, score just below 0.99
	// just out of dead == time.Second, score just above 0.1
	// so range == 900 ms
	// slow scores range from 0.1 to 0.99 (exclusive, both)
	// very basic linear slowing here, can go all out.
	if oldWaitTime == no_wait {
		return s.defaultWaittime
	}

	if newScoreLast10 > 2*newScore {
		// we're improving greatly!
		return oldWaitTime / 2
	}

	if newScoreLast10 < 2*newScore {
		// we degrading fast
		return oldWaitTime * 2
	}

	// linear =100 + 1000 * ((1-x)*(0.99-0.1))
	factor := (1 - newScore) / (ok_threshold - dead_threshold)
	return 100*time.Millisecond + time.Second*time.Duration(factor)
}
