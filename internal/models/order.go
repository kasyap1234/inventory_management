package models

import (
	"time"

	"github.com/google/uuid"
)

// OrderSearchFilter holds search and filter criteria for order queries
type OrderSearchFilter struct {
	Query             string     `json:"query,omitempty"`              // Full-text search across notes, supplier, distributor, product
	Status            *string    `json:"status,omitempty"`             // Status filter (pending, confirmed, delivered, etc.)
	OrderType         *string    `json:"order_type,omitempty"`         // Order type filter (purchase, sale)
	SupplierID        *uuid.UUID `json:"supplier_id,omitempty"`        // Supplier filter
	DistributorID     *uuid.UUID `json:"distributor_id,omitempty"`     // Distributor filter
	ProductID         *uuid.UUID `json:"product_id,omitempty"`         // Product filter
	WarehouseID       *uuid.UUID `json:"warehouse_id,omitempty"`       // Warehouse filter
	MinQuantity       *int       `json:"min_quantity,omitempty"`       // Minimum quantity
	MaxQuantity       *int       `json:"max_quantity,omitempty"`       // Maximum quantity
	MinValue          *float64   `json:"min_value,omitempty"`          // Minimum value (quantity * unit_price)
	MaxValue          *float64   `json:"max_value,omitempty"`          // Maximum value
	OrderDateFrom     *time.Time `json:"order_date_from,omitempty"`    // Order date from
	OrderDateTo       *time.Time `json:"order_date_to,omitempty"`      // Order date to
	ExpectedDeliveryBefore *time.Time `json:"expected_delivery_before,omitempty"`  // Expected delivery before
	ExpectedDeliveryAfter  *time.Time `json:"expected_delivery_after,omitempty"`   // Expected delivery after
	SortBy            string     `json:"sort_by,omitempty"`            // Sort field: order_date, created_at, quantity, unit_price, expected_delivery
	SortOrder         string     `json:"sort_order,omitempty"`         // Sort order: asc, desc
	Limit             int        `json:"limit,omitempty"`              // Page size (default: 50)
	Offset            int        `json:"offset,omitempty"`             // Page offset
}

// OrderBulkStatusUpdate represents bulk order status updates
type OrderBulkStatusUpdate struct {
	OrderIDs         []uuid.UUID `json:"order_ids" validate:"required,min=1"`          // List of order IDs to update
	NewStatus        string      `json:"new_status" validate:"required"`                // New status to apply
	ExpectedDelivery *time.Time  `json:"expected_delivery,omitempty"`                  // New expected delivery date (for shipped status)
	Notes            *string     `json:"notes,omitempty"`                              // Notes for the status change
	ValidationMode   string      `json:"validation_mode"`                             // Mode: "strict", "skip_invalid" - default strict
	TransactionMode  string      `json:"transaction_mode"`                            // Mode: "atomic", "best_effort" - default atomic
}

// OrderBulkCreate represents bulk order creation
type OrderBulkCreate struct {
	Orders           []*Order    `json:"orders" validate:"required,min=1,dive"`        // List of orders to create
	ValidationMode   string      `json:"validation_mode"`                             // Mode: "strict", "skip_invalid" - default strict
	TransactionMode  string      `json:"transaction_mode"`                            // Mode: "atomic", "best_effort" - default atomic
}

// GetTotalValue calculates the total value if quantity and unit_price are present
func (f *OrderSearchFilter) GetTotalValue() float64 {
	if f.MinQuantity != nil && f.MaxQuantity != nil && *f.MinQuantity == *f.MaxQuantity {
		// If exact quantity, could use for value comparison, but for now return 0 to indicate no specific calculation
	}
	return 0
}

type Order struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OrderType         string     `json:"order_type" db:"order_type"`
	SupplierID        *uuid.UUID `json:"supplier_id" db:"supplier_id"`
	DistributorID     *uuid.UUID `json:"distributor_id" db:"distributor_id"`
	ProductID         uuid.UUID  `json:"product_id" db:"product_id"`
	WarehouseID       uuid.UUID  `json:"warehouse_id" db:"warehouse_id"`
	Quantity          int        `json:"quantity" db:"quantity"`
	UnitPrice         float64    `json:"unit_price" db:"unit_price"`
	Status            string     `json:"status" db:"status"`
	OrderDate         time.Time  `json:"order_date" db:"order_date"`
	ExpectedDelivery  *time.Time `json:"expected_delivery" db:"expected_delivery"`
	Notes             *string    `json:"notes" db:"notes"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}