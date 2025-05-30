package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/LIUHUANUCAS/house/config"
	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

var (
	redisClient *redis.Client
	ctx         = context.Background()
)

// Redis key prefixes and structures for fortune data
const (
	// Key prefixes
	FortuneDailyKeyPrefix = "fortune:day"  // Prefix for daily fortune data
	FortuneDaysSetKey     = "fortune:days" // Sorted set of days with fortune data
)

// StoreFortuneData stores fortune data in Redis permanently (no expiration)
func StoreFortuneData(ctx context.Context, day string, data Poem) error {
	// Key format: fortune:day:{day}
	key := formatFortuneKey(day)

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to marshal fortune data")
		return err
	}

	// Store in Redis permanently (no expiration)
	err = redisDB.Set(ctx, key, jsonData, NoExpiration).Err()
	if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to store fortune data in Redis")
		return err
	}

	// Add the day to a sorted set for easy retrieval of recent days
	// Score is Unix timestamp for that day (start of day)
	t, _ := parseDay(day)
	score := float64(t.Unix())

	err = redisDB.ZAdd(ctx, FortuneDaysSetKey, &redis.Z{
		Score:  score,
		Member: day,
	}).Err()

	if err != nil {
		log.Logger.Error().Err(err).Str("day", day).Msg("Failed to add day to sorted set")
		return err
	}

	log.Logger.Debug().Str("key", key).Msg("Fortune data stored in Redis")
	return nil
}

// GetFortuneData retrieves fortune data for a specific day
func GetFortuneData(ctx context.Context, day string) (Poem, bool, error) {
	var poem Poem

	// Key format: fortune:day:{day}
	key := formatFortuneKey(day)

	// Get data from Redis
	jsonData, err := redisDB.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key does not exist
		return poem, false, nil
	} else if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to get fortune data from Redis")
		return poem, false, err
	}

	// Unmarshal JSON data
	err = json.Unmarshal([]byte(jsonData), &poem)
	if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal fortune data")
		return poem, false, err
	}

	return poem, true, nil
}

// GetRecentFortuneDays retrieves the list of days for which we have fortune data
// for the specified number of recent days
func GetRecentFortuneDays(ctx context.Context, days int) ([]string, error) {
	// Get current time
	now := time.Now()

	// Calculate the Unix timestamp for 'days' ago
	daysAgo := now.AddDate(0, 0, -days-1)
	minScore := float64(daysAgo.Unix())
	maxScore := float64(now.Unix())

	// Get days from sorted set
	result, err := redisDB.ZRangeByScore(ctx, FortuneDaysSetKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minScore),
		Max: fmt.Sprintf("%f", maxScore),
	}).Result()

	if err != nil {
		log.Logger.Error().Err(err).Int("days", days).Msg("Failed to get recent fortune days")
		return nil, err
	}

	return result, nil
}

// GetFortuneDataForRecentDays retrieves fortune data for recent days
func GetFortuneDataForRecentDays(ctx context.Context, days int) ([]Poem, error) {
	recentDays, err := GetRecentFortuneDays(ctx, days)
	if err != nil {
		return nil, err
	}

	var poems []Poem
	for _, day := range recentDays {
		poem, found, err := GetFortuneData(ctx, day)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", day).Msg("Error getting fortune data")
			continue
		}
		if found {
			poems = append(poems, poem)
		}
	}

	return poems, nil
}

// Helper function for fortune key formatting
func formatFortuneKey(day string) string {
	return fmt.Sprintf("%s:%s", FortuneDailyKeyPrefix, day)
}

// Redis key prefixes and structures
const (
	// Key prefixes
	HouseDailyKeyPrefix   = "house:daily"   // Prefix for daily house data
	HouseMonthlyKeyPrefix = "house:monthly" // Prefix for monthly house data
	HouseDaysSetKey       = "house:days"    // Sorted set of days with house data

	// No TTL for permanent storage
	NoExpiration = 0 // 0 means no expiration (permanent storage)
)

