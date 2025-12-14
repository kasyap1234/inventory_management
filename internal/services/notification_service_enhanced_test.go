package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationService_ListTemplates(t *testing.T) {
	// Start miniredis for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create Redis client connected to miniredis
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	tenantID := uuid.New()

	t.Run("returns empty slice when no templates exist", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		templates, err := service.ListTemplates(context.Background(), tenantID, "")

		assert.NoError(t, err)
		assert.Empty(t, templates)
	})

	t.Run("lists all templates for tenant", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		// Create test templates
		template1 := &models.NotificationTemplate{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Name:         "Order Confirmation",
			EventType:    "order_created",
			BodyTemplate: "Your order has been confirmed",
		}
		template2 := &models.NotificationTemplate{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Name:         "Shipping Update",
			EventType:    "order_shipped",
			BodyTemplate: "Your order has been shipped",
		}

		// Store templates in Redis
		key1 := "notification_template:" + tenantID.String() + ":" + template1.ID.String()
		data1, _ := json.Marshal(template1)
		err := redisClient.Set(context.Background(), key1, data1, time.Hour).Err()
		require.NoError(t, err)

		key2 := "notification_template:" + tenantID.String() + ":" + template2.ID.String()
		data2, _ := json.Marshal(template2)
		err = redisClient.Set(context.Background(), key2, data2, time.Hour).Err()
		require.NoError(t, err)

		templates, err := service.ListTemplates(context.Background(), tenantID, "")

		assert.NoError(t, err)
		assert.Len(t, templates, 2)
	})

	t.Run("filters by event type", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		otherTenantID := uuid.New()

		// Create templates with different event types
		template1 := &models.NotificationTemplate{
			ID:           uuid.New(),
			TenantID:     otherTenantID,
			Name:         "Order Created",
			EventType:    "order_created",
			BodyTemplate: "Order created template",
		}
		template2 := &models.NotificationTemplate{
			ID:           uuid.New(),
			TenantID:     otherTenantID,
			Name:         "Payment Received",
			EventType:    "payment_received",
			BodyTemplate: "Payment template",
		}

		key1 := "notification_template:" + otherTenantID.String() + ":" + template1.ID.String()
		data1, _ := json.Marshal(template1)
		redisClient.Set(context.Background(), key1, data1, time.Hour)

		key2 := "notification_template:" + otherTenantID.String() + ":" + template2.ID.String()
		data2, _ := json.Marshal(template2)
		redisClient.Set(context.Background(), key2, data2, time.Hour)

		// Filter by order_created
		templates, err := service.ListTemplates(context.Background(), otherTenantID, "order_created")

		assert.NoError(t, err)
		assert.Len(t, templates, 1)
		assert.Equal(t, "order_created", templates[0].EventType)
	})

	t.Run("isolates templates by tenant", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		tenant1ID := uuid.New()
		tenant2ID := uuid.New()

		// Create template for tenant1
		template1 := &models.NotificationTemplate{
			ID:           uuid.New(),
			TenantID:     tenant1ID,
			Name:         "Tenant 1 Template",
			EventType:    "test_event",
			BodyTemplate: "Template for tenant 1",
		}
		key1 := "notification_template:" + tenant1ID.String() + ":" + template1.ID.String()
		data1, _ := json.Marshal(template1)
		redisClient.Set(context.Background(), key1, data1, time.Hour)

		// Create template for tenant2
		template2 := &models.NotificationTemplate{
			ID:           uuid.New(),
			TenantID:     tenant2ID,
			Name:         "Tenant 2 Template",
			EventType:    "test_event",
			BodyTemplate: "Template for tenant 2",
		}
		key2 := "notification_template:" + tenant2ID.String() + ":" + template2.ID.String()
		data2, _ := json.Marshal(template2)
		redisClient.Set(context.Background(), key2, data2, time.Hour)

		// List for tenant1 should only return 1 template
		templates1, err := service.ListTemplates(context.Background(), tenant1ID, "")
		assert.NoError(t, err)
		assert.Len(t, templates1, 1)
		assert.Equal(t, tenant1ID, templates1[0].TenantID)

		// List for tenant2 should only return 1 template
		templates2, err := service.ListTemplates(context.Background(), tenant2ID, "")
		assert.NoError(t, err)
		assert.Len(t, templates2, 1)
		assert.Equal(t, tenant2ID, templates2[0].TenantID)
	})
}

