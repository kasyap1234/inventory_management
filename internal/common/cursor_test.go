package common

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecode(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Microsecond)

	cursor := &Cursor{
		ID:        id,
		CreatedAt: createdAt,
		Direction: CursorNext,
	}

	encoded, err := EncodeCursor(cursor)
	require.NoError(t, err)
	assert.NotEmpty(t, encoded)

	decoded, err := DecodeCursor(encoded)
	require.NoError(t, err)
	require.NotNil(t, decoded)

	assert.Equal(t, cursor.ID, decoded.ID)
	assert.Equal(t, cursor.CreatedAt.Unix(), decoded.CreatedAt.Unix())
	assert.Equal(t, cursor.Direction, decoded.Direction)
}

func TestDecodeCursor_Empty(t *testing.T) {
	decoded, err := DecodeCursor("")
	require.NoError(t, err)
	assert.Nil(t, decoded)
}

func TestDecodeCursor_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"invalid base64", "not-base64!@#$"},
		{"invalid json", "aW52YWxpZC1qc29u"}, // base64 of "invalid-json"
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecodeCursor(tt.input)
			assert.Error(t, err)
		})
	}
}

func TestEncodeCursor_Nil(t *testing.T) {
	encoded, err := EncodeCursor(nil)
	require.NoError(t, err)
	assert.Empty(t, encoded)
}

func TestNewCursor(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()

	cursor := NewCursor(id, createdAt, CursorPrev)

	assert.Equal(t, id, cursor.ID)
	assert.Equal(t, createdAt, cursor.CreatedAt)
	assert.Equal(t, CursorPrev, cursor.Direction)
}

func TestCursorFromProduct(t *testing.T) {
	id := uuid.New()
	createdAt := time.Now()

	encoded := CursorFromProduct(id, createdAt)
	assert.NotEmpty(t, encoded)

	decoded, err := DecodeCursor(encoded)
	require.NoError(t, err)
	assert.Equal(t, id, decoded.ID)
}

func TestCursorPaginationParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  CursorPaginationParams
		wantErr bool
	}{
		{
			name:    "valid first",
			params:  CursorPaginationParams{First: 10},
			wantErr: false,
		},
		{
			name:    "valid last",
			params:  CursorPaginationParams{Last: 10},
			wantErr: false,
		},
		{
			name:    "valid with after cursor",
			params:  CursorPaginationParams{First: 10, After: "cursor"},
			wantErr: false,
		},
		{
			name:    "valid with before cursor",
			params:  CursorPaginationParams{Last: 10, Before: "cursor"},
			wantErr: false,
		},
		{
			name:    "error: both first and last",
			params:  CursorPaginationParams{First: 10, Last: 10},
			wantErr: true,
		},
		{
			name:    "error: both after and before",
			params:  CursorPaginationParams{After: "a", Before: "b"},
			wantErr: true,
		},
		{
			name:    "error: negative first",
			params:  CursorPaginationParams{First: -1},
			wantErr: true,
		},
		{
			name:    "error: exceeds max limit",
			params:  CursorPaginationParams{First: 200},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate(100)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCursorPaginationParams_GetLimit(t *testing.T) {
	tests := []struct {
		name         string
		params       CursorPaginationParams
		defaultLimit int
		expected     int
	}{
		{
			name:         "uses first when set",
			params:       CursorPaginationParams{First: 25},
			defaultLimit: 10,
			expected:     25,
		},
		{
			name:         "uses last when set",
			params:       CursorPaginationParams{Last: 15},
			defaultLimit: 10,
			expected:     15,
		},
		{
			name:         "uses default when neither set",
			params:       CursorPaginationParams{},
			defaultLimit: 20,
			expected:     20,
		},
		{
			name:         "first takes precedence",
			params:       CursorPaginationParams{First: 30},
			defaultLimit: 10,
			expected:     30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.GetLimit(tt.defaultLimit)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCursorPaginationParams_IsBackward(t *testing.T) {
	tests := []struct {
		name     string
		params   CursorPaginationParams
		expected bool
	}{
		{
			name:     "forward with first",
			params:   CursorPaginationParams{First: 10},
			expected: false,
		},
		{
			name:     "backward with last",
			params:   CursorPaginationParams{Last: 10},
			expected: true,
		},
		{
			name:     "backward with before cursor",
			params:   CursorPaginationParams{Before: "cursor"},
			expected: true,
		},
		{
			name:     "forward with after cursor",
			params:   CursorPaginationParams{First: 10, After: "cursor"},
			expected: false,
		},
		{
			name:     "default is forward",
			params:   CursorPaginationParams{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.params.IsBackward()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateCursor(t *testing.T) {
	t.Run("empty string returns nil", func(t *testing.T) {
		cursor, err := ValidateCursor("")
		assert.NoError(t, err)
		assert.Nil(t, cursor)
	})

	t.Run("valid cursor decodes correctly", func(t *testing.T) {
		id := uuid.New()
		encoded := CursorFromProduct(id, time.Now())

		cursor, err := ValidateCursor(encoded)
		assert.NoError(t, err)
		assert.NotNil(t, cursor)
		assert.Equal(t, id, cursor.ID)
	})

	t.Run("invalid cursor returns error", func(t *testing.T) {
		_, err := ValidateCursor("invalid-cursor")
		assert.Error(t, err)
	})
}
