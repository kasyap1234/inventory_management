package models

import (
	"time"
)

// ExportRequest represents the request payload for tally export operations
type ExportRequest struct {
	StartDate string `json:"start_date" validate:"required" example:"2024-01-01"`
	EndDate   string `json:"end_date" validate:"required" example:"2024-01-31"`
	Format    string `json:"format" validate:"omitempty,oneof=csv excel" example:"csv"`
	DataType  string `json:"data_type" validate:"omitempty,oneof=orders invoices" example:"invoices"`
}

// ExportResult represents the response payload for tally export operations
type ExportResult struct {
	FileName        string `json:"file_name" example:"tally_export_12345_2024-01-01_2024-01-31.csv"`
	FileContent     string `json:"file_content,omitempty" example:"..."` // Only present in API responses, not file downloads
	RecordsExported int    `json:"records_exported" example:"150"`
	Message         string `json:"message,omitempty" example:"Export completed successfully"`
}

// ImportRequest represents the request payload for tally import operations
type ImportRequest struct {
	Data     string `json:"data" validate:"required" example:"Order Type,Product ID,Warehouse ID,Quantity,Unit Price,Order Date,Supplier/Distributor ID\npurchase,550e8400-e29b-41d4-a716-446655440000,550e8400-e29b-41d4-a716-446655440001,100,50.00,2024-01-01,550e8400-e29b-41d4-a716-446655440002"`
	DataType string `json:"data_type" validate:"required,oneof=orders invoices" example:"orders"`
}

// ImportResult represents the response payload for tally import operations
type ImportResult struct {
	RecordsProcessed int      `json:"records_processed" example:"10"`
	RecordsImported  int      `json:"records_imported" example:"8"`
	Errors           []string `json:"errors" example:"['Row 9: invalid product ID: invalid UUID format','Row 10: insufficient columns, expected at least 7']"`
	Message          string   `json:"message,omitempty" example:"Import completed with some errors"`
}

// TallyLedger represents ledger entries from Tally
type TallyLedger struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Balance     float64   `json:"balance"`
	LastUpdated time.Time `json:"last_updated"`
	Transactions []TallyTransaction `json:"transactions"`
}

// TallyBalance represents account balances from Tally
type TallyBalance struct {
	AccountName string    `json:"account_name"`
	Balance     float64   `json:"balance"`
	LastUpdated time.Time `json:"last_updated"`
}

// TallyTransaction represents individual ledger transactions
type TallyTransaction struct {
	ID          string    `json:"id"`
	Date        time.Time `json:"date"`
	Type        string    `json:"type"` // debit/credit
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
}