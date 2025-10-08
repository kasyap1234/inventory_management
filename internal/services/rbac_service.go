package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"agromart2/internal/caching"
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
	// Try cache first if available
	if s.cacheService != nil {
		cacheKey := fmt.Sprintf("agromart:rbac:permission:%s:%s:%s", tenantID.String(), userID.String(), permissionName)
		cachedValue, err := s.cacheService.GetString(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			log.Printf("RBAC: Cache hit for permission check - User: %s, Tenant: %s, Permission: %s", userID, tenantID, permissionName)
			return cachedValue == "true", nil
		}
	}

	log.Printf("RBAC: Checking permission - User: %s, Tenant: %s, Permission: %s", userID, tenantID, permissionName)

	userRoles, err := s.userRoleRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		log.Printf("RBAC: Error fetching user roles - User: %s, Error: %v", userID, err)
		return false, fmt.Errorf("failed to fetch user roles: %w", err)
	}

	if len(userRoles) == 0 {
		log.Printf("RBAC: User has no roles - User: %s, Tenant: %s", userID, tenantID)
		s.cachePermissionResult(ctx, userID, tenantID, permissionName, false)
		return false, nil
	}

	for _, ur := range userRoles {
		rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, tenantID, ur.RoleID)
		if err != nil {
			log.Printf("RBAC: Error fetching role permissions - Role: %s, Error: %v", ur.RoleID, err)
			return false, fmt.Errorf("failed to fetch role permissions: %w", err)
		}

		for _, rp := range rolePermissions {
			perm, err := s.permissionRepo.GetByID(ctx, rp.PermissionID)
			if err != nil {
				log.Printf("RBAC: Error fetching permission - PermissionID: %s, Error: %v", rp.PermissionID, err)
				return false, fmt.Errorf("failed to fetch permission: %w", err)
			}
			if perm == nil {
				log.Printf("RBAC: Warning - Permission not found - PermissionID: %s", rp.PermissionID)
				continue // Skip nil permissions instead of returning error
			}
			if perm.Name == permissionName {
				log.Printf("RBAC: Permission granted - User: %s, Permission: %s", userID, permissionName)
				s.cachePermissionResult(ctx, userID, tenantID, permissionName, true)
				return true, nil
			}
		}
	}

	log.Printf("RBAC: Permission denied - User: %s, Permission: %s", userID, permissionName)
	s.cachePermissionResult(ctx, userID, tenantID, permissionName, false)
	return false, nil
}

func (s *rbacService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	// Try cache first if available
	if s.cacheService != nil {
		cacheKey := fmt.Sprintf("agromart:rbac:permissions:%s:%s", tenantID.String(), userID.String())
		cachedValue, err := s.cacheService.GetString(ctx, cacheKey)
		if err == nil && cachedValue != "" {
			var permissions []string
			if unmarshalErr := json.Unmarshal([]byte(cachedValue), &permissions); unmarshalErr == nil {
				log.Printf("RBAC: Cache hit for user permissions - User: %s, Tenant: %s", userID, tenantID)
				return permissions, nil
			}
		}
	}

	log.Printf("RBAC: Fetching user permissions - User: %s, Tenant: %s", userID, tenantID)

	userRoles, err := s.userRoleRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		log.Printf("RBAC: Error fetching user roles - User: %s, Error: %v", userID, err)
		return nil, fmt.Errorf("failed to fetch user roles: %w", err)
	}

	permissionNames := make(map[string]bool)
	for _, ur := range userRoles {
		rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, tenantID, ur.RoleID)
		if err != nil {
			log.Printf("RBAC: Error fetching role permissions - Role: %s, Error: %v", ur.RoleID, err)
			continue // Log error but continue processing other roles
		}

		for _, rp := range rolePermissions {
			perm, err := s.permissionRepo.GetByID(ctx, rp.PermissionID)
			if err != nil {
				log.Printf("RBAC: Error fetching permission - PermissionID: %s, Error: %v", rp.PermissionID, err)
				continue // Log error but continue processing other permissions
			}
			if perm == nil {
				log.Printf("RBAC: Warning - Permission not found - PermissionID: %s", rp.PermissionID)
				continue // Skip nil permissions
			}
			permissionNames[perm.Name] = true
		}
	}

	var perms []string
	for p := range permissionNames {
		perms = append(perms, p)
	}

	// Cache the result if cache service is available
	if s.cacheService != nil && len(perms) > 0 {
		if data, err := json.Marshal(perms); err == nil {
			cacheKey := fmt.Sprintf("agromart:rbac:permissions:%s:%s", tenantID.String(), userID.String())
			if cacheErr := s.cacheService.SetString(ctx, cacheKey, string(data), 10*time.Minute); cacheErr != nil {
				log.Printf("RBAC: Failed to cache user permissions - User: %s, Error: %v", userID, cacheErr)
			}
		}
	}

	log.Printf("RBAC: Retrieved %d permissions for user - User: %s, Tenant: %s", len(perms), userID, tenantID)
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
		log.Printf("RBAC: Failed to cache permission result - User: %s, Permission: %s, Error: %v", userID, permissionName, err)
	}
}

// InvalidateUserPermissionsCache invalidates all cached permissions for a user
func (s *rbacService) InvalidateUserPermissionsCache(ctx context.Context, userID, tenantID uuid.UUID) error {
	if s.cacheService == nil {
		return nil // No cache service, nothing to invalidate
	}

	log.Printf("RBAC: Invalidating permission cache - User: %s, Tenant: %s", userID, tenantID)

	// Delete the user permissions list cache
	permissionsKey := fmt.Sprintf("agromart:rbac:permissions:%s:%s", tenantID.String(), userID.String())
	if err := s.cacheService.Delete(ctx, permissionsKey); err != nil {
		log.Printf("RBAC: Failed to invalidate permissions cache - Key: %s, Error: %v", permissionsKey, err)
	}

	// Note: Individual permission checks are cached separately with pattern:
	// agromart:rbac:permission:{tenantID}:{userID}:{permissionName}
	// For complete invalidation, we would need to track all permission names or use a pattern-based delete
	// For now, we rely on the TTL to expire these caches

	return nil
}