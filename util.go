package main

import (
	"time"
)

func getPreviousDay(hours int) string {

	now := time.Now()

	// Subtract 24 hours to get previous day
	yesterday := now.Add(time.Duration(-hours) * time.Hour)

	// fmt.Println("Yesterday was:", yesterday.Format("2006-1-02"))
	return yesterday.Format("2006-01-02")
}

// func getStartEndTimestamp(hours int) (int64, int64) {

// 	now := time.Now()

// 	// Subtract {hours} hours to get previous day
// 	previousDay := now.Add(time.Duration(hours) * time.Hour)

// 	return previousDay.Unix(), now.Unix()
// }

func getPreviousMonth(month int) string {

	now := time.Now()
	if month == 0 {
		return now.Format("2006-01")
	}

	// 上个月的第一天
	firstOfLastMonth := time.Date(now.Year(), now.Month()+time.Month(month), 1, 0, 0, 0, 0, now.Location())

	// 上个月的最后一天
	lastOfLastMonth := firstOfLastMonth.AddDate(0, 1, -1)

	// fmt.Println("Yesterday was:", lastOfLastMonth.Format("2006-1-02"))
	return lastOfLastMonth.Format("2006-01")
}

func getTodayDay() string {
	now := time.Now()
	return now.Format("2006-01-02")
}

func getPreviousHour(hours int) string {

	now := time.Now()

	// Subtract 24 hours to get previous day
	yesterday := now.Add(time.Duration(hours) * time.Hour)

	// fmt.Println("Yesterday was:", yesterday.Format("2006-1-02"))
	return yesterday.Format("2006-01-02-15")
}

func parseDay(day string) (time.Time, error) {
	format := "2006-01-02"
	if len(day) == 13 { // if it includes hour}
		format = "2006-01-02-15"
	}
	return time.Parse(format, day)
}
