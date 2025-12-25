package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"
	"agromart2/internal/validation"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
)

// twoFAAttemptTracker tracks 2FA verification attempts per user/IP for rate limiting
type twoFAAttemptTracker struct {
	mu          sync.RWMutex
	attempts    map[string]*attemptInfo
	cancel      context.CancelFunc
	cleanupOnce sync.Once
}

type attemptInfo struct {
	count    int
	firstAt  time.Time
	lockedAt *time.Time
}

const (
	// max2FAAttempts is the maximum number of 2FA attempts allowed within the window
	max2FAAttempts = 5
	// twoFAAttemptWindow is the time window for tracking attempts
	twoFAAttemptWindow = 15 * time.Minute
	// twoFALockoutDuration is how long to lock out after max attempts
	twoFALockoutDuration = 30 * time.Minute
)

var globalTwoFATracker = &twoFAAttemptTracker{
	attempts: make(map[string]*attemptInfo),
}

// init starts the cleanup goroutine for the global 2FA tracker
func init() {
	ctx, cancel := context.WithCancel(context.Background())
	globalTwoFATracker.cancel = cancel
	go globalTwoFATracker.startCleanup(ctx)
}

func (t *twoFAAttemptTracker) checkAndIncrement(key string) (allowed bool, remainingAttempts int, lockoutRemaining time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	info, exists := t.attempts[key]

	// Clean up expired entries
	if exists {
		// Check if locked
		if info.lockedAt != nil {
			lockRemaining := twoFALockoutDuration - now.Sub(*info.lockedAt)
			if lockRemaining > 0 {
				return false, 0, lockRemaining
			}
			// Lock expired, reset
			delete(t.attempts, key)
			exists = false
		} else if now.Sub(info.firstAt) > twoFAAttemptWindow {
			// Window expired, reset
			delete(t.attempts, key)
			exists = false
		}
	}

	if !exists {
		t.attempts[key] = &attemptInfo{
			count:   1,
			firstAt: now,
		}
		return true, max2FAAttempts - 1, 0
	}

	info.count++
	if info.count >= max2FAAttempts {
		// Lock the account
		nowCopy := now
		info.lockedAt = &nowCopy
		return false, 0, twoFALockoutDuration
	}

	return true, max2FAAttempts - info.count, 0
}

func (t *twoFAAttemptTracker) clearAttempts(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.attempts, key)
}

// startCleanup runs a periodic cleanup of expired 2FA attempt entries
func (t *twoFAAttemptTracker) startCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			t.cleanup()
		case <-ctx.Done():
			return
		}
	}
}

// Stop stops the cleanup goroutine
func (t *twoFAAttemptTracker) Stop() {
	t.cleanupOnce.Do(func() {
		if t.cancel != nil {
			t.cancel()
		}
	})
}

// cleanup removes expired entries from the 2FA attempt tracker
func (t *twoFAAttemptTracker) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for key, info := range t.attempts {
		// Remove locked entries that have expired
		if info.lockedAt != nil {
			if now.Sub(*info.lockedAt) > twoFALockoutDuration {
				delete(t.attempts, key)
			}
		} else if now.Sub(info.firstAt) > twoFAAttemptWindow {
			// Remove attempt windows that have expired
			delete(t.attempts, key)
		}
	}
}

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService         services.AuthService
	userRepo            repositories.UserRepository
	tenantRepo          repositories.TenantRepository
	roleRepo            repositories.RoleRepository
	userRoleRepo        repositories.UserRoleRepository
	rolePermissionRepo  repositories.RolePermissionRepository
	permissionRepo      repositories.PermissionRepository
	rbacMiddleware      *middleware.RBACMiddleware
	notificationService services.NotificationService
	frontendBaseURL     string
	ssoService          services.SSOService
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(
	authService services.AuthService,
	userRepo repositories.UserRepository,
	tenantRepo repositories.TenantRepository,
	roleRepo repositories.RoleRepository,
	userRoleRepo repositories.UserRoleRepository,
	rolePermissionRepo repositories.RolePermissionRepository,
	permissionRepo repositories.PermissionRepository,
	rbacMiddleware *middleware.RBACMiddleware,
	notificationService services.NotificationService,
	frontendBaseURL string,
	ssoService services.SSOService,
) *AuthHandlers {
	return &AuthHandlers{
		authService:         authService,
		userRepo:            userRepo,
		tenantRepo:          tenantRepo,
		roleRepo:            roleRepo,
		userRoleRepo:        userRoleRepo,
		rolePermissionRepo:  rolePermissionRepo,
		permissionRepo:      permissionRepo,
		rbacMiddleware:      rbacMiddleware,
		notificationService: notificationService,
		frontendBaseURL:     frontendBaseURL,
		ssoService:          ssoService,
	}
}

