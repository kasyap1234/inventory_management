package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductSearchFilter holds search and filter criteria for product queries
type ProductSearchFilter struct {
	Query        string     `json:"query,omitempty"`         // Full-text search across name, description, barcode, category
	CategoryID   *uuid.UUID `json:"category_id,omitempty"`   // Filter by category
	MinQuantity  *int       `json:"min_quantity,omitempty"`  // Minimum stock quantity
	MaxQuantity  *int       `json:"max_quantity,omitempty"`  // Maximum stock quantity
	MinPrice     *float64   `json:"min_price,omitempty"`     // Minimum unit price
	MaxPrice     *float64   `json:"max_price,omitempty"`     // Maximum unit price
	ExpiryBefore *time.Time `json:"expiry_before,omitempty"` // Expiry before date
	ExpiryAfter  *time.Time `json:"expiry_after,omitempty"`  // Expiry after date
	Barcode      *string    `json:"barcode,omitempty"`       // Exact barcode match
	SortBy       string     `json:"sort_by,omitempty"`       // Sort field: name, created_at, quantity, unit_price
	SortOrder    string     `json:"sort_order,omitempty"`    // Sort order: asc, desc
	Limit        int        `json:"limit,omitempty"`         // Page size (default: 50)
	Offset       int        `json:"offset,omitempty"`        // Page offset
}

// ProductBulkUpdate represents a bulk update operation for products
type ProductBulkUpdate struct {
	ProductIDs        []uuid.UUID          `json:"product_ids" validate:"required,min=1"`       // List of product IDs to update
	CategoryID        *uuid.UUID           `json:"category_id,omitempty"`                       // New category for all products
	UnitPriceChange   *float64             `json:"unit_price_change,omitempty"`                 // Price change (absolute or percentage)
	UnitPriceMode     string               `json:"unit_price_mode,omitempty"`                   // Mode: "absolute", "percentage"
	BatchNumber       *string              `json:"batch_number,omitempty"`                      // New batch number for all products
	ExpiryDate        *time.Time           `json:"expiry_date,omitempty"`                       // New expiry date for all products
	UnitOfMeasure     *string              `json:"unit_of_measure,omitempty"`                    // New unit of measure
	Description       *string              `json:"description,omitempty"`                        // New description for all products
	ValidationMode    string               `json:"validation_mode"`                             // Mode: "strict", "skip_invalid" - default strict
	TransactionMode   string               `json:"transaction_mode"`                            // Mode: "atomic", "best_effort" - default atomic
}

// ProductBulkCreate represents bulk product creation
type ProductBulkCreate struct {
	Products         []*Product           `json:"products" validate:"required,min=1,dive"`      // List of products to create
	ValidationMode   string               `json:"validation_mode"`                             // Mode: "strict", "skip_invalid" - default strict
	TransactionMode  string               `json:"transaction_mode"`                            // Mode: "atomic", "best_effort" - default atomic
}

type Product struct {
	ID             uuid.UUID `json:"id" db:"id"`
	TenantID       uuid.UUID `json:"tenant_id" db:"tenant_id"`
	CategoryID     *uuid.UUID `json:"category_id" db:"category_id"`
	Name           string    `json:"name" db:"name"`
	BatchNumber    *string   `json:"batch_number" db:"batch_number"`
	ExpiryDate     *time.Time `json:"expiry_date" db:"expiry_date"`
	Quantity       int       `json:"quantity" db:"quantity"`
	UnitPrice      float64   `json:"unit_price" db:"unit_price"`
	Barcode        *string   `json:"barcode" db:"barcode"`
	UnitOfMeasure  *string   `json:"unit_of_measure" db:"unit_of_measure"`
	Description    *string   `json:"description" db:"description"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}