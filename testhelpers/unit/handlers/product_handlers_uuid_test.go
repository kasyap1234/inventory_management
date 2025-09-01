package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateUUIDEnhanced tests the enhanced UUID validation function
func TestValidateUUIDEnhanced(t *testing.T) {
	// Create a ProductHandlers instance with nil dependencies for testing
	handlers := &ProductHandlers{
		productService: nil, // nil for test purposes
		rbacMiddleware: nil,
	}

	tests := []struct {
		name        string
		input       string
		expectError bool
		errorMsg    string
		expectedUUID uuid.UUID
	}{
		{
			name:        "Valid UUID",
			input:       "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
			expectedUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:        "Valid UUID with whitespaces trimmed",
			input:       " 550e8400-e29b-41d4-a716-446655440000 ",
			expectError: false,
			expectedUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
		{
			name:     "Empty string",
			input:    "",
			expectError: true,
			errorMsg: "Invalid UUID format: empty string",
		},
		{
			name:     "Empty string after trimming",
			input:    "   ",
			expectError: true,
			errorMsg: "Invalid UUID format: empty string",
		},
		{
			name:     "Too short UUID",
			input:    "550e8400-e29b-41d4-a716-44665544000",
			expectError: true,
			errorMsg: "Invalid UUID format: length must be 36 characters (including hyphens)",
		},
		{
			name:     "Too long UUID",
			input:    "550e8400-e29b-41d4-a716-4466554400000",
			expectError: true,
			errorMsg: "Invalid UUID format: length must be 36 characters (including hyphens)",
		},
		{
			name:     "Missing hyphen at position 8",
			input:    "550e8400e29b-41d4-a716-446655440000",
			expectError: true,
			errorMsg: "Invalid UUID format: hyphens must be at positions 8, 13, 18, and 23",
		},
		{
			name:     "Invalid character",
			input:    "550e8400-e29b-41d4-g716-446655440000",
			expectError: true,
			errorMsg: "Invalid UUID format: contains invalid characters or format",
		},
		{
			name:     "All hyphens placed wrong",
			input:    "550e8400e-29b-41d4-a716-446655440000",
			expectError: true,
			errorMsg: "Invalid UUID format: hyphens must be at positions 8, 13, 18, and 23",
		},
		{
			name:     "Case insensitive UUID",
			input:    "550E8400-E29B-41D4-A716-446655440000",
			expectError: false,
			expectedUUID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handlers.validateUUID(tt.input)

			if tt.expectError {
				require.Error(t, err, "Expected an error for input: %s", tt.input)
				var httpErr *echo.HTTPError
				require.ErrorAs(t, err, &httpErr, "Error should be HTTPError")
				assert.Equal(t, tt.errorMsg, httpErr.Message)
				assert.Equal(t, result, uuid.Nil)
			} else {
				require.NoError(t, err, "Did not expect error for input: %s", tt.input)
				assert.Equal(t, tt.expectedUUID, result)
			}
		})
	}
}