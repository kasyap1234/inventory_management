package models

import (
	"time"

	"github.com/google/uuid"
)

// Payment represents a one-time payment collected through Razorpay.
type Payment struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	TenantID          uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OrderID           *uuid.UUID `json:"order_id,omitempty" db:"order_id"`
	RazorpayOrderID   string     `json:"razorpay_order_id" db:"razorpay_order_id"`
	RazorpayPaymentID *string    `json:"razorpay_payment_id,omitempty" db:"razorpay_payment_id"`
	Currency          string     `json:"currency" db:"currency"`
	AmountPaise       int64      `json:"amount_paise" db:"amount_paise"`
	Status            string     `json:"status" db:"status"` // created, authorized, captured, failed, refunded
	Receipt           *string    `json:"receipt,omitempty" db:"receipt"`
	Signature         *string    `json:"signature,omitempty" db:"signature"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
	PaidAt            *time.Time `json:"paid_at,omitempty" db:"paid_at"`
}
