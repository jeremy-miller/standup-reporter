/*
Package configuration handles standup-reporter configuration.
*/
package configuration

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

/*
Configuration defines the configuration parameters of the standup-reporter.
*/
type Configuration struct {
	AuthHeader    string          // Authentication token used to authenticate to Asana.
	Client        http.Client     // HTTP client which will be reused for all requests.
	TodayMidnight time.Time       // Today's date at midnight in the local timezone.
	EarliestDate  string          // Midnight of the day for which Asana tasks will be retrieved.
	WG            *sync.WaitGroup // WaitGroup used to coordinate goroutines.
}

/*
Get returns the current configuration of the standup-reporter.
*/
func Get(days int, asanaToken string) *Configuration {
	if days == 0 {
		days = calculateDays()
	}
	t := time.Now().Local()
	todayMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	return &Configuration{
		AuthHeader: fmt.Sprintf("Bearer %s", asanaToken),
		Client: http.Client{
			Timeout: time.Second * 10,
		},
		TodayMidnight: todayMidnight,
		EarliestDate:  todayMidnight.AddDate(0, 0, -days).Format(time.RFC3339),
		WG:            &wg,
	}
}

func calculateDays() int {
	if time.Now().Weekday() == time.Monday { // account for weekend
		return 3
	}
	return 1
}
