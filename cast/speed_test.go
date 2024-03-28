package cast_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/wormbks/asciinema-edit/cast"
)

func TestSpeed_Validation(t *testing.T) {
	var data *cast.Cast

	t.Run("With nil cast", func(t *testing.T) {
		err := cast.Speed(nil, 1, 2, 3)
		assert.Error(t, err)
	})

	t.Run("With an empty event stream", func(t *testing.T) {
		data = &cast.Cast{
			EventStream: []*cast.Event{},
		}
		err := cast.Speed(data, 1, 1, 2)
		assert.Error(t, err)
	})

	t.Run("With unusual factors", func(t *testing.T) {
		tests := []struct {
			factor float64
		}{
			{12},
			{0.05},
		}

		for _, test := range tests {
			t.Run("Factor "+fmt.Sprint(test.factor), func(t *testing.T) {
				err := cast.Speed(data, test.factor, 2, 3)
				assert.Error(t, err)
			})
		}
	})

	t.Run("With invalid ranges", func(t *testing.T) {
		tests := []struct {
			from, to float64
		}{
			{2, 2},
			{10, 2},
		}

		for _, test := range tests {
			t.Run(fmt.Sprintf("From %f to %f", test.from, test.to), func(t *testing.T) {
				err := cast.Speed(data, 1, test.from, test.to)
				assert.Error(t, err)
			})
		}
	})
}

func setup() *cast.Cast {
	event1 := &cast.Event{
		Time: 1,
		Data: "event1",
	}
	event2 := &cast.Event{
		Time: 2,
		Data: "event2",
	}
	event3 := &cast.Event{
		Time: 3,
		Data: "event3",
	}
	event4 := &cast.Event{
		Time: 4,
		Data: "event4",
	}

	data := &cast.Cast{
		EventStream: []*cast.Event{
			event1,
			event2,
			event3,
			event4,
		},
	}
	return data
}

func TestSpeed_NonEmptyStream(t *testing.T) {

	t.Run("With `from` not found", func(t *testing.T) {
		data := setup()
		err := cast.Speed(data, 2, 1.3, 2)
		assert.Error(t, err)
	})

	t.Run("With `to` not found", func(t *testing.T) {
		data := setup()
		err := cast.Speed(data, 2, 2, 3.3)
		assert.Error(t, err)
	})
}
func TestSpeed_Slowdown(t *testing.T) {
	t.Run("In a small range", func(t *testing.T) {
		data := setup()
		err := cast.Speed(data, 3, 1, 2)
		assert.NoError(t, err)

		assert.Equal(t, float64(1), data.EventStream[0].Time)
		assert.Equal(t, float64(4), data.EventStream[1].Time)
		assert.Equal(t, float64(5), data.EventStream[2].Time)
		assert.Equal(t, float64(6), data.EventStream[3].Time)
	})

	t.Run("In the whole set", func(t *testing.T) {
		data := setup()
		err := cast.Speed(data, 3, 1, 4)
		assert.NoError(t, err)

		assert.Equal(t, float64(1), data.EventStream[0].Time)
		assert.Equal(t, float64(4), data.EventStream[1].Time)
		assert.Equal(t, float64(7), data.EventStream[2].Time)
		assert.Equal(t, float64(10), data.EventStream[3].Time)
	})
}

func TestSpeed_SpeedUp(t *testing.T) {

	data := setup()
	err := cast.Speed(data, 0.5, 1, 2)
	assert.NoError(t, err)

	assert.Equal(t, float64(1), data.EventStream[0].Time)
	assert.Equal(t, float64(1.5), data.EventStream[1].Time)
	assert.Equal(t, float64(2.5), data.EventStream[2].Time)
	assert.Equal(t, float64(3.5), data.EventStream[3].Time)

}
