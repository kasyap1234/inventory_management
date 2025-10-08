package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandlers handles authentication-related HTTP requests
type AuthHandlers struct {
	authService    services.AuthService
	userRepo       repositories.UserRepository
	roleRepo       repositories.RoleRepository
	userRoleRepo   repositories.UserRoleRepository
	rbacMiddleware *middleware.RBACMiddleware
}

// NewAuthHandlers creates a new auth handlers instance
func NewAuthHandlers(authService services.AuthService, userRepo repositories.UserRepository, roleRepo repositories.RoleRepository, userRoleRepo repositories.UserRoleRepository, rbacMiddleware *middleware.RBACMiddleware) *AuthHandlers {
	return &AuthHandlers{
		authService:    authService,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		userRoleRepo:   userRoleRepo,
		rbacMiddleware: rbacMiddleware,
	}
}

// LoginResponse represents the login response
type LoginResponse struct {
	models.TokenResponse
	User *models.User `json:"user"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// Login handles user login with email and password
func (h *AuthHandlers) Login(c echo.Context) error {
	ctx := c.Request().Context()

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.Email == "" || req.Password == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email and password are required")
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

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid password")
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
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=6"`
	FirstName string  `json:"first_name" validate:"required"`
	LastName  string  `json:"last_name" validate:"required"`
	TenantID  *string `json:"tenant_id"` // Optional, for existing tenants
}

// SignupResponse represents the signup response
type SignupResponse struct {
	models.TokenResponse
	User *models.User `json:"user"`
}

// Signup handles user registration
func (h *AuthHandlers) Signup(c echo.Context) error {
	ctx := c.Request().Context()

	var req SignupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Email, password, first name, and last name are required")
	}

	// Generate user ID
	userID := uuid.New()

	var tenantID uuid.UUID

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
		Status:       "active",
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

	// Generate JWT tokens
	tokenResponse, err := h.authService.GenerateTokens(ctx, userID, tenantID, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate tokens")
	}

	response := SignupResponse{
		TokenResponse: *tokenResponse,
		User:          user,
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
	RefreshToken string  `json:"refresh_token" validate:"required"`
	GrantType    string  `json:"grant_type" validate:"required"`
	ClientID     *string `json:"client_id"`
	Scope        *string `json:"scope"`
}

// Refresh handles token refresh
func (h *AuthHandlers) Refresh(c echo.Context) error {
	ctx := c.Request().Context()

	var req RefreshRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if req.RefreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Refresh token is required")
	}

	if req.GrantType != "refresh_token" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid grant type")
	}

	// Refresh tokens
	tokenResponse, err := h.authService.RefreshToken(ctx, req.RefreshToken, req.ClientID)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid or expired refresh token")
	}

	return c.JSON(http.StatusOK, tokenResponse)
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
	// In production, you might want to:
	// 1. Use email domain mapping to tenant
	// 2. Have a dedicated users_index table for fast lookup
	// 3. Cache email->tenant_id mappings
	
	// For now, we'll query the user by email and tenant separately
	// The UserRepository.GetByEmail would need enhancement to support cross-tenant search
	// As a workaround, we can check common/active tenants
	
	// This is a temporary implementation - in production, enhance UserRepository
	// to have a GetByEmailGlobal method that doesn't require tenant_id
	
	// For now, return error to indicate this needs repository enhancement
	return nil, fmt.Errorf("cross-tenant user lookup requires repository enhancement")
}

// getOrCreateTenantForSignup gets existing tenant by email domain or creates new one
func (h *AuthHandlers) getOrCreateTenantForSignup(ctx context.Context, email string) (uuid.UUID, error) {
	// Extract domain from email
	domain := extractDomainFromEmail(email)
	
	// For personal emails (gmail, yahoo, etc.), create individual tenant
	// For business emails, could try to find existing tenant by domain
	
	personalDomains := map[string]bool{
		"gmail.com":    true,
		"yahoo.com":    true,
		"hotmail.com":  true,
		"outlook.com":  true,
		"icloud.com":   true,
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
