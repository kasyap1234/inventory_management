package models

import (
	"time"

	"github.com/google/uuid"
)

// Batch represents a specific batch of a product
type Batch struct {
	ID                uuid.UUID `json:"id" db:"id"`
	ProductID         uuid.UUID `json:"product_id" db:"product_id"`
	BatchNumber       string    `json:"batch_number" db:"batch_number"`
	Quantity          int       `json:"quantity" db:"quantity"`
	ExpiryDate        *time.Time `json:"expiry_date" db:"expiry_date"`
	ManufacturingDate *time.Time `json:"manufacturing_date" db:"manufacturing_date"`
	Location          *string    `json:"location" db:"location"`
	Status            string    `json:"status" db:"status"` // active, expired, quarantined, recalled
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// BatchFilter holds search and filter criteria for batches
type BatchFilter struct {
	ProductID    *uuid.UUID `json:"product_id,omitempty"`
	BatchNumber  string     `json:"batch_number,omitempty"`
	Status       string     `json:"status,omitempty"`
	ExpiryBefore *time.Time `json:"expiry_before,omitempty"`
	ExpiryAfter  *time.Time `json:"expiry_after,omitempty"`
}
