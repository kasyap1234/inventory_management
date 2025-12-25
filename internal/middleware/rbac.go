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

			// Support logical OR using "||" and logical AND using "&&"
			// Logic: (A && B) || (C && D) -> User needs (A AND B) OR (C AND D)
			orGroups := strings.Split(permission, "||")

			for _, group := range orGroups {
				// Check if this group is satisfied (ALL permissions in the group must be present)
				andPerms := strings.Split(group, "&&")
				groupSatisfied := true

				for _, p := range andPerms {
					p = strings.TrimSpace(p)
					if p == "" {
						continue
					}

					hasPermission, err := m.rbacService.UserHasPermission(ctx, userID, tenantID, p)
					if err != nil {
						return echo.NewHTTPError(http.StatusInternalServerError, "Error checking permission")
					}

					if !hasPermission {
						groupSatisfied = false
						break
					}
				}

				// If any group is fully satisfied, allow access
				if groupSatisfied {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
		}
	}
}
