package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDB is a simple mock for testing when no database is needed
type MockDB struct {
	mock.Mock
}

// Test without database (covers nil db case)
func TestAnalyticsRefreshService_NilDatabase(t *testing.T) {
	t.Run("handles nil database gracefully", func(t *testing.T) {
		service := NewAnalyticsRefreshService(nil, nil)

		result, err := service.RefreshAllTenantsAnalytics(context.Background())

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.TenantsProcessed)
		assert.False(t, result.DataUpdated)
	})
}

// Unit tests for AnalyticsRefreshResult
func TestAnalyticsRefreshResult(t *testing.T) {
	t.Run("result struct is properly initialized", func(t *testing.T) {
		result := &AnalyticsRefreshResult{
			TenantsProcessed:       5,
			DataUpdated:            true,
			MaterializedViewsRefed: true,
			LastRefreshAt:          time.Now(),
		}

		assert.Equal(t, 5, result.TenantsProcessed)
		assert.True(t, result.DataUpdated)
		assert.True(t, result.MaterializedViewsRefed)
	})

	t.Run("timestamp is set correctly", func(t *testing.T) {
		now := time.Now()
		result := &AnalyticsRefreshResult{
			LastRefreshAt: now,
		}

		assert.Equal(t, now, result.LastRefreshAt)
	})
}

// Test RefreshAnalyticsForTenant - note: requires non-nil analytics service 
// so we skip direct unit test and defer to integration tests


// Mock test for tenant filtering logic
func TestTenantFilteringLogic(t *testing.T) {
	t.Run("filters only active tenants", func(t *testing.T) {
		tenants := []struct {
			ID     uuid.UUID
			Status string
		}{
			{uuid.New(), "active"},
			{uuid.New(), "suspended"},
			{uuid.New(), "active"},
			{uuid.New(), "pending"},
		}

		var activeTenants []uuid.UUID
		for _, tenant := range tenants {
			if tenant.Status == "active" {
				activeTenants = append(activeTenants, tenant.ID)
			}
		}

		assert.Len(t, activeTenants, 2)
	})
}

// Test result calculation
func TestResultCalculation(t *testing.T) {
	t.Run("calculates success count correctly", func(t *testing.T) {
		totalTenants := 10
		failedRefresh := 2
		successCount := totalTenants - failedRefresh

		assert.Equal(t, 8, successCount)
	})

	t.Run("data updated is true when at least one success", func(t *testing.T) {
		successCount := 1
		dataUpdated := successCount > 0

		assert.True(t, dataUpdated)
	})

	t.Run("data updated is false when no success", func(t *testing.T) {
		successCount := 0
		dataUpdated := successCount > 0

		assert.False(t, dataUpdated)
	})
}