// RedisDB interface for Redis operations (useful for mocking in tests)
type RedisDB interface {
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd
	Get(ctx context.Context, key string) *redis.StringCmd
	ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd
	ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd
}

// ProductionRedisDB Production Redis client that implements RedisDB
type ProductionRedisDB struct {
	client *redis.Client
}

// Set stores a value in Redis with a specified key and TTL
func (db *ProductionRedisDB) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	return db.client.Set(ctx, key, value, ttl)
}

// Get retrieves a value from Redis by key
func (db *ProductionRedisDB) Get(ctx context.Context, key string) *redis.StringCmd {
	return db.client.Get(ctx, key)
}

// ZAdd adds members to a sorted set
func (db *ProductionRedisDB) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	return db.client.ZAdd(ctx, key, members...)
}

// ZRangeByScore retrieves members in a sorted set by score
func (db *ProductionRedisDB) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	return db.client.ZRangeByScore(ctx, key, opt)
}

// Global Redis DB instance
var redisDB RedisDB

// InitRedis initializes the Redis client
func InitRedis(ctx context.Context, redisCfg *config.RedisConfig) RedisDB {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	// Test the connection
	pong, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to connect to Redis")
		panic(fmt.Sprintf("Failed to connect to Redis: %v", err))
	}

	// Initialize the production Redis DB
	redisDB = &ProductionRedisDB{client: redisClient}

	log.Logger.Info().Str("pong", pong).Msg("Connected to Redis")
	return redisDB
}

// StoreHouseData stores house data in Redis permanently (no expiration)
func StoreHouseData(ctx context.Context, day string, data DailyHouseResp, region string) error {
	// Key format: house:daily:{region}:{day}
	key := formatDailyKey(region, day)

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to marshal house data")
		return err
	}

	// Store in Redis permanently (no expiration)
	if err := redisDB.Set(ctx, key, jsonData, NoExpiration).Err(); err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to store house data in Redis")
		return err
	}

	// Add the day to a sorted set for easy retrieval of recent days
	// Score is Unix timestamp for that day (start of day)
	t, _ := parseDay(day)
	score := float64(t.Unix())

	// Use region-specific sorted set
	daysSetKey := formatDaysSetKey(region)
	if err := redisDB.ZAdd(ctx, daysSetKey, &redis.Z{
		Score:  score,
		Member: day,
	}).Err(); err != nil {
		log.Logger.Error().Err(err).Str("day", day).Msg("Failed to add day to sorted set")
		return err
	}

	log.Logger.Debug().Str("key", key).Msg("House data stored in Redis")
	return nil
}

// StoreMonthHouseData stores monthly house data in Redis permanently (no expiration)
func StoreMonthHouseData(ctx context.Context, month string, data MonthHouseResp, region string) error {
	// Key format: house:monthly:{region}:{month}
	key := formatMonthlyKey(region, month)

	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Logger.Error().Err(err).Msg("Failed to marshal month house data")
		return err
	}

	// Store in Redis permanently (no expiration)
	if err := redisDB.Set(ctx, key, jsonData, NoExpiration).Err(); err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to store month house data in Redis")
		return err
	}

	log.Logger.Debug().Str("key", key).Msg("Month house data stored in Redis")
	return nil
}

// GetHouseData retrieves house data for a specific day
func GetHouseData(ctx context.Context, day string, region string) (DailyHouseResp, bool, error) {
	var houseData DailyHouseResp

	// Key format: house:daily:{region}:{day}
	key := formatDailyKey(region, day)

	// Get data from Redis
	jsonData, err := redisDB.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key does not exist
		return houseData, false, nil
	} else if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to get house data from Redis")
		return houseData, false, err
	}

	// Unmarshal JSON data
	err = json.Unmarshal([]byte(jsonData), &houseData)
	if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal house data")
		return houseData, false, err
	}

	return houseData, true, nil
}

