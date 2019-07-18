package configuration

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateDays(t *testing.T) {
	weekday, _ := time.Parse(time.RFC3339, "2019-07-18T15:00:00Z") //nolint:errcheck
	monday, _ := time.Parse(time.RFC3339, "2019-07-15T15:00:00Z")  //nolint:errcheck
	testCases := []struct {
		name     string
		time     time.Time
		expected int
	}{
		{name: "Weekday", time: weekday, expected: 1},
		{name: "Monday", time: monday, expected: 3},
	}
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := calculateDays(tc.time)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
