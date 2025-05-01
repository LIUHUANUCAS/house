package main

import (
	// "fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var m sync.Map
var hours []int = []int{-24, -48}

func main() {
	// Create a new Gin router
	router := gin.Default()
	// Add middleware
	router.Use(loggerMiddleware())
	router.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
	router.GET("/health1", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"msg": "success"})
	})
	v1 := router.Group("/v1")
	{
		// Define routes
		v1.GET("/daily_house", dailyHouse)
		v1.POST("/add_daily_house", AddDailyHouse)
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
	c.JSON(http.StatusNotFound, gin.H{"msg": "data not found"})
}

func AddDailyHouse(c *gin.Context) {
	var req DailyHouse
	// Bind JSON body to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if _, ok := m.Load(req.Day); !ok {
		m.Store(req.Day, req)
	}

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
		println(method, path, status, duration.String())
	}
}
