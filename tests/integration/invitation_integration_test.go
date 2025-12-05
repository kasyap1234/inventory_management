package integration

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"
	"agromart2/testhelpers/containers"

	"github.com/google/uuid"
	"github.com/pquerna/otp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// InvitationIntegrationTestSuite tests invitation operations
type InvitationIntegrationTestSuite struct {
	suite.Suite
	container          *containers.PostgresContainer
	ctx                context.Context
	cancel             context.CancelFunc
	tenantRepo         repositories.TenantRepository
	userRepo           repositories.UserRepository
	roleRepo           repositories.RoleRepository
	userRoleRepo       repositories.UserRoleRepository
	permissionRepo     repositories.PermissionRepository
	rolePermissionRepo repositories.RolePermissionRepository
	invitationRepo     repositories.InvitationRepository
	tenantService      services.TenantService
	invitationService  services.InvitationService
	roleService        services.RoleManagementService
	logger             *common.StructuredLogger
}

func TestInvitationIntegrationTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	suite.Run(t, new(InvitationIntegrationTestSuite))
}

func (s *InvitationIntegrationTestSuite) SetupSuite() {
	s.ctx, s.cancel = context.WithTimeout(context.Background(), 5*time.Minute)

	// Get the path to migrations directory
	_, filename, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(filename), "..", "..")
	migrationsPath := filepath.Join(projectRoot, "migrations")

	config := containers.DefaultPostgresConfig()
	config.MigrationsPath = migrationsPath

	container, err := containers.NewPostgresContainer(s.ctx, config)
	require.NoError(s.T(), err, "Failed to start PostgreSQL container")

	s.container = container
	s.logger = common.NewStructuredLogger()

	// Initialize repositories
	s.tenantRepo = repositories.NewTenantRepo(container.Pool)
	s.userRepo = repositories.NewUserRepo(container.Pool)
	s.roleRepo = repositories.NewRoleRepo(container.Pool)
	s.userRoleRepo = repositories.NewUserRoleRepo(container.Pool)
	s.permissionRepo = repositories.NewPermissionRepo(container.Pool)
	s.rolePermissionRepo = repositories.NewRolePermissionRepo(container.Pool)
	s.invitationRepo = repositories.NewInvitationRepo(container.Pool)

	// Initialize services
	mockNotificationService := &MockNotificationService{}
	mockAuthService := &MockAuthService{}

	s.invitationService = services.NewInvitationService(
		s.invitationRepo,
		s.userRepo,
		s.userRoleRepo,
		s.tenantRepo,
		mockNotificationService,
		mockAuthService,
		"http://localhost:3000",
	)

	s.roleService = services.NewRoleManagementService(
		s.roleRepo,
		s.permissionRepo,
		s.logger,
	)

	s.tenantService = services.NewTenantService(
		s.tenantRepo,
		s.invitationService,
		s.roleRepo,
	)
}

func (s *InvitationIntegrationTestSuite) TearDownSuite() {
	if s.container != nil {
		s.container.Close(s.ctx)
	}
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *InvitationIntegrationTestSuite) SetupTest() {
	// Clean tables before each test
	err := s.container.CleanTables(s.ctx)
	require.NoError(s.T(), err, "Failed to clean tables")
}

func (s *InvitationIntegrationTestSuite) TestInvitationFlow() {
	// 1. Create Tenant manually
	tenantID := uuid.New()
	err := s.tenantRepo.Create(s.ctx, &models.Tenant{
		ID:        tenantID,
		Name:      "Manual Tenant",
		Subdomain: "manual",
		Status:    "active",
	})
	require.NoError(s.T(), err)

	// 2. Create Role manually
	roleID := uuid.New()
	err = s.roleRepo.Create(s.ctx, &models.Role{
		ID:       roleID,
		TenantID: tenantID,
		Name:     "admin",
	})
	require.NoError(s.T(), err)

	// 3. Create Invitation
	inviteReq := &services.CreateInvitationRequest{
		TenantID: tenantID,
		Email:    "user@test.com",
		RoleID:   roleID,
	}

	invitation, err := s.invitationService.CreateInvitation(s.ctx, inviteReq, uuid.New())
	require.NoError(s.T(), err)
	assert.NotEmpty(s.T(), invitation.Token)
	assert.Equal(s.T(), models.InvitationStatusPending, invitation.Status)

	// 4. Accept Invitation
	acceptReq := &services.AcceptInvitationRequest{
		FirstName: "John",
		LastName:  "Doe",
		Password:  "SecurePass123!",
	}

	user, err := s.invitationService.AcceptInvitation(s.ctx, invitation.Token, acceptReq)
	require.NoError(s.T(), err)
	assert.NotNil(s.T(), user)
	assert.Equal(s.T(), "user@test.com", user.Email)
	assert.Equal(s.T(), "active", user.Status)

	// 5. Verify User Role
	roles, err := s.roleRepo.GetUserRoles(s.ctx, tenantID, user.ID)
	require.NoError(s.T(), err)
	assert.Len(s.T(), roles, 1)
	assert.Equal(s.T(), "admin", roles[0].Name)

	// 6. Verify Invitation Status
	updatedInv, err := s.invitationRepo.GetByToken(s.ctx, invitation.Token)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), models.InvitationStatusAccepted, updatedInv.Status)
}

