package cast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wormbks/asciinema-edit/cast"
)

func TestQuantize(t *testing.T) {
	t.Run("Parameter validation", func(t *testing.T) {
		var data *cast.Cast

		setup := func() {
			data = &cast.Cast{
				EventStream: []*cast.Event{
					{},
				},
			}
		}

		t.Run("With nil cast", func(t *testing.T) {
			setup()
			err := cast.Quantize(nil, nil)
			assert.Error(t, err)
		})

		t.Run("With an empty event stream", func(t *testing.T) {
			setup()
			err := cast.Quantize(data, nil)
			assert.Error(t, err)
		})

		t.Run("With a nil range list", func(t *testing.T) {
			setup()
			err := cast.Quantize(data, nil)
			assert.Error(t, err)
		})

		t.Run("With an empty range list", func(t *testing.T) {
			setup()
			err := cast.Quantize(data, []cast.QuantizeRange{})
			assert.Error(t, err)
		})
	})

	t.Run("RangeOverlaps", func(t *testing.T) {
		qRange := &cast.QuantizeRange{
			From: 1,
			To:   2,
		}

		t.Run("Doesn't overlap if no in another range", func(t *testing.T) {
			assert.False(t, qRange.RangeOverlaps(cast.QuantizeRange{
				From: 30,
				To:   40,
			}))
		})

		t.Run("Overlaps if from in another range", func(t *testing.T) {
			assert.True(t, qRange.RangeOverlaps(cast.QuantizeRange{
				From: 1.5,
				To:   3,
			}))
		})

		t.Run("Overlaps if to in another range", func(t *testing.T) {
			assert.True(t, qRange.RangeOverlaps(cast.QuantizeRange{
				From: 0.9,
				To:   1.5,
			}))
		})
	})

	t.Run("InRange", func(t *testing.T) {
		qRange := &cast.QuantizeRange{
			From: 1,
			To:   2,
		}

		t.Run("In range if `from <= x < to`", func(t *testing.T) {
			assert.True(t, qRange.InRange(1.5))
		})

		t.Run("In range if `x == from`", func(t *testing.T) {
			assert.True(t, qRange.InRange(1))
		})

		t.Run("Not in range if `x == to`", func(t *testing.T) {
			assert.False(t, qRange.InRange(2))
		})

		t.Run("Not in range if `x > to`", func(t *testing.T) {
			assert.False(t, qRange.InRange(2.1))
		})

		t.Run("Not in range if `x < from`", func(t *testing.T) {
			assert.False(t, qRange.InRange(0.9))
		})
	})

	t.Run("Having ranges specified", func(t *testing.T) {
		var (
			data                     *cast.Cast
			event1, event2, event5   *cast.Event
			event9, event10, event11 *cast.Event
			err                      error
		)

		setup := func() {
			event1 = &cast.Event{Time: 1}
			event2 = &cast.Event{Time: 2}
			event5 = &cast.Event{Time: 5}
			event9 = &cast.Event{Time: 9}
			event10 = &cast.Event{Time: 10}
			event11 = &cast.Event{Time: 11}

			data = &cast.Cast{
				EventStream: []*cast.Event{
					event1,
					event2,
					event5,
					event9,
					event10,
					event11,
				},
			}
		}

		t.Run("Cuts down delays with a single range", func(t *testing.T) {
			setup()
			ranges := []cast.QuantizeRange{{2, 6}}
			err = cast.Quantize(data, ranges)
			assert.NoError(t, err)

			assert.Equal(t, float64(1), event1.Time)
			assert.Equal(t, float64(2), event2.Time)
			assert.Equal(t, float64(4), event5.Time)
			assert.Equal(t, float64(6), event9.Time)
			assert.Equal(t, float64(7), event10.Time)
			assert.Equal(t, float64(8), event11.Time)
		})
	})
}
