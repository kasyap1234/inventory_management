package middleware

import (
    "context"
    "net/http"
    "strings"

    "agromart2/internal/common"

    "github.com/golang-jwt/jwt/v5"
    "github.com/google/uuid"
    "github.com/labstack/echo/v4"
)






// JWTCustomClaims represents custom JWT claims
type JWTCustomClaims struct {
	UserID   string  `json:"user_id"`
	TenantID string  `json:"tenant_id"`
	Scope    *string `json:"scope,omitempty"`
	TokenID  string  `json:"token_id"`
	ClientID *string `json:"client_id,omitempty"`
	jwt.RegisteredClaims
}

// ParseJWTPayload parses JWT token payload into custom claims
func ParseJWTPayload(c echo.Context, dst *JWTCustomClaims, jwtSecret string) error {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "Missing token")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == authHeader {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token format")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method is HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return err
	}

	if !token.Valid {
		return echo.NewHTTPError(http.StatusUnauthorized, "Token not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid claims")
	}

	// Safe type assertion with checks
	if sub, ok := claims["sub"].(string); ok {
		dst.Subject = sub
	} else {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid subject claim")
	}

	if userID, ok := claims["user_id"].(string); ok {
		dst.UserID = userID
	} else {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid user_id claim")
	}

	if tenantID, ok := claims["tenant_id"].(string); ok {
		dst.TenantID = tenantID
	} else {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid tenant_id claim")
	}

	if scope, ok := claims["scope"].(string); ok {
		scopePtr := &scope
		dst.Scope = scopePtr
	}
	if tokenID, ok := claims["token_id"].(string); ok {
		dst.TokenID = tokenID
	}
	if clientID, ok := claims["client_id"].(string); ok {
		clientIDPtr := &clientID
		dst.ClientID = clientIDPtr
	}

	return nil
}

// JWTMiddleware handles JWT token validation

// Define a minimal interface to avoid importing repositories
type UserTenantResolver interface {
    GetTenantIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error)
}

func JWTMiddleware(userRepo UserTenantResolver, jwtSecret string) echo.MiddlewareFunc {
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
				// Verify the signing method is HMAC
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, echo.NewHTTPError(http.StatusUnauthorized, "Unexpected signing method")
				}
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

			if defaultTenantID == uuid.Nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid tenant ID for user")
			}

			// SECURITY: Tenant context is IMMUTABLE after JWT validation
			// No handler should be able to override the tenant context
			// to prevent cross-tenant data access vulnerabilities

			ctx := context.WithValue(c.Request().Context(), common.UserIDKey, userID)
			ctx = context.WithValue(ctx, common.TenantIDKey, defaultTenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			return next(c)
				}
			}
		}
