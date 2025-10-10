package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AuthService handles OAuth2 and JWT token management
type AuthService interface {
	// Token management
	GenerateTokens(ctx context.Context, userID, tenantID uuid.UUID, scope *string) (*models.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, clientID *string) (*models.TokenResponse, error)
	ValidateToken(ctx context.Context, token string) (*TokenClaims, error)
	RevokeToken(ctx context.Context, token string, tokenType *string) error
	GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error)
	ValidatePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error)
	ConsumePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error)
	GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error)
	ConsumeEmailVerificationToken(ctx context.Context, token string) (uuid.UUID, error)
	IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error)
	RegisterFailedLoginAttempt(ctx context.Context, userID uuid.UUID) (int, bool, error)
	ClearFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error

	// OAuth2 flows
	GenerateAuthorizationCode(ctx context.Context, userID, tenantID uuid.UUID, clientID string, redirectURI, scope *string) (string, error)
	ValidateAuthorizationCode(ctx context.Context, code, clientID, redirectURI string) (*AuthorizationCodeClaims, error)
	MarkAuthorizationCodeUsed(ctx context.Context, code string) error

	// Refresh token management
	GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error)
	RevokeUserTokens(ctx context.Context, userID uuid.UUID) error
	CleanupExpiredTokens(ctx context.Context) error
}

type authService struct {
	cacheSvc             caching.CacheService
	jwtSecret            []byte
	tokenTTL             int // Access token TTL in seconds
	refreshTTL           int // Refresh token TTL in seconds
	passwordResetTTL     time.Duration
	emailVerificationTTL time.Duration
	maxLoginAttempts     int
	lockoutDuration      time.Duration
	loginAttemptWindow   time.Duration
}

// TokenClaims represents JWT claims
type TokenClaims struct {
	UserID   string  `json:"user_id"`
	TenantID string  `json:"tenant_id"`
	Scope    *string `json:"scope,omitempty"`
	TokenID  string  `json:"token_id"`
	ClientID *string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

// AuthorizationCodeClaims represents authorization code claims
type AuthorizationCodeClaims struct {
	Code     string
	CodeHash string
	UserID   string
	TenantID string
	ClientID string
	Scope    *string
}

// NewAuthService creates a new authentication service
func NewAuthService(cacheSvc caching.CacheService, jwtSecret string, tokenTTLSeconds, refreshTTLSeconds int) AuthService {
	return &authService{
		cacheSvc:             cacheSvc,
		jwtSecret:            []byte(jwtSecret),
		tokenTTL:             tokenTTLSeconds,
		refreshTTL:           refreshTTLSeconds,
		passwordResetTTL:     15 * time.Minute,
		emailVerificationTTL: 24 * time.Hour,
		maxLoginAttempts:     5,
		lockoutDuration:      15 * time.Minute,
		loginAttemptWindow:   15 * time.Minute,
	}
}

// GenerateTokens generates access and refresh tokens for a user
func (s *authService) GenerateTokens(ctx context.Context, userID, tenantID uuid.UUID, scope *string) (*models.TokenResponse, error) {
	now := time.Now()
	tokenID := uuid.NewString()

	// Generate JWT access token
	claims := TokenClaims{
		UserID:   userID.String(),
		TenantID: tenantID.String(),
		Scope:    scope,
		TokenID:  tokenID,
		ClientID: nil, // Public client default
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "agromart-auth",
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{"agromart-api"},
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.tokenTTL) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        tokenID,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessTokenString, err := accessToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT: %v", err)
	}

	// Generate refresh token
	refreshToken := s.generateSecureToken()
	refreshTokenHash, err := s.hashToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %v", err)
	}

	// Store refresh token (in production this would be database)
	refreshTokenData := fmt.Sprintf("%s:%s:%s:%d", userID.String(), tenantID.String(), refreshTokenHash, now.Unix()+int64(s.refreshTTL))
	cacheKey := fmt.Sprintf("refresh_token:%s", refreshTokenHash)
	if err := s.cacheSvc.SetString(ctx, cacheKey, refreshTokenData, time.Duration(s.refreshTTL)*time.Second); err != nil {
		log.Printf("Failed to store refresh token: %v", err)
		// Continue - token generation succeeded
	}

	response := &models.TokenResponse{
		AccessToken:  accessTokenString,
		TokenType:    "Bearer",
		ExpiresIn:    s.tokenTTL,
		RefreshToken: refreshToken,
		Scope:        scope,
		UserID:       userID.String(),
		TenantID:     tenantID.String(),
		TokenID:      tokenID,
		IssuedAt:     now,
	}

	return response, nil
}

