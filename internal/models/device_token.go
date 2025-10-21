package models

import (
	"time"

	"github.com/google/uuid"
)

// DeviceToken represents a mobile/web device token for push notifications
type DeviceToken struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	TenantID    uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	DeviceToken string     `json:"device_token" db:"device_token"`
	DeviceType  string     `json:"device_type" db:"device_type"` // android, ios, web
	DeviceName  string     `json:"device_name,omitempty" db:"device_name"`
	AppVersion  string     `json:"app_version,omitempty" db:"app_version"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	LastUsedAt  time.Time  `json:"last_used_at" db:"last_used_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// DeviceTokenRequest represents the request payload for registering a device
type DeviceTokenRequest struct {
	DeviceToken string `json:"device_token" validate:"required"`
	DeviceType  string `json:"device_type" validate:"required,oneof=android ios web"`
	DeviceName  string `json:"device_name"`
	AppVersion  string `json:"app_version"`
}
