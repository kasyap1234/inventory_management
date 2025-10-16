package handlers

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateUUIDEnhanced tests UUID validation logic
func TestValidateUUIDEnhanced(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
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
		},
		{
			name:     "Empty string after trimming",
			input:    "   ",
			expectError: true,
		},
		{
			name:     "Too short UUID",
			input:    "550e8400-e29b-41d4-a716-44665544000",
			expectError: true,
		},
		{
			name:     "Too long UUID",
			input:    "550e8400-e29b-41d4-a716-4466554400000",
			expectError: true,
		},
		{
			name:     "Missing hyphen at position 8",
			input:    "550e8400e29b-41d4-a716-446655440000",
			expectError: true,
		},
		{
			name:     "Invalid character",
			input:    "550e8400-e29b-41d4-g716-446655440000",
			expectError: true,
		},
		{
			name:     "All hyphens placed wrong",
			input:    "550e8400e-29b-41d4-a716-446655440000",
			expectError: true,
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
			result, err := uuid.Parse(tt.input)

			if tt.expectError {
				require.Error(t, err, "Expected an error for input: %s", tt.input)
			} else {
				require.NoError(t, err, "Did not expect error for input: %s", tt.input)
				assert.Equal(t, tt.expectedUUID, result)
			}
		})
	}
}