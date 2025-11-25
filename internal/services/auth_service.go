package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/models"

	"agromart2/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
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

	// User management
	Signup(ctx context.Context, email, password, firstName, lastName string, tenantID *uuid.UUID) (*models.User, error)
	HashPassword(password string) (string, error)
	VerifyPassword(hashedPassword, password string) error

	// TOTP 2FA
	Generate2FASecret(ctx context.Context, userID uuid.UUID) (*otp.Key, error)
	Enable2FA(ctx context.Context, userID uuid.UUID, code string) error
	Disable2FA(ctx context.Context, userID uuid.UUID) error
	Verify2FACode(ctx context.Context, userID uuid.UUID, code string) (bool, error)

	// Google Auth
	GetGoogleAuthURL(state string) string
	HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error)
	CompleteGoogleSignup(ctx context.Context, email, googleID, firstName, lastName, tenantName, subdomain string) (*models.User, error)
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
	userRepo             repositories.UserRepository
	tenantRepo           repositories.TenantRepository
	roleRepo             repositories.RoleRepository
	userRoleRepo         repositories.UserRoleRepository
	rolePermissionRepo   repositories.RolePermissionRepository
	permissionRepo       repositories.PermissionRepository
	googleOAuthConfig    *oauth2.Config
	frontendURL          string
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
func NewAuthService(
	cacheSvc caching.CacheService,
	jwtSecret string,
	tokenTTLSeconds, refreshTTLSeconds int,
	userRepo repositories.UserRepository,
	tenantRepo repositories.TenantRepository,
	roleRepo repositories.RoleRepository,
	userRoleRepo repositories.UserRoleRepository,
	rolePermissionRepo repositories.RolePermissionRepository,
	permissionRepo repositories.PermissionRepository,
	frontendURL string,
	backendURL string,
) AuthService {
	if backendURL == "" {
		backendURL = "http://localhost:8080" // Default for backend if not provided
	}

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
		userRepo:             userRepo,
		tenantRepo:           tenantRepo,
		roleRepo:             roleRepo,
		userRoleRepo:         userRoleRepo,
		rolePermissionRepo:   rolePermissionRepo,
		permissionRepo:       permissionRepo,
		frontendURL:          frontendURL,
		googleOAuthConfig: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  backendURL + "/v1/auth/google/callback",
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
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
		// Verify the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %v", err)
	}

	if claims, ok := jwtToken.Claims.(*TokenClaims); ok && jwtToken.Valid {
		// Check if token is blacklisted
		blacklistKey := fmt.Sprintf("token_blacklist:%s", claims.TokenID)
		if val, err := s.cacheSvc.GetString(ctx, blacklistKey); err == nil && val == "revoked" {
			return nil, fmt.Errorf("token has been revoked")
		}
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
	if err := s.cacheSvc.SetString(ctx, blacklistKey, "revoked", time.Until(claims.ExpiresAt.Time)); err != nil {
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
	if _, err := rand.Read(bytes); err != nil {
		// This should never happen with crypto/rand, but handle it gracefully
		log.Printf("CRITICAL: Failed to generate secure random bytes: %v", err)
		// Fallback: use timestamp + UUID as entropy source (not ideal but better than panic)
		fallback := fmt.Sprintf("%d-%s", time.Now().UnixNano(), uuid.New().String())
		return base64.URLEncoding.EncodeToString([]byte(fallback))
	}
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

// HashPassword hashes a password using bcrypt
func (s *authService) HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %v", err)
	}
	return string(hashedPassword), nil
}

// VerifyPassword verifies a password against its hash
func (s *authService) VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Signup handles user registration with tenant creation/lookup
func (s *authService) Signup(ctx context.Context, email, password, firstName, lastName string, tenantID *uuid.UUID) (*models.User, error) {
	userID := uuid.New()

	var (
		tenant uuid.UUID
		err    error
	)

	if tenantID != nil && *tenantID != uuid.Nil {
		tenant = *tenantID
	} else {
		tenant, err = s.getOrCreateTenantForSignup(ctx, email)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create tenant: %v", err)
		}
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, tenant, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user already exists")
	}

	// Hash password
	hashedPassword, err := s.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &models.User{
		ID:           userID,
		TenantID:     tenant,
		Email:        email,
		PasswordHash: hashedPassword,
		FirstName:    firstName,
		LastName:     lastName,
		Status:       "pending_verification",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("user already exists with this email in the tenant")
		}
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// Ensure default roles and assign user role
	adminRole, userRole, err := s.ensureTenantDefaults(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure default roles: %v", err)
	}

	if userRole == nil || adminRole == nil {
		return nil, fmt.Errorf("default roles not found")
	}

	// Check if this is the first user in the tenant
	// If so, assign admin role. Otherwise, assign user role.
	isFirstUser := false
	existingUsers, err := s.userRepo.List(ctx, tenant, 1, 0) // Check if any users exist
	if err != nil {
		return nil, fmt.Errorf("failed to check existing users: %v", err)
	}
	if len(existingUsers) == 0 {
		isFirstUser = true
	}

	roleID := userRole.ID
	if isFirstUser {
		roleID = adminRole.ID
	}

	newUserRole := &models.UserRole{
		UserID: userID,
		RoleID: roleID,
	}
	if err := s.userRoleRepo.Create(ctx, tenant, newUserRole); err != nil {
		return nil, fmt.Errorf("failed to assign role to user: %v", err)
	}

	// Generate verification token
	token, err := s.GenerateEmailVerificationToken(ctx, userID)
	if err != nil {
		// Log error but don't fail signup, user can request new token later
		log.Printf("Failed to generate verification token for user %s: %v", userID, err)
	} else {
		// Send verification email
		SendVerificationEmailAsync(ctx, email, token, s.frontendURL)
	}

	return user, nil
}

