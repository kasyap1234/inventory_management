package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	// DefaultPageSize is the default number of items per page
	DefaultPageSize = 20

	// MaxPageSize is the maximum allowed items per page to prevent abuse
	MaxPageSize = 100

	// MaxOffset is the maximum allowed offset to prevent deep pagination issues
	MaxOffset = 10000
)

// PaginationMiddleware enforces pagination limits to prevent performance issues
func PaginationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Check and enforce limit parameter
			limitStr := c.QueryParam("limit")
			if limitStr != "" {
				limit, err := strconv.Atoi(limitStr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INVALID_PARAMETER",
							"message": "Invalid limit parameter: must be a valid integer",
							"field":   "limit",
						},
					})
				}

				if limit <= 0 {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INVALID_PARAMETER",
							"message": "Limit must be greater than 0",
							"field":   "limit",
						},
					})
				}

				if limit > MaxPageSize {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "LIMIT_EXCEEDED",
							"message": fmt.Sprintf("Limit exceeds maximum allowed value (%d). Please use a smaller limit or implement pagination.", MaxPageSize),
							"field":   "limit",
							"max":     MaxPageSize,
						},
					})
				}
			}

			// Check and enforce offset parameter
			offsetStr := c.QueryParam("offset")
			if offsetStr != "" {
				offset, err := strconv.Atoi(offsetStr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INVALID_PARAMETER",
							"message": "Invalid offset parameter: must be a valid integer",
							"field":   "offset",
						},
					})
				}

				if offset < 0 {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INVALID_PARAMETER",
							"message": "Offset cannot be negative",
							"field":   "offset",
						},
					})
				}

				if offset > MaxOffset {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "OFFSET_EXCEEDED",
							"message": fmt.Sprintf("Offset exceeds maximum allowed value (%d). Deep pagination is not supported. Please use cursor-based pagination instead.", MaxOffset),
							"field":   "offset",
							"max":     MaxOffset,
						},
					})
				}
			}

			// Check for page parameter (alternative to offset)
			pageStr := c.QueryParam("page")
			if pageStr != "" {
				page, err := strconv.Atoi(pageStr)
				if err != nil {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INVALID_PARAMETER",
							"message": "Invalid page parameter: must be a valid integer",
							"field":   "page",
						},
					})
				}

				if page < 1 {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "INVALID_PARAMETER",
							"message": "Page must be greater than or equal to 1",
							"field":   "page",
						},
					})
				}

				// Calculate equivalent offset and validate
				limit := DefaultPageSize
				if limitStr != "" {
					if l, err := strconv.Atoi(limitStr); err == nil {
						limit = l
					}
				}

				calculatedOffset := (page - 1) * limit
				if calculatedOffset > MaxOffset {
					return echo.NewHTTPError(http.StatusBadRequest, map[string]interface{}{
						"error": map[string]interface{}{
							"code":    "OFFSET_EXCEEDED",
							"message": fmt.Sprintf("Calculated offset (%d) exceeds maximum allowed value (%d). Please use a smaller page number or cursor-based pagination.", calculatedOffset, MaxOffset),
							"field":   "page",
							"max":     MaxOffset,
						},
					})
				}
			}

			return next(c)
		}
	}
}

// GetPaginationParams safely extracts and validates pagination parameters
func GetPaginationParams(c echo.Context) (limit int, offset int) {
	// Default values
	limit = DefaultPageSize
	offset = 0

	// Parse limit
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= MaxPageSize {
			limit = l
		}
	}

	// Parse offset
	if offsetStr := c.QueryParam("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 && o <= MaxOffset {
			offset = o
		}
	}

	// Handle page-based pagination (convert to offset)
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page >= 1 {
			calculatedOffset := (page - 1) * limit
			if calculatedOffset <= MaxOffset {
				offset = calculatedOffset
			}
		}
	}

	return limit, offset
}
