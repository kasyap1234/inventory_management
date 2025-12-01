package middleware

import (
	"context"
	"net/http"
	"strings"

	"agromart2/internal/common"
	"agromart2/internal/services"

	"github.com/labstack/echo/v4"
)

type SubdomainMiddleware struct {
	tenantService services.TenantService
}

func NewSubdomainMiddleware(tenantService services.TenantService) *SubdomainMiddleware {
	return &SubdomainMiddleware{tenantService: tenantService}
}

func (m *SubdomainMiddleware) ResolveTenant(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		host := c.Request().Host
		// Remove port if present
		if strings.Contains(host, ":") {
			host = strings.Split(host, ":")[0]
		}

		parts := strings.Split(host, ".")
		var subdomain string

		// Logic to extract subdomain.
		// Assuming format: subdomain.domain.com (3 parts)
		// Localhost logic: subdomain.localhost (2 parts)
		// Adjust based on environment/config if needed.
		if len(parts) >= 3 {
			subdomain = parts[0]
		} else if len(parts) == 2 && parts[1] == "localhost" {
			subdomain = parts[0]
		}

		// Skip for main domain or www
		if subdomain == "" || subdomain == "www" || subdomain == "api" {
			return next(c)
		}

		ctx := c.Request().Context()
		tenant, err := m.tenantService.GetBySubdomain(ctx, subdomain)
		if err != nil {
			// If subdomain exists but tenant not found, return 404 or redirect to signup
			return echo.NewHTTPError(http.StatusNotFound, "Tenant not found")
		}

		// Set tenant in context
		c.Set("tenant", tenant)
		ctx = context.WithValue(ctx, common.TenantIDKey, tenant.ID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
