package models

import "time"

// Refresh Token models
type RefreshToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Token     string     `json:"-" db:"token"` // Never return in JSON
	TokenHash string     `json:"-" db:"token_hash"`
	ClientID  *string    `json:"client_id" db:"client_id"`
	Scope     *string    `json:"scope" db:"scope"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	IssuedAt  time.Time  `json:"issued_at" db:"issued_at"`
	RevokedAt *time.Time `json:"revoked_at" db:"revoked_at"`
	IsRevoked bool       `json:"is_revoked" db:"is_revoked"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// Access Token Response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
	Scope        *string   `json:"scope"`
	UserID       string    `json:"user_id"`
	TenantID     string    `json:"tenant_id"`
	TokenID      string    `json:"token_id"`
	IssuedAt     time.Time `json:"issued_at"`
}

// Token Refresh Request
type RefreshTokenRequest struct {
	RefreshToken string   `json:"refresh_token"`
	GrantType    string   `json:"grant_type"`
	ClientID     *string  `json:"client_id"`
	Scope        *string  `json:"scope"`
}

// Token Revocation Request
type RevokeTokenRequest struct {
	Token     string    `json:"token"`
	TokenTypeHint *string `json:"token_type_hint"` // "access_token" or "refresh_token"
}

// OAuth2 Client
type OAuth2Client struct {
	ID          string     `json:"id" db:"id"`
	ClientID    string     `json:"client_id" db:"client_id"`
	ClientSecret string    `json:"-" db:"client_secret"`
	ClientSecretHash string `json:"-" db:"client_secret_hash"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	RedirectURIs []string  `json:"redirect_uris" db:"-"`
	Scopes      []string   `json:"scopes" db:"-"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// Authorization Code for OAuth2 flow
type AuthorizationCode struct {
	Code        string     `json:"code" db:"code"`
	CodeHash    string     `json:"-" db:"code_hash"`
	ClientID    string     `json:"client_id" db:"client_id"`
	UserID      string     `json:"user_id" db:"user_id"`
	RedirectURI string     `json:"redirect_uri" db:"redirect_uri"`
	Scope       *string    `json:"scope" db:"scope"`
	ExpiresAt   time.Time  `json:"expires_at" db:"expires_at"`
	Used        bool       `json:"used" db:"used"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
}