// LoginResponse represents the login response
type LoginResponse struct {
	models.TokenResponse
	User *models.User `json:"user"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email,max=255" sanitize:"trim,lower"`
	Password string `json:"password" validate:"required,min=8,max=128" sanitize:"trim"`
}

// Login handles user login with email and password
func (h *AuthHandlers) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Get user by email - search across all tenants dynamically
	// This approach retrieves all users with the given email across tenants
	user, err := h.getUserByEmailAcrossTenants(ctx, req.Email)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	}

	if user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
	}

	// Check if user has a password hash (handle users created with previous bug)
	// This provides a graceful error message for accounts that may have been
	// created before password hashing was properly implemented
	if user.PasswordHash == "" {
		log.Printf("WARNING: User %s has no password hash - account integrity issue", user.ID.String())
		return echo.NewHTTPError(http.StatusUnauthorized, "Account not properly initialized. Please contact support or try signing up again.")
	}

	isLocked, err := h.authService.IsAccountLocked(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to check lock status for user %s: %v", user.ID.String(), err)
	} else if isLocked {
		return echo.NewHTTPError(http.StatusTooManyRequests, "Account temporarily locked due to multiple failed login attempts. Please try again later.")
	}

	// Verify password
	if err := h.authService.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		if _, locked, attemptErr := h.authService.RegisterFailedLoginAttempt(ctx, user.ID); attemptErr != nil {
			log.Printf("Failed to record login attempt for user %s: %v", user.ID.String(), attemptErr)
		} else if locked {
			return echo.NewHTTPError(http.StatusTooManyRequests, "Account temporarily locked due to multiple failed login attempts. Please try again later.")
		}
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
	}

	if err := h.authService.ClearFailedLoginAttempts(ctx, user.ID); err != nil {
		log.Printf("Failed to clear login attempts for user %s: %v", user.ID.String(), err)
	}

	if user.Status == "pending_approval" {
		return echo.NewHTTPError(http.StatusForbidden, "Your account is pending approval by an administrator. You will be notified once approved.")
	}

	if user.Status != "active" {
		return echo.NewHTTPError(http.StatusForbidden, "Email not verified. Please check your inbox for the verification link.")
	}

	// Get tenant ID for the user
	tenantID, err := h.userRepo.GetTenantIDByUserID(ctx, user.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "User tenant not found")
	}

	// Check if user has 2FA enabled
	if user.TwoFactorEnabled {
		// Generate a temporary token for 2FA verification
		tempToken, err := h.authService.GeneratePasswordResetToken(ctx, user.ID)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate temporary token")
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"requires_2fa": true,
			"temp_token":   tempToken,
			"message":      "2FA verification required",
		})
	}

	// Generate our internal JWT tokens
	tokenResponse, err := h.authService.GenerateTokens(ctx, user.ID, tenantID, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate tokens")
	}

	// Set HttpOnly cookie with JWT token
	// Default to secure=true unless explicitly in development mode
	isSecure := os.Getenv("ENV") != "development"
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = tokenResponse.AccessToken
	cookie.HttpOnly = true
	cookie.Secure = isSecure
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Path = "/"
	cookie.MaxAge = tokenResponse.ExpiresIn // Match JWT token expiration
	c.SetCookie(cookie)

	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = tokenResponse.RefreshToken
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = isSecure
	refreshCookie.SameSite = http.SameSiteStrictMode
	refreshCookie.Path = "/"
	refreshCookie.MaxAge = 604800 // 7 days
	c.SetCookie(refreshCookie)

	// Get user role
	userRoles, err := h.userRoleRepo.ListByUser(ctx, tenantID, user.ID)
	if err == nil && len(userRoles) > 0 {
		role, err := h.roleRepo.GetByID(ctx, tenantID, userRoles[0].RoleID)
		if err == nil {
			user.Role = role.Name
		}
	}

	response := LoginResponse{
		TokenResponse: *tokenResponse,
		User:          user,
	}

	return c.JSON(http.StatusOK, response)
}

