package repositories

import (
	"context"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/testhelpers"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type NotificationRepoTestSuite struct {
	suite.Suite
	testDB   *testhelpers.TestDB
	repo     NotificationRepository
	tenantID uuid.UUID
	userID   uuid.UUID
}

func stringPtr(s string) *string {
	return &s
}

func (suite *NotificationRepoTestSuite) SetupSuite() {
	suite.testDB = testhelpers.SetupTestDB(suite.T(), "")
	suite.repo = NewNotificationRepo(suite.testDB.Pool)

	// Create test tenant and user
	suite.tenantID = uuid.New()
	suite.userID = uuid.New()

	ctx := context.Background()

	// Insert test tenant
	_, err := suite.testDB.Pool.Exec(ctx, `
		INSERT INTO tenants (id, name, subdomain, status, created_at, updated_at)
		VALUES ($1, 'Test Tenant', 'test-notification', 'active', NOW(), NOW())
	`, suite.tenantID)
	require.NoError(suite.T(), err)

	// Insert test user
	_, err = suite.testDB.Pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, status, created_at, updated_at)
		VALUES ($1, $2, 'test-notif@example.com', 'hash', 'active', NOW(), NOW())
	`, suite.userID, suite.tenantID)
	require.NoError(suite.T(), err)
}

func (suite *NotificationRepoTestSuite) TearDownSuite() {
	if suite.testDB != nil {
		suite.testDB.Cleanup()
	}
}

func (suite *NotificationRepoTestSuite) SetupTest() {
	// Clean up notifications before each test
	ctx := context.Background()
	_, _ = suite.testDB.Pool.Exec(ctx, "DELETE FROM notifications WHERE tenant_id = $1", suite.tenantID)
}

func (suite *NotificationRepoTestSuite) TestCountUnread_NoNotifications() {
	ctx := context.Background()

	count, err := suite.repo.CountUnread(ctx, suite.tenantID, suite.userID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)
}

func (suite *NotificationRepoTestSuite) TestCountUnread_AllUnread() {
	ctx := context.Background()

	// Create 3 unread notifications
	for i := 0; i < 3; i++ {
		notification := &models.EnhancedNotification{
			ID:               uuid.New(),
			TenantID:         suite.tenantID,
			UserID:           suite.userID,
			Title:            "Test Notification",
			Message:          "Test message",
			NotificationType: "info",
			Priority:         "normal",
			Status:           "pending",
			EventType:        stringPtr("test_event"),
		}
		err := suite.repo.Create(ctx, notification)
		require.NoError(suite.T(), err)
	}

	count, err := suite.repo.CountUnread(ctx, suite.tenantID, suite.userID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, count)
}

func (suite *NotificationRepoTestSuite) TestCountUnread_MixedReadStatus() {
	ctx := context.Background()

	// Create 2 unread notifications
	for i := 0; i < 2; i++ {
		notification := &models.EnhancedNotification{
			ID:               uuid.New(),
			TenantID:         suite.tenantID,
			UserID:           suite.userID,
			Title:            "Unread Notification",
			Message:          "Unread message",
			NotificationType: "info",
			Priority:         "normal",
			Status:           "pending",
			EventType:        stringPtr("test_event"),
		}
		err := suite.repo.Create(ctx, notification)
		require.NoError(suite.T(), err)
	}

	// Create 1 read notification
	readNotification := &models.EnhancedNotification{
		ID:               uuid.New(),
		TenantID:         suite.tenantID,
		UserID:           suite.userID,
		Title:            "Read Notification",
		Message:          "Read message",
		NotificationType: "info",
		Priority:         "normal",
		Status:           "read",
		EventType:        stringPtr("test_event"),
		ReadAt:           func() *time.Time { t := time.Now(); return &t }(),
	}
	err := suite.repo.Create(ctx, readNotification)
	require.NoError(suite.T(), err)

	count, err := suite.repo.CountUnread(ctx, suite.tenantID, suite.userID)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)
}

func (suite *NotificationRepoTestSuite) TestCountUnread_MarkAsReadUpdatesCount() {
	ctx := context.Background()

	// Create unread notification
	notification := &models.EnhancedNotification{
		ID:               uuid.New(),
		TenantID:         suite.tenantID,
		UserID:           suite.userID,
		Title:            "Test Notification",
		Message:          "Test message",
		NotificationType: "info",
		Priority:         "normal",
		Status:           "pending",
		EventType:        stringPtr("test_event"),
	}
	err := suite.repo.Create(ctx, notification)
	require.NoError(suite.T(), err)

	// Verify count is 1
	count, err := suite.repo.CountUnread(ctx, suite.tenantID, suite.userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)

	// Mark as read
	err = suite.repo.MarkAsRead(ctx, suite.tenantID, notification.ID)
	require.NoError(suite.T(), err)

	// Verify count is now 0
	count, err = suite.repo.CountUnread(ctx, suite.tenantID, suite.userID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)
}

func TestNotificationRepoTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	suite.Run(t, new(NotificationRepoTestSuite))
}