// getOrCreateTenantForSignup gets existing tenant by email domain or creates new one
func (s *authService) getOrCreateTenantForSignup(ctx context.Context, email string) (uuid.UUID, error) {
	domain := extractDomainFromEmail(email)

	personalDomains := map[string]struct{}{
		"gmail.com":      {},
		"yahoo.com":      {},
		"hotmail.com":    {},
		"outlook.com":    {},
		"icloud.com":     {},
		"protonmail.com": {},
	}

	if _, isPersonal := personalDomains[domain]; isPersonal {
		localPart := extractLocalPartFromEmail(email)
		name := buildTenantName(localPart, "Personal Workspace")
		return s.createTenant(ctx, name, localPart)
	}

	sanitizedDomain := sanitizeSubdomain(domain)
	if tenant, err := s.tenantRepo.GetBySubdomain(ctx, sanitizedDomain); err == nil {
		if _, _, ensureErr := s.ensureTenantDefaults(ctx, tenant.ID); ensureErr != nil {
			return uuid.Nil, ensureErr
		}
		return tenant.ID, nil
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, err
	}

	name := buildTenantName(domain, "Business Workspace")
	return s.createTenant(ctx, name, domain)
}

// createTenant creates a new tenant with unique subdomain
func (s *authService) createTenant(ctx context.Context, name, baseSubdomain string) (uuid.UUID, error) {
	sanitizedBase := sanitizeSubdomain(baseSubdomain)
	if sanitizedBase == "" {
		sanitizedBase = "tenant"
	}

	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = "Agromart Tenant"
	} else {
		displayName = buildTenantName(displayName, "Agromart Tenant")
	}

	for i := 0; i < 5; i++ {
		candidateSubdomain := sanitizedBase
		if i > 0 {
			candidateSubdomain = fmt.Sprintf("%s-%d", sanitizedBase, i)
		}

		tenant := &models.Tenant{
			ID:        uuid.New(),
			Name:      displayName,
			Subdomain: candidateSubdomain,
			License:   "",
			Status:    "active",
		}

		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				continue
			}
			return uuid.Nil, err
		}

		if _, _, err := s.ensureTenantDefaults(ctx, tenant.ID); err != nil {
			return uuid.Nil, err
		}

		return tenant.ID, nil
	}

	uniqueSubdomain := fmt.Sprintf("%s-%s", sanitizedBase, uuid.New().String()[:8])
	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      displayName,
		Subdomain: uniqueSubdomain,
		License:   "",
		Status:    "active",
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return uuid.Nil, err
	}

	if _, _, err := s.ensureTenantDefaults(ctx, tenant.ID); err != nil {
		return uuid.Nil, err
	}

	return tenant.ID, nil
}

