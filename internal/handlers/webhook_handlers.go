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
	"strings"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/common"
	"agromart2/internal/config"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// WebhookHandlers handles HTTP requests for webhooks
type WebhookHandlers struct {
	subscriptionService services.SubscriptionService
	razorpayService     services.RazorpayService
	invoiceService      services.InvoiceServiceInterface
	notificationService services.NotificationService
	paymentService      services.PaymentService
	webhookRepo         repositories.WebhookEventRepository
	webhookSecret       string
	rbacMiddleware      *middleware.RBACMiddleware
}

// NewWebhookHandlers creates a new webhook handlers instance
func NewWebhookHandlers(
	subscriptionService services.SubscriptionService,
	razorpayService services.RazorpayService,
	invoiceService services.InvoiceServiceInterface,
	notificationService services.NotificationService,
	paymentService services.PaymentService,
	webhookRepo repositories.WebhookEventRepository,
	webhookSecret string,
	rbacMiddleware *middleware.RBACMiddleware,
) *WebhookHandlers {
	return &WebhookHandlers{
		subscriptionService: subscriptionService,
		razorpayService:     razorpayService,
		invoiceService:      invoiceService,
		notificationService: notificationService,
		paymentService:      paymentService,
		webhookRepo:         webhookRepo,
		webhookSecret:       webhookSecret,
		rbacMiddleware:      rbacMiddleware,
	}
}

// verifyRazorpayWebhookSignature verifies the webhook signature
func (h *WebhookHandlers) verifyRazorpayWebhookSignature(signature string, body []byte) bool {
	trimmedSignature := strings.TrimSpace(signature)
	if trimmedSignature == "" || h.webhookSecret == "" {
		return false
	}

	hash := hmac.New(sha256.New, []byte(h.webhookSecret))
	hash.Write(body)
	expected := hash.Sum(nil)

	provided, err := hex.DecodeString(trimmedSignature)
	if err != nil {
		return false
	}

	return hmac.Equal(provided, expected)
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

	// Idempotency guard
	if h.webhookRepo != nil && event != nil && event.ID != "" {
		if processed, _ := h.webhookRepo.AlreadyProcessed(c.Request().Context(), "razorpay", event.ID); processed {
			return c.JSON(http.StatusOK, map[string]string{
				"status": "ignored_duplicate",
				"event":  event.Event,
			})
		}
	}

	// Process webhook based on event type
	err = h.processRazorpayEvent(event)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	if h.webhookRepo != nil && event != nil {
		_ = h.webhookRepo.MarkProcessed(c.Request().Context(), "razorpay", event.ID, signature, event)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "success",
		"event":  event.Event,
	})
}

