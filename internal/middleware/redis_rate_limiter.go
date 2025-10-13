package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

// RedisRateLimiterStore implements the RateLimiterStore interface using Redis
type RedisRateLimiterStore struct {
	client    *redis.Client
	rate      rate.Limit
	burst     int
	expiresIn time.Duration
}

// NewRedisRateLimiterStore creates a new Redis-backed rate limiter store
// This implementation is cluster-aware and shares rate limits across multiple instances
func NewRedisRateLimiterStore(redisAddr, redisPassword string, redisDB int, r rate.Limit, burst int, expiresIn time.Duration) *RedisRateLimiterStore {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		// Log error but don't fail - will gracefully degrade
		fmt.Printf("WARNING: Redis rate limiter connection failed: %v. Falling back to lenient mode.\n", err)
	}

	return &RedisRateLimiterStore{
		client:    client,
		rate:      r,
		burst:     burst,
		expiresIn: expiresIn,
	}
}

// Allow implements the RateLimiterStore interface
// Uses Redis sliding window algorithm for distributed rate limiting
func (s *RedisRateLimiterStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()
	
	// Redis key for this identifier
	key := fmt.Sprintf("ratelimit:%s", identifier)
	
	now := time.Now()
	windowStart := now.Add(-s.expiresIn)
	
	// Use Redis sorted set with timestamps as scores for sliding window
	pipe := s.client.TxPipeline()
	
	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))
	
	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)
	
	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})
	
	// Set expiration
	pipe.Expire(ctx, key, s.expiresIn+time.Minute)
	
	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		// On Redis error, allow the request (fail open)
		return true, nil
	}
	
	// Get count
	count := countCmd.Val()
	
	// Calculate rate based on requests per second
	requestsPerWindow := int64(float64(s.rate) * s.expiresIn.Seconds())
	if requestsPerWindow == 0 {
		requestsPerWindow = 1
	}
	
	// Allow burst
	maxRequests := requestsPerWindow + int64(s.burst)
	
	// Check if under limit
	return count <= maxRequests, nil
}

// Close closes the Redis connection
func (s *RedisRateLimiterStore) Close() error {
	return s.client.Close()
}

// NewRedisRateLimiterStoreWithConfig creates a Redis rate limiter store with config
func NewRedisRateLimiterStoreWithConfig(redisAddr, redisPassword string, redisDB int, config middleware.RateLimiterMemoryStoreConfig) *RedisRateLimiterStore {
	return NewRedisRateLimiterStore(
		redisAddr,
		redisPassword,
		redisDB,
		config.Rate,
		config.Burst,
		config.ExpiresIn,
	)
}

// RedisRateLimiterConfig extends Echo's rate limiter config with Redis support
type RedisRateLimiterConfig struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	Rate          rate.Limit
	Burst         int
	ExpiresIn     time.Duration
	Enabled       bool // Allow disabling Redis rate limiting
}

// CreateRedisRateLimiterMiddleware creates a rate limiter middleware using Redis
func CreateRedisRateLimiterMiddleware(config RedisRateLimiterConfig) middleware.RateLimiterConfig {
	var store middleware.RateLimiterStore
	
	if config.Enabled && config.RedisAddr != "" {
		// Use Redis store for cluster-aware rate limiting
		store = NewRedisRateLimiterStore(
			config.RedisAddr,
			config.RedisPassword,
			config.RedisDB,
			config.Rate,
			config.Burst,
			config.ExpiresIn,
		)
	} else {
		// Fallback to memory store for single-instance deployments
		store = middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      config.Rate,
				Burst:     config.Burst,
				ExpiresIn: config.ExpiresIn,
			},
		)
	}
	
	return middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   store,
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Rate limit by IP address
			id := c.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(429, map[string]string{
				"error": "Too many requests, please try again later",
			})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(429, map[string]string{
				"error": "Rate limit exceeded",
			})
		},
	}
}

// Simple token bucket implementation for Redis
// This uses the "token bucket" algorithm via Lua script for atomic operations
const luaTokenBucketScript = `
local key = KEYS[1]
local max_tokens = tonumber(ARGV[1])
local refill_rate = tonumber(ARGV[2])
local tokens_requested = tonumber(ARGV[3])
local current_time = tonumber(ARGV[4])

local tokens = redis.call('HGET', key, 'tokens')
local last_refill = redis.call('HGET', key, 'last_refill')

if not tokens then
    tokens = max_tokens
    last_refill = current_time
else
    tokens = tonumber(tokens)
    last_refill = tonumber(last_refill)
    
    -- Refill tokens based on time passed
    local time_passed = current_time - last_refill
    local tokens_to_add = math.floor(time_passed * refill_rate)
    tokens = math.min(max_tokens, tokens + tokens_to_add)
    last_refill = current_time
end

-- Check if we have enough tokens
if tokens >= tokens_requested then
    tokens = tokens - tokens_requested
    redis.call('HSET', key, 'tokens', tokens)
    redis.call('HSET', key, 'last_refill', last_refill)
    redis.call('EXPIRE', key, 3600)
    return 1
else
    return 0
end
`

// RedisTokenBucketStore implements rate limiting using token bucket algorithm
type RedisTokenBucketStore struct {
	client       *redis.Client
	maxTokens    int
	refillRate   float64
	scriptSHA    string
}

// NewRedisTokenBucketStore creates a token bucket rate limiter
func NewRedisTokenBucketStore(redisAddr, redisPassword string, redisDB int, maxTokens int, refillRate float64) (*RedisTokenBucketStore, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	// Load Lua script
	sha, err := client.ScriptLoad(ctx, luaTokenBucketScript).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to load lua script: %w", err)
	}

	return &RedisTokenBucketStore{
		client:     client,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		scriptSHA:  sha,
	}, nil
}

// Allow checks if a request should be allowed
func (s *RedisTokenBucketStore) Allow(identifier string) (bool, error) {
	ctx := context.Background()
	key := fmt.Sprintf("ratelimit:tb:%s", identifier)
	
	currentTime := float64(time.Now().UnixNano()) / 1e9
	
	result, err := s.client.EvalSha(ctx, s.scriptSHA, []string{key},
		s.maxTokens,
		s.refillRate,
		1, // tokens requested
		currentTime,
	).Result()
	
	if err != nil {
		// On error, allow the request (fail open)
		return true, nil
	}
	
	allowed, ok := result.(int64)
	if !ok {
		return true, nil
	}
	
	return allowed == 1, nil
}

// Close closes the Redis connection
func (s *RedisTokenBucketStore) Close() error {
	return s.client.Close()
}
