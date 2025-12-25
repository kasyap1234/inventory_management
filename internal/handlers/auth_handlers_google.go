package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
)

// GoogleLogin initiates the Google OAuth2 flow
func (h *AuthHandlers) GoogleLogin(c echo.Context) error {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate state")
	}
	state := base64.URLEncoding.EncodeToString(b)
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

		// Set HttpOnly cookies for tokens
		c.SetCookie(&http.Cookie{
			Name:     "auth_token",
			Value:    tokenResponse.AccessToken,
			HttpOnly: true,
			Secure:   os.Getenv("ENV") == "production",
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   3600,
		})
		c.SetCookie(&http.Cookie{
			Name:     "refresh_token",
			Value:    tokenResponse.RefreshToken,
			HttpOnly: true,
			Secure:   os.Getenv("ENV") == "production",
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   86400 * 7,
		})

		// Redirect to frontend without tokens in URL
		frontendURL := os.Getenv("FRONTEND_URL")
		return c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/auth/google/callback")
	}

	// If user does not exist, redirect to completion page with temp token in cookie
	c.SetCookie(&http.Cookie{
		Name:     "temp_token",
		Value:    tempToken,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   3600,
	})
	frontendURL := os.Getenv("FRONTEND_URL")
	return c.Redirect(http.StatusTemporaryRedirect, frontendURL+"/complete-registration")
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
