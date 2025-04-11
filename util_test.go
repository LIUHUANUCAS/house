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
