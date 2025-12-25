package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// =============================================================================
// Basic Health Endpoints (currently used by main.go)
// =============================================================================

// HealthCheck handles GET /health - simple liveness probe
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"service":   "agromart2",
	})
}

// ReadinessCheck handles GET /health/ready - readiness probe
func ReadinessCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ready",
		"timestamp": time.Now().UTC(),
		"service":   "agromart2",
		"ready_at":  time.Now().UTC(),
	})
}

// HealthCheckDetailed handles GET /health/detailed - basic detailed check
func HealthCheckDetailed(c echo.Context, pool *pgxpool.Pool) error {
	response := map[string]interface{}{
		"service":   "agromart2",
		"timestamp": time.Now().UTC(),
		"checks":    make(map[string]interface{}),
	}

	checks := response["checks"].(map[string]interface{})

	// Database health check
	if pool != nil {
		ctx := c.Request().Context()
		if err := pool.Ping(ctx); err != nil {
			log.Printf("Database ping failed: %v", err)
			checks["database"] = "unhealthy"
			checks["database_error"] = err.Error()
			response["status"] = "degraded"
		} else {
			checks["database"] = "ok"
		}
	} else {
		log.Printf("Database pool is nil - not initialized")
		checks["database"] = "unknown"
	}

	if response["status"] != "degraded" {
		response["status"] = "ok"
	}

	return c.JSON(http.StatusOK, response)
}

// =============================================================================
// Comprehensive Health Check Service (for future use or infrastructure monitoring)
// =============================================================================

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status       string                 `json:"status"` // "healthy", "degraded", "unhealthy"
	Timestamp    time.Time              `json:"timestamp"`
	Details      map[string]interface{} `json:"details,omitempty"`
	ResponseTime time.Duration          `json:"response_time_ms"`
}

// OverallHealth represents the overall health of the system
type OverallHealth struct {
	Status     string                  `json:"status"`
	Timestamp  time.Time               `json:"timestamp"`
	Version    string                  `json:"version"`
	Components map[string]HealthStatus `json:"components"`
}

// HealthCheckService provides comprehensive health check functionality
type HealthCheckService struct {
	db       *pgxpool.Pool
	cache    caching.CacheService
	minioSvc services.MinioService
}

// NewHealthCheckService creates a new health check service
func NewHealthCheckService(db *pgxpool.Pool, cache caching.CacheService, minioSvc services.MinioService) *HealthCheckService {
	return &HealthCheckService{
		db:       db,
		cache:    cache,
		minioSvc: minioSvc,
	}
}

// CheckDatabaseHealth checks the health of the database
func (s *HealthCheckService) CheckDatabaseHealth(ctx context.Context) HealthStatus {
	start := time.Now()
	status := HealthStatus{
		Timestamp: start,
		Details:   make(map[string]interface{}),
	}

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := s.db.Ping(checkCtx); err != nil {
		status.Status = "unhealthy"
		status.Details["error"] = err.Error()
		status.ResponseTime = time.Since(start)
		return status
	}

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

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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

	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

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

// GetOverallHealth returns the overall health status of all components
func (s *HealthCheckService) GetOverallHealth(ctx context.Context) OverallHealth {
	health := OverallHealth{
		Timestamp:  time.Now(),
		Version:    "1.0.0",
		Components: make(map[string]HealthStatus),
	}

	health.Components["database"] = s.CheckDatabaseHealth(ctx)
	health.Components["redis"] = s.CheckRedisHealth(ctx)
	health.Components["minio"] = s.CheckMinIOHealth(ctx)

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

// HealthCheckDetailedHandler returns detailed health information using HealthCheckService
func HealthCheckDetailedHandler(healthSvc *HealthCheckService) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()
		health := healthSvc.GetOverallHealth(ctx)

		statusCode := http.StatusOK
		if health.Status == "unhealthy" {
			statusCode = http.StatusServiceUnavailable
		}

		return c.JSON(statusCode, health)
	}
}

// =============================================================================
// Metrics and Documentation Endpoints
// =============================================================================

// MetricsHandler handles GET /metrics (Prometheus metrics format)
func MetricsHandler(c echo.Context) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	metrics := `# HELP go_gc_duration_seconds A summary of the GC invocation durations.
# TYPE go_gc_duration_seconds summary
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines %d
# HELP go_mem_heap_alloc_bytes Number of heap bytes allocated and still in use.
# TYPE go_mem_heap_alloc_bytes gauge
go_mem_heap_alloc_bytes %d
# HELP health_status Service health status (1 for healthy, 0 for unhealthy)
# TYPE health_status gauge
health_status 1
`

	response := fmt.Sprintf(metrics,
		runtime.NumGoroutine(),
		m.HeapAlloc,
	)

	c.Response().Header().Set("Content-Type", "text/plain; charset=utf-8")
	return c.String(http.StatusOK, response)
}

// DocumentationGuideHandler handles GET /v1/docs/guide
func DocumentationGuideHandler(c echo.Context) error {
	guide := `# Agromart2 API Developer Guide

Welcome to the Agromart2 API! This guide will help you get started with our REST API.

## Base URL
All API requests should be made to: https://api.agromart2.com/v1

## Authentication
Authentication is handled via JWT tokens. Include the token in the Authorization header:
` + "```" + `
Authorization: Bearer your_jwt_token
` + "```" + `

## Common Response Format
All responses follow this structure:
` + "```json" + `
{
	 "message": "Description of the result",
	 "data": { ... }
}
` + "```" + `

## Error Handling
Errors return a status code and message:
` + "```json" + `
{
	 "message": "Error description",
	 "error": "Detailed error information"
}
` + "```" + `

## Rate Limits
- 1000 requests per hour per IP
- 10000 requests per hour per authenticated user

## Support
For support, contact: support@agromart2.com
`

	c.Response().Header().Set("Content-Type", "text/markdown; charset=utf-8")
	return c.String(http.StatusOK, guide)
}

// DocumentationSpecHandler handles GET /v1/docs/spec (returns OpenAPI YAML)
func DocumentationSpecHandler(c echo.Context) error {
	spec := `# OpenAPI 3.0 Specification for Agromart2 API

This endpoint serves the OpenAPI specification in YAML format.
For detailed API specification, visit: /docs/swagger/v1/openapi.yaml

## Quick Start
1. Obtain JWT token via /auth/login
2. Include token in Authorization header
3. Make authenticated API calls

## Available Endpoints
- Authentication: /v1/auth/*
- Users: /v1/users/*
- Orders: /v1/orders/*
- Products: /v1/products/*
- Inventory: /v1/inventory/*
- Invoices: /v1/invoices/*

## Full Specification
The complete OpenAPI specification is available at:
GET /docs/swagger/v1/openapi.yaml
`

	c.Response().Header().Set("Content-Type", "text/markdown; charset=utf-8")
	return c.String(http.StatusOK, spec)
}
