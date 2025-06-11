package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// /v1/daily_new_house
func beijingNewDailyHouse(c *gin.Context) {
	bj := GetInMemDataAccessor(beijing)
	// Try to get data from Redis first
	for _, h := range hours {
		currentDay := getBeijingNewHouseDayKey(getPreviousDay(-h))
		log.Logger.Debug().Str("currentDay", currentDay).Msg("Fetching bj house data for current day")
		houseData, found, err := GetHouseData(ctx, currentDay, beijingKey)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", currentDay).Msg("Error getting house data from Redis")
		} else if found {
			c.JSON(http.StatusOK, houseData)
			return
		}

		// Fallback to in-memory if Redis fails or data not found in Redis
		if v, ok := bj.Load(currentDay); ok {
			// Store in Redis for future use
			dailyResp, ok := v.(DailyHouseResp)
			if ok {
				go func(day string, data DailyHouseResp) {
					if err := StoreHouseData(ctx, day, data, beijingKey); err != nil {
						log.Logger.Error().Err(err).Str("day", day).Msg("Failed to store house data in Redis")
					}
				}(currentDay, dailyResp)
			}
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "bj new house data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

// AddBeijingHouse adds Beijing house data
func addBeijingNewHouse(c *gin.Context) {
	var req DailyHouseResp
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Logger.Error().Err(err).Msg("Failed to bind JSON")
		return
	}

	// Create daily house response
	dailyResp := req

	var dailyInMem bool

	m := GetInMemDataAccessor(beijing)
	// Store in Redis
	if err := StoreHouseData(ctx, req.Day, dailyResp, beijingKey); err != nil {
		log.Logger.Error().Err(err).Str("day", req.Day).Msg("Failed to store house data in Redis")
		dailyInMem = true
	}

	// Also store in memory for backward compatibility
	if dailyInMem {
		if _, ok := m.Load(req.Day); !ok {
			m.Store(req.Day, dailyResp)
		}
	}

	// Create response with both daily and derived monthly data
	response := map[string]interface{}{
		"day":        req.Day,
		"daily_data": req.DailyData,
	}

	log.Logger.Debug().Str("day", req.Day).Msg("Beijing house data added successfully")
	c.JSON(http.StatusOK, response)
}
