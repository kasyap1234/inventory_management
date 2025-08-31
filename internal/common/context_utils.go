package common

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type contextKey string

const (
	UserIDKey   contextKey = "user_id"
	TenantIDKey contextKey = "tenant_id"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error struct {
		Code    string            `json:"code"`
		Message string            `json:"message"`
		Details map[string]string `json:"details,omitempty"`
	} `json:"error"`
}

// CreateErrorResponse creates a standardized error response
func CreateErrorResponse(code string, message string, details map[string]string) *ErrorResponse {
	var resp ErrorResponse
	resp.Error.Code = code
	resp.Error.Message = message
	resp.Error.Details = details
	return &resp
}

// SendValidationError sends a validation error response
func SendValidationError(c echo.Context, field, message string) error {
	details := map[string]string{
		field: message,
	}
	return c.JSON(http.StatusBadRequest, CreateErrorResponse("VALIDATION_ERROR", "Validation failed", details))
}

// SendClientError sends a client error response
func SendClientError(c echo.Context, message string) error {
	return c.JSON(http.StatusBadRequest, CreateErrorResponse("CLIENT_ERROR", message, nil))
}

// SendServerError sends a server error response
func SendServerError(c echo.Context, message string) error {
	return c.JSON(http.StatusInternalServerError, CreateErrorResponse("SERVER_ERROR", message, nil))
}

// SendNotFoundError sends a not found error response
func SendNotFoundError(c echo.Context, resource string) error {
	return c.JSON(http.StatusNotFound, CreateErrorResponse("NOT_FOUND", fmt.Sprintf("%s not found", resource), nil))
}

// SendUnauthorizedError sends an unauthorized error response
func SendUnauthorizedError(c echo.Context) error {
	return c.JSON(http.StatusUnauthorized, CreateErrorResponse("UNAUTHORIZED", "Unauthorized access", nil))
}

// ValidateUUID validates UUID format with comprehensive checks
func ValidateUUID(idStr string, fieldName string) (uuid.UUID, error) {
	if strings.TrimSpace(idStr) == "" {
		return uuid.Nil, fmt.Errorf("%s is required", fieldName)
	}

	// Trim whitespace
	idStr = strings.TrimSpace(idStr)

	// Check exact length
	if len(idStr) != 36 {
		return uuid.Nil, fmt.Errorf("%s must be exactly 36 characters (including hyphens)", fieldName)
	}

	// Check hyphen placement
	expectedHyphens := []int{8, 13, 18, 23}
	for _, pos := range expectedHyphens {
		if pos >= len(idStr) || idStr[pos] != '-' {
			return uuid.Nil, fmt.Errorf("%s has invalid UUID format: hyphens must be at positions 9, 14, 19, and 24", fieldName)
		}
	}

	// Validate with UUID parser
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s contains invalid characters: %v", fieldName, err)
	}

	return id, nil
}

// ValidatePositiveInteger validates positive integer values with upper bounds
func ValidatePositiveInteger(value int, fieldName string, maxValue int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	if value > maxValue {
		return fmt.Errorf("%s cannot exceed %d", fieldName, maxValue)
	}
	return nil
}

// ValidatePositiveFloat validates positive float values with upper bounds
func ValidatePositiveFloat(value float64, fieldName string, maxValue float64) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive", fieldName)
	}
	if value > maxValue {
		return fmt.Errorf("%s cannot exceed %.2f", fieldName, maxValue)
	}
	return nil
}

// ValidateDateFormat validates date strings
func ValidateDateFormat(dateStr, fieldName string) error {
	if strings.TrimSpace(dateStr) == "" {
		return nil // Empty is allowed, will be handled elsewhere
	}

	// Try to parse as YYYY-MM-DD format
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("%s must be in YYYY-MM-DD format", fieldName)
	}

	// Check for reasonable date bounds
	if date.After(time.Now().AddDate(10, 0, 0)) {
		return fmt.Errorf("%s cannot be more than 10 years in the future", fieldName)
	}
	if date.Before(time.Now().AddDate(-100, 0, 0)) {
		return fmt.Errorf("%s cannot be more than 100 years ago", fieldName)
	}

	return nil
}

// ValidateGSTIN validates GSTIN format
func ValidateGSTIN(gstin, fieldName string) error {
	if strings.TrimSpace(gstin) == "" {
		return nil // GSTIN is optional
	}

	// GSTIN format: 22AAAAA1234A1ZA (15 characters)
	if len(gstin) != 15 {
		return fmt.Errorf("%s must be exactly 15 characters", fieldName)
	}

	// Pattern: First 2 digits, then 10 alphanumeric, then 1 alpha, 1 digit, 1 alpha
	pattern := `^[0-9]{2}[A-Z]{10}[0-9]{1}[A-Z]{1}[A-Z0-9]{1}$`
	matched, err := regexp.MatchString(pattern, gstin)
	if err != nil {
		return fmt.Errorf("invalid GSTIN validation pattern")
	}
	if !matched {
		return fmt.Errorf("%s has invalid GSTIN format", fieldName)
	}

	return nil
}

// ValidateRequiredString validates required string fields
func ValidateRequiredString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateOptionalString validates optional string fields
func ValidateOptionalString(value *string, fieldName string, maxLength int) error {
	if value != nil {
		if len(*value) > maxLength {
			return fmt.Errorf("%s cannot exceed %d characters", fieldName, maxLength)
		}
		// Trim whitespace
		*value = strings.TrimSpace(*value)
	}
	return nil
}

