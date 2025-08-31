package services

import (
	"context"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuditLogsServiceTestSuite struct {
	suite.Suite
	mockRepo  *repositories.MockAuditLogsRepository
	service   AuditLogsService
	tenantID  uuid.UUID
	userID    uuid.UUID
	ctx       context.Context
}

func (suite *AuditLogsServiceTestSuite) SetupTest() {
	suite.mockRepo = &repositories.MockAuditLogsRepository{}
	suite.service = NewAuditLogsService(suite.mockRepo)
	suite.tenantID = uuid.New()
	suite.userID = uuid.New()
	suite.ctx = context.Background()
}

func (suite *AuditLogsServiceTestSuite) TestLogActivity_Success() {
	// Arrange
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		TenantID:   suite.tenantID,
		TableName:  "users",
		RecordID:   suite.userID.String(),
		Action:     models.ActionUpdate,
		NewValues:  models.JSONB{"name": "John Doe"},
		ChangedBy:  &suite.userID,
		Deleted:    false,
	}

	suite.mockRepo.On("Create", suite.ctx, auditLog).Return(nil)

	// Act
	err := suite.service.LogActivity(suite.ctx, suite.tenantID, "users", suite.userID.String(),
		models.ActionUpdate, &suite.userID, nil, models.JSONB{"name": "John Doe"})

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuditLogsServiceTestSuite) TestGetAuditLog_Success() {
	// Arrange
	auditLogID := uuid.New()
	expectedLog := &models.AuditLog{
		ID:        auditLogID,
		TenantID:  suite.tenantID,
		TableName: "products",
		Action:    models.ActionInsert,
	}

	suite.mockRepo.On("GetByID", suite.ctx, suite.tenantID, auditLogID).Return(expectedLog, nil)

	// Act
	result, err := suite.service.GetAuditLog(suite.ctx, suite.tenantID, auditLogID)

	// Assert
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedLog.ID, result.ID)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuditLogsServiceTestSuite) TestListAuditLogs_Success() {
	// Arrange
	filters := &models.AuditLogFilters{
		TableName: stringPtr("users"),
		Limit:     10,
		Offset:    0,
	}

	expectedLogs := []*models.AuditLog{
		{ID: uuid.New(), TableName: "users", Action: models.ActionInsert},
	}

	suite.mockRepo.On("List", suite.ctx, suite.tenantID, filters).Return(expectedLogs, nil)

	// Act
	result, err := suite.service.ListAuditLogs(suite.ctx, suite.tenantID, filters)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuditLogsServiceTestSuite) TestGetEntityHistory_Success() {
	// Arrange
	recordID := "record-123"
	expectedLogs := []*models.AuditLog{
		{ID: uuid.New(), TableName: "orders", RecordID: recordID, Action: models.ActionUpdate},
	}

	suite.mockRepo.On("GetByTableAndRecord", suite.ctx, suite.tenantID, "orders", recordID, 50, 0).Return(expectedLogs, nil)

	// Act
	result, err := suite.service.GetEntityHistory(suite.ctx, suite.tenantID, "orders", recordID, 50, 0)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 1)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuditLogsServiceTestSuite) TestLogEntityCreate_Success() {
	// Arrange
	newValues := models.JSONB{"name": "New Product", "price": 100}
	auditLog := &models.AuditLog{
		TenantID:  suite.tenantID,
		TableName: "products",
		RecordID:  "prod-123",
		Action:    models.ActionInsert,
		NewValues: newValues,
		ChangedBy: &suite.userID,
	}

	suite.mockRepo.On("Create", suite.ctx, auditLog).Return(nil)

	// Act
	err := suite.service.LogEntityCreate(suite.ctx, suite.tenantID, "products", "prod-123", &suite.userID, newValues)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertExpectations(suite.T())
}

func (suite *AuditLogsServiceTestSuite) TestValidateAuditFilters_Valid() {
	filters := &models.AuditLogFilters{
		TableName: stringPtr("users"),
		Limit:     50,
		StartDate: &time.Time{},
		EndDate:   &time.Time{},
	}

	// Set valid date range
	start := time.Now().AddDate(0, -1, 0)
	end := time.Now()
	*filters.StartDate = start
	*filters.EndDate = end

	err := suite.service.ValidateAuditFilters(filters)
	assert.NoError(suite.T(), err)
}

func (suite *AuditLogsServiceTestSuite) TestValidateAuditFilters_InvalidDateRange() {
	filters := &models.AuditLogFilters{
		StartDate: &time.Time{},
		EndDate:   &time.Time{},
	}

	// Set invalid date range (end before start)
	start := time.Now()
	end := start.AddDate(0, -1, 0)
	*filters.StartDate = start
	*filters.EndDate = end

	err := suite.service.ValidateAuditFilters(filters)
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "start_date cannot be after end_date")
}

func stringPtr(s string) *string {
	return &s
}

func TestAuditLogsServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuditLogsServiceTestSuite))
}

// Note: In a real test suite, you would also need to create mock implementations for the repository interface