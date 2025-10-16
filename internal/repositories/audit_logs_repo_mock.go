package repositories

import (
	"context"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockAuditLogsRepository is a mock implementation of AuditLogsRepository for testing
type MockAuditLogsRepository struct {
	mock.Mock
}

func (m *MockAuditLogsRepository) Create(ctx context.Context, log *models.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

func (m *MockAuditLogsRepository) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.AuditLog, error) {
	args := m.Called(ctx, tenantID, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogsRepository) List(ctx context.Context, tenantID uuid.UUID, filters *models.AuditLogFilters) ([]*models.AuditLog, error) {
	args := m.Called(ctx, tenantID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogsRepository) ListByUserID(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	args := m.Called(ctx, tenantID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogsRepository) ListByEntity(ctx context.Context, tenantID uuid.UUID, entityType, entityID string, limit, offset int) ([]*models.AuditLog, error) {
	args := m.Called(ctx, tenantID, entityType, entityID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogsRepository) GetActions(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuditLogsRepository) GetByTableAndRecord(ctx context.Context, tenantID uuid.UUID, tableName, recordID string, limit, offset int) ([]*models.AuditLog, error) {
	args := m.Called(ctx, tenantID, tableName, recordID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogsRepository) GetByUser(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	args := m.Called(ctx, tenantID, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditLogsRepository) SoftDelete(ctx context.Context, tenantID, id uuid.UUID) error {
	args := m.Called(ctx, tenantID, id)
	return args.Error(0)
}

func (m *MockAuditLogsRepository) GetSummary(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*models.AuditLogSummary, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuditLogSummary), args.Error(1)
}

func (m *MockAuditLogsRepository) GetTableNames(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
