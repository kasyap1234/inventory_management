package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	TenantID               uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	RazorpaySubscriptionID *string    `json:"razorpay_subscription_id" db:"razorpay_subscription_id"`
	PlanName               string     `json:"plan_name" db:"plan_name"`
	Amount                 float64    `json:"amount" db:"amount"`
	Currency               string     `json:"currency" db:"currency"`
	Status                 string     `json:"status" db:"status"`
	StartDate              time.Time  `json:"start_date" db:"start_date"`
	EndDate                *time.Time `json:"end_date" db:"end_date"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
}