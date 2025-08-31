package middleware

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/services"

	"github.com/labstack/echo/v4"
)

type RBACMiddleware struct {
	rbacService services.RBACService
}

func NewRBACMiddleware(rbacService services.RBACService) *RBACMiddleware {
	return &RBACMiddleware{
		rbacService: rbacService,
	}
}

func (m *RBACMiddleware) RequirePermission(permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			userID, ok := common.GetUserIDFromContext(ctx)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
			}
			tenantID, ok := common.GetTenantIDFromContext(ctx)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
			}

			hasPermission, err := m.rbacService.UserHasPermission(ctx, userID, tenantID, permission)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
			}
			if !hasPermission {
				return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
			}

			return next(c)
		}
	}
}