package models

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Resource    *string                `json:"resource,omitempty" db:"resource"`
	Action      *string                `json:"action,omitempty" db:"action"`
	Conditions  map[string]interface{} `json:"conditions,omitempty" db:"conditions"`
	Description *string                `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}