package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name        string     `json:"name" db:"name"`
	Description string     `json:"description" db:"description"`
	ParentID    *uuid.UUID `json:"parent_id" db:"parent_id"`
	Level       int        `json:"level" db:"level"`
	Path        string     `json:"path" db:"path"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	Subcategories []*Category `json:"subcategories,omitempty" db:"-"` // For nested responses
}

// CategoryBulkUpdate represents bulk update operations for categories
type CategoryBulkUpdate struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty"`
	Name        *string    `json:"name,omitempty"`
	Description *string    `json:"description,omitempty"`
}

// CategoryTree represents a category with its full path and level information
type CategoryTree struct {
	Category
	Path []string `json:"full_path" db:"-"` // Full path from root
	Depth int     `json:"depth" db:"-"`     // Depth in hierarchy
}