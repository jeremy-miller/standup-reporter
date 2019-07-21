package configuration_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jeremy-miller/standup-reporter/internal/configuration"
)

func TestGet1Day(t *testing.T) {
	assert := assert.New(t)
	const days = 1
	config := configuration.Get(days)
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	assert.Equal(midnight, config.TodayMidnight)
	assert.Equal(midnight.AddDate(0, 0, -days).Format(time.RFC3339), config.EarliestDate)
	assert.IsType(&sync.WaitGroup{}, config.WG)
}

func TestGet0Day(t *testing.T) {
	assert := assert.New(t)
	const days = 0
	config := configuration.Get(days)
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	assert.Equal(midnight, config.TodayMidnight)
	assert.Equal(midnight.AddDate(0, 0, -1).Format(time.RFC3339), config.EarliestDate)
	assert.IsType(&sync.WaitGroup{}, config.WG)
}
