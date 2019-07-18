package configuration_test

import (
	"fmt"
	"github.com/jeremy-miller/standup-reporter/internal/configuration"
	"github.com/stretchr/testify/assert"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestGet1Day(t *testing.T) {
	assert := assert.New(t)
	const days = 1
	const asanaToken = "123abc"
	config := configuration.Get(days, asanaToken)
	assert.Equal(fmt.Sprintf("Bearer %s", asanaToken), config.AuthHeader)
	assert.IsType(http.Client{}, config.Client)
	assert.Equal(time.Second * 10, config.Client.Timeout)
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	assert.Equal(midnight, config.TodayMidnight)
	assert.Equal(midnight.AddDate(0, 0, -days).Format(time.RFC3339), config.EarliestDate)
	assert.IsType(&sync.WaitGroup{}, config.WG)
}

func TestGet0Day(t *testing.T) {
	assert := assert.New(t)
	const days = 0
	const asanaToken = "123abc"
	config := configuration.Get(days, asanaToken)
	assert.Equal(fmt.Sprintf("Bearer %s", asanaToken), config.AuthHeader)
	assert.IsType(http.Client{}, config.Client)
	assert.Equal(time.Second * 10, config.Client.Timeout)
	now := time.Now().Local()
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	assert.Equal(midnight, config.TodayMidnight)
	assert.Equal(midnight.AddDate(0, 0, -1).Format(time.RFC3339), config.EarliestDate)
	assert.IsType(&sync.WaitGroup{}, config.WG)
}
