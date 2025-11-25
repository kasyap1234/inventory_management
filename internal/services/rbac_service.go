package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/logging"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type RBACService interface {
	UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permissionName string) (bool, error)
	GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error)
	InvalidateUserPermissionsCache(ctx context.Context, userID, tenantID uuid.UUID) error
}

type rbacService struct {
	userRoleRepo       repositories.UserRoleRepository
	rolePermissionRepo repositories.RolePermissionRepository
	permissionRepo     repositories.PermissionRepository
	cacheService       caching.CacheService
}

func NewRBACService(userRoleRepo repositories.UserRoleRepository, rolePermissionRepo repositories.RolePermissionRepository, permissionRepo repositories.PermissionRepository) RBACService {
	return &rbacService{
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
		cacheService:       nil, // Optional: can be nil for backward compatibility
	}
}

func NewRBACServiceWithCache(userRoleRepo repositories.UserRoleRepository, rolePermissionRepo repositories.RolePermissionRepository, permissionRepo repositories.PermissionRepository, cacheService caching.CacheService) RBACService {
	return &rbacService{
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
		cacheService:       cacheService,
	}
}

func (s *rbacService) UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permissionName string) (bool, error) {
	logger := logging.WithContext(ctx)
	
	// Try cache first if available
	if s.cacheService != nil {
		cacheKey := fmt.Sprintf("agromart:rbac:permission:%s:%s:%s", tenantID.String(), userID.String(), permissionName)
		cachedValue, err := s.cacheService.GetString(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			if logger.IsDebugEnabled() {
				logger.Debug().
					UUID("user_id", userID).
					UUID("tenant_id", tenantID).
					Str("permission", permissionName).
					Msg("RBAC cache hit for permission check")
			}
			return cachedValue == "true", nil
		}
	}

	if logger.IsDebugEnabled() {
		logger.Debug().
			UUID("user_id", userID).
			UUID("tenant_id", tenantID).
			Str("permission", permissionName).
			Msg("RBAC checking permission")
	}

	// Use optimized single-query method that handles role hierarchy
	permissions, err := s.rolePermissionRepo.GetAllUserPermissions(ctx, userID, tenantID)
	if err != nil {
		logger.Error().
			Err(err).
			UUID("user_id", userID).
			Msg("RBAC error fetching user permissions")
		return false, fmt.Errorf("failed to fetch user permissions: %w", err)
	}

	if len(permissions) == 0 {
		if logger.IsDebugEnabled() {
			logger.Debug().
				UUID("user_id", userID).
				UUID("tenant_id", tenantID).
				Msg("RBAC user has no permissions")
		}
		s.cachePermissionResult(ctx, userID, tenantID, permissionName, false)
		return false, nil
	}

	// Check for exact match or wildcard match
	for _, perm := range permissions {
		// Exact match
		if perm.Name == permissionName {
			if logger.IsDebugEnabled() {
				logger.Debug().
					UUID("user_id", userID).
					Str("permission", permissionName).
					Msg("RBAC permission granted (exact match)")
			}
			s.cachePermissionResult(ctx, userID, tenantID, permissionName, true)
			return true, nil
		}
		
		// Wildcard match (e.g., "*", "product.*", "*.read")
		if s.matchesWildcard(perm.Name, permissionName) {
			if logger.IsDebugEnabled() {
				logger.Debug().
					UUID("user_id", userID).
					Str("pattern", perm.Name).
					Str("permission", permissionName).
					Msg("RBAC permission granted (wildcard match)")
			}
			s.cachePermissionResult(ctx, userID, tenantID, permissionName, true)
			return true, nil
		}
	}

	if logger.IsDebugEnabled() {
		logger.Debug().
			UUID("user_id", userID).
			Str("permission", permissionName).
			Msg("RBAC permission denied")
	}
	s.cachePermissionResult(ctx, userID, tenantID, permissionName, false)
	return false, nil
}

