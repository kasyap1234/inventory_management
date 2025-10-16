package models

import (
	"time"

	"github.com/google/uuid"
)


type User struct {
	ID                uuid.UUID `json:"id" db:"id"`
	TenantID          uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Email             string    `json:"email" db:"email"`
	PasswordHash      string    `json:"-" db:"password_hash"` // Never serialize in JSON
	FirstName         string    `json:"first_name" db:"first_name"`
	LastName          string    `json:"last_name" db:"last_name"`
	Status            string    `json:"status" db:"status"`
	TwoFactorSecret   string    `json:"two_factor_secret,omitempty" db:"two_factor_secret"`
	TwoFactorEnabled  bool      `json:"two_factor_enabled" db:"two_factor_enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}