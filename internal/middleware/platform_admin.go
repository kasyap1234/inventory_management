package middleware

import (
	"context"
	"net/http"

	"agromart2/internal/common"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// PlatformAdminMiddleware handles platform admin authorization checks
type PlatformAdminMiddleware struct {
	userRepo UserRepoForPlatformAdmin
}

// UserRepoForPlatformAdmin defines the minimal interface needed for platform admin checks
type UserRepoForPlatformAdmin interface {
	IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

// NewPlatformAdminMiddleware creates a new platform admin middleware
func NewPlatformAdminMiddleware(userRepo UserRepoForPlatformAdmin) *PlatformAdminMiddleware {
	return &PlatformAdminMiddleware{
		userRepo: userRepo,
	}
}

// RequirePlatformAdmin middleware restricts access to platform admins only
// Platform admins are super admins who can manage all tenants
func (m *PlatformAdminMiddleware) RequirePlatformAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			
			userID, ok := common.GetUserIDFromContext(ctx)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
			}

			isPlatformAdmin, err := m.userRepo.IsPlatformAdmin(ctx, userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error checking platform admin status")
			}

			if !isPlatformAdmin {
				return echo.NewHTTPError(http.StatusForbidden, "Platform admin access required")
			}

			return next(c)
		}
	}
}

// AllowPlatformAdminOrTenantAdmin middleware allows access if user is either:
// - A platform admin (super admin)
// - A tenant admin within their own tenant
func (m *PlatformAdminMiddleware) AllowPlatformAdminOrTenantAdmin(rbacService RBACService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			
			userID, ok := common.GetUserIDFromContext(ctx)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
			}

			// Check if platform admin first
			isPlatformAdmin, err := m.userRepo.IsPlatformAdmin(ctx, userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error checking platform admin status")
			}

			if isPlatformAdmin {
				return next(c)
			}

			// Not platform admin, check if tenant admin
			tenantID, ok := common.GetTenantIDFromContext(ctx)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
			}

			// Check for admin permission in tenant
			hasPermission, err := rbacService.UserHasPermission(ctx, userID, tenantID, "user.manage_roles")
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permissions")
			}

			if !hasPermission {
				return echo.NewHTTPError(http.StatusForbidden, "Admin access required")
			}

			return next(c)
		}
	}
}
