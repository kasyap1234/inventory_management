package middleware

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"agromart2/internal/caching"

	"github.com/labstack/echo/v4"
)

// IdempotencyConfig holds configuration for idempotency middleware
type IdempotencyConfig struct {
	CacheService caching.CacheService
	TTL          time.Duration
	HeaderName   string
	SkipMethods  []string
	KeyPrefix    string
}

// IdempotencyMiddleware provides idempotent request handling
type IdempotencyMiddleware struct {
	config IdempotencyConfig
}

// idempotencyResponse represents a cached response
type idempotencyResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}

// NewIdempotencyMiddleware creates a new idempotency middleware
func NewIdempotencyMiddleware(config IdempotencyConfig) *IdempotencyMiddleware {
	if config.HeaderName == "" {
		config.HeaderName = "Idempotency-Key"
	}
	if config.TTL == 0 {
		config.TTL = 24 * time.Hour
	}
	if config.KeyPrefix == "" {
		config.KeyPrefix = "idempotency:"
	}
	if len(config.SkipMethods) == 0 {
		config.SkipMethods = []string{"GET", "HEAD", "OPTIONS", "TRACE"}
	}

	return &IdempotencyMiddleware{
		config: config,
	}
}

// Middleware returns the Echo middleware function
func (m *IdempotencyMiddleware) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip if method is not applicable
			if m.shouldSkip(c.Request().Method) {
				return next(c)
			}

			// Get idempotency key from header
			idempotencyKey := c.Request().Header.Get(m.config.HeaderName)
			if idempotencyKey == "" {
				// No idempotency key provided, proceed normally
				return next(c)
			}

			// Validate idempotency key format
			if len(idempotencyKey) < 16 || len(idempotencyKey) > 255 {
				return echo.NewHTTPError(http.StatusBadRequest, "Invalid idempotency key format")
			}

			// Generate cache key
			cacheKey := m.generateCacheKey(c, idempotencyKey)

			// Check if request was already processed
			ctx := c.Request().Context()
			if cachedResponse, err := m.getCachedResponse(ctx, cacheKey); err == nil {
				// Return cached response
				return m.sendCachedResponse(c, cachedResponse)
			}

			// Capture response
			rec := &responseRecorder{
				ResponseWriter: c.Response().Writer,
				statusCode:     http.StatusOK,
				body:           []byte{},
			}
			c.Response().Writer = rec

			// Process request
			err := next(c)

			// Cache response if successful (2xx status code)
			if err == nil && rec.statusCode >= 200 && rec.statusCode < 300 {
				response := idempotencyResponse{
					StatusCode: rec.statusCode,
					Headers:    rec.headers,
					Body:       rec.body,
				}
				m.cacheResponse(ctx, cacheKey, response)
			}

			return err
		}
	}
}

// shouldSkip checks if the method should skip idempotency check
func (m *IdempotencyMiddleware) shouldSkip(method string) bool {
	for _, skipMethod := range m.config.SkipMethods {
		if strings.EqualFold(method, skipMethod) {
			return true
		}
	}
	return false
}

// generateCacheKey generates a unique cache key for the request
func (m *IdempotencyMiddleware) generateCacheKey(c echo.Context, idempotencyKey string) string {
	// Include method, path, and tenant/user ID in the key to prevent collisions
	h := sha256.New()
	h.Write([]byte(c.Request().Method))
	h.Write([]byte(c.Request().URL.Path))
	h.Write([]byte(idempotencyKey))

	// Include user ID if available
	if userID := c.Request().Header.Get("X-User-ID"); userID != "" {
		h.Write([]byte(userID))
	}

	hash := hex.EncodeToString(h.Sum(nil))
	return fmt.Sprintf("%s%s", m.config.KeyPrefix, hash)
}

// getCachedResponse retrieves a cached response
func (m *IdempotencyMiddleware) getCachedResponse(ctx context.Context, key string) (*idempotencyResponse, error) {
	// For simplicity, using string storage
	// In production, use structured storage (JSON, msgpack, etc.)
	data, err := m.config.CacheService.GetString(ctx, key)
	if err != nil {
		return nil, err
	}

	// Parse cached response (simplified - in production use proper serialization)
	response := &idempotencyResponse{
		StatusCode: http.StatusOK,
		Headers:    make(map[string]string),
		Body:       []byte(data),
	}

	return response, nil
}

// cacheResponse stores a response in cache
func (m *IdempotencyMiddleware) cacheResponse(ctx context.Context, key string, response idempotencyResponse) {
	// Simplified storage - in production use proper serialization
	data := string(response.Body)
	if err := m.config.CacheService.SetString(ctx, key, data, m.config.TTL); err != nil {
		// Log error but don't fail the request
		log.Printf("Failed to cache idempotent response: %v", err)
	}
}

// sendCachedResponse sends a cached response to the client
func (m *IdempotencyMiddleware) sendCachedResponse(c echo.Context, response *idempotencyResponse) error {
	// Set headers
	for key, value := range response.Headers {
		c.Response().Header().Set(key, value)
	}

	// Add header to indicate cached response
	c.Response().Header().Set("X-Idempotency-Cached", "true")

	// Send response
	return c.JSONBlob(response.StatusCode, response.Body)
}

// responseRecorder captures the response for caching
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	headers    map[string]string
	body       []byte
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// Hijack implements http.Hijacker
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("hijack not supported")
}

// Flush implements http.Flusher
func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Push implements http.Pusher
func (r *responseRecorder) Push(target string, opts *http.PushOptions) error {
	if p, ok := r.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return fmt.Errorf("push not supported")
}

// Note: The above imports for net, bufio need to be added to the import section
