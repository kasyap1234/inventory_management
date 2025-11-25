package common

import (
	"fmt"
	"strings"
)

// SafeStringDeref safely dereferences a string pointer with a default value
func SafeStringDeref(s *string, defaultVal string) string {
	if s == nil {
		return defaultVal
	}
	return *s
}

// SafeFloatDeref safely dereferences a float64 pointer with a default value
func SafeFloatDeref(f *float64, defaultVal float64) float64 {
	if f == nil {
		return defaultVal
	}
	return *f
}

// SafeIntDeref safely dereferences an int pointer with a default value
func SafeIntDeref(i *int, defaultVal int) int {
	if i == nil {
		return defaultVal
	}
	return *i
}

// SafeBoolDeref safely dereferences a bool pointer with a default value
func SafeBoolDeref(b *bool, defaultVal bool) bool {
	if b == nil {
		return defaultVal
	}
	return *b
}

// TruncateString safely truncates a string to maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// SanitizeFilename removes dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Replace dangerous characters with underscores
	dangerous := []string{"/", "\\", "..", "<", ">", ":", "\"", "|", "?", "*", "\x00"}
	result := filename
	for _, char := range dangerous {
		result = strings.ReplaceAll(result, char, "_")
	}
	
	// Trim whitespace and dots from start/end
	result = strings.Trim(result, " .")
	
	// Ensure filename is not empty
	if result == "" {
		return "unnamed_file"
	}
	
	// Limit length to prevent filesystem issues
	return TruncateString(result, 255)
}

// ValidateEmailDomain performs basic email domain validation
func ValidateEmailDomain(email string) error {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}
	
	domain := parts[1]
	if domain == "" {
		return fmt.Errorf("email domain cannot be empty")
	}
	
	// Check for common typos and invalid domains
	invalidDomains := []string{
		"example.com", "test.com", "localhost", 
		"invalid.invalid", "temp.temp",
	}
	
	domainLower := strings.ToLower(domain)
	for _, invalid := range invalidDomains {
		if domainLower == invalid {
			return fmt.Errorf("email domain '%s' is not allowed", domain)
		}
	}
	
	return nil
}

// EscapeLikePattern escapes SQL LIKE special characters to prevent pattern injection
// Special characters: % (wildcard), _ (single char), \ (escape char)
// Uses \ as the escape character - queries should include ESCAPE '\\'
func EscapeLikePattern(input string) string {
	// Order matters: escape backslashes first
	result := strings.ReplaceAll(input, "\\", "\\\\")
	result = strings.ReplaceAll(result, "%", "\\%")
	result = strings.ReplaceAll(result, "_", "\\_")
	return result
}

// SafeSliceAccess safely accesses a slice element with bounds checking
func SafeSliceAccess[T any](slice []T, index int, defaultVal T) T {
	if index < 0 || index >= len(slice) {
		return defaultVal
	}
	return slice[index]
}

// IsEmptyOrWhitespace checks if a string pointer is nil or contains only whitespace
func IsEmptyOrWhitespace(s *string) bool {
	if s == nil {
		return true
	}
	return strings.TrimSpace(*s) == ""
}

// CoalesceString returns the first non-empty string from the provided values
func CoalesceString(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// SafeAppend safely appends to a slice, initializing if nil
func SafeAppend[T any](slice []T, items ...T) []T {
	if slice == nil {
		slice = make([]T, 0, len(items))
	}
	return append(slice, items...)
}