// MockNotificationService
type MockNotificationService struct{}

func (m *MockNotificationService) SendNotification(ctx context.Context, tenantID uuid.UUID, notification *models.Notification) error {
	return nil
}
func (m *MockNotificationService) SendEmail(ctx context.Context, tenantID uuid.UUID, recipient, subject, body string) error {
	return nil
}
func (m *MockNotificationService) SendSMS(ctx context.Context, tenantID uuid.UUID, recipient, message string) error {
	return nil
}
func (m *MockNotificationService) SendWebhook(ctx context.Context, tenantID uuid.UUID, webhook *models.WebhookSubscription, payload map[string]interface{}) error {
	return nil
}
func (m *MockNotificationService) SendLowStockAlerts(ctx context.Context, tenantID uuid.UUID, products []models.Product, userRepo repositories.UserRepository) error {
	return nil
}
func (m *MockNotificationService) ListNotifications(ctx context.Context, tenantID uuid.UUID, notificationType, eventType, status string) ([]*models.Notification, error) {
	return nil, nil
}
func (m *MockNotificationService) GetNotification(ctx context.Context, tenantID uuid.UUID, notificationID string) (*models.Notification, error) {
	return nil, nil
}
func (m *MockNotificationService) MarkAsRead(ctx context.Context, tenantID uuid.UUID, notificationID string) error {
	return nil
}
func (m *MockNotificationService) MarkAllAsRead(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}
func (m *MockNotificationService) DeleteNotification(ctx context.Context, tenantID uuid.UUID, notificationID string) error {
	return nil
}
func (m *MockNotificationService) ArchiveNotification(ctx context.Context, tenantID uuid.UUID, notificationID string) error {
	return nil
}
func (m *MockNotificationService) CreateTemplate(ctx context.Context, tenantID uuid.UUID, template *models.NotificationTemplate) error {
	return nil
}
func (m *MockNotificationService) UpdateTemplate(ctx context.Context, tenantID uuid.UUID, template *models.NotificationTemplate) error {
	return nil
}
func (m *MockNotificationService) DeleteTemplate(ctx context.Context, tenantID uuid.UUID, templateID string) error {
	return nil
}
func (m *MockNotificationService) GetTemplate(ctx context.Context, tenantID uuid.UUID, templateID string) (*models.NotificationTemplate, error) {
	return nil, nil
}
func (m *MockNotificationService) ListTemplates(ctx context.Context, tenantID uuid.UUID, eventType string) ([]*models.NotificationTemplate, error) {
	return nil, nil
}
func (m *MockNotificationService) UpdateNotificationConfig(ctx context.Context, tenantID uuid.UUID, config *models.NotificationConfig) error {
	return nil
}
func (m *MockNotificationService) GetNotificationConfig(ctx context.Context, tenantID uuid.UUID, notificationType models.NotificationType) (*models.NotificationConfig, error) {
	return nil, nil
}
func (m *MockNotificationService) CreateAlert(ctx context.Context, tenantID uuid.UUID, alert *models.Alert) error {
	return nil
}
func (m *MockNotificationService) UpdateAlertStatus(ctx context.Context, tenantID uuid.UUID, alertID string, status string) error {
	return nil
}
func (m *MockNotificationService) ProcessAlert(ctx context.Context, tenantID uuid.UUID, alertID string) error {
	return nil
}
func (m *MockNotificationService) CheckAndTriggerAlerts(ctx context.Context, tenantID uuid.UUID) error {
	return nil
}
func (m *MockNotificationService) CreateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error {
	return nil
}
func (m *MockNotificationService) UpdateWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscription *models.WebhookSubscription) error {
	return nil
}
func (m *MockNotificationService) DeleteWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscriptionID string) error {
	return nil
}
func (m *MockNotificationService) ListWebhookSubscriptions(ctx context.Context, tenantID uuid.UUID) ([]*models.WebhookSubscription, error) {
	return nil, nil
}
func (m *MockNotificationService) GetWebhookSubscription(ctx context.Context, tenantID uuid.UUID, subscriptionID string) (*models.WebhookSubscription, error) {
	return nil, nil
}
func (m *MockNotificationService) UpdateAlertConfig(ctx context.Context, tenantID uuid.UUID, config *models.AlertConfig) error {
	return nil
}
func (m *MockNotificationService) GetAlertConfig(ctx context.Context, tenantID uuid.UUID, alertType models.AlertType) (*models.AlertConfig, error) {
	return nil, nil
}
func (m *MockNotificationService) RenderTemplate(template *models.NotificationTemplate, data map[string]interface{}) (string, error) {
	return "", nil
}
func (m *MockNotificationService) RetryFailedNotifications(ctx context.Context) error { return nil }
func (m *MockNotificationService) SendEmailWithAttachment(ctx context.Context, tenantID uuid.UUID, recipient, subject, body string, attachmentName string, attachmentData []byte) error {
	return nil
}

