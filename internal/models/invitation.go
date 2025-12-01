package models

import (
	"time"

	"github.com/google/uuid"
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
	InvitationStatusRevoked  InvitationStatus = "revoked"
)

type Invitation struct {
	ID          uuid.UUID        `json:"id" db:"id"`
	TenantID    uuid.UUID        `json:"tenant_id" db:"tenant_id"`
	Email       string           `json:"email" db:"email"`
	RoleID      uuid.UUID        `json:"role_id" db:"role_id"`
	Token       string           `json:"token" db:"token"`
	Status      InvitationStatus `json:"status" db:"status"`
	Permissions []string         `json:"permissions,omitempty" db:"permissions"` // Stored as JSONB
	ExpiresAt   time.Time        `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at" db:"updated_at"`
}
