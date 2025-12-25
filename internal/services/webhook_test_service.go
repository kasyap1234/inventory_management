package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"agromart2/internal/common"

	"github.com/google/uuid"
)

// WebhookTestParams represents the inputs to execute a test webhook delivery
type WebhookTestParams struct {
	TargetURL        string
	Method           string // POST (default), PUT, PATCH
	Headers          map[string]string
	Payload          map[string]interface{}
	Secret           string // resolved secret
	TenantID         uuid.UUID
	EventType        *string
	TimeoutMs        int
	MaxResponseBytes int
	AllowedSchemes   []string
	AllowHTTPInDev   bool
	MaxRedirects     int
}

// WebhookTestResultDTO is the transport object returned to handler
type WebhookTestResultDTO struct {
	Success             bool              `json:"success"`
	TargetStatus        *int              `json:"target_status,omitempty"`
	ResponseHeaders     map[string]string `json:"response_headers,omitempty"`
	ResponseBodySnippet *string           `json:"response_body_snippet,omitempty"`
	DurationMs          int64             `json:"duration_ms"`
	Signature           struct {
		Algorithm  string `json:"algorithm"`
		HeaderName string `json:"header_name"`
	} `json:"signature"`
	Error *string `json:"error,omitempty"`
}

// WebhookTestService provides test webhook delivery execution
type WebhookTestService interface {
	ExecuteTest(ctx context.Context, p WebhookTestParams) (*WebhookTestResultDTO, error)
}

type webhookTestService struct{}

// NewWebhookTestService creates a new WebhookTestService
func NewWebhookTestService() WebhookTestService {
	return &webhookTestService{}
}

func (s *webhookTestService) ExecuteTest(ctx context.Context, p WebhookTestParams) (*WebhookTestResultDTO, error) {
	start := time.Now()
	res := &WebhookTestResultDTO{
		Success:    false,
		DurationMs: 0,
	}
	res.Signature.Algorithm = "HMAC-SHA256"
	res.Signature.HeaderName = "X-Webhook-Signature"

	// Basic validation
	method := strings.ToUpper(strings.TrimSpace(p.Method))
	if method == "" {
		method = "POST"
	}
	switch method {
	case "POST", "PUT", "PATCH":
	default:
		errStr := "invalid_method"
		res.Error = &errStr
		return res, fmt.Errorf("invalid method: %s", method)
	}

	if err := common.ValidateOutgoingURLForWebhook(p.TargetURL, p.AllowHTTPInDev, p.AllowedSchemes); err != nil {
		errStr := err.Error()
		res.Error = &errStr
		return res, err
	}

	// Prepare JSON body
	if p.Payload == nil {
		p.Payload = map[string]interface{}{}
	}
	bodyBytes, err := json.Marshal(p.Payload)
	if err != nil {
		errStr := "payload_encoding_failed"
		res.Error = &errStr
		return res, fmt.Errorf("failed to encode payload: %w", err)
	}

	// Build request
	req, err := http.NewRequestWithContext(ctx, method, p.TargetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		errStr := "request_build_failed"
		res.Error = &errStr
		return res, fmt.Errorf("failed to build request: %w", err)
	}

	// Default headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "AgroMart-Webhook/1.0")
	req.Header.Set("X-Tenant-ID", p.TenantID.String())
	if p.EventType != nil && strings.TrimSpace(*p.EventType) != "" {
		req.Header.Set("X-Event-Type", strings.TrimSpace(*p.EventType))
	}

	// Signature header (if secret provided)
	if strings.TrimSpace(p.Secret) != "" {
		signature := common.ComputeWebhookSignature(bodyBytes, p.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Merge custom headers after sanitization (Authorization and X-Webhook-Signature not allowed)
	for k, v := range p.Headers {
		lk := strings.ToLower(strings.TrimSpace(k))
		if lk == "" {
			continue
		}
		if lk == "authorization" || lk == "x-webhook-signature" {
			continue
		}
		req.Header.Set(k, v)
	}

	// Timeout and redirect policy
	timeout := time.Duration(p.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	validateRedirect := func(u *url.URL) error {
		return common.ValidateOutgoingURLForWebhook(u.String(), p.AllowHTTPInDev, p.AllowedSchemes)
	}
	client := common.BuildNoRedirectClient(timeout, p.MaxRedirects, validateRedirect)

	// Execute
	resp, doErr := client.Do(req)
	elapsed := time.Since(start)
	res.DurationMs = elapsed.Milliseconds()

	// Handle network error
	if doErr != nil {
		errStr := doErr.Error()
		res.Error = &errStr
		// success remains false
		return res, doErr
	}
	defer resp.Body.Close()

	// Read and truncate body
	maxBytes := p.MaxResponseBytes
	if maxBytes <= 0 {
		maxBytes = 2048
	}
	limitedReader := io.LimitReader(resp.Body, int64(maxBytes))
	respBytes, _ := io.ReadAll(limitedReader)
	snippet := common.TruncateBodySnippet(respBytes, maxBytes)

	status := resp.StatusCode
	res.Success = true // attempted and reached target
	res.TargetStatus = &status
	if snippet != "" {
		res.ResponseBodySnippet = &snippet
	}
	res.ResponseHeaders = common.ExtractWhitelistedHeaders(resp)

	return res, nil
}
