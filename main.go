package main

import (
	// "fmt"
	"net/http"
	"time"

	"github.com/LIUHUANUCAS/house/config"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// date to serve data
var (
	hours   []int = []int{-24 * previousDay, -24 * prePreviousDay}
	shHours []int = []int{today, -24 * previousDay}
)

var monthScope []int = []int{currentMonth, -previousMonth, -prePreviousMonth}

func main() {

	logger := setupLogger()
	// replace global logger
	log.Logger = logger
	cfg := config.GetConfig()
	InitInMemoryDB()

	// Initialize Redis
	redisDB = InitRedis(ctx, &cfg.RedisConfig)

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
		v1.GET("/daily_new_house", beijingNewDailyHouse)
		v1.GET("/month_house", monthHouse)
		v1.POST("/add_daily_house", addDailyHouse)
		v1.POST("/add_beijing_new_house", addBeijingNewHouse)
		v1.POST("/force_house", forceAddHouse)

		// Time-based retrieval endpoints
		v1.GET("/house_period/:days", getHousePeriod)
	}
	// shanghai data API
	v2 := router.Group("/v2/sh")
	{
		// Define routes
		v2.GET("/new_daily_house", shNewDailyHouse)
		v2.GET("/old_daily_house", shOldDailyHouse)
		v2.POST("/add_new_daily_house", addShNewDailyHouse)
		v2.POST("/add_old_daily_house", addShOldDailyHouse)

		// Time-based retrieval endpoint
		v2.GET("/house_period/:days", getShHousePeriod)
	}

	v3 := router.Group("/v3/fortune")
	{
		// Define routes
		v3.GET("/daily", dailyFortune)
		v3.POST("/add_daily", addDailyFortune)

	}

	// Run the server
	router.Run(":8080")
}

func dailyHouse(c *gin.Context) {
	// Try to get data from Redis first
	m := GetInMemDataAccessor(beijing)
	for _, h := range hours {
		previousDate := getPreviousDay(-h)
		houseData, found, err := GetHouseData(ctx, previousDate, beijingKey)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", previousDate).Msg("Error getting house data from Redis")
		} else if found {
			c.JSON(http.StatusOK, houseData)
			return
		}

		// Fallback to in-memory if Redis fails or data not found in Redis
		if v, ok := m.Load(previousDate); ok {
			// Store in Redis for future use
			dailyResp, ok := v.(DailyHouseResp)
			if ok {
				go func(day string, data DailyHouseResp) {
					if err := StoreHouseData(ctx, day, data, beijingKey); err != nil {
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

func monthHouse(c *gin.Context) {
	m := GetInMemDataAccessor(beijing)
	// Try to get data from Redis first
	for _, mon := range monthScope {
		previousDate := getPreviousMonth(mon)
		monthData, found, err := GetMonthHouseData(ctx, previousDate, beijingKey)
		if err != nil {
			log.Logger.Error().Err(err).Str("month", previousDate).Msg("Error getting month house data from Redis")
		} else if found {
			c.JSON(http.StatusOK, monthData)
			return
		}

		// Fallback to in-memory if Redis fails or data not found in Redis
		if v, ok := m.Load(previousDate); ok {
			// Store in Redis for future use
			monthResp, ok := v.(MonthHouseResp)
			if ok {
				go func(month string, data MonthHouseResp) {
					if err := StoreMonthHouseData(ctx, month, data, beijingKey); err != nil {
						log.Logger.Error().Err(err).Str("month", month).Msg("Failed to store month house data in Redis")
					}
				}(previousDate, monthResp)
			}
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

	// Create daily house response
	dailyResp := DailyHouseResp{
		Day:       req.Day,
		DailyData: req.DailyData,
	}

	// Create monthly house response
	monthResp := MonthHouseResp{
		Month:     req.Month,
		MonthData: req.MonthData,
	}

	var dailyInMem bool
	var monthInMem bool

	m := GetInMemDataAccessor(beijing)
	// Store in Redis
	if err := StoreHouseData(ctx, req.Day, dailyResp, beijingKey); err != nil {
		log.Logger.Error().Err(err).Str("day", req.Day).Msg("Failed to store house data in Redis")
		dailyInMem = true
	}

	if err := StoreMonthHouseData(ctx, req.Month, monthResp, beijingKey); err != nil {
		log.Logger.Error().Err(err).Str("month", req.Month).Msg("Failed to store month house data in Redis")
		monthInMem = true
	}

	// Also store in memory for backward compatibility
	if dailyInMem {
		if _, ok := m.Load(req.Day); !ok {
			m.Store(req.Day, dailyResp)
		}
	}
	if monthInMem {
		if _, ok := m.Load(req.Month); !ok {
			m.Store(req.Month, monthResp)
		}
	}

	log.Logger.Debug().Str("day", req.Day).Msg("Data added successfully")
	c.JSON(http.StatusOK, req)
}

// ForceAddHouse force add daily house data
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

	// Create daily house response
	dailyResp := DailyHouseResp{
		Day:       req.Day,
		DailyData: req.DailyData,
	}

	// Create monthly house response
	monthResp := MonthHouseResp{
		Month:     req.Month,
		MonthData: req.MonthData,
	}
	m := GetInMemDataAccessor(beijing)

	// Store in Redis (force overwrite)
	if err := StoreHouseData(ctx, req.Day, dailyResp, beijingKey); err != nil {
		log.Logger.Error().Err(err).Str("day", req.Day).Msg("Failed to force store house data in Redis")
	}

	if err := StoreMonthHouseData(ctx, req.Month, monthResp, beijingKey); err != nil {
		log.Logger.Error().Err(err).Str("month", req.Month).Msg("Failed to force store month house data in Redis")
	}

	// Always store in memory
	m.Store(req.Day, dailyResp)
	m.Store(req.Month, monthResp)

	c.JSON(http.StatusOK, req)
}

// getHousePeriod retrieves house data for a specific period (1, 7, or 30 days)
func getHousePeriod(c *gin.Context) {
	// Get period from URL parameter
	daysParam := c.Param("days")
	var period int

	// Parse period
	switch daysParam {
	case "1":
		period = 1
	case "7":
		period = 7
	case "30":
		period = 30
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period (must be 1, 7, or 30)"})
		return
	}

	// Get region from query parameter (default to beijing)
	region := c.DefaultQuery("region", beijingKey)

	// Get data for the specified period
	data, err := GetHouseDataForPeriod(ctx, period, region)
	if err != nil {
		log.Logger.Error().Err(err).Int("period", period).Str("region", region).Msg("Failed to get house data for period")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get house data"})
		return
	}

	// Return data
	if len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "no data found for the specified period"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": period,
		"region": region,
		"data":   data,
	})
}

// getShHousePeriod retrieves Shanghai house data for a specific period (1, 7, or 30 days)
func getShHousePeriod(c *gin.Context) {
	// Get period from URL parameter
	daysParam := c.Param("days")
	var period int

	// Parse period
	switch daysParam {
	case "1":
		period = 1
	case "7":
		period = 7
	case "30":
		period = 30
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid period (must be 1, 7, or 30)"})
		return
	}

	// Get data for the specified period
	data, err := GetHouseDataForPeriod(ctx, period, shanghaiKey)
	if err != nil {
		log.Logger.Error().Err(err).Int("period", period).Msg("Failed to get Shanghai house data for period")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get Shanghai house data"})
		return
	}

	// Return data
	if len(data) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"msg": "no Shanghai house data found for the specified period"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period": period,
		"region": shanghaiKey,
		"data":   data,
	})
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
