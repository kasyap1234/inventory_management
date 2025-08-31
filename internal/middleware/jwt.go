package middleware

import (
	"context"
	"net/http"
	"strings"

	"agromart2/internal/common"
	"agromart2/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)






// JWTMiddleware handles JWT token validation

func JWTMiddleware(userRepo repositories.UserRepository, jwtSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			if !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Token not valid")
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid claims")
			}

			sub, ok := claims["sub"].(string)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing user_id in token")
			}

			userID, err := uuid.Parse(sub)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user_id format")
			}

			defaultTenantID, err := userRepo.GetTenantIDByUserID(c.Request().Context(), userID)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "User not found")
			}

			// Check for explicit tenant_id override in request context (set by handlers)
			// Handle both direct UUID and uuid.UUID types
			if explicitTenantID := c.Get("explicit_tenant_id"); explicitTenantID != nil {
				if tenantUUID, ok := explicitTenantID.(uuid.UUID); ok {
					// Use explicit tenant_id for cross-tenant operations
					defaultTenantID = tenantUUID
				}
			}

			ctx := context.WithValue(c.Request().Context(), common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, defaultTenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
				}
			}
		}
		
		// GetTenantIDFromContext extracts tenant ID from request context
		func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
			tenantID, ok := ctx.Value(common.TenantIDKey).(uuid.UUID)
			return tenantID, ok
		}