func TestNotificationService_RetryFailedNotifications(t *testing.T) {
	// Start miniredis for testing
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	// Create Redis client connected to miniredis
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer redisClient.Close()

	t.Run("processes empty queue without error", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		err := service.RetryFailedNotifications(context.Background())

		assert.NoError(t, err)
	})

	t.Run("moves to dead letter after max retries", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		// Create a failed notification with max retries
		failedNotif := struct {
			TenantID      uuid.UUID              `json:"tenant_id"`
			Recipient     string                 `json:"recipient"`
			Subject       string                 `json:"subject"`
			Body          string                 `json:"body"`
			Channel       string                 `json:"channel"`
			RetryCount    int                    `json:"retry_count"`
			LastError     string                 `json:"last_error"`
			OriginalTime  time.Time              `json:"original_time"`
			LastRetryTime time.Time              `json:"last_retry_time"`
			Metadata      map[string]interface{} `json:"metadata"`
		}{
			TenantID:      uuid.New(),
			Recipient:     "test@example.com",
			Subject:       "Test",
			Body:          "Test body",
			Channel:       "email",
			RetryCount:    3, // Max retries exceeded
			LastError:     "connection failed",
			OriginalTime:  time.Now().Add(-1 * time.Hour),
			LastRetryTime: time.Now().Add(-10 * time.Minute),
		}

		data, _ := json.Marshal(failedNotif)
		key := "failed_notification:test-max-retry"
		redisClient.Set(context.Background(), key, data, time.Hour)

		err := service.RetryFailedNotifications(context.Background())

		assert.NoError(t, err)

		// Original key should be gone
		exists := redisClient.Exists(context.Background(), key).Val()
		assert.Equal(t, int64(0), exists)

		// Should be in dead letter queue
		deadLetterKey := "dead_letter:test-max-retry"
		deadLetterExists := redisClient.Exists(context.Background(), deadLetterKey).Val()
		assert.Equal(t, int64(1), deadLetterExists)
	})

	t.Run("respects exponential backoff", func(t *testing.T) {
		service := &notificationService{
			redisClient: redisClient,
		}

		// Create a failed notification that was just retried
		failedNotif := struct {
			TenantID      uuid.UUID              `json:"tenant_id"`
			Recipient     string                 `json:"recipient"`
			Subject       string                 `json:"subject"`
			Body          string                 `json:"body"`
			Channel       string                 `json:"channel"`
			RetryCount    int                    `json:"retry_count"`
			LastError     string                 `json:"last_error"`
			OriginalTime  time.Time              `json:"original_time"`
			LastRetryTime time.Time              `json:"last_retry_time"`
			Metadata      map[string]interface{} `json:"metadata"`
		}{
			TenantID:      uuid.New(),
			Recipient:     "test@example.com",
			Subject:       "Test",
			Body:          "Test body",
			Channel:       "email",
			RetryCount:    1,
			LastError:     "connection failed",
			OriginalTime:  time.Now().Add(-1 * time.Hour),
			LastRetryTime: time.Now(), // Just retried - should skip
		}

		data, _ := json.Marshal(failedNotif)
		key := "failed_notification:test-backoff"
		redisClient.Set(context.Background(), key, data, time.Hour)

		err := service.RetryFailedNotifications(context.Background())

		assert.NoError(t, err)

		// Key should still exist (was skipped due to backoff)
		exists := redisClient.Exists(context.Background(), key).Val()
		assert.Equal(t, int64(1), exists)
	})
}
