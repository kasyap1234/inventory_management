package middleware

import (
	"context"
	"time"

	"agromart2/internal/common"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EnhancedErrorMiddleware provides enhanced error handling and logging
func EnhancedErrorMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Generate request ID if not present
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Set request ID in context and response header
			c.Set("request_id", requestID)
			c.Response().Header().Set("X-Request-ID", requestID)

			// Add HTTP context information for audit logging
			httpContext := map[string]interface{}{
				"ip_address": c.RealIP(),
				"user_agent": c.Request().UserAgent(),
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
			}

			// Create new context with HTTP information
			ctx := context.WithValue(c.Request().Context(), "http_context", httpContext)
			ctx = context.WithValue(ctx, "request_id", requestID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Record start time for performance logging
			start := time.Now()

			// Execute the handler
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Add status code to HTTP context
			httpContext["status_code"] = c.Response().Status

			// Log performance metrics
			logger := common.GetGlobalLogger()
			if logger != nil {
				metadata := map[string]interface{}{
					"method":      c.Request().Method,
					"path":        c.Request().URL.Path,
					"status_code": c.Response().Status,
					"request_id":  requestID,
				}

				// Add user and tenant info if available
				if userID, ok := common.GetUserIDFromContext(ctx); ok {
					metadata["user_id"] = userID
				}
				if tenantID, ok := common.GetTenantIDFromContext(ctx); ok {
					metadata["tenant_id"] = tenantID
				}

				operation := c.Request().Method + " " + c.Request().URL.Path
				logger.LogPerformance(ctx, operation, duration, metadata)
			}

			return err
		}
	}
}

// RequestLoggingMiddleware logs all incoming requests
func RequestLoggingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger := common.GetGlobalLogger()
			if logger == nil {
				return next(c)
			}

			ctx := c.Request().Context()

			// Log request start
			logger.InfoWithContext(ctx, "Incoming request", map[string]interface{}{
				"method":     c.Request().Method,
				"path":       c.Request().URL.Path,
				"query":      c.Request().URL.RawQuery,
				"ip_address": c.RealIP(),
				"user_agent": c.Request().UserAgent(),
			})

			// Execute handler
			err := next(c)

			// Log request completion
			level := "info"
			message := "Request completed"

			if err != nil {
				level = "error"
				message = "Request failed"
			} else if c.Response().Status >= 400 {
				level = "warn"
				message = "Request completed with error status"
			}

			logData := map[string]interface{}{
				"method":        c.Request().Method,
				"path":          c.Request().URL.Path,
				"status_code":   c.Response().Status,
				"response_size": c.Response().Size,
			}

			if err != nil {
				logData["error"] = err.Error()
			}

			switch level {
			case "error":
				logger.ErrorWithContext(ctx, message, err, logData)
			case "warn":
				logger.WarnWithContext(ctx, message, logData)
			default:
				logger.InfoWithContext(ctx, message, logData)
			}

			return err
		}
	}
}

// ErrorNotificationMiddleware sends notifications for critical errors
func ErrorNotificationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)

			if err != nil {
				// Check if it's a critical error that needs notification
				if shouldNotifyError(err, c.Response().Status) {
					go sendErrorNotification(c.Request().Context(), err, c)
				}
			}

			return err
		}
	}
}

// shouldNotifyError determines if an error should trigger notifications
func shouldNotifyError(err error, statusCode int) bool {
	// Notify for server errors (5xx)
	if statusCode >= 500 {
		return true
	}

	// Check for specific error types that should be notified
	if enhancedErr, ok := err.(*common.EnhancedError); ok {
		return enhancedErr.Severity == common.ErrorSeverityCritical ||
			enhancedErr.Severity == common.ErrorSeverityHigh
	}

	return false
}

// sendErrorNotification sends notifications for critical errors
func sendErrorNotification(ctx context.Context, err error, c echo.Context) {
	logger := common.GetGlobalLogger()
	if logger == nil {
		return
	}

	// Create notification data
	notificationData := map[string]interface{}{
		"error_type":    "critical_system_error",
		"error_message": err.Error(),
		"method":        c.Request().Method,
		"path":          c.Request().URL.Path,
		"status_code":   c.Response().Status,
		"ip_address":    c.RealIP(),
		"user_agent":    c.Request().UserAgent(),
		"timestamp":     time.Now(),
	}

	// Add user context if available
	if userID, ok := common.GetUserIDFromContext(ctx); ok {
		notificationData["user_id"] = userID
	}
	if tenantID, ok := common.GetTenantIDFromContext(ctx); ok {
		notificationData["tenant_id"] = tenantID
	}

	// Log critical alert
	logger.InfoWithContext(ctx, "CRITICAL ERROR NOTIFICATION", notificationData)

	// Publish event for downstream delivery service
	common.PublishEvent(ctx, "critical_system_error", notificationData)
}

// RecoveryMiddleware provides panic recovery with enhanced logging
func RecoveryMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			defer func() {
				if r := recover(); r != nil {
					logger := common.GetGlobalLogger()
					ctx := c.Request().Context()

					// Create enhanced error for panic
					panicErr := &common.EnhancedError{
						Code:     "PANIC_RECOVERED",
						Message:  "Application panic recovered",
						Severity: common.ErrorSeverityCritical,
						Context: common.ErrorContext{
							RequestID: c.Get("request_id").(string),
							Operation: "panic_recovery",
							Method:    c.Request().Method,
							Path:      c.Request().URL.Path,
							IPAddress: c.RealIP(),
							UserAgent: c.Request().UserAgent(),
							Timestamp: time.Now(),
						},
						Details: map[string]interface{}{
							"panic_value": r,
						},
					}

					// Add user context if available
					if userID, ok := common.GetUserIDFromContext(ctx); ok {
						panicErr.Context.UserID = &userID
					}
					if tenantID, ok := common.GetTenantIDFromContext(ctx); ok {
						panicErr.Context.TenantID = &tenantID
					}

					// Log the panic
					if logger != nil {
						logger.LogCriticalError(ctx, panicErr)
					}

					// Send error response
					if !c.Response().Committed {
						c.JSON(500, map[string]interface{}{
							"error": map[string]interface{}{
								"code":       "SERVER_ERROR",
								"message":    "Internal server error",
								"request_id": panicErr.Context.RequestID,
								"timestamp":  panicErr.Context.Timestamp,
							},
						})
					}
				}
			}()

			return next(c)
		}
	}
}
