package main

import (
	"sync"
)

func getDB(factory string) *sync.Map {

	switch factory {
	case "daily":
		return &m
	case "sh":
		return &sh
	case "fortune":
		return &fortune
	}

	return &m
}
