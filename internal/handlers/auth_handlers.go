package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

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
	Password string `json:"password" validate:"required,min=6,max=128" sanitize:"trim"`
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

	response := LoginResponse{
		TokenResponse: *tokenResponse,
		User:          user,
	}

	return c.JSON(http.StatusOK, response)
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

	// Generate a test token for demonstration
	testToken := "test-token-12345"

	errCh := services.SendVerificationEmailAsync(ctx, req.Email, testToken, h.frontendBaseURL)
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

// Logout handles user logout by revoking tokens
func (h *AuthHandlers) Logout(c echo.Context) error {
	ctx := c.Request().Context()

	_, ok := common.GetUserIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}

	// Get the token from Authorization header
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Authorization header missing")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	var req LogoutRequest
	if err := c.Bind(&req); err != nil {
		// Bind is optional for logout, but we'll proceed with access token revocation
		req.TokenTypeHint = nil
	}

	// Revoke the access token (and optionally refresh token)
	if err := h.authService.RevokeToken(ctx, tokenString, req.TokenTypeHint); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to revoke token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Logged out successfully",
	})
}

// RefreshRequest represents the token refresh request payload
type RefreshRequest struct {
	RefreshToken string  `json:"refresh_token" validate:"required" sanitize:"trim"`
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

	if err := h.userRepo.UpdateStatus(ctx, tenantID, userID, "active"); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to activate account")
	}

	if err := h.authService.ClearFailedLoginAttempts(ctx, userID); err != nil {
		log.Printf("Failed to clear login attempts for user %s after verification: %v", userID.String(), err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Email verified successfully. You can now sign in.",
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
func (h *AuthHandlers) Verify2FA(c echo.Context) error {
	ctx := c.Request().Context()

	var req Verify2FARequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	validation.SanitizeStruct(&req)

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Validate the temporary token to get user ID
	userID, err := h.authService.ValidatePasswordResetToken(ctx, req.Token)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
	}

	// Verify the 2FA code
	valid, err := h.authService.Verify2FACode(ctx, userID, req.Code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to verify 2FA code")
	}

	if !valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid 2FA code")
	}

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
