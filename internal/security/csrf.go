package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	csrfDefaultTokenLength = 32
	csrfTokenParts         = 3
)

// CSRFTokenManager handles stateless CSRF token generation and validation using HMAC signatures.
type CSRFTokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewCSRFTokenManager creates a new CSRF token manager with the provided secret and TTL.
// If the secret is empty, a random secret will be generated.
func NewCSRFTokenManager(secret string, ttl time.Duration) *CSRFTokenManager {
	if ttl <= 0 {
		ttl = time.Hour
	}
	if secret == "" {
		generated := make([]byte, csrfDefaultTokenLength)
		if _, err := rand.Read(generated); err != nil {
			// Critical security error: fall back to time-based entropy (not ideal but better than empty)
			timeStr := fmt.Sprintf("%d", time.Now().UnixNano())
			secret = base64.RawURLEncoding.EncodeToString([]byte(timeStr))
		} else {
			secret = base64.RawURLEncoding.EncodeToString(generated)
		}
	}

	return &CSRFTokenManager{
		secret: []byte(secret),
		ttl:    ttl,
	}
}

// GenerateToken produces a signed CSRF token and its expiration time.
func (m *CSRFTokenManager) GenerateToken() (string, time.Time, error) {
	nonce := make([]byte, csrfDefaultTokenLength)
	if _, err := rand.Read(nonce); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate nonce: %w", err)
	}

	nonceStr := base64.RawURLEncoding.EncodeToString(nonce)
	expiresAt := time.Now().Add(m.ttl).UTC()
	payload := fmt.Sprintf("%s.%d", nonceStr, expiresAt.Unix())
	signature := m.sign(payload)

	token := fmt.Sprintf("%s.%s", payload, signature)
	return token, expiresAt, nil
}

// ValidateToken verifies the provided CSRF token.
func (m *CSRFTokenManager) ValidateToken(token string) error {
	parts := strings.Split(token, ".")
	if len(parts) != csrfTokenParts {
		return errors.New("invalid CSRF token format")
	}

	payload := strings.Join(parts[:2], ".")
	providedSignature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return fmt.Errorf("invalid CSRF token signature: %w", err)
	}

	expiresAtUnix, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid CSRF token timestamp: %w", err)
	}

	expiresAt := time.Unix(expiresAtUnix, 0).UTC()
	if time.Now().UTC().After(expiresAt) {
		return errors.New("CSRF token expired")
	}

	expectedSignature, err := base64.RawURLEncoding.DecodeString(m.sign(payload))
	if err != nil {
		return fmt.Errorf("failed to decode expected signature: %w", err)
	}

	if !hmac.Equal(providedSignature, expectedSignature) {
		return errors.New("CSRF token signature mismatch")
	}

	return nil
}

func (m *CSRFTokenManager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.secret)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
