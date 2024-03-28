package cast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wormbks/asciinema-edit/cast"
)

func TestCut_Validation(t *testing.T) {
	data := &cast.Cast{
		EventStream: []*cast.Event{},
	}

	t.Run("With nil cast", func(t *testing.T) {
		err := cast.Cut(nil, 1, 2)
		assert.Error(t, err)
	})

	t.Run("With an empty event stream", func(t *testing.T) {
		err := cast.Cut(data, 1, 2)
		assert.Error(t, err)
	})

	t.Run("With `from` > `to`", func(t *testing.T) {
		err := cast.Cut(data, 3, 2)
		assert.Error(t, err)
	})
}

func TestCut_Stream(t *testing.T) {
	t.Run("With non-empty event stream", func(t *testing.T) {
		var (
			err                                error
			data                               *cast.Cast
			initialNumberOfEvents              int
			event1, event1_2, event1_6, event2 *cast.Event
		)

		setup := func() {
			event1 = &cast.Event{
				Time: 1,
				Data: "event1",
			}
			event1_2 = &cast.Event{
				Time: 1.2,
				Data: "event1_2",
			}
			event1_6 = &cast.Event{
				Time: 1.6,
				Data: "event1_6",
			}
			event2 = &cast.Event{
				Time: 2,
				Data: "event2",
			}

			data = &cast.Cast{
				EventStream: []*cast.Event{
					event1,
					event1_2,
					event1_6,
					event2,
				},
			}

			initialNumberOfEvents = len(data.EventStream)
		}

		t.Run("With `from` not found", func(t *testing.T) {
			setup()
			err = cast.Cut(data, 1.1, 2)
			assert.Error(t, err)
		})

		t.Run("With `to` not found", func(t *testing.T) {
			setup()
			err = cast.Cut(data, 2, 3.3)
			assert.Error(t, err)
		})

		t.Run("Cutting a single frame when `from` == `to`", func(t *testing.T) {
			setup()
			err = cast.Cut(data, 1.2, 1.2)
			assert.NoError(t, err)

			assert.Contains(t, data.EventStream, event1)
			assert.NotContains(t, data.EventStream, event1_2)
			assert.Contains(t, data.EventStream, event1_6)
			assert.Contains(t, data.EventStream, event2)

			assert.Len(t, data.EventStream, initialNumberOfEvents-1)

			assert.Equal(t, float64(1), event1.Time)
			assert.Equal(t, float64(1.2), event1_6.Time)
			assert.Equal(t, float64(1.6), event2.Time)
		})

		t.Run("Cutting range without bounds included", func(t *testing.T) {
			setup()
			err = cast.Cut(data, 1.2, 1.6)
			assert.NoError(t, err)

			assert.Contains(t, data.EventStream, event1)
			assert.NotContains(t, data.EventStream, event1_2)
			assert.NotContains(t, data.EventStream, event1_6)
			assert.Contains(t, data.EventStream, event2)

			assert.Len(t, data.EventStream, initialNumberOfEvents-2)

			assert.Equal(t, float64(1), event1.Time)
			assert.Equal(t, float64(1.2), event2.Time)
		})

		t.Run("Cuts frames in range containing last element", func(t *testing.T) {
			setup()
			err = cast.Cut(data, 1.2, 2)
			assert.NoError(t, err)

			assert.Contains(t, data.EventStream, event1)
			assert.NotContains(t, data.EventStream, event1_2)
			assert.NotContains(t, data.EventStream, event1_6)
			assert.NotContains(t, data.EventStream, event2)

			assert.Len(t, data.EventStream, initialNumberOfEvents-3)
		})
	})
}
