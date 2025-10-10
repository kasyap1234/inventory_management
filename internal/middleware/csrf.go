package middleware

import (
	"log"
	"net/http"
	"strings"

	"agromart2/internal/security"

	"github.com/labstack/echo/v4"
)

// CSRFProtection validates CSRF tokens on state-changing requests using the provided token manager.
// Paths listed in skipPaths are excluded from validation.
func CSRFProtection(manager *security.CSRFTokenManager, skipPaths map[string]struct{}) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()
			if _, ok := skipPaths[request.URL.Path]; ok {
				return next(c)
			}

			method := strings.ToUpper(request.Method)
			if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
				return next(c)
			}

			token := request.Header.Get("X-CSRF-Token")
			if token == "" {
				return echo.NewHTTPError(http.StatusForbidden, "CSRF token missing")
			}

			if err := manager.ValidateToken(token); err != nil {
				log.Printf("CSRF validation failed for %s %s: %v", request.Method, request.URL.Path, err)
				return echo.NewHTTPError(http.StatusForbidden, "Invalid CSRF token")
			}

			return next(c)
		}
	}
}
