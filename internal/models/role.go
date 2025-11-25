package models

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	TenantID     uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name         string     `json:"name" db:"name"`
	Description  *string    `json:"description,omitempty" db:"description"`
	IsActive     bool       `json:"is_active" db:"is_active"`
	ParentRoleID *uuid.UUID `json:"parent_role_id,omitempty" db:"parent_role_id"` // For role hierarchy inheritance
	Priority     int        `json:"priority" db:"priority"`                       // Higher priority takes precedence in permission conflicts
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}