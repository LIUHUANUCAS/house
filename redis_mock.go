package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/rs/zerolog/log"
)

// MockRedisDB is a mock implementation of RedisDB for testing
type MockRedisDB struct {
	data       map[string]string
	sortedSets map[string]map[string]float64
	mu         sync.RWMutex
}

// NewMockRedisDB creates a new MockRedisDB instance
func NewMockRedisDB() *MockRedisDB {
	return &MockRedisDB{
		data:       make(map[string]string),
		sortedSets: make(map[string]map[string]float64),
	}
}

// Set implements RedisDB.Set for the mock
func (m *MockRedisDB) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) *redis.StatusCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Convert value to string if it's not already
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		// Try to marshal to JSON
		data, err := json.Marshal(value)
		if err != nil {
			log.Logger.Error().Err(err).Msg("Failed to marshal value in mock Redis")
			return redis.NewStatusResult("", err)
		}
		strValue = string(data)
	}

	m.data[key] = strValue
	return redis.NewStatusResult("OK", nil)
}

// Get implements RedisDB.Get for the mock
func (m *MockRedisDB) Get(ctx context.Context, key string) *redis.StringCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if value, ok := m.data[key]; ok {
		return redis.NewStringResult(value, nil)
	}
	return redis.NewStringResult("", redis.Nil)
}

// ZAdd implements RedisDB.ZAdd for the mock
func (m *MockRedisDB) ZAdd(ctx context.Context, key string, members ...*redis.Z) *redis.IntCmd {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.sortedSets[key]; !ok {
		m.sortedSets[key] = make(map[string]float64)
	}

	for _, member := range members {
		memberStr, ok := member.Member.(string)
		if !ok {
			// Try to convert to string
			memberStr = fmt.Sprintf("%v", member.Member)
		}
		m.sortedSets[key][memberStr] = member.Score
	}

	return redis.NewIntResult(int64(len(members)), nil)
}

// ZRangeByScore implements RedisDB.ZRangeByScore for the mock
func (m *MockRedisDB) ZRangeByScore(ctx context.Context, key string, opt *redis.ZRangeBy) *redis.StringSliceCmd {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if set, ok := m.sortedSets[key]; ok {
		// Parse min and max scores
		var minScore, maxScore float64
		fmt.Sscanf(opt.Min, "%f", &minScore)
		fmt.Sscanf(opt.Max, "%f", &maxScore)

		// Filter members by score
		var result []string
		for member, score := range set {
			if score >= minScore && score <= maxScore {
				result = append(result, member)
			}
		}

		return redis.NewStringSliceResult(result, nil)
	}

	return redis.NewStringSliceResult([]string{}, nil)
}

// EnableMockRedisForTesting replaces the global redisDB with a mock implementation for testing
func EnableMockRedisForTesting() *MockRedisDB {
	mockDB := NewMockRedisDB()
	redisDB = mockDB
	return mockDB
}
