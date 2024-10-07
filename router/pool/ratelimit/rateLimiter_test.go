package ratelimit

import (
	"testing"
	"time"
)

func TestScore(t *testing.T) {
	slowThreshold := time.Millisecond * 200
	tests := map[string]struct {
		durations []time.Duration
		wantScore float64
	}{
		"no durations": {
			wantScore: 1,
		},
		"one fast duration": {
			durations: []time.Duration{slowThreshold - time.Millisecond},
			wantScore: 1,
		},
		"one slow duration": { // as if it was all fast + one slow
			durations: []time.Duration{slowThreshold + time.Millisecond},
			wantScore: 0.99,
		},
		"all fast + one slow": {
			durations: append(repeat(time.Millisecond, windowSize+10), slowThreshold+time.Millisecond),
			wantScore: 0.99,
		},
		"all slow": {
			durations: repeat(slowThreshold+time.Millisecond, windowSize+10),
			wantScore: 0,
		},
		"10% fast, last 90% slow": {
			durations: append(repeat(time.Millisecond, windowSize*0.1), repeat(slowThreshold+time.Millisecond, windowSize*0.9)...),
			wantScore: 0.1,
		},
		"50/50 slow/fast": {
			durations: alternate(time.Millisecond, slowThreshold+time.Millisecond, windowSize),
			wantScore: 0.5,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rl := NewRateLimiter(slowThreshold)
			for _, d := range test.durations {
				rl.TrackNewDuration(d)
			}
			gotScore := rl.score()
			if gotScore != test.wantScore {
				t.Errorf("score mismatch. got %f want %f", gotScore, test.wantScore)
			}
		})
	}
}

func repeat(t time.Duration, count int) []time.Duration {
	var res []time.Duration
	for i := 0; i < count; i++ {
		res = append(res, t)
	}
	return res
}

func alternate(t1, t2 time.Duration, count int) []time.Duration {
	var res []time.Duration
	for i := 0; i < count; i++ {
		res = append(res, t1, t2)
	}
	return res
}

func TestScoreLastN(t *testing.T) {
	slowThreshold := time.Millisecond * 200
	tests := map[string]struct {
		durations      []time.Duration
		wantLastNScore float64
	}{
		"no durations": {
			wantLastNScore: 1,
		},
		"one fast duration": {
			durations:      []time.Duration{slowThreshold - time.Millisecond},
			wantLastNScore: 1,
		},
		"one slow duration": { // as if it was all fast + one slow at the end
			durations:      []time.Duration{slowThreshold + time.Millisecond},
			wantLastNScore: 0.9,
		},
		"all fast + one slow": {
			durations:      append(repeat(time.Millisecond, windowSize+10), slowThreshold+time.Millisecond),
			wantLastNScore: 0.9,
		},
		"all slow": {
			durations:      repeat(slowThreshold+time.Millisecond, windowSize+10),
			wantLastNScore: 0,
		},
		"10% fast, last 90% slow": {
			durations:      append(repeat(time.Millisecond, windowSize*0.1), repeat(slowThreshold+time.Millisecond, windowSize*0.9)...),
			wantLastNScore: 0,
		},
		"50/50 slow/fast": {
			durations:      alternate(time.Millisecond, slowThreshold+time.Millisecond, windowSize),
			wantLastNScore: 0.5,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rl := NewRateLimiter(slowThreshold)
			for _, d := range test.durations {
				rl.TrackNewDuration(d)
			}
			gotScore := rl.scoreLastN(10)
			if gotScore != test.wantLastNScore {
				t.Errorf("score mismatch. got %f want %f", gotScore, test.wantLastNScore)
			}
		})
	}
}

func TestTrackNewDuration(t *testing.T) {
	slowThreshold := time.Millisecond * 200
	tests := map[string]struct {
		durations    []time.Duration
		wantStage    waitTimeCalculator
		wantWaitTime time.Duration
	}{
		"no durations": {
			wantStage: stage_ok,
		},
		"one fast duration": {
			durations:    []time.Duration{slowThreshold - time.Millisecond},
			wantStage:    stage_ok,
			wantWaitTime: stage_ok.defaultWaittime,
		},
		"one slow duration": { // as if it was all fast + one slow at the end
			durations:    []time.Duration{slowThreshold + time.Millisecond},
			wantStage:    stage_ok,
			wantWaitTime: stage_ok.defaultWaittime,
		},
		"all fast + one slow": {
			durations:    append(repeat(time.Millisecond, windowSize+10), slowThreshold+time.Millisecond),
			wantStage:    stage_ok,
			wantWaitTime: stage_ok.defaultWaittime,
		},
		"all slow": {
			durations:    repeat(slowThreshold+time.Millisecond, windowSize+10),
			wantStage:    stage_dead,
			wantWaitTime: stage_dead.defaultWaittime,
		},
		"10% fast, last 90% slow": {
			durations:    append(repeat(time.Millisecond, windowSize*0.1), repeat(slowThreshold+time.Millisecond, windowSize*0.9)...),
			wantStage:    stage_dead,
			wantWaitTime: stage_dead.defaultWaittime,
		},
		"50/50 slow/fast": {
			durations:    alternate(time.Millisecond, slowThreshold+time.Millisecond, windowSize),
			wantStage:    stage_slow,
			wantWaitTime: time.Millisecond * 200,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			rl := NewRateLimiter(slowThreshold)
			for i, d := range test.durations {
				_ = i
				rl.TrackNewDuration(d)
			}
			if rl.currentStage != test.wantStage {
				t.Errorf("state mismatch. got %s want %s", rl.currentStage, test.wantStage)
			}
		})
	}
}