// GetMonthHouseData retrieves monthly house data
func GetMonthHouseData(ctx context.Context, month string, region string) (MonthHouseResp, bool, error) {
	var monthData MonthHouseResp

	// Key format: house:monthly:{region}:{month}
	key := formatMonthlyKey(region, month)

	// Get data from Redis
	jsonData, err := redisDB.Get(ctx, key).Result()
	if err == redis.Nil {
		// Key does not exist
		return monthData, false, nil
	} else if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to get month house data from Redis")
		return monthData, false, err
	}

	// Unmarshal JSON data
	err = json.Unmarshal([]byte(jsonData), &monthData)
	if err != nil {
		log.Logger.Error().Err(err).Str("key", key).Msg("Failed to unmarshal month house data")
		return monthData, false, err
	}

	return monthData, true, nil
}

// GetRecentHouseDays retrieves the list of days for which we have house data
// for the specified number of recent days
func GetRecentHouseDays(ctx context.Context, days int, region string) ([]string, error) {
	// Get current time
	now := time.Now()

	// Calculate the Unix timestamp for 'days' ago
	daysAgo := now.AddDate(0, 0, -days-1)
	minScore := float64(daysAgo.Unix())
	maxScore := float64(now.Unix())

	// Use region-specific sorted set
	daysSetKey := formatDaysSetKey(region)

	// Get days from sorted set
	result, err := redisDB.ZRangeByScore(ctx, daysSetKey, &redis.ZRangeBy{
		Min: fmt.Sprintf("%f", minScore),
		Max: fmt.Sprintf("%f", maxScore),
	}).Result()

	if err != nil {
		log.Logger.Error().Err(err).Int("days", days).Msg("Failed to get recent house days")
		return nil, err
	}
	if len(result) > days {
		result = result[:days]
	}
	return result, nil
}

// GetHouseDataForRecentDays retrieves house data for recent days
func GetHouseDataForRecentDays(ctx context.Context, days int, region string) ([]DailyHouseResp, error) {
	recentDays, err := GetRecentHouseDays(ctx, days, region)
	if err != nil {
		return nil, err
	}

	var houseDataList []DailyHouseResp
	for _, day := range recentDays {
		houseData, found, err := GetHouseData(ctx, day, region)
		if err != nil {
			log.Logger.Error().Err(err).Str("day", day).Msg("Error getting house data")
			continue
		}
		if found {
			houseDataList = append(houseDataList, houseData)
		}
	}

	return houseDataList, nil
}

// GetHouseDataForPeriod retrieves house data for a specific period (1, 7, or 30 days)
func GetHouseDataForPeriod(ctx context.Context, period int, region string) ([]DailyHouseResp, error) {
	// Validate period
	if err := validatePeriod(period); err != nil {
		return nil, err
	}

	// Get data for the specified period
	return GetHouseDataForRecentDays(ctx, period, region)
}

func validatePeriod(period int) error {

	if period != oneDay && period != sevenDay && period != aMonth {
		return fmt.Errorf("invalid period: %d (must be 1, 7, or 30)", period)
	}
	return nil
}

// GetFortuneDataForPeriod retrieves fortune data for a specific period (1, 7, or 30 days)
func GetFortuneDataForPeriod(ctx context.Context, period int) ([]Poem, error) {
	// Validate period
	if err := validatePeriod(period); err != nil {
		return nil, err
	}

	// Get data for the specified period
	return GetFortuneDataForRecentDays(ctx, period)
}

// Helper functions for key formatting
func formatDailyKey(region, day string) string {
	return fmt.Sprintf("%s:%s:%s", HouseDailyKeyPrefix, region, day)
}

func formatMonthlyKey(region, month string) string {
	return fmt.Sprintf("%s:%s:%s", HouseMonthlyKeyPrefix, region, month)
}

func formatDaysSetKey(region string) string {
	return fmt.Sprintf("%s:%s", HouseDaysSetKey, region)
}