// StartSSO initiates the SSO login flow by returning the provider authorization URL.
func (h *AuthHandlers) StartSSO(c echo.Context) error {
	if h.ssoService == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "SSO is not configured")
	}

	ctx := c.Request().Context()
	tenantID := c.QueryParam("tenant_id")
	if tenantID == "" {
		if tid, ok := common.GetTenantIDFromContext(ctx); ok {
			tenantID = tid.String()
		}
	}
	if tenantID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required")
	}

	authURL, err := h.ssoService.GetAuthURL(ctx, tenantID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"auth_url": authURL,
	})
}

// HandleSSOCallback completes the SSO flow and issues internal JWT tokens.
func (h *AuthHandlers) HandleSSOCallback(c echo.Context) error {
	if h.ssoService == nil {
		return echo.NewHTTPError(http.StatusNotImplemented, "SSO is not configured")
	}

	ctx := c.Request().Context()
	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "authorization code is required")
	}

	tenantID := c.QueryParam("tenant_id")
	if tenantID == "" {
		if tid, ok := common.GetTenantIDFromContext(ctx); ok {
			tenantID = tid.String()
		}
	}
	if tenantID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "tenant_id is required")
	}

	user, err := h.ssoService.HandleCallback(ctx, tenantID, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	tokenResponse, err := h.authService.GenerateTokens(ctx, user.ID, user.TenantID, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate tokens")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token": tokenResponse,
		"user":  user,
	})
}

// TestEmailRequest represents the test email request payload
type TestEmailRequest struct {
	Email string `json:"email" validate:"required,email" sanitize:"trim,lower"`
}

// TestEmailSending tests email functionality by sending a test email
func (h *AuthHandlers) TestEmailSending(c echo.Context) error {
	ctx := c.Request().Context()

	var req TestEmailRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Generate a secure random token for email verification
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return common.SendServerError(c, "Failed to generate verification token")
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	errCh := services.SendVerificationEmailAsync(ctx, req.Email, token, h.frontendBaseURL)
	go func(recipient string) {
		if err := <-errCh; err != nil {
			log.Printf("Test email failed for %s: %v", recipient, err)
		} else {
			log.Printf("Test email sent successfully to %s", recipient)
		}
	}(req.Email)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Test email sent. Please check your inbox and spam folder.",
		"email":   req.Email,
		"note":    "This is a test email to verify your email configuration is working correctly.",
	})
}

// SignupRequest represents the signup request payload
type SignupRequest struct {
	Email     string  `json:"email" validate:"required,email,max=255" sanitize:"trim,lower"`
	Password  string  `json:"password" validate:"required,min=8,max=128" sanitize:"trim"`
	FirstName string  `json:"first_name" validate:"required,min=2,max=100" sanitize:"trim,html"`
	LastName  string  `json:"last_name" validate:"required,min=2,max=100" sanitize:"trim,html"`
	TenantID  *string `json:"tenant_id" validate:"omitempty,uuid4" sanitize:"trim"`
}

// SignupResponse represents the signup response
type SignupResponse struct {
	User                 *models.User `json:"user"`
	Message              string       `json:"message"`
	VerificationRequired bool         `json:"verification_required"`
}

// ForgotPasswordRequest represents password reset initiation request
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" sanitize:"trim,lower"`
}

// ResetPasswordRequest represents password reset confirmation
type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required" sanitize:"trim"`
	Password        string `json:"password" validate:"required,min=8,max=128" sanitize:"trim"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password" sanitize:"trim"`
}

// VerifyEmailRequest represents email verification payload
type VerifyEmailRequest struct {
	Token string `json:"token" validate:"required" sanitize:"trim"`
}

