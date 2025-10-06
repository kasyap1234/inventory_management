package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"

	"agromart2/internal/middleware"
	"agromart2/internal/models"
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
	// Extract Razorpay subscription ID from event data
	var razorpayID string
	
	// Check for subscription_id in different possible locations
	if subID, ok := event.Data["subscription_id"].(string); ok {
		razorpayID = subID
	} else if payload, ok := event.Data["payload"].(map[string]interface{}); ok {
		if subscription, ok := payload["subscription"].(map[string]interface{}); ok {
			if id, ok := subscription["entity"].(map[string]interface{})["id"].(string); ok {
				razorpayID = id
			}
		}
	}
	
	if razorpayID == "" {
		return nil // Skip if no subscription ID found
	}

	// Find the subscription by Razorpay ID using a direct repository query
	// This queries across tenants to find the matching subscription
	ctx := context.Background()
	subscription, tenantID, err := h.findSubscriptionByRazorpayID(ctx, razorpayID)
	if err != nil {
		// Log but don't fail - webhook might be for a non-existent subscription
		log.Printf("Failed to find subscription for Razorpay ID %s: %v", razorpayID, err)
		return nil
	}

	// Update the subscription status
	subscription.Status = status
	if err := h.subscriptionService.Update(ctx, tenantID, subscription); err != nil {
		return fmt.Errorf("failed to update subscription status: %v", err)
	}

	// Log the event for auditing
	log.Printf("Updated subscription %s for tenant %s to status %s via webhook", subscription.ID.String(), tenantID.String(), status)

	return nil
}

// Helper method to find subscription by Razorpay subscription ID
func (h *WebhookHandlers) findSubscriptionByRazorpayID(ctx context.Context, razorpayID string) (*models.Subscription, uuid.UUID, error) {
	// Query across all tenants using repository method
	// This uses an indexed query on razorpay_subscription_id
	subscription, err := h.subscriptionService.FindByRazorpayIDCrossTenant(ctx, razorpayID)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("subscription not found for Razorpay ID %s: %v", razorpayID, err)
	}
	
	return subscription, subscription.TenantID, nil
}