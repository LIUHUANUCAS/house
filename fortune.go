package main

import (
	// "fmt"

	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func dailyFortune(c *gin.Context) {
	db := GetInMemDataAccessor(fortune)
	// Try to get data from Redis first
	for _, day := range []string{getTodayDay(), getPreviousDay(24)} {
		poem, found, err := GetFortuneData(ctx, day)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", day).Msg("Error getting fortune data from Redis")
		} else if found {
			c.JSON(http.StatusOK, poem)
			return
		}

		// Fallback to in-memory if Redis fails or data not found in Redis
		if v, ok := db.Load(day); ok {
			// Store in Redis for future use
			poem, ok := v.(Poem)
			if ok {
				go func(day string, data Poem) {
					if err := StoreFortuneData(ctx, day, data); err != nil {
						log.Logger.Error().Err(err).Str("day", day).Msg("Failed to store fortune data in Redis")
					}
				}(day, poem)
			}
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "daily data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

// addDailyFortune add daily fortune data
func addDailyFortune(c *gin.Context) {
	var req Poem
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Logger.Error().Err(err).Msg("Failed to bind JSON")
		return
	}
	log.Logger.Debug().Any("fortune", req).Msg("add data")

	// Store in Redis
	forceUpdate := false
	if v, ok := c.GetQuery("force"); ok && v == "fortune" {
		forceUpdate = true
	}

	// Check if we should store the data
	shouldStore := forceUpdate
	if !shouldStore {
		// Check if data already exists in Redis
		_, found, err := GetFortuneData(ctx, req.Day)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", req.Day).Msg("Error checking if fortune data exists in Redis")
			shouldStore = true // Store anyway if we can't check
		} else {
			shouldStore = !found // Store if not found
		}
	}

	if shouldStore {
		// Store in Redis
		if err := StoreFortuneData(ctx, req.Day, req); err != nil {
			log.Logger.Error().Err(err).Str("day", req.Day).Msg("Failed to store fortune data in Redis")
		}
	}

	// Also store in memory for backward compatibility
	db := GetInMemDataAccessor(fortune)
	_, ok := db.Load(req.Day)
	if forceUpdate || !ok {
		db.Store(req.Day, req)
	}

	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")
	c.JSON(http.StatusOK, req)
}
