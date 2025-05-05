package main

import (
	// "fmt"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func shNewDailyHouse(c *gin.Context) {
	for _, h := range shHours {
		currentDay := getPreviousDay(-h)
		fmt.Println("current day", currentDay)
		if v, ok := sh.Load(currentDay); ok {
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "daily data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

func shOldDailyHouse(c *gin.Context) {
	for _, h := range hours {
		previousDate := getPreviousDay(-h)
		if v, ok := sh.Load(previousDate); ok {
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
	// "todaySign_area":14360.1,"todaySign_ts":203
	if _, ok := sh.Load(req.Day); !ok {
		sh.Store(req.Day, DailyHouseResp{
			Day: req.Day,
			DailyData: DailyData{
				TotalCount: req.DailyData.HouseCount,
				TotalArea:  req.DailyData.HouseArea,
				HouseCount: req.DailyData.HouseCount,
				HouseArea:  req.DailyData.HouseArea,
			},
		})
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
	if _, ok := sh.Load(req.Day); !ok {
		sh.Store(req.Day, DailyHouseResp{
			Day: req.Day,
			DailyData: DailyData{
				TotalCount: req.DailyData.HouseCount,
				TotalArea:  req.DailyData.HouseArea,
				HouseCount: req.DailyData.HouseCount,
				HouseArea:  req.DailyData.HouseArea,
				HousePrice: req.DailyData.HousePrice,
				TotalPrice: req.DailyData.HousePrice,
			},
		})
	}

	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")

	c.JSON(http.StatusOK, req)
}
