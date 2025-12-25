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
	IsSystemRole bool       `json:"is_system_role" db:"is_system_role"`           // Cannot be deleted by tenant admin
	ParentRoleID *uuid.UUID `json:"parent_role_id,omitempty" db:"parent_role_id"` // For role hierarchy inheritance
	Priority     int        `json:"priority" db:"priority"`                       // Higher priority takes precedence in permission conflicts
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// CanBeAssignedBy checks if this role can be assigned by a user with the given priority.
// A user can assign roles with priority equal to or lower than their own.
func (r *Role) CanBeAssignedBy(assignerPriority int) bool {
	return assignerPriority >= r.Priority
}

// CanBeCreatedBy checks if a role with this priority can be created by a user.
// A user can only create roles with priority strictly lower than their own.
func (r *Role) CanBeCreatedBy(creatorPriority int) bool {
	return creatorPriority > r.Priority
}

// CanBeModifiedBy checks if this role can be modified by a user with the given priority.
// System roles cannot be modified. Non-system roles require higher priority to modify.
func (r *Role) CanBeModifiedBy(userPriority int, isPlatformAdmin bool) bool {
	if isPlatformAdmin {
		return true
	}
	if r.IsSystemRole {
		return false
	}
	return userPriority > r.Priority
}
