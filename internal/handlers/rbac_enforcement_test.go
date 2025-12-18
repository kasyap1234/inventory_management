package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRBACService for testing
type MockRBACService struct {
	mock.Mock
}

func (m *MockRBACService) UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, permission)
	return args.Bool(0), args.Error(1)
}

// Mock repositories
type MockUserRepo struct {
	mock.Mock
	repositories.UserRepository
}

type MockTenantRepo struct {
	mock.Mock
	repositories.TenantRepository
}

type MockRoleService struct {
	mock.Mock
}

func TestCreateUser_RBACEnforcement(t *testing.T) {
	e := echo.New()

	userID := uuid.New()
	tenantID := uuid.New()

	// NOTE: This test validates RBAC middleware behavior ONLY, not the full CreateUser handler.
	// The middleware only checks the permission specified in RequirePermission().
	// Additional permission checks (user.manage_roles, role.create, etc.) happen inside
	// the CreateUser handler and require integration tests with mocked repositories.
	tests := []struct {
		name           string
		hasCreatePerm  bool
		requestBody    map[string]interface{}
		expectedStatus int
		description    string
	}{
		{
			name:          "User with user.create permission passes middleware",
			hasCreatePerm: true,
			requestBody: map[string]interface{}{
				"email":      "test@example.com",
				"first_name": "Test",
				"last_name":  "User",
				"tenant_id":  tenantID.String(),
			},
			expectedStatus: http.StatusCreated,
			description:    "Should allow request when user.create permission exists",
		},
		{
			name:          "User without user.create permission is denied by middleware",
			hasCreatePerm: false,
			requestBody: map[string]interface{}{
				"email":      "test@example.com",
				"first_name": "Test",
				"last_name":  "User",
				"tenant_id":  tenantID.String(),
			},
			expectedStatus: http.StatusForbidden,
			description:    "Should deny request when user.create permission is missing",
		},
		{
			// Middleware doesn't inspect request body - role_id check happens in handler
			name:          "User with user.create passes middleware even with role_id in request",
			hasCreatePerm: true,
			requestBody: map[string]interface{}{
				"email":      "test@example.com",
				"first_name": "Test",
				"last_name":  "User",
				"tenant_id":  tenantID.String(),
				"role_id":    uuid.New().String(),
			},
			expectedStatus: http.StatusCreated,
			description:    "Middleware passes with user.create; user.manage_roles check is inside handler",
		},
		{
			// Middleware doesn't inspect request body - permissions check happens in handler
			name:          "User with user.create passes middleware even with permissions in request",
			hasCreatePerm: true,
			requestBody: map[string]interface{}{
				"email":       "test@example.com",
				"first_name":  "Test",
				"last_name":   "User",
				"tenant_id":   tenantID.String(),
				"permissions": []string{"product.read", "product.create"},
			},
			expectedStatus: http.StatusCreated,
			description:    "Middleware passes with user.create; role.create check is inside handler",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRBAC := new(MockRBACService)
			mockRBAC.On("UserHasPermission", mock.Anything, userID, tenantID, "user.create").Return(tt.hasCreatePerm, nil)

			rbacMiddleware := middleware.NewRBACMiddleware(mockRBAC)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set context values
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			// Test RBAC middleware directly - only checks user.create
			handler := rbacMiddleware.RequirePermission("user.create")(func(c echo.Context) error {
				return c.NoContent(http.StatusCreated)
			})

			err := handler(c)

			if tt.expectedStatus == http.StatusCreated {
				assert.NoError(t, err, tt.description)
			} else {
				assert.Error(t, err, tt.description)
				if err != nil {
					he, ok := err.(*echo.HTTPError)
					assert.True(t, ok, tt.description)
					if ok && he != nil {
						assert.Equal(t, tt.expectedStatus, he.Code, tt.description)
					}
				}
			}

			mockRBAC.AssertExpectations(t)
		})
	}
}

