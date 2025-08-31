package models

import (
	"time"
	"github.com/google/uuid"
)

// BulkOperationResult represents the result of a bulk operation
type BulkOperationResult struct {
	OperationID    string                  `json:"operation_id"`               // Unique operation ID
	Status         string                  `json:"status"`                     // Status: "pending", "processing", "completed", "failed", "partial"
	TotalItems     int                     `json:"total_items"`                // Total items to process
	ProcessedItems int                     `json:"processed_items"`            // Successfully processed items
	FailedItems    int                     `json:"failed_items"`              // Failed items
	Progress       float64                 `json:"progress"`                   // Progress percentage (0-100)
	StartTime      time.Time               `json:"start_time"`                // Operation start time
	CompletionTime *time.Time              `json:"completion_time,omitempty"` // Operation completion time
	Errors         []BulkOperationError    `json:"errors,omitempty"`          // List of errors encountered
	Items          []BulkOperationItem     `json:"items,omitempty"`           // Results per item
}

// BulkOperationError represents an error for a specific item in bulk operation
type BulkOperationError struct {
	ItemIndex int    `json:"item_index"` // Index of the item that failed
	ItemID    string `json:"item_id"`    // ID of the item that failed
	Error     string `json:"error"`      // Error message
}

// BulkOperationItem represents the result for a specific item
type BulkOperationItem struct {
	ItemIndex int       `json:"item_index"` // Index of the item
	ItemID    string    `json:"item_id"`    // ID of the item
	Status    string    `json:"status"`     // Status: "success", "failed"
	Error     *string   `json:"error,omitempty"` // Error message if failed
}

// BulkOperationQueue represents a queued bulk operation
type BulkOperationQueue struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	TenantID       uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	OperationType  string      `json:"operation_type" db:"operation_type"` // e.g., "product_bulk_update", "order_bulk_create"
	Payload        string      `json:"payload" db:"payload"`               // JSON payload
	Status         string      `json:"status" db:"status"`                 // Status: "queued", "processing", "completed", "failed"
	Priority       int         `json:"priority" db:"priority"`             // Priority: 1=low, 2=normal, 3=high, 4=critical
	Progress       float64     `json:"progress" db:"progress"`             // Progress percentage
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
	CompletedAt    *time.Time  `json:"completed_at,omitempty" db:"completed_at"`
	ErrorMessage   *string     `json:"error_message,omitempty" db:"error_message"`
	UserID         uuid.UUID   `json:"user_id" db:"user_id"`               // User who initiated the operation
}

// InventoryBulkAdjust represents bulk inventory adjustments
type InventoryBulkAdjust struct {
	Adjustments     []InventoryAdjustment `json:"adjustments" validate:"required,min=1,dive"` // List of adjustments
	ValidationMode  string                 `json:"validation_mode"`                            // Mode: "strict", "skip_invalid" - default strict
	TransactionMode string                 `json:"transaction_mode"`                           // Mode: "atomic", "best_effort" - default atomic
}

// InventoryAdjustment represents a single inventory adjustment
type InventoryAdjustment struct {
	WarehouseID uuid.UUID `json:"warehouse_id" validate:"required"`
	ProductID   uuid.UUID `json:"product_id" validate:"required"`
	QuantityChange int    `json:"quantity_change"`                          // Positive for addition, negative for deduction
	Reason      string    `json:"reason" validate:"required"`               // Reason for adjustment
}

// InventoryBulkTransfer represents bulk inventory transfers between warehouses
type InventoryBulkTransfer struct {
	Transfers       []InventoryTransfer `json:"transfers" validate:"required,min=1,dive"`    // List of transfers
	ValidationMode  string               `json:"validation_mode"`                             // Mode: "strict", "skip_invalid" - default strict
	TransactionMode string               `json:"transaction_mode"`                            // Mode: "atomic", "best_effort" - default atomic
}

// InventoryTransfer represents a single inventory transfer
type InventoryTransfer struct {
	ProductID         uuid.UUID `json:"product_id" validate:"required"`
	FromWarehouseID   uuid.UUID `json:"from_warehouse_id" validate:"required"`
	ToWarehouseID     uuid.UUID `json:"to_warehouse_id" validate:"required"`
	Quantity          int       `json:"quantity" validate:"required,min=1"`
	Reason            string    `json:"reason" validate:"required"`
}