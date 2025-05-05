package main

import (
	// "fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// db to store data
var m sync.Map
var sh sync.Map

// date to serve data
var hours []int = []int{-24, -48}
var shHours []int = []int{0, -24}
var month []int = []int{0, -1, -2}

func main() {

	logger := setupLogger()
	// replace global logger
	log.Logger = logger

	// Create a new Gin router
	router := gin.New()
	// Add middleware
	router.Use(loggerMiddleware())

	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "success"})
	})

	//Beijing data API
	v1 := router.Group("/v1")
	{
		// Define routes
		v1.GET("/daily_house", dailyHouse)
		v1.GET("/month_house", monthHouse)
		v1.POST("/add_daily_house", addDailyHouse)
		v1.POST("/force_house", forceAddHouse)
	}
	// shanghai data API
	v2 := router.Group("/v2/sh")
	{
		// Define routes
		v2.GET("/new_daily_house", shNewDailyHouse)
		v2.GET("/old_daily_house", shOldDailyHouse)
		v2.POST("/add_new_daily_house", addShNewDailyHouse)
		v2.POST("/add_old_daily_house", addShOldDailyHouse)
	}

	// Run the server
	router.Run(":8080")
}

func dailyHouse(c *gin.Context) {
	for _, h := range hours {
		previousDate := getPreviousDay(-h)
		if v, ok := m.Load(previousDate); ok {
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "daily data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

func monthHouse(c *gin.Context) {
	for _, mon := range month {
		previousDate := getPreviousMonth(mon)
		if v, ok := m.Load(previousDate); ok {
			c.JSON(http.StatusOK, v)
			return
		}
	}
	log.Logger.Error().Str("msg", "month data not found").Msg("Data not found")
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

// AddDailyHouse add daily house data
func addDailyHouse(c *gin.Context) {
	var req DailyHouse
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		log.Logger.Error().Err(err).Msg("Failed to bind JSON")
		return
	}
	if _, ok := m.Load(req.Day); !ok {
		m.Store(req.Day, DailyHouseResp{
			Day:       req.Day,
			DailyData: req.DailyData,
		})
	}

	if _, ok := m.Load(req.Month); !ok {
		m.Store(req.Month, MonthHouseResp{
			Month:     req.Month,
			MonthData: req.MonthData,
		})
	}
	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")

	c.JSON(http.StatusOK, req)
}

// AddDailyHouse add daily house data
func forceAddHouse(c *gin.Context) {
	var req DailyHouse
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	key, ok := c.GetQuery("key")
	if !ok || key != "huan_house" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid key"})
		return
	}
	m.Store(req.Day, DailyHouseResp{
		Day:       req.Day,
		DailyData: req.DailyData,
	})
	m.Store(req.Month, MonthHouseResp{
		Month:     req.Month,
		MonthData: req.MonthData,
	})
	c.JSON(http.StatusOK, req)
}

// Middleware
func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Log after request is processed
		duration := time.Since(start)
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path

		// You would typically log this to a file or logging service
		// For this example, we'll just print it
		log.Logger.Debug().Str("method", method).Str("path", path).Int("status", status).Str("duration", duration.String()).Msg("Request processed")
	}
}
