package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/common"
	"agromart2/internal/config"
	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// stubCacheService implements caching.CacheService with minimal behavior needed for tests
type stubCacheService struct {
	limited bool
}

var _ caching.CacheService = (*stubCacheService)(nil)

func (s *stubCacheService) GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*models.Product, error) { return nil, nil }
func (s *stubCacheService) SetProduct(ctx context.Context, tenantID uuid.UUID, product *models.Product, ttl time.Duration) error {
	return nil
}
func (s *stubCacheService) DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error { return nil }
func (s *stubCacheService) GetInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	return nil, nil
}
func (s *stubCacheService) SetInventory(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory, ttl time.Duration) error {
	return nil
}
func (s *stubCacheService) DeleteInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) error { return nil }
func (s *stubCacheService) GetCategory(ctx context.Context, tenantID, categoryID uuid.UUID) (*models.Category, error) {
	return nil, nil
}
func (s *stubCacheService) SetCategory(ctx context.Context, tenantID uuid.UUID, category *models.Category, ttl time.Duration) error {
	return nil
}
func (s *stubCacheService) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error { return nil }
func (s *stubCacheService) GetTenantAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	return nil, nil
}
func (s *stubCacheService) SetTenantAnalytics(ctx context.Context, tenantID uuid.UUID, analytics map[string]interface{}, ttl time.Duration) error {
	return nil
}
func (s *stubCacheService) InvalidateTenantCache(ctx context.Context, tenantID uuid.UUID) error { return nil }
func (s *stubCacheService) InvalidateAllCache(ctx context.Context) error                     { return nil }
func (s *stubCacheService) SetSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	return nil
}
func (s *stubCacheService) GetSession(ctx context.Context, sessionID string) (string, error) { return "", nil }
func (s *stubCacheService) DeleteSession(ctx context.Context, sessionID string) error        { return nil }
func (s *stubCacheService) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	return s.limited, nil
}
func (s *stubCacheService) IncrementRateLimit(ctx context.Context, key string, window time.Duration) error { return nil }
func (s *stubCacheService) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return nil
}
func (s *stubCacheService) GetString(ctx context.Context, key string) (string, error) { return "", nil }
func (s *stubCacheService) Delete(ctx context.Context, key string) error              { return nil }
func (s *stubCacheService) DeleteByPattern(ctx context.Context, pattern string) error { return nil }

func makeEchoWithAuthContext() (*echo.Echo, echo.Context, uuid.UUID, uuid.UUID) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	tenantID := uuid.New()
	userID := uuid.New()
	ctx := context.WithValue(req.Context(), common.TenantIDKey, tenantID)
	ctx = context.WithValue(ctx, common.UserIDKey, userID) // fixed: use userID not tenantID
	c.SetRequest(req.WithContext(ctx))
	return e, c, tenantID, userID
}

func TestWebhookTestHandlers_HeaderSanitization(t *testing.T) {
	// Target server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return OK; we will inspect headers via handler result only (service ignores forbidden headers)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	cache := &stubCacheService{limited: false}
	webhookRepoSvc := services.NewWebhookSubscriptionService(nil, common.GetGlobalLogger())
	testSvc := services.NewWebhookTestService()
	cfg := &config.WebhookTestConfig{
		TimeoutMs:        1000,
		MaxResponseBytes: 2048,
		AllowedSchemes:   []string{"http", "https"},
		AllowHTTPInDev:   true,
		RateLimitPerMin:  5,
		MaxRedirects:     0,
	}
	h := NewWebhookTestHandlers(cache, webhookRepoSvc, testSvc, cfg)

	_, c, _, _ := makeEchoWithAuthContext()

	// Build request payload with forbidden headers
	body := map[string]interface{}{
		"target_url": ts.URL,
		"method":     "POST",
		"headers": map[string]string{
			"Authorization":      "Bearer attempt",
			"X-Webhook-Signature": "override",
			"X-Allowed":          "ok",
		},
		"payload": map[string]interface{}{"x": 1},
	}
	b, _ := json.Marshal(body)
	c.Request().Body = io.NopCloser(bytes.NewReader(b))
	c.Request().Header.Set("Content-Type", "application/json")

	err := h.TestWebhook(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusBadRequest, httpErr.Code)
	}
}

func TestWebhookTestHandlers_RateLimit(t *testing.T) {
	cache := &stubCacheService{limited: true}
	webhookRepoSvc := services.NewWebhookSubscriptionService(nil, common.GetGlobalLogger())
	testSvc := services.NewWebhookTestService()
	cfg := &config.WebhookTestConfig{
		TimeoutMs:        1000,
		MaxResponseBytes: 2048,
		AllowedSchemes:   []string{"https"},
		AllowHTTPInDev:   false,
		RateLimitPerMin:  5,
		MaxRedirects:     0,
	}
	h := NewWebhookTestHandlers(cache, webhookRepoSvc, testSvc, cfg)

	_, c, _, _ := makeEchoWithAuthContext()

	// Use an example URL that won't be resolved since we expect early 429
	body := map[string]interface{}{
		"target_url": "https://example.com/webhook",
		"method":     "POST",
		"headers":    map[string]string{},
		"payload":    map[string]interface{}{"x": 1},
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/v1/webhooks/test", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	// Preserve the auth context from makeEchoWithAuthContext
	req = req.WithContext(c.Request().Context())
	c.SetRequest(req)

	err := h.TestWebhook(c)
	if assert.Error(t, err) {
		httpErr, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		if httpErr.Code != http.StatusTooManyRequests {
			t.Logf("Error message: %v", httpErr.Message)
		}
		assert.Equal(t, http.StatusTooManyRequests, httpErr.Code)
	}
}

// ioNopCloser provides a no-op closer for setting request bodies
type ioNopCloser struct {
	*bytes.Reader
}

func (io ioNopCloser) Close() error { return nil }