package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type RateLimiter struct {
	limiters map[string]*limiterEntry
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
	stopCh   chan struct{}
	wg       sync.WaitGroup
}

func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	rl := &RateLimiter{
		limiters: make(map[string]*limiterEntry),
		rate:     r,
		burst:    b,
		stopCh:   make(chan struct{}),
	}
	rl.wg.Add(1)
	go rl.cleanup()
	return rl
}

// NewRateLimiterWithContext creates a rate limiter that stops when context is cancelled
func NewRateLimiterWithContext(ctx context.Context, r rate.Limit, b int) *RateLimiter {
	rl := NewRateLimiter(r, b)
	go func() {
		<-ctx.Done()
		rl.Stop()
	}()
	return rl
}

// Stop gracefully shuts down the rate limiter's cleanup goroutine
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
	rl.wg.Wait()
}

func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.limiters[ip]
	if !exists {
		entry = &limiterEntry{
			limiter:  rate.NewLimiter(rl.rate, rl.burst),
			lastSeen: time.Now(),
		}
		rl.limiters[ip] = entry
	} else {
		entry.lastSeen = time.Now()
	}
	return entry.limiter
}

func (rl *RateLimiter) cleanup() {
	defer rl.wg.Done()
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, entry := range rl.limiters {
				if now.Sub(entry.lastSeen) > 1*time.Hour {
					delete(rl.limiters, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

func (rl *RateLimiter) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			limiter := rl.getLimiter(ip)
			if !limiter.Allow() {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "Too many requests",
				})
			}
			return next(c)
		}
	}
}
