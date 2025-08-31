package middleware

import (
	"context"
	"reflect"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuditMiddleware provides automatic audit logging for HTTP requests
type AuditMiddleware struct {
	auditService services.AuditLogsService
}

// NewAuditMiddleware creates a new audit middleware instance
func NewAuditMiddleware(auditService services.AuditLogsService) *AuditMiddleware {
	return &AuditMiddleware{
		auditService: auditService,
	}
}

// AuditRequest audits HTTP requests with configurable sensitivity levels
func (m *AuditMiddleware) AuditRequest(sensitivityLevel string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Process the request
			err := next(c)

			ctx := c.Request().Context()
			tenantID, ok := GetTenantIDFromContext(ctx)
			if !ok {
				// Skip auditing if no tenant context
				return err
			}

			userID, ok := common.GetUserIDFromContext(ctx)
			var userPtr *uuid.UUID
			if ok {
				userPtr = &userID
			}

			// Log based on sensitivity level
			switch sensitivityLevel {
			case "high":
				m.auditHighSensitivity(c, tenantID, userPtr, err)
			case "medium":
				m.auditMediumSensitivity(c, tenantID, userPtr, err)
			default:
				// Low sensitivity - only log errors and critical operations
				m.auditLowSensitivity(c, tenantID, userPtr, err)
			}

			return err
		}
	}
}

// auditLowSensitivity logs only critical operations and errors
func (m *AuditMiddleware) auditLowSensitivity(c echo.Context, tenantID uuid.UUID, userID *uuid.UUID, reqErr error) {
	method := c.Request().Method
	path := c.Path()

	// Only log certain HTTP methods or when there are errors
	if !m.shouldLogLowSensitivity(method, path, reqErr) {
		return
	}

	ctx := c.Request().Context()
	action := method + " " + path

	data := map[string]interface{}{
		"method":     method,
		"path":       path,
		"user_agent": c.Request().UserAgent(),
		"ip":         c.RealIP(),
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	if reqErr != nil {
		data["error"] = reqErr.Error()
	}

	// Use a generic audit log entry
	if err := m.auditService.LogActivity(ctx, tenantID, "http_requests", path, action, userID, nil, data); err != nil {
		// Log audit failure but don't fail the request
		c.Logger().Errorf("Failed to log audit activity: %v", err)
	}
}

// auditMediumSensitivity logs business operations
func (m *AuditMiddleware) auditMediumSensitivity(c echo.Context, tenantID uuid.UUID, userID *uuid.UUID, reqErr error) {
	method := c.Request().Method
	path := c.Path()

	// Skip static assets, health checks, etc.
	if m.shouldSkipLogging(method, path) {
		return
	}

	ctx := c.Request().Context()
	action := method + " " + path

	data := map[string]interface{}{
		"method":     method,
		"path":       path,
		"user_agent": c.Request().UserAgent(),
		"ip":         c.RealIP(),
		"timestamp":  time.Now().Format(time.RFC3339),
		"query_params": c.QueryParams(),
	}

	if reqErr != nil {
		data["error"] = reqErr.Error()
	}

	// Log with more details for medium sensitivity
	if err := m.auditService.LogActivity(ctx, tenantID, "http_requests", path, action, userID, nil, data); err != nil {
		c.Logger().Errorf("Failed to log audit activity: %v", err)
	}
}

// auditHighSensitivity logs detailed information for sensitive operations
func (m *AuditMiddleware) auditHighSensitivity(c echo.Context, tenantID uuid.UUID, userID *uuid.UUID, reqErr error) {
	method := c.Request().Method
	path := c.Path()

	ctx := c.Request().Context()
	action := method + " " + path

	data := map[string]interface{}{
		"method":       method,
		"path":         path,
		"user_agent":   c.Request().UserAgent(),
		"ip":           c.RealIP(),
		"timestamp":    time.Now().Format(time.RFC3339),
		"query_params": c.QueryParams(),
		"headers":      m.sanitizeHeaders(c.Request().Header),
	}

	if reqErr != nil {
		data["error"] = reqErr.Error()
	}

	if err := m.auditService.LogActivity(ctx, tenantID, "http_requests_sensitive", path, action, userID, nil, data); err != nil {
		c.Logger().Errorf("Failed to log audit activity: %v", err)
	}
}

// shouldLogLowSensitivity determines if a request should be logged at low sensitivity
func (m *AuditMiddleware) shouldLogLowSensitivity(method, path string, reqErr error) bool {
	// Always log errors
	if reqErr != nil {
		return true
	}

	// Log specific HTTP methods
	if method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE" {
		return true
	}

	// Log specific sensitive paths
	sensitivePaths := []string{"/auth/", "/users/", "/admin/", "/settings/"}
	for _, sensitive := range sensitivePaths {
		if len(path) >= len(sensitive) && path[:len(sensitive)] == sensitive {
			return true
		}
	}

	return false
}

