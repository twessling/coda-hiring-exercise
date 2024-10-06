package ratelimit

import (
	"mrbarrel/lib/interval"
	"time"
)

type waitTimeCalculator interface {
	calculateNewWaitTime(oldStage waitTimeCalculator, newScore, newScoreLast10 float64) time.Duration
	contains(f float64) bool
}
type stage struct {
	interval        *interval.Interval
	defaultWaittime time.Duration
}

type deadStage struct {
	stage
}

func (s *deadStage) contains(score float64) bool {
	return s.interval.Contains(score)
}
func (s *deadStage) calculateNewWaitTime(oldStage waitTimeCalculator, newScore, newScoreLast10 float64) time.Duration {
	return s.defaultWaittime
}

type okStage struct {
	stage
}

func (s *okStage) contains(score float64) bool {
	return s.interval.Contains(score)
}
func (s *okStage) calculateNewWaitTime(oldStage waitTimeCalculator, newScore, newScoreLast10 float64) time.Duration {
	return s.defaultWaittime
}

type slowStage struct {
	stage
}

func (s *slowStage) contains(score float64) bool {
	return s.interval.Contains(score)
}
func (s *slowStage) calculateNewWaitTime(oldStage waitTimeCalculator, newScore, newScoreLast10 float64) time.Duration {

	// transitions & slow stage handling
	switch oldStage {
	case stage_ok:
		// we
	case stage_dead:
	case stage_slow:
	}

	return s.defaultWaittime
}

var (
	stage_ok   = &okStage{stage: stage{interval: interval.New(0.99, true, 1.0, true), defaultWaittime: 0}}
	stage_slow = &slowStage{stage: stage{interval: interval.New(0.10, true, 0.99, false), defaultWaittime: 100 * time.Millisecond}}
	stage_dead = &deadStage{stage: stage{interval: interval.New(0, true, 0.10, false), defaultWaittime: time.Minute}}
	all_stages = []waitTimeCalculator{stage_ok, stage_slow, stage_dead}
)
