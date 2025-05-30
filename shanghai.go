package main

import (
	// "fmt"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func shNewDailyHouse(c *gin.Context) {
	sh := GetInMemDataAccessor(shanghai)
	// Try to get data from Redis first
	for _, h := range []int{-current, -previous, -prePrevious} {
		currentDay := getPreviousHour(h)
		log.Logger.Debug().Str("currentDay", currentDay).Msg("Fetching sh house data for current day")
		houseData, found, err := GetHouseData(ctx, currentDay, shanghaiKey)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", currentDay).Msg("Error getting house data from Redis")
		} else if found {
			c.JSON(http.StatusOK, houseData)
			return
		}

		// Fallback to in-memory if Redis fails or data not found in Redis
		if v, ok := sh.Load(currentDay); ok {
			// Store in Redis for future use
			dailyResp, ok := v.(DailyHouseResp)
			if ok {
				go func(day string, data DailyHouseResp) {
					if err := StoreHouseData(ctx, day, data, shanghaiKey); err != nil {
						log.Logger.Error().Err(err).Str("day", day).Msg("Failed to store house data in Redis")
					}
				}(currentDay, dailyResp)
			}
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "sh new house data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

func shOldDailyHouse(c *gin.Context) {
	sh := GetInMemDataAccessor(shanghai)
	// Try to get data from Redis first
	for _, h := range hours {
		previousDate := getPreviousDay(-h)
		houseData, found, err := GetHouseData(ctx, previousDate, shanghaiKey)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", previousDate).Msg("Error getting house data from Redis")
		} else if found {
			c.JSON(http.StatusOK, houseData)
			return
		}

		// Fallback to in-memory if Redis fails or data not found in Redis
		if v, ok := sh.Load(previousDate); ok {
			// Store in Redis for future use
			dailyResp, ok := v.(DailyHouseResp)
			if ok {
				go func(day string, data DailyHouseResp) {
					if err := StoreHouseData(ctx, day, data, shanghaiKey); err != nil {
						log.Logger.Error().Err(err).Str("day", day).Msg("Failed to store house data in Redis")
					}
				}(previousDate, dailyResp)
			}
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "daily data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

// addShNewDailyHouse add daily house data
func addShNewDailyHouse(c *gin.Context) {
	var req DailyHouse
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Logger.Error().Err(err).Msg("Failed to bind JSON")
		return
	}
	fmt.Println("req", req)
	log.Logger.Debug().Any("sh-data", req).Msg("add data")

	// Create daily house response
	dailyResp := DailyHouseResp{
		Day: req.Day,
		DailyData: DailyData{
			TotalCount: req.DailyData.HouseCount,
			TotalArea:  req.DailyData.HouseArea,
			HouseCount: req.DailyData.HouseCount,
			HouseArea:  req.DailyData.HouseArea,
		},
	}

	// Store in Redis
	if err := StoreHouseData(ctx, req.Day, dailyResp, shanghaiKey); err != nil {
		log.Logger.Error().Err(err).Str("day", req.Day).Msg("Failed to store house data in Redis")
	}
	sh := GetInMemDataAccessor(shanghai)
	// Also store in memory for backward compatibility
	if _, ok := sh.Load(req.Day); !ok {
		sh.Store(req.Day, dailyResp)
	}

	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")
	c.JSON(http.StatusOK, req)
}

// addShOldDailyHouse add daily house data
func addShOldDailyHouse(c *gin.Context) {
	var req DailyHouse
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Logger.Error().Err(err).Msg("Failed to bind JSON")
		return
	}

	// Create daily house response
	dailyResp := DailyHouseResp{
		Day: req.Day,
		DailyData: DailyData{
			TotalCount: req.DailyData.HouseCount,
			TotalArea:  req.DailyData.HouseArea,
			HouseCount: req.DailyData.HouseCount,
			HouseArea:  req.DailyData.HouseArea,
			HousePrice: req.DailyData.HousePrice,
			TotalPrice: req.DailyData.HousePrice,
		},
	}

	// Store in Redis
	if err := StoreHouseData(ctx, req.Day, dailyResp, shanghaiKey); err != nil {
		log.Logger.Error().Err(err).Str("day", req.Day).Msg("Failed to store house data in Redis")
	}
	sh := GetInMemDataAccessor(shanghai)
	// Also store in memory for backward compatibility
	if _, ok := sh.Load(req.Day); !ok {
		sh.Store(req.Day, dailyResp)
	}

	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")
	c.JSON(http.StatusOK, req)
}
