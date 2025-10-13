package testhelpers

import (
	"context"
	"time"

	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockCacheService is a mock implementation of the CacheService interface
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	args := m.Called(mock.Anything, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) GetString(ctx context.Context, key string) (string, error) {
	args := m.Called(mock.Anything, key)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) Delete(ctx context.Context, key string) error {
	args := m.Called(mock.Anything, key)
	return args.Error(0)
}

func (m *MockCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	args := m.Called(ctx, key)
	return args.Get(0), args.Error(1)
}

func (m *MockCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *MockCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCacheService) GetCategory(ctx context.Context, tenantID, categoryID uuid.UUID) (*models.Category, error) {
	args := m.Called(ctx, tenantID, categoryID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Category), args.Error(1)
}

func (m *MockCacheService) SetCategory(ctx context.Context, tenantID uuid.UUID, category *models.Category, ttl time.Duration) error {
	args := m.Called(ctx, tenantID, category, ttl)
	return args.Error(0)
}

func (m *MockCacheService) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error {
	args := m.Called(ctx, tenantID, categoryID)
	return args.Error(0)
}

func (m *MockCacheService) GetInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tenantID, warehouseID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Inventory), args.Error(1)
}

func (m *MockCacheService) SetInventory(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory, ttl time.Duration) error {
	args := m.Called(ctx, tenantID, inventory, ttl)
	return args.Error(0)
}

func (m *MockCacheService) DeleteInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) error {
	args := m.Called(ctx, tenantID, warehouseID, productID)
	return args.Error(0)
}

func (m *MockCacheService) GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, tenantID, productID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockCacheService) SetProduct(ctx context.Context, tenantID uuid.UUID, product *models.Product, ttl time.Duration) error {
	args := m.Called(ctx, tenantID, product, ttl)
	return args.Error(0)
}

func (m *MockCacheService) DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error {
	args := m.Called(ctx, tenantID, productID)
	return args.Error(0)
}

func (m *MockCacheService) GetTenantAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCacheService) SetTenantAnalytics(ctx context.Context, tenantID uuid.UUID, analytics map[string]interface{}, ttl time.Duration) error {
	args := m.Called(ctx, tenantID, analytics, ttl)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateTenantCache(ctx context.Context, tenantID uuid.UUID) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateAllCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCacheService) SetSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	args := m.Called(ctx, sessionID, userID, ttl)
	return args.Error(0)
}

func (m *MockCacheService) GetSession(ctx context.Context, sessionID string) (string, error) {
	args := m.Called(ctx, sessionID)
	return args.String(0), args.Error(1)
}

func (m *MockCacheService) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockCacheService) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Error(1)
}

func (m *MockCacheService) IncrementRateLimit(ctx context.Context, key string, window time.Duration) error {
	args := m.Called(ctx, key, window)
	return args.Error(0)
}
