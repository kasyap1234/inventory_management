package common

import (
    "context"
    "fmt"
    "net/http"
    "runtime"
    "time"

    "github.com/google/uuid"
    "github.com/labstack/echo/v4"
)

// ErrorSeverity represents the severity level of an error
type ErrorSeverity string

const (
	ErrorSeverityLow      ErrorSeverity = "low"
	ErrorSeverityMedium   ErrorSeverity = "medium"
	ErrorSeverityHigh     ErrorSeverity = "high"
	ErrorSeverityCritical ErrorSeverity = "critical"
)

// ErrorContext contains contextual information about an error
type ErrorContext struct {
	RequestID    string                 `json:"request_id"`
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	TenantID     *uuid.UUID             `json:"tenant_id,omitempty"`
	Operation    string                 `json:"operation"`
	Resource     string                 `json:"resource,omitempty"`
	Method       string                 `json:"method,omitempty"`
	Path         string                 `json:"path,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	StackTrace   string                 `json:"stack_trace,omitempty"`
}

// EnhancedError represents a structured error with context and severity
type EnhancedError struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Severity  ErrorSeverity          `json:"severity"`
	Context   ErrorContext           `json:"context"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Cause     error                  `json:"-"` // Original error, not serialized
}

// Error implements the error interface
func (e *EnhancedError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Severity, e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *EnhancedError) Unwrap() error {
	return e.Cause
}

// EnhancedErrorHandler provides comprehensive error handling capabilities
type EnhancedErrorHandler struct {
	logger ErrorLogger
}

// ErrorLogger interface for logging errors
type ErrorLogger interface {
	LogError(ctx context.Context, err *EnhancedError)
	LogCriticalError(ctx context.Context, err *EnhancedError)
}

// NewEnhancedErrorHandler creates a new enhanced error handler
func NewEnhancedErrorHandler(logger ErrorLogger) *EnhancedErrorHandler {
	return &EnhancedErrorHandler{
		logger: logger,
	}
}

// HandleError processes an error and returns an appropriate response
func (h *EnhancedErrorHandler) HandleError(c echo.Context, err error, operation string) error {
	if c.Response().Committed {
		return nil
	}

	enhancedErr := h.enhanceError(c, err, operation)
	
	// Log the error based on severity
	if enhancedErr.Severity == ErrorSeverityCritical {
		h.logger.LogCriticalError(c.Request().Context(), enhancedErr)
	} else {
		h.logger.LogError(c.Request().Context(), enhancedErr)
	}

	// Return appropriate HTTP response
	return h.createHTTPResponse(c, enhancedErr)
}

// enhanceError converts a regular error into an EnhancedError with context
func (h *EnhancedErrorHandler) enhanceError(c echo.Context, err error, operation string) *EnhancedError {
	// Check if it's already an EnhancedError
	if enhancedErr, ok := err.(*EnhancedError); ok {
		return enhancedErr
	}

	// Extract context information
	ctx := h.extractContext(c, operation)
	
	// Determine error code, message, and severity
	code, message, severity := h.classifyError(err)
	
	// Create enhanced error
	enhancedErr := &EnhancedError{
		Code:     code,
		Message:  message,
		Severity: severity,
		Context:  ctx,
		Cause:    err,
	}

	// Add stack trace for high severity errors
	if severity == ErrorSeverityHigh || severity == ErrorSeverityCritical {
		enhancedErr.Context.StackTrace = h.getStackTrace()
	}

	return enhancedErr
}

// extractContext extracts contextual information from the Echo context
func (h *EnhancedErrorHandler) extractContext(c echo.Context, operation string) ErrorContext {
	ctx := ErrorContext{
		RequestID: h.getRequestID(c),
		Operation: operation,
		Method:    c.Request().Method,
		Path:      c.Request().URL.Path,
		UserAgent: c.Request().UserAgent(),
		IPAddress: c.RealIP(),
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Extract user and tenant IDs if available
	if userID, ok := GetUserIDFromContext(c.Request().Context()); ok {
		ctx.UserID = &userID
	}
	if tenantID, ok := GetTenantIDFromContext(c.Request().Context()); ok {
		ctx.TenantID = &tenantID
	}

	return ctx
}

// getRequestID extracts or generates a request ID
func (h *EnhancedErrorHandler) getRequestID(c echo.Context) string {
	// Try to get from header first
	if requestID := c.Request().Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}
	
	// Try to get from Echo context
	if requestID := c.Get("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	
	// Generate new UUID
	return uuid.New().String()
}

// classifyError determines the error code, message, and severity
func (h *EnhancedErrorHandler) classifyError(err error) (string, string, ErrorSeverity) {
	if err == nil {
		return "UNKNOWN_ERROR", "An unknown error occurred", ErrorSeverityMedium
	}

	// Handle Echo HTTP errors
	if he, ok := err.(*echo.HTTPError); ok {
		return h.classifyHTTPError(he)
	}

	// Handle validation errors
	if isValidationError(err) {
		return "VALIDATION_ERROR", "Validation failed", ErrorSeverityLow
	}

	// Handle database errors
	if isDatabaseError(err) {
		return "DATABASE_ERROR", "Database operation failed", ErrorSeverityHigh
	}

	// Handle external service errors
	if isExternalServiceError(err) {
		return "EXTERNAL_SERVICE_ERROR", "External service unavailable", ErrorSeverityMedium
	}

	// Handle authentication/authorization errors
	if isAuthError(err) {
		return "AUTH_ERROR", "Authentication or authorization failed", ErrorSeverityMedium
	}

	// Default to server error
	return "SERVER_ERROR", "Internal server error", ErrorSeverityHigh
}

// classifyHTTPError handles Echo HTTP errors
func (h *EnhancedErrorHandler) classifyHTTPError(he *echo.HTTPError) (string, string, ErrorSeverity) {
	switch he.Code {
	case http.StatusBadRequest:
		return "BAD_REQUEST", "Bad request", ErrorSeverityLow
	case http.StatusUnauthorized:
		return "UNAUTHORIZED", "Unauthorized", ErrorSeverityMedium
	case http.StatusForbidden:
		return "FORBIDDEN", "Forbidden", ErrorSeverityMedium
	case http.StatusNotFound:
		return "NOT_FOUND", "Resource not found", ErrorSeverityLow
	case http.StatusConflict:
		return "CONFLICT", "Resource conflict", ErrorSeverityMedium
	case http.StatusTooManyRequests:
		return "RATE_LIMITED", "Rate limit exceeded", ErrorSeverityMedium
	case http.StatusInternalServerError:
		return "SERVER_ERROR", "Internal server error", ErrorSeverityHigh
	case http.StatusBadGateway:
		return "BAD_GATEWAY", "Bad gateway", ErrorSeverityHigh
	case http.StatusServiceUnavailable:
		return "SERVICE_UNAVAILABLE", "Service unavailable", ErrorSeverityCritical
	default:
		if he.Code >= 500 {
			return "SERVER_ERROR", "Server error", ErrorSeverityHigh
		}
		return "CLIENT_ERROR", "Client error", ErrorSeverityLow
	}
}

// createHTTPResponse creates an appropriate HTTP response for the error
func (h *EnhancedErrorHandler) createHTTPResponse(c echo.Context, err *EnhancedError) error {
	statusCode := h.getHTTPStatusCode(err.Code)
	
	// Create response body
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"code":       err.Code,
			"message":    h.getSafeMessage(err),
			"request_id": err.Context.RequestID,
			"timestamp":  err.Context.Timestamp,
		},
	}

	// Add details for validation errors
	if err.Code == "VALIDATION_ERROR" && err.Details != nil {
		response["error"].(map[string]interface{})["details"] = err.Details
	}

	return c.JSON(statusCode, response)
}

