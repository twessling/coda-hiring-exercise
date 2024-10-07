package interval

import "testing"

func TestIntervalContains(t *testing.T) {
	tests := map[string]struct {
		i          *Interval
		allowed    []float64
		disallowed []float64
	}{
		"nothing contains": {
			i:          &Interval{2, false, 2, false},
			disallowed: []float64{2, 1.999, 2.00001, 3, 1, 100, -12},
		},
		"only one number contains": {
			i:          &Interval{2, true, 2, true},
			allowed:    []float64{2},
			disallowed: []float64{1.999, 2.00001, 3, 1, 100, -12},
		},
		"range contains": {
			i:          &Interval{2, true, 5, true},
			allowed:    []float64{1, 3, 5, 2.00001, 4.9999},
			disallowed: []float64{1.999, 1, 100, -12, 5.0000001, 0},
		},
		"including left": {
			i:          &Interval{2, true, 5, false},
			allowed:    []float64{2, 3, 2.00001, 4.9999},
			disallowed: []float64{1.999, 5, 1, 100, -12, 5.0000001, 0},
		},
		"including right": {
			i:          &Interval{2, false, 5, true},
			allowed:    []float64{3, 5, 2.00001, 4.9999},
			disallowed: []float64{1.999, 1, 100, -12, 5.0000001, 0},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			for _, n := range test.allowed {
				if !test.i.Contains(n) {
					t.Errorf("interval %s does not contain %f", test.i, n)
				}
			}
			for _, n := range test.disallowed {
				if test.i.Contains(n) {
					t.Errorf("interval %s contains %f", test.i, n)
				}
			}
		})
	}
}
