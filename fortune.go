package main

import (
	// "fmt"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func dailyFortune(c *gin.Context) {
	db := getDB("fortune")
	for _, day := range []string{getTodayDay(), getPreviousDay(24)} {
		if v, ok := db.Load(day); ok {
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "daily data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

// addDailyFortune add daily house data
func addDailyFortune(c *gin.Context) {
	var req Poem
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Logger.Error().Err(err).Msg("Failed to bind JSON")
		return
	}
	fmt.Println("req", req)
	log.Logger.Debug().Any("fortune", req).Msg("add data")
	db := getDB("fortune")
	if _, ok := db.Load(req.Day); !ok {
		db.Store(req.Day, req)
	}
	if v, ok := c.GetQuery("force"); ok && v == "fortune" {
		db.Store(req.Day, req)
	}
	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")
	c.JSON(http.StatusOK, req)
}