func TestRoleHandlers_RBACEnforcement(t *testing.T) {
	e := echo.New()
	userID := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		permission     string
		hasPermission  bool
		expectedStatus int
		description    string
	}{
		{
			name:           "ListRoles requires role.list",
			endpoint:       "/roles",
			method:         http.MethodGet,
			permission:     "role.list",
			hasPermission:  true,
			expectedStatus: http.StatusOK,
			description:    "Should allow listing roles with role.list permission",
		},
		{
			name:           "ListRoles denied without role.list",
			endpoint:       "/roles",
			method:         http.MethodGet,
			permission:     "role.list",
			hasPermission:  false,
			expectedStatus: http.StatusForbidden,
			description:    "Should deny listing roles without role.list permission",
		},
		{
			name:           "CreateRole requires role.create",
			endpoint:       "/roles",
			method:         http.MethodPost,
			permission:     "role.create",
			hasPermission:  true,
			expectedStatus: http.StatusCreated,
			description:    "Should allow creating roles with role.create permission",
		},
		{
			name:           "CreateRole denied without role.create",
			endpoint:       "/roles",
			method:         http.MethodPost,
			permission:     "role.create",
			hasPermission:  false,
			expectedStatus: http.StatusForbidden,
			description:    "Should deny creating roles without role.create permission",
		},
		{
			name:           "AssignPermissionsToRole requires role.manage_permissions",
			endpoint:       "/roles/123/permissions",
			method:         http.MethodPost,
			permission:     "role.manage_permissions",
			hasPermission:  true,
			expectedStatus: http.StatusOK,
			description:    "Should allow assigning permissions with role.manage_permissions",
		},
		{
			name:           "AssignPermissionsToRole denied without role.manage_permissions",
			endpoint:       "/roles/123/permissions",
			method:         http.MethodPost,
			permission:     "role.manage_permissions",
			hasPermission:  false,
			expectedStatus: http.StatusForbidden,
			description:    "Should deny assigning permissions without role.manage_permissions",
		},
		{
			name:           "AssignRolesToUser requires user.manage_roles",
			endpoint:       "/users/123/roles",
			method:         http.MethodPost,
			permission:     "user.manage_roles",
			hasPermission:  true,
			expectedStatus: http.StatusOK,
			description:    "Should allow assigning roles to users with user.manage_roles",
		},
		{
			name:           "AssignRolesToUser denied without user.manage_roles",
			endpoint:       "/users/123/roles",
			method:         http.MethodPost,
			permission:     "user.manage_roles",
			hasPermission:  false,
			expectedStatus: http.StatusForbidden,
			description:    "Should deny assigning roles without user.manage_roles",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRBAC := new(MockRBACService)
			mockRBAC.On("UserHasPermission", mock.Anything, userID, tenantID, tt.permission).Return(tt.hasPermission, nil)

			rbacMiddleware := middleware.NewRBACMiddleware(mockRBAC)

			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set context values
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			handler := rbacMiddleware.RequirePermission(tt.permission)(func(c echo.Context) error {
				return c.NoContent(tt.expectedStatus)
			})

			err := handler(c)

			if tt.expectedStatus == http.StatusForbidden {
				assert.Error(t, err, tt.description)
				he, ok := err.(*echo.HTTPError)
				assert.True(t, ok, tt.description)
				assert.Equal(t, http.StatusForbidden, he.Code, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}

			mockRBAC.AssertExpectations(t)
		})
	}
}

func TestInvitationHandlers_RBACEnforcement(t *testing.T) {
	e := echo.New()
	userID := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name           string
		endpoint       string
		method         string
		permission     string
		hasPermission  bool
		expectedStatus int
		description    string
	}{
		{
			name:           "CreateInvitation requires user.invite",
			endpoint:       "/invitations",
			method:         http.MethodPost,
			permission:     "user.invite",
			hasPermission:  true,
			expectedStatus: http.StatusCreated,
			description:    "Should allow creating invitations with user.invite permission",
		},
		{
			name:           "CreateInvitation denied without user.invite",
			endpoint:       "/invitations",
			method:         http.MethodPost,
			permission:     "user.invite",
			hasPermission:  false,
			expectedStatus: http.StatusForbidden,
			description:    "Should deny creating invitations without user.invite permission",
		},
		{
			name:           "ListInvitations requires user.list",
			endpoint:       "/invitations",
			method:         http.MethodGet,
			permission:     "user.list",
			hasPermission:  true,
			expectedStatus: http.StatusOK,
			description:    "Should allow listing invitations with user.list permission",
		},
		{
			name:           "RevokeInvitation requires user.invite",
			endpoint:       "/invitations/123",
			method:         http.MethodDelete,
			permission:     "user.invite",
			hasPermission:  true,
			expectedStatus: http.StatusOK,
			description:    "Should allow revoking invitations with user.invite permission",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRBAC := new(MockRBACService)
			mockRBAC.On("UserHasPermission", mock.Anything, userID, tenantID, tt.permission).Return(tt.hasPermission, nil)

			rbacMiddleware := middleware.NewRBACMiddleware(mockRBAC)

			req := httptest.NewRequest(tt.method, tt.endpoint, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set context values
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			handler := rbacMiddleware.RequirePermission(tt.permission)(func(c echo.Context) error {
				return c.NoContent(tt.expectedStatus)
			})

			err := handler(c)

			if tt.expectedStatus == http.StatusForbidden {
				assert.Error(t, err, tt.description)
				he, ok := err.(*echo.HTTPError)
				assert.True(t, ok, tt.description)
				assert.Equal(t, http.StatusForbidden, he.Code, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}

			mockRBAC.AssertExpectations(t)
		})
	}
}
