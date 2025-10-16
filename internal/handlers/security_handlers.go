package handlers

import (
	"net/http"
	"time"

	"agromart2/internal/security"

	"github.com/labstack/echo/v4"
)

// SecurityHandlers provides endpoints related to application security utilities.
type SecurityHandlers struct {
	csrfManager *security.CSRFTokenManager
}

// NewSecurityHandlers creates a new security handlers instance.
func NewSecurityHandlers(csrfManager *security.CSRFTokenManager) *SecurityHandlers {
	return &SecurityHandlers{csrfManager: csrfManager}
}

// GetCSRFToken handles GET /security/csrf and returns a signed CSRF token for subsequent requests.
func (h *SecurityHandlers) GetCSRFToken(c echo.Context) error {
	token, expiresAt, err := h.csrfManager.GenerateToken()
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate CSRF token")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}