func (s *rbacService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	logger := logging.WithContext(ctx)
	
	// Try cache first if available
	if s.cacheService != nil {
		cacheKey := fmt.Sprintf("agromart:rbac:permissions:%s:%s", tenantID.String(), userID.String())
		cachedValue, err := s.cacheService.GetString(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			var permissions []string
			if unmarshalErr := json.Unmarshal([]byte(cachedValue), &permissions); unmarshalErr == nil {
				if logger.IsDebugEnabled() {
					logger.Debug().
						UUID("user_id", userID).
						UUID("tenant_id", tenantID).
						Msg("RBAC cache hit for user permissions")
				}
				return permissions, nil
			}
		}
	}

	if logger.IsDebugEnabled() {
		logger.Debug().
			UUID("user_id", userID).
			UUID("tenant_id", tenantID).
			Msg("RBAC fetching user permissions")
	}

	// Use optimized single-query method that handles role hierarchy
	permissions, err := s.rolePermissionRepo.GetAllUserPermissions(ctx, userID, tenantID)
	if err != nil {
		logger.Error().
			Err(err).
			UUID("user_id", userID).
			Msg("RBAC error fetching user permissions")
		return nil, fmt.Errorf("failed to fetch user permissions: %w", err)
	}

	// Extract permission names
	var perms []string
	for _, p := range permissions {
		perms = append(perms, p.Name)
	}

	// Cache the result if cache service is available
	if s.cacheService != nil && len(perms) > 0 {
		if data, err := json.Marshal(perms); err == nil {
			cacheKey := fmt.Sprintf("agromart:rbac:permissions:%s:%s", tenantID.String(), userID.String())
			if cacheErr := s.cacheService.SetString(ctx, cacheKey, string(data), 10*time.Minute); cacheErr != nil {
				logger.Warn().
					Err(cacheErr).
					UUID("user_id", userID).
					Msg("RBAC failed to cache user permissions")
			}
		}
	}

	if logger.IsDebugEnabled() {
		logger.Debug().
			UUID("user_id", userID).
			UUID("tenant_id", tenantID).
			Int("permission_count", len(perms)).
			Msg("RBAC retrieved user permissions")
	}
	
	return perms, nil
}

// cachePermissionResult caches the result of a permission check
func (s *rbacService) cachePermissionResult(ctx context.Context, userID, tenantID uuid.UUID, permissionName string, hasPermission bool) {
	if s.cacheService == nil {
		return
	}

	cacheKey := fmt.Sprintf("agromart:rbac:permission:%s:%s:%s", tenantID.String(), userID.String(), permissionName)
	value := "false"
	if hasPermission {
		value = "true"
	}

	// Cache for 10 minutes - shorter TTL for security-sensitive data
	if err := s.cacheService.SetString(ctx, cacheKey, value, 10*time.Minute); err != nil {
		logger := logging.WithContext(ctx)
		logger.Warn().
			Err(err).
			UUID("user_id", userID).
			Str("permission", permissionName).
			Msg("RBAC failed to cache permission result")
	}
}

// InvalidateUserPermissionsCache invalidates all cached permissions for a user
func (s *rbacService) InvalidateUserPermissionsCache(ctx context.Context, userID, tenantID uuid.UUID) error {
	logger := logging.WithContext(ctx)
	
	if s.cacheService == nil {
		return nil // No cache service, nothing to invalidate
	}

	if logger.IsDebugEnabled() {
		logger.Debug().
			UUID("user_id", userID).
			UUID("tenant_id", tenantID).
			Msg("RBAC invalidating permission cache")
	}

	// Delete the user permissions list cache
	permissionsKey := fmt.Sprintf("agromart:rbac:permissions:%s:%s", tenantID.String(), userID.String())
	if err := s.cacheService.Delete(ctx, permissionsKey); err != nil {
		logger.Warn().
			Err(err).
			Str("key", permissionsKey).
			Msg("RBAC failed to invalidate permissions cache")
	}

	// Note: Individual permission checks are cached separately with pattern:
	// agromart:rbac:permission:{tenantID}:{userID}:{permissionName}
	// For complete invalidation, we would need to track all permission names or use a pattern-based delete
	// For now, we rely on the TTL to expire these caches

	return nil
}

// matchesWildcard checks if a permission matches a wildcard pattern
// Supports patterns like:
// - "*" (matches all)
// - "product.*" (matches all product permissions)
// - "*.read" (matches all read permissions)
func (s *rbacService) matchesWildcard(pattern, permission string) bool {
	// Full wildcard
	if pattern == "*" {
		return true
	}

	// No wildcard in pattern
	if !strings.Contains(pattern, "*") {
		return false
	}

	// Split both pattern and permission
	patternParts := strings.Split(pattern, ".")
	permParts := strings.Split(permission, ".")

	// Must have same number of parts
	if len(patternParts) != len(permParts) {
		return false
	}

	// Check each part
	for i := range patternParts {
		if patternParts[i] == "*" {
			continue // Wildcard matches anything
		}
		if patternParts[i] != permParts[i] {
			return false
		}
	}

	return true
}
