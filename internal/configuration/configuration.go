/*
Package configuration handles shared standup-reporter configuration.
*/
package configuration

import (
	"sync"
	"time"
)

/*
Configuration defines the shared configuration parameters of the standup-reporter.
*/
type Configuration struct {
	TodayMidnight time.Time       // Today's date at midnight in the local timezone.
	EarliestDate  string          // Midnight of the day for which Asana tasks will be retrieved.
	WG            *sync.WaitGroup // WaitGroup used to coordinate goroutines.
}

/*
Get returns the current configuration of the standup-reporter.
*/
func Get(days int) *Configuration {
	t := time.Now().Local()
	if days == 0 {
		days = calculateDays(t)
	}
	todayMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	var wg sync.WaitGroup
	return &Configuration{
		TodayMidnight: todayMidnight,
		EarliestDate:  todayMidnight.AddDate(0, 0, -days).Format(time.RFC3339),
		WG:            &wg,
	}
}

func calculateDays(t time.Time) int {
	if t.Weekday() == time.Monday { // account for weekend
		return 3
	}
	return 1
}
