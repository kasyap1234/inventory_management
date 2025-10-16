package middleware

import (
    "context"
    "net/http"
    "strings"

    "agromart2/internal/common"
    "github.com/google/uuid"

    "github.com/labstack/echo/v4"
)

type RBACMiddleware struct {
    rbacService RBACService
}

// Define a narrow interface to avoid importing services
type RBACService interface {
    UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permission string) (bool, error)
}

func NewRBACMiddleware(rbacService RBACService) *RBACMiddleware {
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

			// Support logical OR using "||" between permission names (e.g., "webhooks:test||notifications:manage")
			perms := []string{strings.TrimSpace(permission)}
			if strings.Contains(permission, "||") {
				raw := strings.Split(permission, "||")
				perms = make([]string, 0, len(raw))
				for _, p := range raw {
					p = strings.TrimSpace(p)
					if p != "" {
						perms = append(perms, p)
					}
				}
			}

			// Evaluate OR - allow if any permission is granted
			for _, p := range perms {
				hasPermission, err := m.rbacService.UserHasPermission(ctx, userID, tenantID, p)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
				}
				if hasPermission {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}