// ValidateOrderStatus validates order status values
func ValidateOrderStatus(status string) error {
	validStatuses := map[string]bool{
		"pending": true, "approved": true, "processing": true,
		"shipped": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("order status must be one of: pending, approved, processing, shipped, delivered, cancelled")
	}
	return nil
}

// ValidateInvoiceStatus validates invoice status
func ValidateInvoiceStatus(status string) error {
	validStatuses := map[string]bool{
		"unpaid": true, "paid": true, "overdue": true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invoice status must be one of: unpaid, paid, overdue")
	}
	return nil
}

// ValidateOrderType validates order types
func ValidateOrderType(orderType string) error {
	if orderType != "purchase" && orderType != "sales" {
		return fmt.Errorf("order type must be either 'purchase' or 'sales'")
	}
	return nil
}

// SafeString safely handles string pointer operations
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SafeFloat64 safely handles float64 pointer operations
func SafeFloat64(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

// GetUserIDFromContext extracts the user ID from the request context
func GetUserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

// GetTenantIDFromContext extracts the tenant ID from the request context
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return tenantID, ok
}
// SanitizeHTMLElement escapes HTML characters to prevent XSS attacks
func SanitizeHTMLElement(input string) string {
	return html.EscapeString(input)
}

// SanitizeHTMLField sanitizes string pointer fields for HTML display
func SanitizeHTMLField(field *string, fieldName string) error {
	if field != nil && *field != "" {
		// Check for potentially dangerous content
		sanitized := SanitizeHTMLElement(*field)

		// Limit length to prevent abuse
		if len(sanitized) > 1000 {
			return fmt.Errorf("%s content exceeds maximum allowed length", fieldName)
		}

		*field = sanitized
	}
	return nil
}

// SanitizeSearchQuery prevents SQL injection through LIKE queries
func SanitizeSearchQuery(query string) string {
	if strings.TrimSpace(query) == "" {
		return ""
	}

	// Remove dangerous characters that could be used for SQL injection
	// While we use parameterized queries, this provides additional defense
	query = strings.ReplaceAll(query, "%", "")  // Remove wildcards
	query = strings.ReplaceAll(query, "_", "")  // Remove single char wildcards
	query = strings.ReplaceAll(query, "'", "''") // Escape single quotes

	// Limit query length
	if len(query) > 100 {
		query = query[:100]
	}

	return strings.TrimSpace(query)
}

// ValidateOrderBusinessRules validates business rules for order creation
func ValidateOrderBusinessRules(quantity int, unitPrice float64, orderType string) error {
	// Validate quantity
	if quantity <= 0 {
		return fmt.Errorf("quantity must be positive")
	}
	if quantity > 1000000 {
		return fmt.Errorf("quantity cannot exceed 1,000,000 units")
	}

	// Validate unit price
	if unitPrice <= 0 {
		return fmt.Errorf("unit price must be positive")
	}
	if unitPrice > 10000000.00 {
		return fmt.Errorf("unit price cannot exceed ₹10,000,000")
	}

	// Validate total value (prevent overflow)
	totalValue := float64(quantity) * unitPrice
	if totalValue > 1000000000.00 {
		return fmt.Errorf("total order value cannot exceed ₹1,000,000,000")
	}

	// Validate order type
	if orderType != "purchase" && orderType != "sales" {
		return fmt.Errorf("order type must be either 'purchase' or 'sales'")
	}

	return nil
}

// SecureErrorMessage creates standardized error messages to prevent information leakage
func SecureErrorMessage(operation string, err error) error {
	if err == nil {
		return nil
	}

	// Log the full error details internally (caller should handle logging)
	// Return a generic message to the user
	return fmt.Errorf("failed to %s: operation could not be completed", operation)
}

// ValidateSortField validates and secures sort field parameters
func ValidateSortField(sortField string) string {
	// Define allowed sort fields to prevent injection
	allowedFields := map[string]bool{
		"order_date":      true,
		"created_at":      true,
		"quantity":        true,
		"unit_price":      true,
		"expected_delivery": true,
	}

	if allowedFields[sortField] {
		return "o." + sortField // Add table prefix for security
	}

	// Default to safe field
	return "o.order_date"
}

// ValidateSortOrder validates sort order parameters
func ValidateSortOrder(sortOrder string) string {
	order := strings.ToLower(sortOrder)
	if order == "asc" {
		return "ASC"
	}
	return "DESC" // Default to DESC
}

// ValidatePaginationParams validates pagination parameters
func ValidatePaginationParams(limit, offset int) (int, int, error) {
	// Validate limit
	if limit <= 0 {
		limit = 50 // Default
	}
	if limit > 1000 {
		limit = 1000 // Maximum
	}

	// Validate offset
	if offset < 0 {
		offset = 0
	}
	if offset > 1000000 {
		return 0, 0, fmt.Errorf("offset cannot exceed 1,000,000")
	}

	return limit, offset, nil
}

// ValidateDateRange validates date ranges to prevent abuse
func ValidateDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return fmt.Errorf("end date cannot be before start date")
	}

	// Prevent querying unreasonably large date ranges
	duration := endDate.Sub(startDate)
	maxDuration := time.Hour * 24 * 365 * 10 // 10 years
	if duration > maxDuration {
		return fmt.Errorf("date range cannot exceed 10 years")
	}

	return nil
}