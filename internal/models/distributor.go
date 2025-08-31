package models

import (
	"time"

	"github.com/google/uuid"
)

type Distributor struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name           string    `json:"name" db:"name"`
	ContactEmail   *string   `json:"contact_email" db:"contact_email"`
	ContactPhone   *string   `json:"contact_phone" db:"contact_phone"`
	Address        *string   `json:"address" db:"address"`
	LicenseNumber  *string   `json:"license_number" db:"license_number"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}