// RefreshToken validates and uses refresh token to generate new tokens
func (s *authService) RefreshToken(ctx context.Context, refreshToken string, clientID *string) (*models.TokenResponse, error) {
	refreshTokenHash, err := s.hashToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to hash refresh token: %v", err)
	}

	cacheKey := fmt.Sprintf("refresh_token:%s", refreshTokenHash)
	tokenData, err := s.cacheSvc.GetString(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Parse token data
	parts := strings.Split(tokenData, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid token data")
	}

	userIDStr, tenantIDStr, tokenHash, expiryStr := parts[0], parts[1], parts[2], parts[3]
	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid token expiry")
	}

	// Check if expired
	if time.Now().Unix() > expiry {
		s.cacheSvc.Delete(ctx, cacheKey)
		return nil, fmt.Errorf("refresh token expired")
	}

	// Verify hash
	if tokenHash != refreshTokenHash {
		return nil, fmt.Errorf("invalid refresh token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID in token")
	}

	// Generate new tokens
	return s.GenerateTokens(ctx, userID, tenantID, nil)
}

// ValidateToken validates JWT access token
func (s *authService) ValidateToken(ctx context.Context, token string) (*TokenClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	if claims, ok := jwtToken.Claims.(*TokenClaims); ok && jwtToken.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

// RevokeToken revokes an access or refresh token
func (s *authService) RevokeToken(ctx context.Context, token string, tokenType *string) error {
	if tokenType != nil && *tokenType == "refresh_token" {
		// Handle refresh token revocation
		refreshTokenHash, err := s.hashToken(token)
		if err != nil {
			return fmt.Errorf("failed to hash refresh token: %v", err)
		}

		cacheKey := fmt.Sprintf("refresh_token:%s", refreshTokenHash)
		return s.cacheSvc.Delete(ctx, cacheKey)
	}

	// For access tokens, invalidate in cache
	claims, err := s.ValidateToken(ctx, token)
	if err != nil {
		return fmt.Errorf("cannot revoke invalid token: %v", err)
	}

	// Blacklist the token
	blacklistKey := fmt.Sprintf("token_blacklist:%s", claims.TokenID)
	if err := s.cacheSvc.SetString(ctx, blacklistKey, "revoked", claims.ExpiresAt.Sub(time.Now())); err != nil {
		log.Printf("Failed to blacklist token: %v", err)
	}

	return nil
}

// GenerateAuthorizationCode generates OAuth2 authorization code
func (s *authService) GenerateAuthorizationCode(ctx context.Context, userID, tenantID uuid.UUID, clientID string, redirectURI, scope *string) (string, error) {
	code := s.generateSecureToken()
	codeHash, err := s.hashToken(code)
	if err != nil {
		return "", fmt.Errorf("failed to hash authorization code: %v", err)
	}

	// Store authorization code
	now := time.Now()
	expiresAt := now.Add(10 * time.Minute) // 10 minute expiry
	codeData := fmt.Sprintf("%s:%s:%s:%s:%d", codeHash, userID.String(), tenantID.String(), clientID, expiresAt.Unix())

	if scope != nil {
		codeData += ":" + *scope
	}

	cacheKey := fmt.Sprintf("auth_code:%s", codeHash)
	if err := s.cacheSvc.SetString(ctx, cacheKey, codeData, 10*time.Minute); err != nil {
		return "", fmt.Errorf("failed to store authorization code: %v", err)
	}

	return code, nil
}

// ValidateAuthorizationCode validates OAuth2 authorization code
func (s *authService) ValidateAuthorizationCode(ctx context.Context, code, clientID, redirectURI string) (*AuthorizationCodeClaims, error) {
	codeHash, err := s.hashToken(code)
	if err != nil {
		return nil, fmt.Errorf("failed to hash authorization code: %v", err)
	}

	cacheKey := fmt.Sprintf("auth_code:%s", codeHash)
	codeData, err := s.cacheSvc.GetString(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("invalid authorization code")
	}

	parts := strings.Split(codeData, ":")
	if len(parts) < 5 {
		return nil, fmt.Errorf("invalid authorization code data")
	}

	hash, userIDStr, tenantIDStr, storedClientID, expiryStr := parts[0], parts[1], parts[2], parts[3], parts[4]

	// Verify code
	if hash != codeHash {
		return nil, fmt.Errorf("invalid authorization code")
	}

	// Verify client ID
	if storedClientID != clientID {
		return nil, fmt.Errorf("client ID mismatch")
	}

	// Check expiry
	expiry, err := strconv.ParseInt(expiryStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid expiry in authorization code")
	}

	if time.Now().Unix() > expiry {
		return nil, fmt.Errorf("authorization code expired")
	}

	var scope *string
	if len(parts) > 5 {
		scope = &parts[5]
	}

	return &AuthorizationCodeClaims{
		Code:     code,
		CodeHash: codeHash,
		UserID:   userIDStr,
		TenantID: tenantIDStr,
		ClientID: clientID,
		Scope:    scope,
	}, nil
}

// MarkAuthorizationCodeUsed marks authorization code as used
func (s *authService) MarkAuthorizationCodeUsed(ctx context.Context, code string) error {
	codeHash, err := s.hashToken(code)
	if err != nil {
		return fmt.Errorf("failed to hash authorization code: %v", err)
	}

	cacheKey := fmt.Sprintf("auth_code:%s", codeHash)
	return s.cacheSvc.Delete(ctx, cacheKey)
}

// GetRefreshToken retrieves refresh token from storage
func (s *authService) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	cacheKey := fmt.Sprintf("refresh_token:%s", tokenHash)
	tokenData, err := s.cacheSvc.GetString(ctx, cacheKey)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found")
	}

	// Parse and convert to RefreshToken struct
	parts := strings.Split(tokenData, ":")
	if len(parts) != 4 {
		return nil, fmt.Errorf("invalid refresh token data")
	}

	userIDStr, _, tokenHash, expiryStr := parts[0], parts[1], parts[2], parts[3]
	expiry, _ := strconv.ParseInt(expiryStr, 10, 64)

	return &models.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    userIDStr,
		Token:     tokenHash,
		TokenHash: tokenHash,
		ExpiresAt: time.Unix(expiry, 0),
		CreatedAt: time.Now(),
	}, nil
}

