package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// PerformanceMiddleware contains all performance-related middleware configurations
type PerformanceMiddleware struct {
	gzipConfig        middleware.GzipConfig
	rateLimiterConfig middleware.RateLimiterConfig
}

// NewPerformanceMiddleware creates a new performance middleware instance
func NewPerformanceMiddleware() *PerformanceMiddleware {
	return &PerformanceMiddleware{
		gzipConfig: middleware.GzipConfig{
			Level: 5, // Balanced compression level (1-9, 5 is good balance)
			Skipper: func(c echo.Context) bool {
				// Skip compression for WebSocket and already compressed content
				if c.Request().Header.Get(echo.HeaderUpgrade) == "websocket" {
					return true
				}
				// Skip metrics endpoint for faster response
				if c.Path() == "/metrics" || c.Path() == "/health" {
					return true
				}
				return false
			},
		},
		rateLimiterConfig: middleware.RateLimiterConfig{
			Skipper: middleware.DefaultSkipper,
			Store: middleware.NewRateLimiterMemoryStoreWithConfig(
				middleware.RateLimiterMemoryStoreConfig{
					Rate:      100,             // 100 requests
					Burst:     150,             // Burst up to 150
					ExpiresIn: 1 * time.Minute, // Per minute
				},
			),
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
		},
	}
}

// Gzip returns the configured gzip middleware
func (pm *PerformanceMiddleware) Gzip() echo.MiddlewareFunc {
	return middleware.GzipWithConfig(pm.gzipConfig)
}

// RateLimiter returns the configured rate limiter middleware
func (pm *PerformanceMiddleware) RateLimiter() echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(pm.rateLimiterConfig)
}

// EndpointRateLimiter returns a configurable endpoint-specific rate limiter middleware
// NOTE: Currently uses in-memory store. For production cluster deployments,
// consider implementing Redis-backed rate limiting for distributed enforcement.
func (pm *PerformanceMiddleware) EndpointRateLimiter(requests int, window time.Duration, burst int) echo.MiddlewareFunc {
	limit := rate.Every(window / time.Duration(requests))
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      limit,
				Burst:     burst,
				ExpiresIn: window,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			id := c.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(429, map[string]string{
				"error": "Too many requests to this endpoint, please try again later",
			})
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(429, map[string]string{
				"error": "Rate limit exceeded for this endpoint",
			})
		},
	}
	return middleware.RateLimiterWithConfig(config)
}

// Timeout returns middleware for request timeout
func (pm *PerformanceMiddleware) Timeout(timeout time.Duration) echo.MiddlewareFunc {
	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: timeout,
		Skipper: func(c echo.Context) bool {
			// Skip timeout for file uploads and long-running operations
			if c.Path() == "/v1/products/:id/images" && c.Request().Method == "POST" {
				return true
			}
			if c.Path() == "/v1/api/tally/export" || c.Path() == "/v1/api/tally/import" {
				return true
			}
			return false
		},
	})
}

// BodyLimit returns middleware for limiting request body size
func (pm *PerformanceMiddleware) BodyLimit(size string) echo.MiddlewareFunc {
	return middleware.BodyLimit(size)
}
