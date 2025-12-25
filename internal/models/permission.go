package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID            uuid.UUID              `json:"id" db:"id"`
	Name          string                 `json:"name" db:"name"`
	Resource      *string                `json:"resource,omitempty" db:"resource"`
	Action        *string                `json:"action,omitempty" db:"action"`
	Conditions    map[string]interface{} `json:"conditions,omitempty" db:"conditions"`
	ConditionsRaw []byte                 `json:"-" db:"-"` // Raw JSON for deferred parsing
	Description   *string                `json:"description,omitempty" db:"description"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
}

// GetConditions returns the parsed conditions map, parsing from raw JSON if needed
func (p *Permission) GetConditions() map[string]interface{} {
	if p.Conditions != nil {
		return p.Conditions
	}
	if len(p.ConditionsRaw) > 0 {
		var conditions map[string]interface{}
		if err := json.Unmarshal(p.ConditionsRaw, &conditions); err == nil {
			p.Conditions = conditions
			return conditions
		}
	}
	return nil
}

// HasDataScope checks if the permission has a specific data scope
func (p *Permission) HasDataScope(scope string) bool {
	conditions := p.GetConditions()
	if conditions == nil {
		return false
	}
	if dataScope, ok := conditions["data_scope"].(string); ok {
		return dataScope == scope
	}
	return false
}
