package interval

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
