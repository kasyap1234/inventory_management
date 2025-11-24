package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"agromart2/internal/common"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRBACService is a mock implementation of RBACService
type MockRBACService struct {
	mock.Mock
}

func (m *MockRBACService) UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error) {
	args := m.Called(ctx, userID, tenantID, permission)
	return args.Bool(0), args.Error(1)
}

func TestRBACMiddleware_RequirePermission(t *testing.T) {
	e := echo.New()
	userID := uuid.New()
	tenantID := uuid.New()

	tests := []struct {
		name           string
		permission     string
		userPerms      map[string]bool
		expectedStatus int
	}{
		{
			name:           "Single permission - Granted",
			permission:     "perm1",
			userPerms:      map[string]bool{"perm1": true},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Single permission - Denied",
			permission:     "perm1",
			userPerms:      map[string]bool{"perm2": true},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "OR logic - First granted",
			permission:     "perm1 || perm2",
			userPerms:      map[string]bool{"perm1": true},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OR logic - Second granted",
			permission:     "perm1 || perm2",
			userPerms:      map[string]bool{"perm2": true},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "OR logic - Neither granted",
			permission:     "perm1 || perm2",
			userPerms:      map[string]bool{"perm3": true},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "AND logic - Both granted",
			permission:     "perm1 && perm2",
			userPerms:      map[string]bool{"perm1": true, "perm2": true},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "AND logic - One missing",
			permission:     "perm1 && perm2",
			userPerms:      map[string]bool{"perm1": true},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Complex logic - (A && B) || C - First group satisfied",
			permission:     "permA && permB || permC",
			userPerms:      map[string]bool{"permA": true, "permB": true},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Complex logic - (A && B) || C - Second group satisfied",
			permission:     "permA && permB || permC",
			userPerms:      map[string]bool{"permC": true},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Complex logic - (A && B) || C - Neither satisfied",
			permission:     "permA && permB || permC",
			userPerms:      map[string]bool{"permA": true},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockRBACService)
			middleware := NewRBACMiddleware(mockService)

			// Setup mock expectations
			// Parse all permissions from the permission string
			allPerms := make(map[string]bool)

			// Split by OR operator
			orGroups := strings.Split(tt.permission, "||")
			for _, group := range orGroups {
				// Split by AND operator
				andPerms := strings.Split(group, "&&")
				for _, p := range andPerms {
					p = strings.TrimSpace(p)
					if p != "" {
						allPerms[p] = false // Default to false
					}
				}
			}

			// Override with user permissions
			for perm, allowed := range tt.userPerms {
				if _, exists := allPerms[perm]; exists {
					allPerms[perm] = allowed
				}
			}

			// Mock each permission explicitly
			for perm, allowed := range allPerms {
				mockService.On("UserHasPermission", mock.Anything, mock.Anything, mock.Anything, perm).Return(allowed, nil)
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Set context values
			ctx := c.Request().Context()
			ctx = context.WithValue(ctx, common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, tenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			handler := middleware.RequirePermission(tt.permission)(func(c echo.Context) error {
				return c.NoContent(http.StatusOK)
			})

			err := handler(c)

			if tt.expectedStatus == http.StatusOK {
				assert.NoError(t, err)
				assert.Equal(t, http.StatusOK, rec.Code)
			} else {
				if err != nil {
					he, ok := err.(*echo.HTTPError)
					assert.True(t, ok)
					assert.Equal(t, tt.expectedStatus, he.Code)
				} else {
					// If handler returned nil but we expected forbidden, that's a fail
					// But wait, if handler returns nil, it means it called next(), which sets status to 200 (in our dummy handler)
					// So we check rec.Code
					assert.Equal(t, tt.expectedStatus, rec.Code)
				}
			}
		})
	}
}
