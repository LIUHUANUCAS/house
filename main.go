package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
	v1 := router.Group("/v1")
	{
		// Define routes
		v1.GET("/daily_house", dailyHouse)
	}

	// Run the server
	router.Run(":8080")
}

func dailyHouse(c *gin.Context) {
	c.JSON(http.StatusOK, getDefaultDailyHouse())
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

type DailyHouse struct {
	// MonthData MonthData `json:"month_data"`
	Day       string    `json:"day"`
	DailyData DailyData `json:"daily_data"`
}
type MonthData struct {
	TotalCount float64 `json:"total_count"`
	TotalArea  float64 `json:"total_area"`
	HouseCount float64 `json:"house_count"`
	HouseArea  float64 `json:"house_area"`
}
type DailyData struct {
	TotalCount float64 `json:"total_count"`
	TotalArea  float64 `json:"total_area"`
	HouseCount float64 `json:"house_count"`
	HouseArea  float64 `json:"house_area"`
}

func getDefaultDailyHouse() DailyHouse {
	return DailyHouse{
		Day: "2025-04-08",
		DailyData: DailyData{
			TotalCount: 744,
			TotalArea:  64840,
			HouseCount: 619,
			HouseArea:  58754.18,
		},
	}
}
