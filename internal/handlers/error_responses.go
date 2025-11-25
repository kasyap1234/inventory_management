package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// StandardErrorResponse provides a consistent JSON structure for API errors
type StandardErrorResponse struct {
	Error   string                 `json:"error"`
	Code    string                 `json:"code,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// BadRequest returns a 400 error with consistent structure
func BadRequest(c echo.Context, message string, details ...map[string]interface{}) error {
	response := StandardErrorResponse{
		Error: message,
		Code:  "BAD_REQUEST",
	}
	if len(details) > 0 {
		response.Details = details[0]
	}
	return c.JSON(http.StatusBadRequest, response)
}

// NotFound returns a 404 error with consistent structure
func NotFound(c echo.Context, resource string) error {
	response := StandardErrorResponse{
		Error: resource + " not found",
		Code:  "NOT_FOUND",
	}
	return c.JSON(http.StatusNotFound, response)
}

// Unauthorized returns a 401 error with consistent structure
func Unauthorized(c echo.Context, message string) error {
	if message == "" {
		message = "Unauthorized access"
	}
	response := StandardErrorResponse{
		Error: message,
		Code:  "UNAUTHORIZED",
	}
	return c.JSON(http.StatusUnauthorized, response)
}

// Forbidden returns a 403 error with consistent structure
func Forbidden(c echo.Context, message string) error {
	if message == "" {
		message = "Access denied"
	}
	response := StandardErrorResponse{
		Error: message,
		Code:  "FORBIDDEN",
	}
	return c.JSON(http.StatusForbidden, response)
}

// InternalError returns a 500 error with consistent structure
// Note: Does not expose internal error details to prevent information leakage
func InternalError(c echo.Context, logErr error) error {
	// Log the actual error internally (caller should do this)
	response := StandardErrorResponse{
		Error: "An internal error occurred. Please try again later.",
		Code:  "INTERNAL_ERROR",
	}
	return c.JSON(http.StatusInternalServerError, response)
}

// ValidationError returns a 400 error for validation failures
func ValidationError(c echo.Context, field, message string) error {
	response := StandardErrorResponse{
		Error: "Validation failed",
		Code:  "VALIDATION_ERROR",
		Details: map[string]interface{}{
			"field":   field,
			"message": message,
		},
	}
	return c.JSON(http.StatusBadRequest, response)
}

// MultipleValidationErrors returns a 400 error for multiple validation failures
func MultipleValidationErrors(c echo.Context, errors map[string]string) error {
	details := make(map[string]interface{})
	for field, msg := range errors {
		details[field] = msg
	}
	response := StandardErrorResponse{
		Error:   "Validation failed",
		Code:    "VALIDATION_ERROR",
		Details: details,
	}
	return c.JSON(http.StatusBadRequest, response)
}

// Conflict returns a 409 error for resource conflicts
func Conflict(c echo.Context, message string) error {
	response := StandardErrorResponse{
		Error: message,
		Code:  "CONFLICT",
	}
	return c.JSON(http.StatusConflict, response)
}

// ParseUUID parses a UUID string and returns a 400 error if invalid
// This helper returns the parsed UUID and nil error on success, or uuid.Nil and an error response on failure
func ParseUUID(c echo.Context, value, paramName string) (uuid.UUID, error) {
	if value == "" {
		return uuid.Nil, BadRequest(c, paramName+" is required", map[string]interface{}{
			"field": paramName,
		})
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, BadRequest(c, "Invalid "+paramName+" format: must be a valid UUID", map[string]interface{}{
			"field":    paramName,
			"received": value,
		})
	}

	return id, nil
}

// ParsePathUUID is a convenience wrapper for parsing UUID from path parameters
func ParsePathUUID(c echo.Context, paramName string) (uuid.UUID, error) {
	return ParseUUID(c, c.Param(paramName), paramName)
}

// ParseQueryUUID is a convenience wrapper for parsing UUID from query parameters
func ParseQueryUUID(c echo.Context, paramName string) (uuid.UUID, bool, error) {
	value := c.QueryParam(paramName)
	if value == "" {
		return uuid.Nil, false, nil // Not provided, not an error
	}

	id, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, true, BadRequest(c, "Invalid "+paramName+" format: must be a valid UUID", map[string]interface{}{
			"field":    paramName,
			"received": value,
		})
	}

	return id, true, nil
}
