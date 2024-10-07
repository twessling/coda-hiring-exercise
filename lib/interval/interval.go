package interval

import "fmt"

type Interval struct {
	min          float64
	minInclusive bool
	max          float64
	maxInclusive bool
}

func (i *Interval) Contains(f float64) bool {
	if f < i.min || f > i.max {
		return false
	}
	if f > i.min && f < i.max {
		return true
	}
	if f == i.min && i.minInclusive {
		return true
	}
	if f == i.max && i.maxInclusive {
		return true
	}
	return false
}

func New(min float64, minInc bool, max float64, maxInc bool) *Interval {
	return &Interval{min: min, minInclusive: minInc, max: max, maxInclusive: maxInc}
}

func (i *Interval) String() string {
	prefix := "["
	postfix := "]"
	if i.minInclusive {
		prefix = "("
	}
	if i.maxInclusive {
		postfix = ")"
	}
	return fmt.Sprintf("%s%.4f,%.4f%s", prefix, i.min, i.max, postfix)
}
