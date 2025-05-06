package main

import (
	"encoding/json"
	"testing"
)

func TestGetPreviousDay(t *testing.T) {
	out := getPreviousDay(48)
	t.Logf("previous:%s\n", out)
	bye, _ := json.Marshal(getDefaultDailyHouse())
	t.Logf("%s", bye)
}

func TestGetPreviousMonth(t *testing.T) {
	out := getPreviousMonth(-2)
	t.Logf("previous:%s\n", out)
}

func TestGetPreviousHour(t *testing.T) {
	hours := []int{0, -1, -2}
	for _, hour := range hours {
		out := getPreviousHour(hour)
		t.Logf("previous:%s\n", out)
	}
}
