package cast

import (
	"github.com/pkg/errors"
)

// Speed updates the cast speed by multiplying all of the
// timestamps in a given range by a given factor.
func Speed(c *Cast, factor, from, to float64) error {
	if c == nil {
		return errors.Errorf("cast must not be nil")
	}

	if len(c.EventStream) == 0 {
		return errors.Errorf("event stream must be nonempty")
	}

	if factor > 10 || factor < 0.1 {
		return errors.Errorf("factor must be within 0.1 and 10 range")
	}

	if from >= to {
		return errors.Errorf("`from` must not be greater or equal than `to`")
	}

	fromIdx := -1
	toIdx := -1

	for idx, ev := range c.EventStream {
		if ev.Time == from {
			fromIdx = idx
		}

		if ev.Time == to {
			toIdx = idx
		}
	}

	if fromIdx == -1 {
		return errors.Errorf("couldn't find initial frame")
	}

	if toIdx == -1 {
		return errors.Errorf("couldn't find final frame")
	}

	var (
		i                int
		k                int
		delta            float64
		newDelta         float64
		accumulatedDelta float64
		deltas           = make([]float64, toIdx-fromIdx)
	)

	k = 0
	for i = fromIdx; i < toIdx; i++ {
		delta = c.EventStream[i+1].Time - c.EventStream[i].Time
		newDelta = delta * factor
		accumulatedDelta += (newDelta - delta)

		deltas[k] = newDelta
		k++
	}

	k = 0
	for i = fromIdx; i < toIdx; i++ {
		c.EventStream[i+1].Time = c.EventStream[i].Time + deltas[k]
		k++
	}

	if toIdx+1 < len(c.EventStream) {
		for _, remainingElem := range c.EventStream[toIdx+1:] {
			remainingElem.Time += accumulatedDelta
		}
	}

	return nil
}
