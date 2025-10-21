package handlers

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DeviceTokenHandlers handles HTTP requests for device token management
type DeviceTokenHandlers struct {
	repo   repositories.DeviceTokenRepository
	logger *common.StructuredLogger
}

// NewDeviceTokenHandlers creates a new device token handlers instance
func NewDeviceTokenHandlers(repo repositories.DeviceTokenRepository, logger *common.StructuredLogger) *DeviceTokenHandlers {
	return &DeviceTokenHandlers{
		repo:   repo,
		logger: logger,
	}
}

// RegisterDevice registers a new device token for push notifications
// @Summary Register device token
// @Description Register a device token for receiving push notifications
// @Tags notifications
// @Accept json
// @Produce json
// @Param device body models.DeviceTokenRequest true "Device token registration"
// @Success 200 {object} map[string]interface{} "Device registered successfully"
// @Failure 400 {object} common.ErrorResponse "Invalid request"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/devices [post]
func (h *DeviceTokenHandlers) RegisterDevice(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	var req models.DeviceTokenRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	// Validate device type
	if req.DeviceType != "android" && req.DeviceType != "ios" && req.DeviceType != "web" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid device type. Must be 'android', 'ios', or 'web'")
	}

	// Validate device token
	if req.DeviceToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Device token is required")
	}

	token := &models.DeviceToken{
		TenantID:    tenantID,
		UserID:      userID,
		DeviceToken: req.DeviceToken,
		DeviceType:  req.DeviceType,
		DeviceName:  req.DeviceName,
		AppVersion:  req.AppVersion,
		IsActive:    true,
	}

	if err := h.repo.RegisterToken(ctx, token); err != nil {
		h.logger.ErrorWithContext(ctx, "Failed to register device token", err, map[string]interface{}{
			"tenant_id": tenantID,
			"user_id":   userID,
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to register device")
	}

	h.logger.InfoWithContext(ctx, "Device token registered", map[string]interface{}{
		"tenant_id":   tenantID,
		"user_id":     userID,
		"device_type": req.DeviceType,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":    "Device registered successfully",
		"device_id":  token.ID,
		"registered": true,
	})
}

// UnregisterDevice unregisters a device token
// @Summary Unregister device token
// @Description Remove a device token from receiving push notifications
// @Tags notifications
// @Produce json
// @Param token path string true "Device Token"
// @Success 200 {object} map[string]interface{} "Device unregistered successfully"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/devices/{token} [delete]
func (h *DeviceTokenHandlers) UnregisterDevice(c echo.Context) error {
	ctx := c.Request().Context()
	deviceToken := c.Param("token")
	tenantID := c.Get("tenant_id").(uuid.UUID)

	if deviceToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Device token is required")
	}

	if err := h.repo.DeleteToken(ctx, tenantID, deviceToken); err != nil {
		h.logger.ErrorWithContext(ctx, "Failed to unregister device token", err, map[string]interface{}{
			"tenant_id":    tenantID,
			"device_token": deviceToken,
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to unregister device")
	}

	h.logger.InfoWithContext(ctx, "Device token unregistered", map[string]interface{}{
		"tenant_id":    tenantID,
		"device_token": deviceToken,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":      "Device unregistered successfully",
		"unregistered": true,
	})
}

// ListDevices lists all registered devices for the current user
// @Summary List user devices
// @Description Get all registered device tokens for the current user
// @Tags notifications
// @Produce json
// @Success 200 {array} models.DeviceToken "List of devices"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/devices [get]
func (h *DeviceTokenHandlers) ListDevices(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID := c.Get("tenant_id").(uuid.UUID)
	userID := c.Get("user_id").(uuid.UUID)

	tokens, err := h.repo.GetTokensByUser(ctx, tenantID, userID)
	if err != nil {
		h.logger.ErrorWithContext(ctx, "Failed to get device tokens", err, map[string]interface{}{
			"tenant_id": tenantID,
			"user_id":   userID,
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get devices")
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"devices": tokens,
		"count":   len(tokens),
	})
}

// DeactivateDevice deactivates a device token without deleting it
// @Summary Deactivate device token
// @Description Temporarily disable push notifications for a device
// @Tags notifications
// @Produce json
// @Param token path string true "Device Token"
// @Success 200 {object} map[string]interface{} "Device deactivated successfully"
// @Failure 500 {object} common.ErrorResponse "Internal server error"
// @Router /api/v1/notifications/devices/{token}/deactivate [put]
func (h *DeviceTokenHandlers) DeactivateDevice(c echo.Context) error {
	ctx := c.Request().Context()
	deviceToken := c.Param("token")
	tenantID := c.Get("tenant_id").(uuid.UUID)

	if deviceToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Device token is required")
	}

	if err := h.repo.DeactivateToken(ctx, tenantID, deviceToken); err != nil {
		h.logger.ErrorWithContext(ctx, "Failed to deactivate device token", err, map[string]interface{}{
			"tenant_id":    tenantID,
			"device_token": deviceToken,
		})
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to deactivate device")
	}

	h.logger.InfoWithContext(ctx, "Device token deactivated", map[string]interface{}{
		"tenant_id":    tenantID,
		"device_token": deviceToken,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":     "Device deactivated successfully",
		"deactivated": true,
	})
}
