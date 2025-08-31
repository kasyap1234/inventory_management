package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"agromart2/internal/middleware"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// WebhookHandlers handles HTTP requests for webhooks
type WebhookHandlers struct {
	subscriptionService services.SubscriptionService
	razorpayService     services.RazorpayService
	webhookSecret       string
	rbacMiddleware      *middleware.RBACMiddleware
}

// NewWebhookHandlers creates a new webhook handlers instance
func NewWebhookHandlers(
	subscriptionService services.SubscriptionService,
	razorpayService services.RazorpayService,
	webhookSecret string,
	rbacMiddleware *middleware.RBACMiddleware,
) *WebhookHandlers {
	return &WebhookHandlers{
		subscriptionService: subscriptionService,
		razorpayService:     razorpayService,
		webhookSecret:       webhookSecret,
		rbacMiddleware:      rbacMiddleware,
	}
}

// verifyRazorpayWebhookSignature verifies the webhook signature
func (h *WebhookHandlers) verifyRazorpayWebhookSignature(signature string, body []byte) bool {
	hash := hmac.New(sha256.New, []byte(h.webhookSecret))
	hash.Write(body)
	expectedSignature := hex.EncodeToString(hash.Sum(nil))

	// Use constant time comparison to prevent timing attacks
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// RazorpayWebhook handles POST /webhooks/razorpay
func (h *WebhookHandlers) RazorpayWebhook(c echo.Context) error {
	// Read the raw body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Failed to read request body")
	}

	// Get signature from headers
	signature := c.Request().Header.Get("X-Razorpay-Signature")
	if signature == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing Razorpay signature")
	}

	// Verify webhook signature
	if !h.verifyRazorpayWebhookSignature(signature, body) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid webhook signature")
	}

	// Parse webhook event
	event, err := h.razorpayService.WebhookVerify(c.Request().Context(), body, signature)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Process webhook based on event type
	err = h.processRazorpayEvent(event)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
		"event":  event.Event,
	})
}

// processRazorpayEvent processes different Razorpay webhook events
func (h *WebhookHandlers) processRazorpayEvent(event *services.WebhookEvent) error {
	switch event.Event {
	case "subscription.activated":
		return h.handleSubscriptionActivated(event)
	case "subscription.charged":
		return h.handleSubscriptionCharged(event)
	case "subscription.cancelled":
		return h.handleSubscriptionCancelled(event)
	case "subscription.paused":
		return h.handleSubscriptionPaused(event)
	case "subscription.resumed":
		return h.handleSubscriptionResumed(event)
	case "subscription.pending":
		return h.handleSubscriptionPending(event)
	case "subscription.halted":
		return h.handleSubscriptionHalted(event)
	default:
		// Log unknown events but don't return error
		return nil
	}
}

// handleSubscriptionActivated handles subscription activation events
func (h *WebhookHandlers) handleSubscriptionActivated(event *services.WebhookEvent) error {
	return h.handleEvent(event, "active")
}

// handleSubscriptionCharged handles successful payment events
func (h *WebhookHandlers) handleSubscriptionCharged(event *services.WebhookEvent) error {
	return h.handleEvent(event, "charged")
}

// handleSubscriptionCancelled handles subscription cancellation events
func (h *WebhookHandlers) handleSubscriptionCancelled(event *services.WebhookEvent) error {
	return h.handleEvent(event, "cancelled")
}

// handleSubscriptionPaused handles subscription pause events
func (h *WebhookHandlers) handleSubscriptionPaused(event *services.WebhookEvent) error {
	return h.handleEvent(event, "paused")
}

// handleSubscriptionResumed handles subscription resume events
func (h *WebhookHandlers) handleSubscriptionResumed(event *services.WebhookEvent) error {
	return h.handleEvent(event, "active")
}

// handleSubscriptionPending handles subscription pending events
func (h *WebhookHandlers) handleSubscriptionPending(event *services.WebhookEvent) error {
	return h.handleEvent(event, "pending")
}

// handleSubscriptionHalted handles subscription halted events
func (h *WebhookHandlers) handleSubscriptionHalted(event *services.WebhookEvent) error {
	return h.handleEvent(event, "halted")
}

// handleEvent is a helper method to handle common webhook event processing
func (h *WebhookHandlers) handleEvent(event *services.WebhookEvent, status string) error {
	razorpayID, ok := event.Data["subscription_id"].(string)
	if !ok {
		return nil // Skip if no subscription ID
	}

	// TODO: Find subscription by Razorpay ID and update status
	// This is a placeholder implementation that logs the event
	_ = razorpayID
	_ = status

	// In a real implementation, you would:
	// 1. Find the subscription by razorpayID across all tenants
	// 2. Update the local subscription status
	// 3. Send notifications if needed

	return nil
}

// Helper method to find tenant by Razorpay subscription ID
func (h *WebhookHandlers) findTenantByRazorpaySubscriptionID(razorpayID string) (uuid.UUID, error) {
	// TODO: Implement a way to find the tenant by Razorpay subscription ID
	// This might require maintaining a mapping or querying across tenants
	return uuid.Nil, echo.NewHTTPError(http.StatusNotFound, "Tenant not found for Razorpay subscription ID")
}