// RevokeUserTokens revokes all tokens for a user
func (s *authService) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error {
	// In production, this would query database for all user tokens
	log.Printf("Revoking all tokens for user %s", userID.String())
	// Implementation would batch delete all user tokens from cache/database
	return nil
}

// CleanupExpiredTokens removes expired tokens from storage
func (s *authService) CleanupExpiredTokens(ctx context.Context) error {
	log.Println("Cleaning up expired tokens")
	// In production, this would query and delete expired tokens
	return nil
}

// GeneratePasswordResetToken creates a time-limited password reset token for a user
func (s *authService) GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token := s.generateSecureToken()
	tokenHash, err := s.hashToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to hash password reset token: %v", err)
	}

	cacheKey := fmt.Sprintf("password_reset:%s", tokenHash)
	if err := s.cacheSvc.SetString(ctx, cacheKey, userID.String(), s.passwordResetTTL); err != nil {
		return "", fmt.Errorf("failed to store password reset token: %v", err)
	}

	return token, nil
}

// ValidatePasswordResetToken validates a password reset token without consuming it
func (s *authService) ValidatePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	tokenHash, err := s.hashToken(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password reset token: %v", err)
	}

	cacheKey := fmt.Sprintf("password_reset:%s", tokenHash)
	userIDStr, err := s.cacheSvc.GetString(ctx, cacheKey)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid or expired password reset token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in password reset token")
	}

	return userID, nil
}

