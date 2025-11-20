package handlers

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// GoogleLogin initiates the Google OAuth2 flow
func (h *AuthHandlers) GoogleLogin(c echo.Context) error {
	// Generate a random state string for security (in production use a proper CSRF token)
	state := "random-state-string"
	url := h.authService.GetGoogleAuthURL(state)
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

// GoogleCallback handles the callback from Google
func (h *AuthHandlers) GoogleCallback(c echo.Context) error {
	ctx := c.Request().Context()
	code := c.QueryParam("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Code not found")
	}

	user, tempToken, err := h.authService.HandleGoogleCallback(ctx, code)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to authenticate with Google: "+err.Error())
	}

	// If user exists, login and redirect to dashboard
	if user != nil {
		// Generate tokens
		tokenResponse, err := h.authService.GenerateTokens(ctx, user.ID, user.TenantID, nil)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate tokens")
		}

		// Redirect to frontend with tokens
		// In a real app, you might set a cookie or redirect to a page that handles the token
		// For this implementation, we'll redirect to a frontend route that processes the token
		frontendURL := os.Getenv("FRONTEND_URL")
		redirectURL := frontendURL + "/auth/google/callback?access_token=" + tokenResponse.AccessToken + "&refresh_token=" + tokenResponse.RefreshToken
		return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
	}

	// If user does not exist, redirect to completion page with temp token
	frontendURL := os.Getenv("FRONTEND_URL")
	redirectURL := frontendURL + "/complete-registration?token=" + tempToken
	return c.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

// CompleteGoogleSignupRequest represents the request to complete Google signup
type CompleteGoogleSignupRequest struct {
	Token      string `json:"token" validate:"required"`
	TenantName string `json:"tenant_name" validate:"required,min=2,max=100"`
	Subdomain  string `json:"subdomain" validate:"required,min=2,max=63,alphanum"`
}

// CompleteGoogleSignup completes the registration for a Google user
func (h *AuthHandlers) CompleteGoogleSignup(c echo.Context) error {
	ctx := c.Request().Context()

	var req CompleteGoogleSignupRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Decode temp token to get Google info
	tokenBytes, err := base64.URLEncoding.DecodeString(req.Token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token format")
	}
	tokenData := string(tokenBytes)
	parts := strings.Split(tokenData, ":")
	if len(parts) != 4 {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid token data")
	}

	email, googleID, firstName, lastName := parts[0], parts[1], parts[2], parts[3]

	// Call service to complete signup
	user, err := h.authService.CompleteGoogleSignup(ctx, email, googleID, firstName, lastName, req.TenantName, req.Subdomain)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to complete signup: "+err.Error())
	}

	// Generate tokens
	tokenResponse, err := h.authService.GenerateTokens(ctx, user.ID, user.TenantID, nil)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate tokens")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user":   user,
		"tokens": tokenResponse,
	})
}
