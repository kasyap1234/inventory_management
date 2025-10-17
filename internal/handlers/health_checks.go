package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// HealthCheckService provides health check functionality
type HealthCheckService struct {
	db       *pgxpool.Pool
	cache    caching.CacheService
	minioSvc *services.MinioService
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(db *pgxpool.Pool, cache caching.CacheService, minioSvc *services.MinioService) *HealthCheckService {
	return &HealthCheckService{
		db:       db,
		cache:    cache,
		minioSvc: minioSvc,
	}
}

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status      string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Timestamp   time.Time              `json:"timestamp"`
	Details     map[string]interface{} `json:"details,omitempty"`
	ResponseTime time.Duration         `json:"response_time_ms"`
}

// OverallHealth represents the overall health of the system
type OverallHealth struct {
	Status     string                  `json:"status"`
	Timestamp  time.Time               `json:"timestamp"`
	Version    string                  `json:"version"`
	Components map[string]HealthStatus `json:"components"`
}

// CheckDatabaseHealth checks the health of the database
func (s *HealthCheckService) CheckDatabaseHealth(ctx context.Context) HealthStatus {
	start := time.Now()
	status := HealthStatus{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Create a timeout context for the health check
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Ping the database
	if err := s.db.Ping(checkCtx); err != nil {
		status.Status = "unhealthy"
		status.Details["error"] = err.Error()
		status.ResponseTime = time.Since(start)
		return status
	}

	// Get pool stats
	stats := s.db.Stat()
	status.Status = "healthy"
	status.Details["total_connections"] = stats.TotalConns()
	status.Details["idle_connections"] = stats.IdleConns()
	status.Details["acquired_connections"] = stats.AcquiredConns()
	status.ResponseTime = time.Since(start)

	return status
}

// CheckRedisHealth checks the health of Redis
func (s *HealthCheckService) CheckRedisHealth(ctx context.Context) HealthStatus {
	start := time.Now()
	status := HealthStatus{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Create a timeout context
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to set and get a test value
	testKey := fmt.Sprintf("health_check:%d", time.Now().UnixNano())
	testValue := "ok"

	if err := s.cache.SetString(checkCtx, testKey, testValue, 10*time.Second); err != nil {
		status.Status = "unhealthy"
		status.Details["error"] = fmt.Sprintf("failed to write: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}

	retrievedValue, err := s.cache.GetString(checkCtx, testKey)
	if err != nil {
		status.Status = "unhealthy"
		status.Details["error"] = fmt.Sprintf("failed to read: %v", err)
		status.ResponseTime = time.Since(start)
		return status
	}

	// Clean up
	s.cache.Delete(checkCtx, testKey)

	if retrievedValue != testValue {
		status.Status = "unhealthy"
		status.Details["error"] = "value mismatch"
		status.ResponseTime = time.Since(start)
		return status
	}

	status.Status = "healthy"
	status.ResponseTime = time.Since(start)
	return status
}

// CheckMinIOHealth checks the health of MinIO
func (s *HealthCheckService) CheckMinIOHealth(ctx context.Context) HealthStatus {
	start := time.Now()
	status := HealthStatus{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	// Create a timeout context
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Try to list buckets to verify connectivity
	if err := s.minioSvc.HealthCheck(checkCtx); err != nil {
		status.Status = "unhealthy"
		status.Details["error"] = err.Error()
		status.ResponseTime = time.Since(start)
		return status
	}

	status.Status = "healthy"
	status.ResponseTime = time.Since(start)
	return status
}

// GetOverallHealth returns the overall health status
func (s *HealthCheckService) GetOverallHealth(ctx context.Context) OverallHealth {
	health := OverallHealth{
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		Components: make(map[string]HealthStatus),
	}

	// Check all components
	health.Components["database"] = s.CheckDatabaseHealth(ctx)
	health.Components["redis"] = s.CheckRedisHealth(ctx)
	health.Components["minio"] = s.CheckMinIOHealth(ctx)

	// Determine overall status
	unhealthyCount := 0
	for _, component := range health.Components {
		if component.Status == "unhealthy" {
			unhealthyCount++
		}
	}

	if unhealthyCount == 0 {
		health.Status = "healthy"
	} else if unhealthyCount == len(health.Components) {
		health.Status = "unhealthy"
	} else {
		health.Status = "degraded"
	}

	return health
}

// HealthCheckDetailedHandler returns detailed health information
func HealthCheckDetailedHandler(healthSvc *HealthCheckService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		health := healthSvc.GetOverallHealth(ctx)

		statusCode := http.StatusOK
		if health.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		} else if health.Status == "degraded" {
			statusCode = http.StatusOK // Still operational
		}

		return c.JSON(statusCode, health)
	}
}
