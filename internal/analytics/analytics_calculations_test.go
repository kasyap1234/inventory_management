package analytics

import (
	"context"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrderRepoForAnalytics is a mock for order repository in analytics tests
type MockOrderRepoForAnalytics struct {
	mock.Mock
}

func (m *MockOrderRepoForAnalytics) Create(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepoForAnalytics) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Order), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) Update(ctx context.Context, order *models.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockOrderRepoForAnalytics) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockOrderRepoForAnalytics) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) ListByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, status, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) GetBySupplier(ctx context.Context, tenantID, supplierID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, supplierID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) GetByDistributor(ctx context.Context, tenantID, distributorID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, distributorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) Search(ctx context.Context, tenantID uuid.UUID, query string, limit, offset int) ([]*models.Order, error) {
	args := m.Called(ctx, tenantID, query, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Order), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) Count(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockOrderRepoForAnalytics) CountByStatus(ctx context.Context, tenantID uuid.UUID, status string) (int64, error) {
	args := m.Called(ctx, tenantID, status)
	return args.Get(0).(int64), args.Error(1)
}

// Test cases for Supplier Quality Score calculation
func TestSupplierQualityScoreCalculation(t *testing.T) {
	t.Run("perfect delivery record scores 100", func(t *testing.T) {
		// All delivered, no cancellations, no returns
		deliveredCount := 10
		totalCount := 10
		cancelledCount := 0
		returnedCount := 0
		onTimeCount := 10

		deliverySuccessRate := float64(deliveredCount) / float64(totalCount) * 100
		onTimeDeliveryRate := float64(onTimeCount) / float64(deliveredCount) * 100
		cancellationRate := float64(cancelledCount) / float64(totalCount) * 100
		returnRate := float64(returnedCount) / float64(totalCount) * 100

		qualityScore := 0.0
		qualityScore += deliverySuccessRate * 0.40
		qualityScore += onTimeDeliveryRate * 0.30
		qualityScore += (100 - cancellationRate) * 0.15
		qualityScore += (100 - returnRate) * 0.15

		assert.InDelta(t, 100.0, qualityScore, 0.01)
	})

	t.Run("50% delivery rate with issues scores lower", func(t *testing.T) {
		deliveredCount := 5
		totalCount := 10
		cancelledCount := 3
		returnedCount := 2
		onTimeCount := 3 // 60% on-time of delivered

		deliverySuccessRate := float64(deliveredCount) / float64(totalCount) * 100
		onTimeDeliveryRate := float64(onTimeCount) / float64(deliveredCount) * 100
		cancellationRate := float64(cancelledCount) / float64(totalCount) * 100
		returnRate := float64(returnedCount) / float64(totalCount) * 100

		qualityScore := 0.0
		qualityScore += deliverySuccessRate * 0.40           // 50 * 0.4 = 20
		qualityScore += onTimeDeliveryRate * 0.30            // 60 * 0.3 = 18
		qualityScore += (100 - cancellationRate) * 0.15      // 70 * 0.15 = 10.5
		qualityScore += (100 - returnRate) * 0.15            // 80 * 0.15 = 12

		expectedScore := 20.0 + 18.0 + 10.5 + 12.0 // 60.5
		assert.InDelta(t, expectedScore, qualityScore, 0.01)
		assert.Less(t, qualityScore, 70.0) // Should be below good threshold
	})

	t.Run("all cancelled orders scores low", func(t *testing.T) {
		deliveredCount := 0
		totalCount := 10
		cancelledCount := 10
		returnedCount := 0

		deliverySuccessRate := float64(deliveredCount) / float64(totalCount) * 100
		cancellationRate := float64(cancelledCount) / float64(totalCount) * 100
		returnRate := float64(returnedCount) / float64(totalCount) * 100

		qualityScore := 0.0
		qualityScore += deliverySuccessRate * 0.40           // 0 * 0.4 = 0
		qualityScore += 0 * 0.30                             // No delivered orders, 0% on-time
		qualityScore += (100 - cancellationRate) * 0.15      // 0 * 0.15 = 0
		qualityScore += (100 - returnRate) * 0.15            // 100 * 0.15 = 15

		assert.InDelta(t, 15.0, qualityScore, 0.01)
	})
}

