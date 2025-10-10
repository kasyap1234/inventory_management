package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// RazorpayService handles all Razorpay API interactions with placeholders
type RazorpayService interface {
	CreateSubscription(ctx context.Context, planID string, tenantID uuid.UUID, customerEmail string) (*CreateSubscriptionResponse, error)
	CancelSubscription(ctx context.Context, subscriptionID string) (*CancelSubscriptionResponse, error)
	PauseSubscription(ctx context.Context, subscriptionID string) (*PauseSubscriptionResponse, error)
	ResumeSubscription(ctx context.Context, subscriptionID string) (*ResumeSubscriptionResponse, error)
	UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) (*UpdateSubscriptionResponse, error)
	WebhookVerify(ctx context.Context, rawData []byte, signature string) (*WebhookEvent, error)
}

type razorpayService struct {
	apiKey        string
	apiSecret     string
	webhookSecret string
	baseURL       string
	http          *http.Client
}

// Plan configurations
type PlanDetails struct {
	ID     string
	Name   string
	Amount float64
	Period string // monthly, yearly, etc.
}

type CreateSubscriptionRequest struct {
	PlanID        string `json:"plan_id"`
	CustomerEmail string `json:"customer_email"`
	StartAt       int64  `json:"start_at,omitempty"`
	EndAt         int64  `json:"end_at,omitempty"`
	Quantity      int    `json:"quantity,omitempty"`
	OfferID       string `json:"offer_id,omitempty"`
}

type CreateSubscriptionResponse struct {
	ID      string `json:"id"`
	Entity  string `json:"entity"`
	Status  string `json:"status"`
	StartAt int64  `json:"start_at"`
	EndAt   int64  `json:"end_at"`
}

type CancelSubscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type PauseSubscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type ResumeSubscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type UpdateSubscriptionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type WebhookEvent struct {
	ID      string                 `json:"id"`
	Event   string                 `json:"event"`
	Data    map[string]interface{} `json:"data"`
	Created int64                  `json:"created"`
}

// NewRazorpayService creates a new Razorpay service instance
func NewRazorpayService(apiKey, apiSecret, webhookSecret string) RazorpayService {
	return &razorpayService{
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		webhookSecret: webhookSecret,
		baseURL:       "https://api.razorpay.com/v1", // Razorpay API base URL
		http:          &http.Client{},
	}
}

// CreateSubscription creates a subscription via Razorpay API
func (s *razorpayService) CreateSubscription(ctx context.Context, planID string, tenantID uuid.UUID, customerEmail string) (*CreateSubscriptionResponse, error) {
	req := CreateSubscriptionRequest{
		PlanID:        planID,
		CustomerEmail: customerEmail,
		Quantity:      1,
	}

	respBytes, err := s.makeRequest(ctx, "POST", "/subscriptions", req)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	var resp CreateSubscriptionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse subscription response: %w", err)
	}

	return &resp, nil
}

// CancelSubscription cancels a subscription via Razorpay API
func (s *razorpayService) CancelSubscription(ctx context.Context, subscriptionID string) (*CancelSubscriptionResponse, error) {
	path := fmt.Sprintf("/subscriptions/%s/cancel", subscriptionID)
	respBytes, err := s.makeRequest(ctx, "POST", path, map[string]interface{}{"cancel_at_cycle_end": 0})
	if err != nil {
		return nil, fmt.Errorf("failed to cancel subscription: %w", err)
	}

	var resp CancelSubscriptionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse cancel response: %w", err)
	}

	return &resp, nil
}

// PauseSubscription pauses a subscription via Razorpay API
func (s *razorpayService) PauseSubscription(ctx context.Context, subscriptionID string) (*PauseSubscriptionResponse, error) {
	path := fmt.Sprintf("/subscriptions/%s/pause", subscriptionID)
	respBytes, err := s.makeRequest(ctx, "POST", path, map[string]interface{}{"pause_at": "now"})
	if err != nil {
		return nil, fmt.Errorf("failed to pause subscription: %w", err)
	}

	var resp PauseSubscriptionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse pause response: %w", err)
	}

	return &resp, nil
}

// ResumeSubscription resumes a subscription via Razorpay API
func (s *razorpayService) ResumeSubscription(ctx context.Context, subscriptionID string) (*ResumeSubscriptionResponse, error) {
	path := fmt.Sprintf("/subscriptions/%s/resume", subscriptionID)
	respBytes, err := s.makeRequest(ctx, "POST", path, map[string]interface{}{"resume_at": "now"})
	if err != nil {
		return nil, fmt.Errorf("failed to resume subscription: %w", err)
	}

	var resp ResumeSubscriptionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse resume response: %w", err)
	}

	return &resp, nil
}

// UpdateSubscription updates subscription details via Razorpay API
func (s *razorpayService) UpdateSubscription(ctx context.Context, subscriptionID string, updates map[string]interface{}) (*UpdateSubscriptionResponse, error) {
	path := fmt.Sprintf("/subscriptions/%s", subscriptionID)
	respBytes, err := s.makeRequest(ctx, "PATCH", path, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	var resp UpdateSubscriptionResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse update response: %w", err)
	}

	return &resp, nil
}

// WebhookVerify verifies webhook signature (HMAC)
func (s *razorpayService) WebhookVerify(ctx context.Context, rawData []byte, signature string) (*WebhookEvent, error) {
	trimmedSignature := strings.TrimSpace(signature)
	if trimmedSignature == "" {
		return nil, fmt.Errorf("missing webhook signature")
	}
	if s.webhookSecret == "" {
		return nil, fmt.Errorf("razorpay webhook secret is not configured")
	}

	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(rawData)
	expected := mac.Sum(nil)

	provided, err := hex.DecodeString(trimmedSignature)
	if err != nil {
		return nil, fmt.Errorf("invalid webhook signature format: %w", err)
	}

	if !hmac.Equal(provided, expected) {
		return nil, fmt.Errorf("invalid webhook signature")
	}

	var event WebhookEvent
	if err := json.Unmarshal(rawData, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook data: %v", err)
	}

	return &event, nil
}

// Helper methods for actual API calls

func (s *razorpayService) makeRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	var req *http.Request
	var err error

	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}

		req, err = http.NewRequestWithContext(ctx, method, s.baseURL+path, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	} else {
		req, err = http.NewRequestWithContext(ctx, method, s.baseURL+path, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
	}

	req.SetBasicAuth(s.apiKey, s.apiSecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("razorpay API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
