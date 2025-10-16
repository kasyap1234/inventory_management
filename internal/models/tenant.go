package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID             uuid.UUID `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	Subdomain      string    `json:"subdomain" db:"subdomain"`
	License        string    `json:"license" db:"license_number"`
	Status         string    `json:"status" db:"status"`
	ContactEmail   *string   `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone   *string   `json:"contact_phone,omitempty" db:"contact_phone"`
	SupportEmail   *string   `json:"support_email,omitempty" db:"support_email"`
	SupportPhone   *string   `json:"support_phone,omitempty" db:"support_phone"`
	Address        *string   `json:"address,omitempty" db:"address"`
	City           *string   `json:"city,omitempty" db:"city"`
	State          *string   `json:"state,omitempty" db:"state"`
	Country        *string   `json:"country,omitempty" db:"country"`
	PostalCode     *string   `json:"postal_code,omitempty" db:"postal_code"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}