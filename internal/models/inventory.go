package models

import (
	"time"

	"github.com/google/uuid"
)

// InventorySearchFilter holds search and filter criteria for inventory queries
type InventorySearchFilter struct {
	Query          string     `json:"query,omitempty"`              // Full-text search across product name, warehouse name
	WarehouseID    *uuid.UUID `json:"warehouse_id,omitempty"`       // Warehouse filter
	ProductID      *uuid.UUID `json:"product_id,omitempty"`         // Product filter
	MinQuantity    *int       `json:"min_quantity,omitempty"`       // Minimum stock quantity
	MaxQuantity    *int       `json:"max_quantity,omitempty"`       // Maximum stock quantity
	StockThreshold *int       `json:"stock_threshold,omitempty"`    // Stock threshold filter (< threshold for low stock alerts)
	MinStock       *int       `json:"min_stock,omitempty"`          // Minimum stock level
	MaxStock       *int       `json:"max_stock,omitempty"`          // Maximum stock level
	LastUpdatedFrom *time.Time `json:"last_updated_from,omitempty"`  // Last updated from
	LastUpdatedTo   *time.Time `json:"last_updated_to,omitempty"`    // Last updated to
	SortBy         string     `json:"sort_by,omitempty"`            // Sort field: quantity, last_updated, product_name, warehouse_name
	SortOrder      string     `json:"sort_order,omitempty"`         // Sort order: asc, desc
	Limit          int        `json:"limit,omitempty"`              // Page size (default: 50)
	Offset         int        `json:"offset,omitempty"`             // Page offset
}

type Inventory struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TenantID   uuid.UUID `json:"tenant_id" db:"tenant_id"`
	WarehouseID uuid.UUID `json:"warehouse_id" db:"warehouse_id"`
	ProductID  uuid.UUID `json:"product_id" db:"product_id"`
	Quantity   int       `json:"quantity" db:"quantity"`
	LastUpdated time.Time `json:"last_updated" db:"last_updated"`
}