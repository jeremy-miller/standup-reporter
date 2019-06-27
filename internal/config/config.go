package config

import (
	"fmt"
	"net/http"
	"time"
)

type Config struct {
	AuthHeader    string
	Client        http.Client
	TodayMidnight time.Time
	EarliestDate  string
}

func Get(days int, asanaToken string) *Config {
	if days == 0 {
		days = calculateDays()
	}
	t := time.Now().Local()
	todayMidnight := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return &Config{
		AuthHeader:    fmt.Sprintf("Bearer %s", asanaToken),
		Client:        http.Client{},
		TodayMidnight: todayMidnight,
		EarliestDate:  todayMidnight.AddDate(0, 0, -days).Format(time.RFC3339),
	}
}

func calculateDays() int {
	if time.Now().Weekday() == time.Monday { // account for weekend
		return 3
	}
	return 1
}
