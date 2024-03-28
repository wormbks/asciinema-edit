package commands

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuantizeRange(t *testing.T) {
	t.Run("Empty input", func(t *testing.T) {
		input := ""
		_, err := ParseQuantizeRange(input)
		assert.Error(t, err)
	})

	t.Run("Invalid characters", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"Non-decimal characters", "1a"},
			{"Starts with non-numeric character", "a"},
			{"Ends with non-numeric character", "1,a"},
			{"Leading comma", ",1"},
			{"Trailing comma", "1,"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				_, err := ParseQuantizeRange(test.input)
				assert.Error(t, err)
			})
		}
	})

	t.Run("Valid Unbounded", func(t *testing.T) {
		input := "1.2"
		qRange, err := ParseQuantizeRange(input)
		assert.NoError(t, err)
		assert.Equal(t, float64(1.2), qRange.From)
		assert.Equal(t, math.MaxFloat64, qRange.To)

		input = "-1.2"
		_, err = ParseQuantizeRange(input)
		assert.Error(t, err)
	})

	t.Run("Valid Bounded", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"From greater than to", "2,1"},
			{"From equal to to", "2,2"},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				_, err := ParseQuantizeRange(test.input)
				assert.Error(t, err)
			})
		}
	})
}
