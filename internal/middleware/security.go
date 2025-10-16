package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds common security headers to every response
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			res := c.Response()
			req := c.Request()

			res.Header().Set("X-Content-Type-Options", "nosniff")
			res.Header().Set("X-Frame-Options", "DENY")
			res.Header().Set("X-XSS-Protection", "1; mode=block")
			res.Header().Set("Referrer-Policy", "no-referrer")
			res.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
			res.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self' data:; connect-src 'self'")

			if req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") {
				res.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			}

			return next(c)
		}
	}
}

// EnforceHTTPS redirects HTTP requests to HTTPS when enabled
func EnforceHTTPS(enabled bool) echo.MiddlewareFunc {
	if !enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				return next(c)
			}
		}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			if req.TLS != nil || strings.EqualFold(req.Header.Get("X-Forwarded-Proto"), "https") {
				return next(c)
			}

			host := req.Host
			if strings.HasPrefix(host, "localhost") || strings.HasPrefix(host, "127.0.0.1") {
				return next(c)
			}

			target := "https://" + host + req.RequestURI
			return c.Redirect(http.StatusMovedPermanently, target)
		}
	}
}
