package models

import (
	"time"

	"github.com/google/uuid"
)

type Invoice struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TenantID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OrderID          uuid.UUID  `json:"order_id" db:"order_id"`
	InvoiceNumber    string     `json:"invoice_number" db:"invoice_number"`
	GSTIN            *string    `json:"gstin" db:"gstin"`
	HSNSAC           *string    `json:"hsn_sac" db:"hsn_sac"`
	TaxableAmount    *float64   `json:"taxable_amount" db:"taxable_amount"`
	GSTRate          *float64   `json:"gst_rate" db:"gst_rate"`
	CGST             *float64   `json:"cgst" db:"cgst"`
	SGST             *float64   `json:"sgst" db:"sgst"`
	IGST             *float64   `json:"igst" db:"igst"`
	TotalAmount      float64    `json:"total_amount" db:"total_amount"`
	Status           string     `json:"status" db:"status"`
	IssuedDate       time.Time  `json:"issued_date" db:"issued_date"`
	PaidDate         *time.Time `json:"paid_date" db:"paid_date"`
	DueDate          time.Time  `json:"due_date" db:"due_date"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}