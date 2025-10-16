package models

import (
	"time"

	"github.com/google/uuid"
)

type RolePermission struct {
	ID            uuid.UUID `json:"id" db:"id"`
	RoleID        uuid.UUID `json:"role_id" db:"role_id"`
	PermissionID  uuid.UUID `json:"permission_id" db:"permission_id"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}