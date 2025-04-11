package main

import (
	"fmt"
	"time"
)

func getPreviousDay(hours int) string {

	now := time.Now()

	// Subtract 24 hours to get previous day
	yesterday := now.Add(time.Duration(-hours) * time.Hour)

	fmt.Println("Yesterday was:", yesterday.Format("2006-1-02"))
	return yesterday.Format("2006-01-02")
}
