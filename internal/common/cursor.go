package common

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// CursorDirection indicates pagination direction
type CursorDirection string

const (
	CursorNext CursorDirection = "next"
	CursorPrev CursorDirection = "prev"
)

// Cursor represents a pagination cursor for keyset pagination
// Cursor-based pagination is more performant than offset-based for large datasets
// because it doesn't require counting/skipping rows
type Cursor struct {
	// ID is the last seen record's unique identifier
	ID uuid.UUID `json:"id"`
	// CreatedAt is the last seen record's creation timestamp
	// Used for stable ordering when combined with ID
	CreatedAt time.Time `json:"created_at"`
	// Direction indicates whether we're paginating forward or backward
	Direction CursorDirection `json:"direction,omitempty"`
}

// CursorPageInfo provides metadata about the current page of results
type CursorPageInfo struct {
	// HasNextPage indicates if there are more results after this page
	HasNextPage bool `json:"has_next_page"`
	// HasPreviousPage indicates if there are results before this page
	HasPreviousPage bool `json:"has_previous_page"`
	// StartCursor is the cursor for the first item in the current page
	StartCursor string `json:"start_cursor,omitempty"`
	// EndCursor is the cursor for the last item in the current page
	EndCursor string `json:"end_cursor,omitempty"`
	// TotalCount is the total number of items (optional, may be omitted for performance)
	TotalCount *int `json:"total_count,omitempty"`
}

// CursorPaginatedResult wraps paginated results with cursor information
type CursorPaginatedResult[T any] struct {
	// Items is the list of results for the current page
	Items []T `json:"items"`
	// PageInfo contains pagination metadata
	PageInfo CursorPageInfo `json:"page_info"`
}

// EncodeCursor encodes a cursor to a base64 string for use in URLs
func EncodeCursor(c *Cursor) (string, error) {
	if c == nil {
		return "", nil
	}
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

// DecodeCursor decodes a base64 cursor string back to a Cursor struct
func DecodeCursor(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, nil
	}
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, errors.New("invalid cursor format")
	}
	var cursor Cursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return nil, errors.New("invalid cursor data")
	}
	return &cursor, nil
}

// NewCursor creates a new cursor from an ID and timestamp
func NewCursor(id uuid.UUID, createdAt time.Time, direction CursorDirection) *Cursor {
	return &Cursor{
		ID:        id,
		CreatedAt: createdAt,
		Direction: direction,
	}
}

// CursorFromProduct creates a cursor from product data
// This is a helper for the common case of product pagination
func CursorFromProduct(id uuid.UUID, createdAt time.Time) string {
	cursor := NewCursor(id, createdAt, CursorNext)
	encoded, _ := EncodeCursor(cursor)
	return encoded
}

// ValidateCursor validates a cursor string and returns the decoded cursor
func ValidateCursor(cursorStr string) (*Cursor, error) {
	if cursorStr == "" {
		return nil, nil
	}
	return DecodeCursor(cursorStr)
}

// CursorPaginationParams holds parameters for cursor-based pagination
type CursorPaginationParams struct {
	// First returns the first N items (forward pagination)
	First int `json:"first,omitempty"`
	// After is the cursor to start after (forward pagination)
	After string `json:"after,omitempty"`
	// Last returns the last N items (backward pagination)
	Last int `json:"last,omitempty"`
	// Before is the cursor to end before (backward pagination)
	Before string `json:"before,omitempty"`
}

// Validate checks that the pagination parameters are valid
func (p *CursorPaginationParams) Validate(maxLimit int) error {
	if p.First != 0 && p.Last != 0 {
		return errors.New("cannot specify both 'first' and 'last'")
	}
	if p.After != "" && p.Before != "" {
		return errors.New("cannot specify both 'after' and 'before'")
	}
	if p.First < 0 || p.Last < 0 {
		return errors.New("pagination limit cannot be negative")
	}
	if p.First > maxLimit || p.Last > maxLimit {
		return errors.New("pagination limit exceeds maximum")
	}
	return nil
}

// GetLimit returns the effective limit (first or last, with default)
func (p *CursorPaginationParams) GetLimit(defaultLimit int) int {
	if p.First > 0 {
		return p.First
	}
	if p.Last > 0 {
		return p.Last
	}
	return defaultLimit
}

// IsBackward returns true if paginating backward
func (p *CursorPaginationParams) IsBackward() bool {
	return p.Last > 0 || p.Before != ""
}