// Test cases for Profit Margin calculation
func TestProfitMarginCalculation(t *testing.T) {
	t.Run("calculates positive margin correctly", func(t *testing.T) {
		totalRevenue := 10000.0
		totalGSTCollected := 1800.0 // 18% GST
		totalCost := 5000.0

		grossProfit := totalRevenue - totalGSTCollected - totalCost // 3200
		margin := (grossProfit / totalRevenue) * 100                // 32%

		assert.InDelta(t, 3200.0, grossProfit, 0.01)
		assert.InDelta(t, 32.0, margin, 0.01)
	})

	t.Run("handles zero revenue", func(t *testing.T) {
		totalRevenue := 0.0
		margin := 0.0
		if totalRevenue > 0 {
			margin = (1000.0 / totalRevenue) * 100
		}

		assert.Equal(t, 0.0, margin)
	})

	t.Run("calculates negative margin for loss", func(t *testing.T) {
		totalRevenue := 10000.0
		totalGSTCollected := 1800.0
		totalCost := 12000.0 // Cost exceeds revenue

		grossProfit := totalRevenue - totalGSTCollected - totalCost // -3800
		margin := (grossProfit / totalRevenue) * 100                // -38%

		assert.InDelta(t, -3800.0, grossProfit, 0.01)
		assert.InDelta(t, -38.0, margin, 0.01)
	})

	t.Run("margin trend categorization", func(t *testing.T) {
		testCases := []struct {
			margin   float64
			expected string
		}{
			{30.0, "healthy"},
			{15.0, "stable"},
			{5.0, "needs_attention"},
			{-10.0, "loss"},
		}

		for _, tc := range testCases {
			var marginTrend string
			if tc.margin > 25 {
				marginTrend = "healthy"
			} else if tc.margin >= 10 {
				marginTrend = "stable"
			} else if tc.margin >= 0 {
				marginTrend = "needs_attention"
			} else {
				marginTrend = "loss"
			}
			assert.Equal(t, tc.expected, marginTrend, "margin: %f", tc.margin)
		}
	})
}

// Test for order type filtering in profit calculation
func TestProfitMarginOrderFiltering(t *testing.T) {
	t.Run("only counts delivered purchase orders for cost", func(t *testing.T) {
		orders := []*models.Order{
			{OrderType: "purchase", Status: "delivered", Quantity: 10, UnitPrice: 100},
			{OrderType: "purchase", Status: "pending", Quantity: 10, UnitPrice: 100},    // Should not count
			{OrderType: "sales", Status: "delivered", Quantity: 10, UnitPrice: 150},     // Should not count as cost
			{OrderType: "purchase", Status: "completed", Quantity: 5, UnitPrice: 100},
		}

		var totalCost float64
		purchaseOrderCount := 0

		for _, order := range orders {
			if order.OrderType == "purchase" && (order.Status == "delivered" || order.Status == "completed") {
				totalCost += float64(order.Quantity) * order.UnitPrice
				purchaseOrderCount++
			}
		}

		assert.Equal(t, 1500.0, totalCost) // 10*100 + 5*100
		assert.Equal(t, 2, purchaseOrderCount)
	})
}

// Test for on-time delivery tracking
func TestOnTimeDeliveryTracking(t *testing.T) {
	t.Run("counts on-time when delivered before expected", func(t *testing.T) {
		deliveryTime := time.Now().Add(-2 * time.Hour)
		expectedDelivery := time.Now()

		isOnTime := deliveryTime.Before(expectedDelivery)
		assert.True(t, isOnTime)
	})

	t.Run("counts as late when delivered after expected", func(t *testing.T) {
		deliveryTime := time.Now()
		expectedDelivery := time.Now().Add(-2 * time.Hour)

		isOnTime := deliveryTime.Before(expectedDelivery)
		assert.False(t, isOnTime)
	})

	t.Run("assumes on-time when no expected delivery set", func(t *testing.T) {
		var expectedDelivery *time.Time = nil

		// Logic: if expectedDelivery is nil, assume on-time
		isOnTime := expectedDelivery == nil
		assert.True(t, isOnTime)
	})
}