// Signup handles user registration
func (h *AuthHandlers) Signup(c echo.Context) error {
	ctx := c.Request().Context()

	var req SignupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)
	if req.TenantID != nil && *req.TenantID == "" {
		req.TenantID = nil
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate password strength
	if err := common.ValidatePassword(req.Password); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Parse tenant ID if provided
	var tenantID *uuid.UUID
	if req.TenantID != nil && *req.TenantID != "" {
		tid, err := uuid.Parse(*req.TenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID format")
		}
		if tid == uuid.Nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID")
		}
		tenantID = &tid
	}

	// Call the auth service to handle signup
	user, err := h.authService.Signup(ctx, req.Email, req.Password, req.FirstName, req.LastName, tenantID)
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			return echo.NewHTTPError(http.StatusConflict, "User already exists")
		}
		log.Printf("Failed to create user account for %s: %v", req.Email, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create account")
	}

	verificationToken, err := h.authService.GenerateEmailVerificationToken(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to create verification token for user %s: %v", user.Email, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate email verification")
	}

	errCh := services.SendVerificationEmailAsync(ctx, user.Email, verificationToken, h.frontendBaseURL)
	go func(recipient string) {
		if err := <-errCh; err != nil {
			log.Printf("CRITICAL: Verification email dispatch failed for %s: %v", recipient, err)
			// Email service now implements retry mechanism with 3 attempts and exponential backoff
			// Admin is automatically notified on failure if ADMIN_EMAIL is configured
			// User account is still created and user can request a new verification email later
		} else {
			log.Printf("Verification email sent successfully to %s", recipient)
		}
	}(user.Email)

	// For auto-approved users (e.g., first user/admin), generate tokens and set cookies
	if user.Status == "active" {
		// Get tenant ID
		userTenantID, err := h.userRepo.GetTenantIDByUserID(ctx, user.ID)
		if err == nil {
			tokenResponse, err := h.authService.GenerateTokens(ctx, user.ID, userTenantID, nil)
			if err == nil {
				// Set HttpOnly cookie with JWT token
				cookie := new(http.Cookie)
				cookie.Name = "auth_token"
				cookie.Value = tokenResponse.AccessToken
				cookie.HttpOnly = true
				cookie.Secure = os.Getenv("ENV") != "development"
				cookie.SameSite = http.SameSiteStrictMode
				cookie.Path = "/"
				cookie.MaxAge = tokenResponse.ExpiresIn
				c.SetCookie(cookie)

				// Set refresh token cookie
				refreshCookie := new(http.Cookie)
				refreshCookie.Name = "refresh_token"
				refreshCookie.Value = tokenResponse.RefreshToken
				refreshCookie.HttpOnly = true
				refreshCookie.Secure = os.Getenv("ENV") != "development"
				refreshCookie.SameSite = http.SameSiteStrictMode
				refreshCookie.Path = "/"
				refreshCookie.MaxAge = 604800 // 7 days
				c.SetCookie(refreshCookie)
			}
		}
	}

	response := SignupResponse{
		User:                 user,
		Message:              "Account created. Please verify your email to activate access.",
		VerificationRequired: true,
	}

	return c.JSON(http.StatusCreated, response)
}

// LogoutRequest represents the logout request payload
type LogoutRequest struct {
	TokenTypeHint *string `json:"token_type_hint"` // "access_token" or "refresh_token"
}

// Logout handles user logout by revoking tokens and clearing cookies
func (h *AuthHandlers) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Get the token from Authorization header or cookie
	var tokenString string
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		tokenString = strings.TrimPrefix(authHeader, "Bearer ")
	} else {
		// Try to get token from cookie
		cookie, err := c.Cookie("auth_token")
		if err == nil && cookie != nil {
			tokenString = cookie.Value
		}
	}

	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		// Bind is optional for logout, but we'll proceed with access token revocation
		req.TokenTypeHint = nil
	}

	// Revoke the access token if present
	if tokenString != "" {
		if err := h.authService.RevokeToken(ctx, tokenString, req.TokenTypeHint); err != nil {
			log.Printf("Failed to revoke token during logout: %v", err)
			// Continue with logout even if revocation fails
		}
	}

	// Clear auth_token cookie
	authCookie := new(http.Cookie)
	authCookie.Name = "auth_token"
	authCookie.Value = ""
	authCookie.HttpOnly = true
	authCookie.Secure = os.Getenv("ENV") != "development"
	authCookie.SameSite = http.SameSiteStrictMode
	authCookie.Path = "/"
	authCookie.MaxAge = -1 // Delete cookie
	c.SetCookie(authCookie)

	// Clear refresh_token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = ""
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = os.Getenv("ENV") != "development"
	refreshCookie.SameSite = http.SameSiteStrictMode
	refreshCookie.Path = "/"
	refreshCookie.MaxAge = -1 // Delete cookie
	c.SetCookie(refreshCookie)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// RefreshRequest represents the token refresh request payload
