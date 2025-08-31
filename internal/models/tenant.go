package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Subdomain    string    `json:"subdomain" db:"subdomain"`
	License      string    `json:"license" db:"license_number"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}