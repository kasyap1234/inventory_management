package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole struct {
	ID         uuid.UUID `json:"id" db:"id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	RoleID     uuid.UUID `json:"role_id" db:"role_id"`
	TenantID   uuid.UUID `json:"tenant_id" db:"tenant_id"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	AssignedAt time.Time `json:"assigned_at" db:"assigned_at"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}