type RefreshRequest struct {
	RefreshToken string  `json:"refresh_token" sanitize:"trim"`
	GrantType    string  `json:"grant_type" validate:"required,oneof=refresh_token" sanitize:"trim"`
	ClientID     *string `json:"client_id" sanitize:"trim"`
	Scope        *string `json:"scope" sanitize:"trim"`
}

// Refresh handles token refresh
func (h *AuthHandlers) Refresh(c echo.Context) error {
	ctx := c.Request().Context()

	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	// If refresh_token not in body, try to get from cookie (for browser clients)
	if req.RefreshToken == "" {
		cookie, err := c.Cookie("refresh_token")
		if err == nil && cookie.Value != "" {
			req.RefreshToken = cookie.Value
		}
	}

	// Validate refresh token is present
	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing refresh token")
	}

	if req.ClientID != nil && *req.ClientID == "" {
		req.ClientID = nil
	}
	if req.Scope != nil && *req.Scope == "" {
		req.Scope = nil
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Refresh tokens
	tokenResponse, err := h.authService.RefreshToken(ctx, req.RefreshToken, req.ClientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token")
	}

	// Set new HttpOnly cookies with refreshed tokens
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = tokenResponse.AccessToken
	cookie.HttpOnly = true
	cookie.Secure = os.Getenv("ENV") != "development"
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Path = "/"
	cookie.MaxAge = tokenResponse.ExpiresIn
	c.SetCookie(cookie)

	// Set new refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = tokenResponse.RefreshToken
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = os.Getenv("ENV") != "development"
	refreshCookie.SameSite = http.SameSiteStrictMode
	refreshCookie.Path = "/"
	refreshCookie.MaxAge = 604800 // 7 days
	c.SetCookie(refreshCookie)

	return c.JSON(http.StatusOK, tokenResponse)
}

// ForgotPassword initiates password reset flow
func (h *AuthHandlers) ForgotPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var req ForgotPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	email := req.Email

	user, err := h.getUserByEmailAcrossTenants(ctx, email)
	if err != nil {
		log.Printf("Password reset lookup failed for %s: %v", email, err)
		return c.JSON(http.StatusOK, map[string]string{
			"message": "If the account exists, password reset instructions have been sent.",
		})
	}

	if user == nil {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "If the account exists, password reset instructions have been sent.",
		})
	}

	token, err := h.authService.GeneratePasswordResetToken(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to generate password reset token for %s: %v", user.Email, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate password reset")
	}

	resetURL := h.buildPasswordResetURL(token)
	emailBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>You requested to reset your Agromart password. Click the link below to set a new password. If you did not request this, you can safely ignore this email.</p><p><a href=\"%s\">Reset Password</a></p><p>This link will expire in 15 minutes.</p>",
		user.FirstName,
		resetURL,
	)

	if err := h.notificationService.SendEmail(ctx, user.TenantID, user.Email, "Password Reset Instructions", emailBody); err != nil {
		log.Printf("Failed to send password reset email to %s: %v", user.Email, err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "If the account exists, password reset instructions have been sent.",
	})
}

