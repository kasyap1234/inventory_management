package models

import (
	"time"

	"github.com/google/uuid"
)

type ProductImage struct {
	ID        uuid.UUID `json:"id" db:"id"`
	TenantID  uuid.UUID `json:"tenant_id" db:"tenant_id"`
	ProductID uuid.UUID `json:"product_id" db:"product_id"`
	ImageURL  string    `json:"image_url" db:"image_url"`
	AltText   *string   `json:"alt_text" db:"alt_text"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}