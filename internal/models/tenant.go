package models

import (
	"time"

	"github.com/google/uuid"
)

type Tenant struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	Name          string     `json:"name" db:"name"`
	Subdomain     string     `json:"subdomain" db:"subdomain"`
	License       string     `json:"license" db:"license_number"`
	Status        string     `json:"status" db:"status"`
	DefaultRoleID *uuid.UUID `json:"default_role_id,omitempty" db:"default_role_id"` // Role to assign to new users
	ContactEmail  *string    `json:"contact_email,omitempty" db:"contact_email"`
	ContactPhone  *string    `json:"contact_phone,omitempty" db:"contact_phone"`
	SupportEmail  *string    `json:"support_email,omitempty" db:"support_email"`
	SupportPhone  *string    `json:"support_phone,omitempty" db:"support_phone"`
	Address       *string    `json:"address,omitempty" db:"address"`
	City          *string    `json:"city,omitempty" db:"city"`
	State         *string    `json:"state,omitempty" db:"state"`
	Country       *string    `json:"country,omitempty" db:"country"`
	PostalCode    *string    `json:"postal_code,omitempty" db:"postal_code"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	SSOConfig     *SSOConfig `json:"sso_config,omitempty" db:"sso_config"`
}

type SSOConfig struct {
	Provider     string `json:"provider"` // saml, oidc
	IssuerURL    string `json:"issuer_url"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	MetadataURL  string `json:"metadata_url,omitempty"` // For SAML
	EnforceSSO   bool   `json:"enforce_sso"`
}
