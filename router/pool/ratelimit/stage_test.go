package ratelimit

import (
	"testing"
	"time"
)

func TestCalculateSlowNewWaitTime(t *testing.T) {
	tests := map[string]struct {
		oldStage        waitTimeCalculator
		oldWaitTime     time.Duration
		newScore        float64
		newScoreLast10  float64
		wantNewDuration time.Duration
	}{
		"old stage OK": {
			oldStage:        stage_ok,
			wantNewDuration: stage_slow.defaultWaittime,
		},
		"old stage DEAD": {
			oldStage:        stage_dead,
			wantNewDuration: time.Second,
		},
		"old stage SLOW from no_wait": {
			oldStage:        stage_slow,
			oldWaitTime:     no_wait,
			wantNewDuration: stage_slow.defaultWaittime,
		},
		"old stage SLOW improving hard": {
			oldStage:        stage_slow,
			oldWaitTime:     stage_slow.defaultWaittime,
			newScore:        0.2,
			newScoreLast10:  1,
			wantNewDuration: stage_slow.defaultWaittime / 2,
		},
		"old stage SLOW degrading hard": {
			oldStage:        stage_slow,
			oldWaitTime:     stage_slow.defaultWaittime,
			newScore:        0.8,
			newScoreLast10:  0.2,
			wantNewDuration: stage_slow.defaultWaittime * 2,
		},
		"old stage SLOW random values": {
			oldStage:        stage_slow,
			oldWaitTime:     stage_slow.defaultWaittime,
			newScore:        0.84,
			newScoreLast10:  0.76,
			wantNewDuration: time.Duration(279.77528 * float64(time.Millisecond)), //279.77528ms
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			newDuration := stage_slow.calculateNewWaitTime(test.oldStage, test.oldWaitTime, test.newScore, test.newScoreLast10)
			if newDuration != test.wantNewDuration {
				t.Errorf("duration mismatch. got %v want %v", newDuration, test.wantNewDuration)
			}
		})
	}
}