// processRazorpayEvent processes different Razorpay webhook events
func (h *WebhookHandlers) processRazorpayEvent(event *services.WebhookEvent) error {
	switch event.Event {
	case "payment.authorized", "payment.captured", "payment.failed":
		return h.handlePaymentEvent(event)
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

// handlePaymentEvent processes one-time payment webhooks (order payments)
func (h *WebhookHandlers) handlePaymentEvent(event *services.WebhookEvent) error {
	if h.paymentService == nil {
		return nil
	}

	payload, ok := event.Data["payload"].(map[string]interface{})
	if !ok {
		return nil
	}
	paymentMap, ok := payload["payment"].(map[string]interface{})
	if !ok {
		return nil
	}
	entity, ok := paymentMap["entity"].(map[string]interface{})
	if !ok {
		return nil
	}

	orderID, _ := entity["order_id"].(string)
	paymentID, _ := entity["id"].(string)
	status := event.Event
	if parts := strings.Split(event.Event, "."); len(parts) == 2 {
		status = parts[1]
	}

	// Extract tenant ID from notes (we add tenant_id when creating orders)
	var tenantID uuid.UUID
	if notes, ok := entity["notes"].(map[string]interface{}); ok {
		if tid, ok := notes["tenant_id"].(string); ok {
			if parsed, err := uuid.Parse(tid); err == nil {
				tenantID = parsed
			}
		}
	}
	if tenantID == uuid.Nil || orderID == "" {
		return nil
	}

	var paidAt *time.Time
	if status == "captured" {
		now := time.Now()
		paidAt = &now
	}

	return h.paymentService.MarkPaymentStatus(context.Background(), tenantID, orderID, status, &paymentID, nil, paidAt)
}

// handleSubscriptionActivated handles subscription activation events
func (h *WebhookHandlers) handleSubscriptionActivated(event *services.WebhookEvent) error {
	return h.handleEvent(event, "active")
}

// handleSubscriptionCharged handles successful payment events
// handleSubscriptionCharged handles successful payment events
func (h *WebhookHandlers) handleSubscriptionCharged(event *services.WebhookEvent) error {
	// First update status
	if err := h.handleEvent(event, "charged"); err != nil {
		return err
	}

	// Extract Razorpay subscription ID from event data
	var razorpayID string
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
		return nil // Should have been handled by handleEvent, but just in case
	}

	// Find subscription
	ctx := context.Background()
	subscription, _, err := h.findSubscriptionByRazorpayID(ctx, razorpayID)
	if err != nil {
		// Log but don't fail - webhook might be for a non-existent subscription
		log.Printf("Failed to find subscription for Razorpay ID %s: %v", razorpayID, err)
		return nil
	}

	// Extract payment details
	var paymentDetails map[string]interface{}
	if payload, ok := event.Data["payload"].(map[string]interface{}); ok {
		if payment, ok := payload["payment"].(map[string]interface{}); ok {
			if entity, ok := payment["entity"].(map[string]interface{}); ok {
				paymentDetails = entity
			}
		}
	}

	if paymentDetails == nil {
		log.Printf("Payment details not found in webhook event for subscription %s", razorpayID)
		return nil
	}

	// Generate and email invoice
	if err := h.invoiceService.EmailSubscriptionInvoice(ctx, subscription, paymentDetails); err != nil {
		log.Printf("Failed to email subscription invoice: %v", err)
		// Don't fail the webhook, as the main action (status update) succeeded
	}

	return nil
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

// -------------------- Test Webhook Endpoint --------------------

// WebhookTestRequest represents the JSON payload for testing a webhook
type WebhookTestRequest struct {
	TargetURL string                 `json:"target_url"`
	Method    string                 `json:"method,omitempty"` // POST default; allow PUT, PATCH
	Headers   map[string]string      `json:"headers,omitempty"`
	Payload   map[string]interface{} `json:"payload,omitempty"`
	SecretID  *string                `json:"secret_id,omitempty"` // UUID
	Secret    *string                `json:"secret,omitempty"`
	EventType *string                `json:"event_type,omitempty"`
}

// WebhookTestHandlers encapsulates dependencies for the test webhook endpoint
type WebhookTestHandlers struct {
	cacheSvc      caching.CacheService
	webhookSubSvc services.WebhookSubscriptionService
	testSvc       services.WebhookTestService
	cfg           *config.WebhookTestConfig
	logger        *common.StructuredLogger
}

// NewWebhookTestHandlers constructs handlers for POST /v1/webhooks/test
func NewWebhookTestHandlers(
	cacheSvc caching.CacheService,
	webhookSubSvc services.WebhookSubscriptionService,
	testSvc services.WebhookTestService,
	cfg *config.WebhookTestConfig,
) *WebhookTestHandlers {
	return &WebhookTestHandlers{
		cacheSvc:      cacheSvc,
		webhookSubSvc: webhookSubSvc,
		testSvc:       testSvc,
		cfg:           cfg,
		logger:        common.GetGlobalLogger(),
	}
}

// TestWebhook handles POST /v1/webhooks/test
func (h *WebhookTestHandlers) TestWebhook(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind request
	var req WebhookTestRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}

	// Validation: target_url required
	if strings.TrimSpace(req.TargetURL) == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "target_url is required")
	}

	// Validation: method
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = "POST"
	}
	switch method {
	case "POST", "PUT", "PATCH":
	default:
		return echo.NewHTTPError(http.StatusBadRequest, "Unsupported method; only POST, PUT, PATCH are allowed")
	}

	// Validation: headers - reject Authorization and X-Webhook-Signature
	for k := range req.Headers {
		lk := strings.ToLower(strings.TrimSpace(k))
		if lk == "authorization" || lk == "x-webhook-signature" {
			return echo.NewHTTPError(http.StatusBadRequest, "Forbidden header override: "+lk)
		}
	}

	// Rate limit per (tenantID,userID): check first to avoid unnecessary preflight/network
	tenantID, okT := common.GetTenantIDFromContext(ctx)
	userID, okU := common.GetUserIDFromContext(ctx)
	if !okT || !okU {
		return echo.NewHTTPError(http.StatusUnauthorized, "Authentication context missing")
	}
	rlKey := "webhooktest:" + tenantID.String() + ":" + userID.String()
	limited, rlErr := h.cacheSvc.IsRateLimited(ctx, rlKey, h.cfg.RateLimitPerMin, time.Minute)
	if rlErr != nil {
		// Internal error checking rate limit
		return echo.NewHTTPError(http.StatusInternalServerError, "Rate limit check failed")
	}
	if limited {
		return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
	}

	// SSRF guard preflight
	if err := common.ValidateOutgoingURLForWebhook(req.TargetURL, h.cfg.AllowHTTPInDev, h.cfg.AllowedSchemes); err != nil {
		// Map SSRF and URL errors to 400
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Resolve signing secret
	var resolvedSecret string
	if req.SecretID != nil && strings.TrimSpace(*req.SecretID) != "" {
		secretUUID, err := uuid.Parse(strings.TrimSpace(*req.SecretID))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "secret_id must be a valid UUID")
		}
		sub, err := h.webhookSubSvc.GetWebhook(ctx, tenantID, secretUUID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "secret_id not found for this tenant")
		}
		if sub.Secret != nil {
			resolvedSecret = strings.TrimSpace(*sub.Secret)
		}
	} else if req.Secret != nil {
		resolvedSecret = strings.TrimSpace(*req.Secret)
	}

	// Execute test via service
	params := services.WebhookTestParams{
		TargetURL:        req.TargetURL,
		Method:           method,
		Headers:          req.Headers,
		Payload:          req.Payload,
		Secret:           resolvedSecret,
		TenantID:         tenantID,
		EventType:        req.EventType,
		TimeoutMs:        h.cfg.TimeoutMs,
		MaxResponseBytes: h.cfg.MaxResponseBytes,
		AllowedSchemes:   h.cfg.AllowedSchemes,
		AllowHTTPInDev:   h.cfg.AllowHTTPInDev,
		MaxRedirects:     h.cfg.MaxRedirects,
	}

	result, err := h.testSvc.ExecuteTest(ctx, params)

	// Structured logging (no secrets)
	h.logger.InfoWithContext(ctx, "Webhook test executed", map[string]interface{}{
		"target_url":    req.TargetURL,
		"method":        method,
		"event_type":    common.SafeString(req.EventType),
		"duration_ms":   result.DurationMs,
		"target_status": result.TargetStatus,
	})

	// Audit event (webhook.test)
	if al := common.GetGlobalAuditLogger(); al != nil {
		al.LogCustomEvent(ctx, "webhook.test", "webhook", "", map[string]interface{}{
			"target_url": req.TargetURL,
			"method":     method,
			"event_type": common.SafeString(req.EventType),
			"success":    result.Success,
			"status":     result.TargetStatus,
		})
	}

	// Error mapping for preflight/network failures
	if err != nil {
		msg := err.Error()
		// Timeout mapping
		if strings.Contains(strings.ToLower(msg), "timeout") {
			return c.JSON(http.StatusGatewayTimeout, result) // 504
		}
		// DNS/TLS errors mapping to 502
		lmsg := strings.ToLower(msg)
		if strings.Contains(lmsg, "dns") || strings.Contains(lmsg, "no such host") || strings.Contains(lmsg, "x509") || strings.Contains(lmsg, "tls") {
			return c.JSON(http.StatusBadGateway, result) // 502
		}
		// Default network error
		return c.JSON(http.StatusBadGateway, result)
	}

	// Non-2xx target responses still return 200 with success=true
	return c.JSON(http.StatusOK, result)
}