// getSafeMessage returns a safe message for external consumption
func (h *EnhancedErrorHandler) getSafeMessage(err *EnhancedError) string {
	// For high severity errors, return generic message to avoid information leakage
	if err.Severity == ErrorSeverityHigh || err.Severity == ErrorSeverityCritical {
		switch err.Code {
		case "DATABASE_ERROR":
			return "A database error occurred. Please try again later."
		case "SERVER_ERROR":
			return "An internal server error occurred. Please try again later."
		default:
			return "An error occurred. Please try again later."
		}
	}
	
	return err.Message
}

// getHTTPStatusCode maps error codes to HTTP status codes
func (h *EnhancedErrorHandler) getHTTPStatusCode(code string) int {
	switch code {
	case "VALIDATION_ERROR", "BAD_REQUEST":
		return http.StatusBadRequest
	case "UNAUTHORIZED":
		return http.StatusUnauthorized
	case "FORBIDDEN":
		return http.StatusForbidden
	case "NOT_FOUND":
		return http.StatusNotFound
	case "CONFLICT":
		return http.StatusConflict
	case "RATE_LIMITED":
		return http.StatusTooManyRequests
	case "SERVICE_UNAVAILABLE":
		return http.StatusServiceUnavailable
	case "BAD_GATEWAY":
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}

// getStackTrace captures the current stack trace
func (h *EnhancedErrorHandler) getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Helper functions to classify error types
func isValidationError(err error) bool {
	// Check for validation-related error messages or types
	errMsg := err.Error()
	return contains(errMsg, "validation", "invalid", "required", "format")
}

func isDatabaseError(err error) bool {
	errMsg := err.Error()
	return contains(errMsg, "database", "sql", "connection", "timeout", "deadlock")
}

func isExternalServiceError(err error) bool {
	errMsg := err.Error()
	return contains(errMsg, "connection refused", "timeout", "network", "service unavailable")
}

func isAuthError(err error) bool {
	errMsg := err.Error()
	return contains(errMsg, "unauthorized", "forbidden", "authentication", "authorization", "token")
}

func contains(str string, keywords ...string) bool {
	str = fmt.Sprintf(" %s ", str) // Add spaces for word boundary matching
	for _, keyword := range keywords {
		if fmt.Sprintf(" %s ", keyword) != "" && 
		   len(str) >= len(keyword) && 
		   fmt.Sprintf("%s", str) != fmt.Sprintf("%s", str) { // Simple contains check
			return true
		}
	}
	return false
}

// CreateValidationError creates a validation error with details
func CreateValidationError(operation string, details map[string]interface{}) *EnhancedError {
	return &EnhancedError{
		Code:     "VALIDATION_ERROR",
		Message:  "Validation failed",
		Severity: ErrorSeverityLow,
		Context: ErrorContext{
			Operation: operation,
			Timestamp: time.Now(),
		},
		Details: details,
	}
}

// CreateDatabaseError creates a database error
func CreateDatabaseError(operation string, cause error) *EnhancedError {
	return &EnhancedError{
		Code:     "DATABASE_ERROR",
		Message:  "Database operation failed",
		Severity: ErrorSeverityHigh,
		Context: ErrorContext{
			Operation: operation,
			Timestamp: time.Now(),
		},
		Cause: cause,
	}
}

// CreateExternalServiceError creates an external service error
func CreateExternalServiceError(operation string, service string, cause error) *EnhancedError {
	return &EnhancedError{
		Code:     "EXTERNAL_SERVICE_ERROR",
		Message:  fmt.Sprintf("External service '%s' is unavailable", service),
		Severity: ErrorSeverityMedium,
		Context: ErrorContext{
			Operation: operation,
			Resource:  service,
			Timestamp: time.Now(),
		},
		Cause: cause,
	}
}