// ConsumePasswordResetToken validates and removes a password reset token
func (s *authService) ConsumePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	userID, err := s.ValidatePasswordResetToken(ctx, token)
	if err != nil {
		return uuid.Nil, err
	}

	tokenHash, hashErr := s.hashToken(token)
	if hashErr != nil {
		return uuid.Nil, fmt.Errorf("failed to hash password reset token: %v", hashErr)
	}

	cacheKey := fmt.Sprintf("password_reset:%s", tokenHash)
	if err := s.cacheSvc.Delete(ctx, cacheKey); err != nil {
		log.Printf("Failed to delete password reset token from cache: %v", err)
	}

	return userID, nil
}

// GenerateEmailVerificationToken creates a verification token for email confirmation
func (s *authService) GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	token := s.generateSecureToken()
	tokenHash, err := s.hashToken(token)
	if err != nil {
		return "", fmt.Errorf("failed to hash email verification token: %v", err)
	}

	cacheKey := fmt.Sprintf("email_verify:%s", tokenHash)
	if err := s.cacheSvc.SetString(ctx, cacheKey, userID.String(), s.emailVerificationTTL); err != nil {
		return "", fmt.Errorf("failed to store email verification token: %v", err)
	}

	return token, nil
}

// ConsumeEmailVerificationToken validates and removes an email verification token
func (s *authService) ConsumeEmailVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	tokenHash, err := s.hashToken(token)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to hash email verification token: %v", err)
	}

	cacheKey := fmt.Sprintf("email_verify:%s", tokenHash)
	userIDStr, err := s.cacheSvc.GetString(ctx, cacheKey)
	if err != nil || userIDStr == "" {
		return uuid.Nil, fmt.Errorf("invalid or expired email verification token")
	}

	if err := s.cacheSvc.Delete(ctx, cacheKey); err != nil {
		log.Printf("Failed to delete email verification token from cache: %v", err)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in email verification token")
	}

	return userID, nil
}

// IsAccountLocked returns whether the user account is currently locked due to repeated failures
func (s *authService) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	lockKey := fmt.Sprintf("login_lock:%s", userID.String())
	val, err := s.cacheSvc.GetString(ctx, lockKey)
	if err != nil {
		return false, err
	}
	return val != "", nil
}

// RegisterFailedLoginAttempt increments failed attempt counters and locks the account if threshold reached
func (s *authService) RegisterFailedLoginAttempt(ctx context.Context, userID uuid.UUID) (int, bool, error) {
	attemptsKey := fmt.Sprintf("login_attempts:%s", userID.String())
	current := 0
	if val, err := s.cacheSvc.GetString(ctx, attemptsKey); err == nil && val != "" {
		if parsed, parseErr := strconv.Atoi(val); parseErr == nil {
			current = parsed
		}
	} else if err != nil {
		return 0, false, err
	}

	current++
	if err := s.cacheSvc.SetString(ctx, attemptsKey, strconv.Itoa(current), s.loginAttemptWindow); err != nil {
		return current, false, err
	}

	if current >= s.maxLoginAttempts {
		lockKey := fmt.Sprintf("login_lock:%s", userID.String())
		if err := s.cacheSvc.SetString(ctx, lockKey, "locked", s.lockoutDuration); err != nil {
			return current, true, err
		}
		return current, true, nil
	}

	return current, false, nil
}

// ClearFailedLoginAttempts clears counters and removes any lock for the user
func (s *authService) ClearFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	attemptsKey := fmt.Sprintf("login_attempts:%s", userID.String())
	lockKey := fmt.Sprintf("login_lock:%s", userID.String())

	if err := s.cacheSvc.Delete(ctx, attemptsKey); err != nil {
		return err
	}
	if err := s.cacheSvc.Delete(ctx, lockKey); err != nil {
		return err
	}
	return nil
}

// Helper methods

// generateSecureToken generates a cryptographically secure random token
func (s *authService) generateSecureToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// hashToken creates a SHA-256 hash of the token for secure storage
func (s *authService) hashToken(token string) (string, error) {
	hasher := sha256.New()
	_, err := hasher.Write([]byte(token))
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