// ResetPassword completes password reset with provided token
func (h *AuthHandlers) ResetPassword(c echo.Context) error {
	ctx := c.Request().Context()

	var req ResetPasswordRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate password strength
	if err := common.ValidatePassword(req.Password); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	token := req.Token
	password := req.Password

	userID, err := h.authService.ConsumePasswordResetToken(ctx, token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired password reset token")
	}

	tenantID, err := h.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		log.Printf("Failed to get tenant for user %s: %v", userID.String(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}

	hashedPassword, err := h.authService.HashPassword(password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}

	if err := h.userRepo.UpdatePassword(ctx, tenantID, userID, string(hashedPassword)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to reset password")
	}

	if err := h.authService.RevokeUserTokens(ctx, userID); err != nil {
		log.Printf("Failed to revoke tokens for user %s: %v", userID.String(), err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Password has been reset successfully.",
	})
}

// VerifyEmail confirms a user's email address using a verification token
func (h *AuthHandlers) VerifyEmail(c echo.Context) error {
	ctx := c.Request().Context()

	var req VerifyEmailRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	token := req.Token

	userID, err := h.authService.ConsumeEmailVerificationToken(ctx, token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid or expired verification token")
	}

	tenantID, err := h.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		log.Printf("Failed to resolve tenant for user %s: %v", userID.String(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify email")
	}

	// Check if user has admin role
	isAdmin := false
	userRoles, err := h.userRoleRepo.ListByUser(ctx, tenantID, userID)
	if err == nil {
		for _, ur := range userRoles {
			role, err := h.roleRepo.GetByID(ctx, tenantID, ur.RoleID)
			if err == nil && role.Name == "admin" {
				isAdmin = true
				break
			}
		}
	}

	// Fallback: If no admin role found but this is the first user in tenant,
	// they should be admin (handles race condition or role assignment failure)
	if !isAdmin {
		isFirstUser, err := h.userRepo.IsFirstUserInTenant(ctx, tenantID)
		if err != nil {
			log.Printf("Warning: Failed to check first user status for tenant %s: %v", tenantID.String(), err)
		} else if isFirstUser {
			isAdmin = true
			log.Printf("First user in tenant %s - granting admin access on email verification", tenantID.String())
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				adminRole, err := h.roleRepo.GetByName(ctx, tenantID, "admin")
				if err != nil {
					log.Printf("Failed to get admin role for tenant %s: %v", tenantID.String(), err)
					return
				}
				newUserRole := &models.UserRole{
					UserID: userID,
					RoleID: adminRole.ID,
				}
				if err := h.userRoleRepo.Create(ctx, tenantID, newUserRole); err != nil {
					log.Printf("Failed to assign admin role to first user %s: %v (may already exist)", userID.String(), err)
				}
			}()
		}
	}

	newStatus := "pending_approval"
	if isAdmin {
		newStatus = "active"
	}

	if err := h.userRepo.UpdateStatus(ctx, tenantID, userID, newStatus); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to activate account")
	}

	if err := h.authService.ClearFailedLoginAttempts(ctx, userID); err != nil {
		log.Printf("Failed to clear login attempts for user %s after verification: %v", userID.String(), err)
	}

	message := "Email verified successfully. Your account is pending approval by an administrator."
	if isAdmin {
		message = "Email verified successfully. You can now sign in."
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": message,
	})
}

// Me handles getting current user profile
func (h *AuthHandlers) Me(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	// Get user details
	user, err := h.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	// Get user roles
	userRoles, err := h.userRoleRepo.ListByUser(ctx, tenantID, userID)
	if err == nil && len(userRoles) > 0 {
		role, err := h.roleRepo.GetByID(ctx, tenantID, userRoles[0].RoleID)
		if err == nil {
			user.Role = role.Name
		}
	}

	return c.JSON(http.StatusOK, user)
}

// Helper methods

// getUserByEmailAcrossTenants searches for a user by email across all tenants
// This is secure because we still validate password after finding the user
func (h *AuthHandlers) getUserByEmailAcrossTenants(ctx context.Context, email string) (*models.User, error) {
	user, err := h.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (h *AuthHandlers) buildPasswordResetURL(token string) string {
	base := strings.TrimRight(h.frontendBaseURL, "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return fmt.Sprintf("%s/reset-password?token=%s", base, url.QueryEscape(token))
}

func (h *AuthHandlers) buildEmailVerificationURL(token string) string {
	base := strings.TrimRight(h.frontendBaseURL, "/")
	if base == "" {
		base = "http://localhost:3000"
	}
	return fmt.Sprintf("%s/verify-email?token=%s", base, url.QueryEscape(token))
}

// Generate2FARequest represents the 2FA generation request
type Generate2FARequest struct{}

// Generate2FAResponse represents the 2FA generation response
type Generate2FAResponse struct {
	QRCodeURL string `json:"qr_code_url"`
}

// Generate2FA generates a new TOTP secret and returns QR code URL
func (h *AuthHandlers) Generate2FA(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	key, err := h.authService.Generate2FASecret(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate 2FA secret")
	}

	response := Generate2FAResponse{
		QRCodeURL: key.URL(),
	}

	return c.JSON(http.StatusOK, response)
}

// Enable2FARequest represents the 2FA enable request
type Enable2FARequest struct {
	Code string `json:"code" validate:"required,len=6" sanitize:"trim"`
}

// Enable2FA enables 2FA for the authenticated user
func (h *AuthHandlers) Enable2FA(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	var req Enable2FARequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	if err := h.authService.Enable2FA(ctx, userID, req.Code); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to enable 2FA")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA enabled successfully",
	})
}