// shouldSkipLogging determines if a path should be skipped from logging
func (m *AuditMiddleware) shouldSkipLogging(method, path string) bool {
	// Skip GET requests to static assets and health checks
	if method == "GET" {
		skipPrefixes := []string{
			"/health",
			"/metrics",
			"/api/docs",
			"/swagger",
			"/static/",
			"/favicon",
			"/robots.txt",
		}

		for _, prefix := range skipPrefixes {
			if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
				return true
			}
		}
	}

	return false
}

// sanitizeHeaders removes sensitive headers before logging
func (m *AuditMiddleware) sanitizeHeaders(headers map[string][]string) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, values := range headers {
		// Skip sensitive headers
		if m.isSensitiveHeader(key) {
			sanitized[key] = "[REDACTED]"
			continue
		}

		sanitized[key] = values
	}

	return sanitized
}

// isSensitiveHeader checks if a header contains sensitive information
func (m *AuditMiddleware) isSensitiveHeader(header string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
		"proxy-authorization",
	}

	for _, sensitive := range sensitiveHeaders {
		if strings.ToLower(header) == sensitive {
			return true
		}
	}

	return false
}

// AuditEntityChange audits changes to business entities
func (m *AuditMiddleware) AuditEntityChange(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, tableName, recordID, action string, oldEntity, newEntity interface{}) error {
	var oldValues, newValues models.JSONB

	// Convert entities to JSON for logging
	if oldEntity != nil {
		var err error
		oldValues, err = m.normalizeEntity(oldEntity)
		if err != nil {
			return err
		}
	}

	if newEntity != nil {
		var err error
		newValues, err = m.normalizeEntity(newEntity)
		if err != nil {
			return err
		}
	}

	return m.auditService.LogActivity(ctx, tenantID, tableName, recordID, action, userID, oldValues, newValues)
}

// normalizeEntity converts an entity to JSONB format, handling common types
func (m *AuditMiddleware) normalizeEntity(entity interface{}) (models.JSONB, error) {
	if entity == nil {
		return nil, nil
	}

	// Handle different entity types
	switch v := entity.(type) {
	case *models.User:
		return map[string]interface{}{
			"id":         v.ID,
			"email":      v.Email,
			"first_name": v.FirstName,
			"last_name":  v.LastName,
			"status":     v.Status,
		}, nil

	case *models.Product:
		return map[string]interface{}{
			"id":           v.ID,
			"name":         v.Name,
			"quantity":     v.Quantity,
			"unit_price":   v.UnitPrice,
			"category_id":  v.CategoryID,
			"barcode":      v.Barcode,
			"description":  v.Description,
		}, nil

	case *models.Order:
		return map[string]interface{}{
			"id":              v.ID,
			"order_type":      v.OrderType,
			"supplier_id":     v.SupplierID,
			"distributor_id":  v.DistributorID,
			"product_id":      v.ProductID,
			"warehouse_id":    v.WarehouseID,
			"quantity":        v.Quantity,
			"unit_price":      v.UnitPrice,
			"status":          v.Status,
		}, nil

	default:
		// Use reflection for generic entity handling
		result := make(map[string]interface{})
		val := reflect.ValueOf(entity)

		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}

		if val.Kind() == reflect.Struct {
			typ := val.Type()

			for i := 0; i < val.NumField(); i++ {
				field := typ.Field(i)
				fieldVal := val.Field(i)

				// Skip unexported fields and sensitive fields
				if field.PkgPath != "" || m.isSensitiveField(field.Name) {
					continue
				}

				// Convert field value
				if fieldVal.CanInterface() {
					result[field.Name] = fieldVal.Interface()
				}
			}
		}

		return result, nil
	}
}

// isSensitiveField checks if a field name represents sensitive data
func (m *AuditMiddleware) isSensitiveField(fieldName string) bool {
	sensitiveFields := []string{
		"PasswordHash",
		"TokenHash",
		"Secret",
	}

	for _, sensitive := range sensitiveFields {
		if fieldName == sensitive {
			return true
		}
	}

	return false
}

// AuditDecorator adds audit logging to function calls
func (m *AuditMiddleware) AuditDecorator(operation string, getEntity func() (interface{}, error)) func() error {
	return func() error {
		// Get current context - this would need to be passed in different ways
		// For now, use background context
		ctx := context.Background()

		tenantID, ok := GetTenantIDFromContext(ctx)
		if !ok {
			// Can't audit without tenant context
			entity, err := getEntity()
			if err != nil {
				return err
			}
			_ = entity // Use entity if needed for custom logging
			return nil
		}

		userID, _ := common.GetUserIDFromContext(ctx)
		var userPtr *uuid.UUID
		if userID != uuid.Nil {
			userPtr = &userID
		}

		// Get entity data before operation
		entity, err := getEntity()
		if err != nil {
			return err
		}

		// Log the operation (this is a simplified approach)
		data := map[string]interface{}{
			"operation": operation,
			"timestamp": time.Now().Format(time.RFC3339),
			"entity_type": reflect.TypeOf(entity).String(),
		}

		_ = entity // entity data could be used for more detailed logging

		if err := m.auditService.LogActivity(ctx, tenantID, "system_operations", operation, operation, userPtr, nil, data); err != nil {
			// Log error but don't fail the operation
		}

		return nil
	}
}