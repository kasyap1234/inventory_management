package middleware

import (
	"context"
	"os"
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

		// Skip tenant resolution for health checks and metrics
		path := c.Request().URL.Path
		if strings.HasPrefix(path, "/health") || strings.HasPrefix(path, "/metrics") || path == "/v1/security/csrf" {
			return next(c)
		}

		// Get the configured main domain
		mainDomain := strings.ToLower(os.Getenv("DOMAIN_NAME"))
		hostLower := strings.ToLower(host)

		// If the host is exactly the main domain, there is no tenant subdomain
		if hostLower == mainDomain || mainDomain == "" {
			return next(c)
		}

		// Check if the host ends with .mainDomain (indicating a subdomain)
		if !strings.HasSuffix(hostLower, "."+mainDomain) {
			// Not a subdomain of our main domain, skip resolution
			return next(c)
		}

		// Extract the subdomain part
		subdomain := strings.TrimSuffix(hostLower, "."+mainDomain)

		// Skip for common non-tenant subdomains
		if subdomain == "" || subdomain == "www" || subdomain == "api" || subdomain == "app" {
			return next(c)
		}

		ctx := c.Request().Context()
		tenant, err := m.tenantService.GetBySubdomain(ctx, subdomain)
		if err != nil {
			// If subdomain exists but tenant not found, we don't set the tenant context.
			// Handlers that require a tenant will fail later with a clear error.
			return next(c)
		}

		// Set tenant in context
		c.Set("tenant", tenant)
		ctx = context.WithValue(ctx, common.TenantIDKey, tenant.ID)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}
