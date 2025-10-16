package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// APIVersion represents API version information
type APIVersion struct {
	Version     string    `json:"version"`
	Status      string    `json:"status"`      // "active", "deprecated", "sunset"
	SunsetDate  *time.Time `json:"sunset_date,omitempty"`
	Message     string    `json:"message,omitempty"`
}

// VersionMiddleware provides API versioning functionality
type VersionMiddleware struct {
	supportedVersions map[string]APIVersion
	defaultVersion   string
}

// NewVersionMiddleware creates a new version middleware instance
func NewVersionMiddleware() *VersionMiddleware {
	supportedVersions := map[string]APIVersion{
		"v1": {
			Version: "v1",
			Status:  "active",
			Message: "Current stable API version",
		},
	}

	return &VersionMiddleware{
		supportedVersions: supportedVersions,
		defaultVersion:    "v1",
	}
}

// VersionHeader adds version information to response headers
func (vm *VersionMiddleware) VersionHeader(version string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Add API version headers
			c.Response().Header().Set("X-API-Version", version)

			// Add versioning information
			if ver, exists := vm.supportedVersions[version]; exists {
				if ver.Status == "deprecated" {
					c.Response().Header().Set("X-API-Deprecated", "true")
					c.Response().Header().Set("X-API-Sunset", ver.SunsetDate.Format(time.RFC3339))
					c.Response().Header().Set("Warning", "299 agromart2 \"This API version is deprecated and will be removed on "+ver.SunsetDate.Format("2006-01-02")+"\"")
				}
				c.Response().Header().Set("X-API-Message", ver.Message)
			}

			return next(c)
		}
	}
}

// VersionRoute creates a version-specific route group
func (vm *VersionMiddleware) VersionRoute(e *echo.Echo, version string) *echo.Group {
	group := e.Group("/" + version)

	// Apply version middleware
	group.Use(vm.VersionHeader(version))

	return group
}

// APIVersionResolver resolves the API version from the request
func (vm *VersionMiddleware) APIVersionResolver() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestPath := c.Request().URL.Path

			// Check if path starts with a version prefix
			version := vm.extractVersionFromPath(requestPath)
			if version != "" {
				if _, supported := vm.supportedVersions[version]; !supported {
					return c.JSON(http.StatusNotFound, map[string]string{
						"error": "Unsupported API version",
						"supported_versions": strings.Join(vm.getSupportedVersions(), ", "),
					})
				}
				c.Set("api_version", version)
			} else {
				// No version specified, use default
				c.Set("api_version", vm.defaultVersion)
			}

			return next(c)
		}
	}
}

// extractVersionFromPath extracts the API version from the URL path
func (vm *VersionMiddleware) extractVersionFromPath(path string) string {
	// Check for patterns like /v1, /v2, etc.
	if len(path) >= 3 && path[0] == '/' {
		if path[1] == 'v' {
			// Try to parse version number
			if versionNum, err := strconv.Atoi(path[2:3]); err == nil && versionNum > 0 {
				return "v" + strconv.Itoa(versionNum)
			}
		}
	}
	return ""
}

// getSupportedVersions returns a list of supported API versions
func (vm *VersionMiddleware) getSupportedVersions() []string {
	var versions []string
	for version, info := range vm.supportedVersions {
		if info.Status == "active" || info.Status == "deprecated" {
			versions = append(versions, version)
		}
	}
	return versions
}

// DeprecationNotice adds deprecation notice for specific endpoints
func (vm *VersionMiddleware) DeprecationNotice(version string, message string, sunsetDate *time.Time) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if ver, exists := vm.supportedVersions[version]; exists {
				// Set deprecation headers
				c.Response().Header().Set("X-API-Deprecated", "true")
				if sunsetDate != nil {
					c.Response().Header().Set("X-API-Sunset", sunsetDate.Format(time.RFC3339))
					c.Response().Header().Set("Warning", "299 agromart2 \"This endpoint is deprecated and will be removed on "+sunsetDate.Format("2006-01-02")+"\"")
				} else if ver.SunsetDate != nil {
					c.Response().Header().Set("X-API-Sunset", ver.SunsetDate.Format(time.RFC3339))
					c.Response().Header().Set("Warning", "299 agromart2 \"This endpoint is deprecated and will be removed on "+ver.SunsetDate.Format("2006-01-02")+"\"")
				}
				if message != "" {
					c.Response().Header().Set("X-API-Deprecation-Message", message)
				}
			}

			return next(c)
		}
	}
}

// GetCurrentVersion returns the current active API version
func (vm *VersionMiddleware) GetCurrentVersion() string {
	return vm.defaultVersion
}

// GetSupportedVersions returns all supported API versions
func (vm *VersionMiddleware) GetSupportedVersions() map[string]APIVersion {
	return vm.supportedVersions
}

// AddVersion adds a new API version with its configuration
func (vm *VersionMiddleware) AddVersion(version string, status string, message string, sunsetDate *time.Time) {
	vm.supportedVersions[version] = APIVersion{
		Version:    version,
		Status:     status,
		SunsetDate: sunsetDate,
		Message:    message,
	}
}