// Disable2FA disables 2FA for the authenticated user
func (h *AuthHandlers) Disable2FA(c echo.Context) error {
	ctx := c.Request().Context()

	userID, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	if err := h.authService.Disable2FA(ctx, userID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to disable 2FA")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "2FA disabled successfully",
	})
}

// Verify2FARequest represents the 2FA verification request
type Verify2FARequest struct {
	Token string `json:"token" validate:"required" sanitize:"trim"`
	Code  string `json:"code" validate:"required,len=6" sanitize:"trim"`
}

// Verify2FAResponse represents the 2FA verification response
type Verify2FAResponse struct {
	models.TokenResponse
	User *models.User `json:"user"`
}

// Verify2FA verifies the 2FA code and returns JWT tokens
// This endpoint is rate-limited to prevent brute-force attacks on 6-digit TOTP codes
func (h *AuthHandlers) Verify2FA(c echo.Context) error {
	ctx := c.Request().Context()
	clientIP := c.RealIP()

	var req Verify2FARequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate the temporary token to get user ID first (before rate limiting)
	userID, err := h.authService.ValidatePasswordResetToken(ctx, req.Token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
	}

	// Rate limiting key combines user ID and IP to prevent both per-user and per-IP brute force
	rateLimitKey := fmt.Sprintf("2fa:%s:%s", userID.String(), clientIP)

	// Check rate limit before processing
	allowed, remaining, lockoutRemaining := globalTwoFATracker.checkAndIncrement(rateLimitKey)
	if !allowed {
		log.Printf("2FA rate limit exceeded for user %s from IP %s, lockout remaining: %v", userID.String(), clientIP, lockoutRemaining)
		return echo.NewHTTPError(http.StatusTooManyRequests, map[string]interface{}{
			"error":            "Too many failed 2FA attempts. Please try again later.",
			"retry_after_secs": int(lockoutRemaining.Seconds()),
		})
	}

	// Verify the 2FA code
	valid, err := h.authService.Verify2FACode(ctx, userID, req.Code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify 2FA code")
	}

	if !valid {
		log.Printf("Invalid 2FA code for user %s from IP %s, %d attempts remaining", userID.String(), clientIP, remaining)
		return echo.NewHTTPError(http.StatusUnauthorized, map[string]interface{}{
			"error":              "Invalid 2FA code",
			"attempts_remaining": remaining,
		})
	}

	// Success - clear the rate limit tracker
	globalTwoFATracker.clearAttempts(rateLimitKey)

	// Get tenant ID
	tenantID, err := h.userRepo.GetTenantIDByUserID(ctx, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user tenant")
	}

	// Get user details
	user, err := h.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user details")
	}

	// Generate JWT tokens
	tokenResponse, err := h.authService.GenerateTokens(ctx, userID, tenantID, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate tokens")
	}

	// Set HttpOnly cookies after 2FA verification
	cookie := new(http.Cookie)
	cookie.Name = "auth_token"
	cookie.Value = tokenResponse.AccessToken
	cookie.HttpOnly = true
	cookie.Secure = os.Getenv("ENV") != "development"
	cookie.SameSite = http.SameSiteStrictMode
	cookie.Path = "/"
	cookie.MaxAge = tokenResponse.ExpiresIn
	c.SetCookie(cookie)

	// Set refresh token cookie
	refreshCookie := new(http.Cookie)
	refreshCookie.Name = "refresh_token"
	refreshCookie.Value = tokenResponse.RefreshToken
	refreshCookie.HttpOnly = true
	refreshCookie.Secure = os.Getenv("ENV") != "development"
	refreshCookie.SameSite = http.SameSiteStrictMode
	refreshCookie.Path = "/"
	refreshCookie.MaxAge = 604800 // 7 days
	c.SetCookie(refreshCookie)

	// Consume the temporary token
	if _, err := h.authService.ConsumePasswordResetToken(ctx, req.Token); err != nil {
		log.Printf("Failed to consume temporary token: %v", err)
	}

	response := Verify2FAResponse{
		TokenResponse: *tokenResponse,
		User:          user,
	}

	return c.JSON(http.StatusOK, response)
}
