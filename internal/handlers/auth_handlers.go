package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	"golang.org/x/crypto/bcrypt"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService         services.AuthService
	userRepo            repositories.UserRepository
	roleRepo            repositories.RoleRepository
	userRoleRepo        repositories.UserRoleRepository
	rbacMiddleware      *middleware.RBACMiddleware
	notificationService services.NotificationService
	frontendBaseURL     string
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(
	authService services.AuthService,
	userRepo repositories.UserRepository,
	roleRepo repositories.RoleRepository,
	userRoleRepo repositories.UserRoleRepository,
	rbacMiddleware *middleware.RBACMiddleware,
	notificationService services.NotificationService,
	frontendBaseURL string,
) *AuthHandlers {
	return &AuthHandlers{
		authService:         authService,
		userRepo:            userRepo,
		roleRepo:            roleRepo,
		userRoleRepo:        userRoleRepo,
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
	if user.PasswordHash == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Account not properly initialized. Please contact support or try signing up again.")
	}

	isLocked, err := h.authService.IsAccountLocked(ctx, user.ID)
	if err != nil {
		log.Printf("Failed to check lock status for user %s: %v", user.ID.String(), err)
	} else if isLocked {
		return echo.NewHTTPError(http.StatusTooManyRequests, "Account temporarily locked due to multiple failed login attempts. Please try again later.")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
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

	// Generate user ID
	userID := uuid.New()

	var (
		tenantID uuid.UUID
		err      error
	)

	// If tenant_id provided, use it; otherwise create a new tenant for the user
	if req.TenantID != nil && *req.TenantID != "" {
		tid, err := uuid.Parse(*req.TenantID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID format")
		}
		if tid == uuid.Nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid tenant ID")
		}
		tenantID = tid
	} else {
		// For signup without tenant_id, derive from email domain or create new tenant
		tenantID, err = h.getOrCreateTenantForSignup(ctx, req.Email)
		if err != nil {
			log.Printf("Failed to get/create tenant for email %s: %v", req.Email, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initialize tenant")
		}
	}

	// Check if user already exists
	existingUser, err := h.userRepo.GetByEmail(ctx, tenantID, req.Email)
	if err == nil && existingUser != nil {
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to hash password")
	}

	// Create new user
	user := &models.User{
		ID:           userID,
		TenantID:     tenantID,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       "pending_verification",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.userRepo.Create(ctx, user); err != nil {
		// Log user creation error
		log.Printf("Failed to create user %s for tenant %s: %v", user.Email, tenantID.String(), err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	// Assign default 'user' role to the new user
	userRole, err := h.roleRepo.GetByName(ctx, tenantID, "user")
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get default user role")
	}

	if userRole == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Default user role not found")
	}

	newUserRole := &models.UserRole{
		UserID: userID,
		RoleID: userRole.ID,
	}
	if err := h.userRoleRepo.Create(ctx, tenantID, newUserRole); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to assign role to user")
	}

	verificationToken, err := h.authService.GenerateEmailVerificationToken(ctx, userID)
	if err != nil {
		log.Printf("Failed to create verification token for user %s: %v", user.Email, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to initiate email verification")
	}

	verificationURL := h.buildEmailVerificationURL(verificationToken)
	emailBody := fmt.Sprintf(
		"<p>Hello %s,</p><p>Welcome to Agromart! Please confirm your email address to activate your account.</p><p><a href=\"%s\">Verify Email</a></p><p>This link expires in 24 hours.</p>",
		user.FirstName,
		verificationURL,
	)

	if err := h.notificationService.SendEmail(ctx, tenantID, user.Email, "Verify your Agromart account", emailBody); err != nil {
		log.Printf("Failed to send verification email to %s: %v", user.Email, err)
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
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

// getOrCreateTenantForSignup gets existing tenant by email domain or creates new one
func (h *AuthHandlers) getOrCreateTenantForSignup(ctx context.Context, email string) (uuid.UUID, error) {
	// Extract domain from email
	domain := extractDomainFromEmail(email)

	// For personal emails (gmail, yahoo, etc.), create individual tenant
	// For business emails, could try to find existing tenant by domain

	personalDomains := map[string]bool{
		"gmail.com":      true,
		"yahoo.com":      true,
		"hotmail.com":    true,
		"outlook.com":    true,
		"icloud.com":     true,
		"protonmail.com": true,
	}

	if personalDomains[domain] {
		// Create individual tenant for personal email
		return uuid.New(), nil
	}

	// For business domains, create a new tenant
	// In production, you might want to check if tenant exists for this domain
	return uuid.New(), nil
}

// extractDomainFromEmail extracts domain from email address
func extractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ""
	}
	return strings.ToLower(parts[1])
}
