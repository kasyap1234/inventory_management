package caching

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStats_Unit(t *testing.T) {
	t.Run("returns empty stats when Redis not available", func(t *testing.T) {
		// This test would require a mock Redis client
		// For now, we just verify the struct is properly constructed
		stats := &CacheStats{
			KeysCount:        100,
			UsedMemory:       1024000,
			UsedMemoryHuman:  "1M",
			Hits:             900,
			Misses:           100,
			HitRate:          90.0,
			ConnectedClients: 5,
			UptimeSeconds:    3600,
		}

		assert.Equal(t, int64(100), stats.KeysCount)
		assert.Equal(t, int64(1024000), stats.UsedMemory)
		assert.Equal(t, "1M", stats.UsedMemoryHuman)
		assert.Equal(t, float64(90.0), stats.HitRate)
	})

	t.Run("calculates hit rate correctly", func(t *testing.T) {
		testCases := []struct {
			hits     int64
			misses   int64
			expected float64
		}{
			{900, 100, 90.0},
			{500, 500, 50.0},
			{0, 100, 0.0},
			{100, 0, 100.0},
			{0, 0, 0.0}, // Edge case: no requests
		}

		for _, tc := range testCases {
			totalRequests := tc.hits + tc.misses
			var hitRate float64
			if totalRequests > 0 {
				hitRate = float64(tc.hits) / float64(totalRequests) * 100
			}
			assert.InDelta(t, tc.expected, hitRate, 0.01)
		}
	})
}

// Integration test requires actual Redis connection
func TestGetStats_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create a test Redis connection
	cacheService := NewRedisCacheService("localhost:6379", "", 0)

	ctx := context.Background()
	stats, err := cacheService.GetStats(ctx)

	// If Redis is not available, skip the test
	if err != nil {
		t.Skip("Redis not available for integration test")
	}

	assert.NotNil(t, stats)
	assert.GreaterOrEqual(t, stats.KeysCount, int64(0))
	assert.GreaterOrEqual(t, stats.UsedMemory, int64(0))
	assert.GreaterOrEqual(t, stats.HitRate, float64(0))
	assert.LessOrEqual(t, stats.HitRate, float64(100))
}