// ensureTenantDefaults ensures default roles and permissions exist for a tenant
func (s *authService) ensureTenantDefaults(ctx context.Context, tenantID uuid.UUID) (*models.Role, *models.Role, error) {
	// 1. Ensure 'user' role exists
	userRole, err := s.roleRepo.GetByName(ctx, tenantID, "user")
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) && !strings.Contains(err.Error(), "role not found") {
			return nil, nil, err
		}

		description := "Regular user role"
		userRole = &models.Role{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "user",
			Description: &description,
		}

		if err := s.roleRepo.Create(ctx, userRole); err != nil {
			return nil, nil, err
		}
	}

	// 2. Ensure 'admin' role exists
	adminRole, err := s.roleRepo.GetByName(ctx, tenantID, "admin")
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) && !strings.Contains(err.Error(), "role not found") {
			return nil, nil, err
		}

		description := "Administrator role with full access"
		adminRole = &models.Role{
			ID:          uuid.New(),
			TenantID:    tenantID,
			Name:        "admin",
			Description: &description,
		}

		if err := s.roleRepo.Create(ctx, adminRole); err != nil {
			return nil, nil, err
		}
	}

	// 3. Assign permissions
	if s.permissionRepo != nil && s.rolePermissionRepo != nil {
		allPermissions, err := s.permissionRepo.ListPermissions(ctx)
		if err != nil {
			return nil, nil, err
		}

		for _, permission := range allPermissions {
			name := permission.Name

			// --- Admin Role Permissions ---
			// Admin gets EVERYTHING
			if err := s.rolePermissionRepo.Create(ctx, tenantID, &models.RolePermission{
				RoleID:       adminRole.ID,
				PermissionID: permission.ID,
			}); err != nil {
				// Ignore duplicate key errors (permission already assigned)
				var pgErr *pgconn.PgError
				if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
					return nil, nil, err
				}
			}

			// --- User Role Permissions ---
			// Grant specific permissions for core business entities to default user role
			shouldGrantToUser := false

			// Core modules that users should have full access to
			// UPDATED: Using singular + dot notation to match DB seeds
			coreModules := []string{
				"product.", "inventory.", "warehouse.",
				"distributor.", "supplier.", "order.", "invoice.",
				"category.",
			}

			for _, module := range coreModules {
				if strings.HasPrefix(name, module) {
					shouldGrantToUser = true
					break
				}
			}

			// Explicitly EXCLUDE sensitive modules for regular users
			sensitiveModules := []string{
				"user.", "tenant.", "analytics.", "role.", "permission.",
			}
			for _, module := range sensitiveModules {
				if strings.HasPrefix(name, module) {
					shouldGrantToUser = false
					break
				}
			}

			// Also keep legacy read_ prefix support but be careful
			if !shouldGrantToUser && strings.HasPrefix(name, "read_") {
				shouldGrantToUser = true
			}

			if shouldGrantToUser {
				if err := s.rolePermissionRepo.Create(ctx, tenantID, &models.RolePermission{
					RoleID:       userRole.ID,
					PermissionID: permission.ID,
				}); err != nil {
					// Ignore duplicate key errors
					var pgErr *pgconn.PgError
					if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
						return nil, nil, err
					}
				}
			}
		}
	}

	return adminRole, userRole, nil
}

// Utility functions

// extractDomainFromEmail extracts domain from email address
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}

// extractLocalPartFromEmail extracts local part from email address
func extractLocalPartFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

// sanitizeSubdomain sanitizes a string for use as a subdomain
func sanitizeSubdomain(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	if value == "" {
		return ""
	}

	var builder strings.Builder
	lastWasHyphen := false

	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
			lastWasHyphen = false
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
			lastWasHyphen = false
		case r == '-' || r == '_' || r == '.' || r == ' ':
			if !lastWasHyphen {
				builder.WriteRune('-')
				lastWasHyphen = true
			}
		}
	}

	sanitized := strings.Trim(builder.String(), "-")
	return sanitized
}

// buildTenantName builds a tenant name from source string with fallback
func buildTenantName(source string, fallback string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return fallback
	}

	parts := strings.FieldsFunc(source, func(r rune) bool {
		return r == '.' || r == '-' || r == '_' || r == '@' || r == ' '
	})

	if len(parts) == 0 {
		return fallback
	}

	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		lower := strings.ToLower(part)
		parts[i] = strings.ToUpper(lower[:1]) + lower[1:]
	}

	return strings.Join(parts, " ")
}

// Generate2FASecret generates a new TOTP secret for a user
func (s *authService) Generate2FASecret(ctx context.Context, userID uuid.UUID) (*otp.Key, error) {
	tenantID, err := s.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant ID: %v", err)
	}

	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if user.TwoFactorEnabled {
		return nil, fmt.Errorf("2FA is already enabled for this user")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Agromart",
		AccountName: user.Email,
		SecretSize:  32,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP secret: %v", err)
	}

	// Save the secret to the user record
	user.TwoFactorSecret = key.Secret()
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save TOTP secret: %v", err)
	}

	return key, nil
}

// Enable2FA enables 2FA for a user after verifying the TOTP code
func (s *authService) Enable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	tenantID, err := s.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get tenant ID: %v", err)
	}

	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.TwoFactorEnabled {
		return fmt.Errorf("2FA is already enabled for this user")
	}

	if user.TwoFactorSecret == "" {
		return fmt.Errorf("no TOTP secret found for user")
	}

	valid := totp.Validate(code, user.TwoFactorSecret)
	if !valid {
		return fmt.Errorf("invalid TOTP code")
	}

	user.TwoFactorEnabled = true
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to enable 2FA: %v", err)
	}

	return nil
}

// Disable2FA disables 2FA for a user
func (s *authService) Disable2FA(ctx context.Context, userID uuid.UUID) error {
	tenantID, err := s.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get tenant ID: %v", err)
	}

	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	user.TwoFactorEnabled = false
	user.TwoFactorSecret = ""
	user.UpdatedAt = time.Now()
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to disable 2FA: %v", err)
	}

	return nil
}

// Verify2FACode verifies a TOTP code for a user
func (s *authService) Verify2FACode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	tenantID, err := s.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get tenant ID: %v", err)
	}

	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %v", err)
	}
	if user == nil {
		return false, fmt.Errorf("user not found")
	}

	if !user.TwoFactorEnabled || user.TwoFactorSecret == "" {
		return false, fmt.Errorf("2FA is not enabled for this user")
	}

	valid := totp.Validate(code, user.TwoFactorSecret)
	return valid, nil
}