// MockAuthService
type MockAuthService struct{}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	return "hashed_" + password, nil
}
func (m *MockAuthService) VerifyPassword(hashedPassword, password string) error { return nil }
func (m *MockAuthService) GenerateTokens(ctx context.Context, userID, tenantID uuid.UUID, scope *string) (*models.TokenResponse, error) {
	return nil, nil
}
func (m *MockAuthService) RefreshToken(ctx context.Context, refreshToken string, clientID *string) (*models.TokenResponse, error) {
	return nil, nil
}
func (m *MockAuthService) ValidateToken(ctx context.Context, token string) (*services.TokenClaims, error) {
	return nil, nil
}
func (m *MockAuthService) RevokeToken(ctx context.Context, token string, tokenType *string) error {
	return nil
}
func (m *MockAuthService) GeneratePasswordResetToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return "", nil
}
func (m *MockAuthService) ValidatePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *MockAuthService) ConsumePasswordResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *MockAuthService) GenerateEmailVerificationToken(ctx context.Context, userID uuid.UUID) (string, error) {
	return "", nil
}
func (m *MockAuthService) ConsumeEmailVerificationToken(ctx context.Context, token string) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *MockAuthService) IsAccountLocked(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *MockAuthService) RegisterFailedLoginAttempt(ctx context.Context, userID uuid.UUID) (int, bool, error) {
	return 0, false, nil
}
func (m *MockAuthService) ClearFailedLoginAttempts(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *MockAuthService) GenerateAuthorizationCode(ctx context.Context, userID, tenantID uuid.UUID, clientID string, redirectURI, scope *string) (string, error) {
	return "", nil
}
func (m *MockAuthService) ValidateAuthorizationCode(ctx context.Context, code, clientID, redirectURI string) (*services.AuthorizationCodeClaims, error) {
	return nil, nil
}
func (m *MockAuthService) MarkAuthorizationCodeUsed(ctx context.Context, code string) error {
	return nil
}
func (m *MockAuthService) GetRefreshToken(ctx context.Context, tokenHash string) (*models.RefreshToken, error) {
	return nil, nil
}
func (m *MockAuthService) RevokeUserTokens(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockAuthService) CleanupExpiredTokens(ctx context.Context) error               { return nil }
func (m *MockAuthService) Signup(ctx context.Context, email, password, firstName, lastName string, tenantID *uuid.UUID) (*models.User, error) {
	return nil, nil
}
func (m *MockAuthService) Generate2FASecret(ctx context.Context, userID uuid.UUID) (*otp.Key, error) {
	return nil, nil
}
func (m *MockAuthService) Enable2FA(ctx context.Context, userID uuid.UUID, code string) error {
	return nil
}
func (m *MockAuthService) Disable2FA(ctx context.Context, userID uuid.UUID) error { return nil }
func (m *MockAuthService) Verify2FACode(ctx context.Context, userID uuid.UUID, code string) (bool, error) {
	return true, nil
}
func (m *MockAuthService) GetGoogleAuthURL(state string) string { return "" }
func (m *MockAuthService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error) {
	return nil, "", nil
}
func (m *MockAuthService) CompleteGoogleSignup(ctx context.Context, email, googleID, firstName, lastName, tenantName, subdomain string) (*models.User, error) {
	return nil, nil
}
func (m *MockAuthService) CreateTenant(ctx context.Context, name, subdomain string) (*models.Tenant, error) {
	return &models.Tenant{
		ID:        uuid.New(),
		Name:      name,
		Subdomain: subdomain,
		Status:    "active",
	}, nil
}
func (m *MockAuthService) GetRoleByName(ctx context.Context, tenantID uuid.UUID, roleName string) (*models.Role, error) {
	return &models.Role{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     roleName,
	}, nil
}
func (m *MockAuthService) SeedSuperAdmin(ctx context.Context, email, password string) error {
	return nil
}
