package handlers

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// HealthCheck handles GET /health
func HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"service":   "agromart2",
	})
}

// ReadinessCheck handles GET /health/ready
func ReadinessCheck(c echo.Context) error {
	// Basic readiness check - service is ready if it can respond
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":     "ready",
		"timestamp":  time.Now().UTC(),
		"service":    "agromart2",
		"ready_at":   time.Now().UTC(),
	})
}

// HealthCheckDetailed handles GET /health/detailed
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
			log.Printf("DEBUG: Database ping failed: %v", err)
			checks["database"] = "unhealthy"
			checks["database_error"] = err.Error()
			response["status"] = "degraded"
		} else {
			log.Printf("DEBUG: Database connection successful")
			checks["database"] = "ok"
		}
	} else {
		log.Printf("DEBUG: Database pool is nil - not initialized")
		checks["database"] = "unknown"
	}

	// Memory usage (basic)
	// We'll add more checks here as needed

	// Overall status
	if response["status"] != "degraded" {
		response["status"] = "ok"
	}

	return c.JSON(http.StatusOK, response)
}

// MetricsHandler handles GET /metrics (Prometheus metrics format)
func MetricsHandler(c echo.Context) error {
	// Basic metrics - in a real application, you'd integrate with a metrics library
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