package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"agromart2/internal/caching"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type RBACService interface {
	UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permissionName string) (bool, error)
	GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error)
	InvalidateUserPermissionsCache(ctx context.Context, userID, tenantID uuid.UUID) error

	// Role CRUD operations
	CreateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error
	GetRoleByID(ctx context.Context, tenantID, roleID uuid.UUID) (*models.Role, error)
	GetRoleByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error)
	UpdateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error
	DeleteRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error
	ListRoles(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Role, error)

	// Permission CRUD operations
	CreatePermission(ctx context.Context, permission *models.Permission) error
	GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error)
	GetPermissionByName(ctx context.Context, name string) (*models.Permission, error)
	UpdatePermission(ctx context.Context, permission *models.Permission) error
	DeletePermission(ctx context.Context, permissionID uuid.UUID) error
	ListPermissions(ctx context.Context, limit, offset int) ([]*models.Permission, error)

	// Role-Permission associations
	AssignPermissionToRole(ctx context.Context, tenantID, roleID, permissionID uuid.UUID) error
	RevokePermissionFromRole(ctx context.Context, tenantID, roleID, permissionID uuid.UUID) error
	GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.Permission, error)
}

type rbacService struct {
	userRoleRepo       repositories.UserRoleRepository
	roleRepo           repositories.RoleRepository
	rolePermissionRepo repositories.RolePermissionRepository
	permissionRepo     repositories.PermissionRepository
	cacheService       caching.CacheService
}

func NewRBACService(userRoleRepo repositories.UserRoleRepository, roleRepo repositories.RoleRepository, rolePermissionRepo repositories.RolePermissionRepository, permissionRepo repositories.PermissionRepository) RBACService {
	return &rbacService{
		userRoleRepo:       userRoleRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
		cacheService:       nil, // Optional: can be nil for backward compatibility
	}
}

func NewRBACServiceWithCache(userRoleRepo repositories.UserRoleRepository, roleRepo repositories.RoleRepository, rolePermissionRepo repositories.RolePermissionRepository, permissionRepo repositories.PermissionRepository, cacheService caching.CacheService) RBACService {
	return &rbacService{
		userRoleRepo:       userRoleRepo,
		roleRepo:           roleRepo,
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

    // Maintain an index of cached permission keys per user to allow precise invalidation
    indexKey := fmt.Sprintf("agromart:rbac:permission:index:%s:%s", tenantID.String(), userID.String())
    if existing, err := s.cacheService.GetString(ctx, indexKey); err == nil {
        var m map[string]bool
        if existing == "" {
            m = map[string]bool{}
        } else {
            if uerr := json.Unmarshal([]byte(existing), &m); uerr != nil {
                m = map[string]bool{}
            }
        }
        if !m[permissionName] {
            m[permissionName] = true
            if data, merr := json.Marshal(m); merr == nil {
                _ = s.cacheService.SetString(ctx, indexKey, string(data), 10*time.Minute)
            }
        }
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

    // Invalidate individual permission check keys using the maintained index
    indexKey := fmt.Sprintf("agromart:rbac:permission:index:%s:%s", tenantID.String(), userID.String())
    if existing, err := s.cacheService.GetString(ctx, indexKey); err == nil && existing != "" {
        var m map[string]bool
        if uerr := json.Unmarshal([]byte(existing), &m); uerr == nil {
            for perm := range m {
                k := fmt.Sprintf("agromart:rbac:permission:%s:%s:%s", tenantID.String(), userID.String(), perm)
                if delErr := s.cacheService.Delete(ctx, k); delErr != nil {
                    log.Printf("RBAC: Failed to delete permission cache key %s: %v", k, delErr)
                }
            }
        }
    }
    // Delete the index key itself
    if err := s.cacheService.Delete(ctx, indexKey); err != nil {
        log.Printf("RBAC: Failed to delete permission index key %s: %v", indexKey, err)
    }

	return nil
}

// Role CRUD operations
func (s *rbacService) CreateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error {
	if role.TenantID != tenantID {
		return fmt.Errorf("role tenant_id does not match provided tenant_id")
	}
	return s.roleRepo.Create(ctx, role)
}

func (s *rbacService) GetRoleByID(ctx context.Context, tenantID, roleID uuid.UUID) (*models.Role, error) {
	return s.roleRepo.GetByID(ctx, tenantID, roleID)
}

func (s *rbacService) GetRoleByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error) {
	return s.roleRepo.GetByName(ctx, tenantID, name)
}

func (s *rbacService) UpdateRole(ctx context.Context, tenantID uuid.UUID, role *models.Role) error {
	if role.TenantID != tenantID {
		return fmt.Errorf("role tenant_id does not match provided tenant_id")
	}
	return s.roleRepo.Update(ctx, role)
}

func (s *rbacService) DeleteRole(ctx context.Context, tenantID uuid.UUID, roleID uuid.UUID) error {
	return s.roleRepo.Delete(ctx, tenantID, roleID)
}

func (s *rbacService) ListRoles(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Role, error) {
	return s.roleRepo.List(ctx, tenantID, limit, offset)
}

// Permission CRUD operations
func (s *rbacService) CreatePermission(ctx context.Context, permission *models.Permission) error {
	return s.permissionRepo.Create(ctx, permission)
}

func (s *rbacService) GetPermissionByID(ctx context.Context, permissionID uuid.UUID) (*models.Permission, error) {
	return s.permissionRepo.GetByID(ctx, permissionID)
}

func (s *rbacService) GetPermissionByName(ctx context.Context, name string) (*models.Permission, error) {
	return s.permissionRepo.GetByName(ctx, name)
}

func (s *rbacService) UpdatePermission(ctx context.Context, permission *models.Permission) error {
	return s.permissionRepo.Update(ctx, permission)
}

func (s *rbacService) DeletePermission(ctx context.Context, permissionID uuid.UUID) error {
	return s.permissionRepo.Delete(ctx, permissionID)
}

func (s *rbacService) ListPermissions(ctx context.Context, limit, offset int) ([]*models.Permission, error) {
	return s.permissionRepo.List(ctx, limit, offset)
}

// Role-Permission associations
func (s *rbacService) AssignPermissionToRole(ctx context.Context, tenantID, roleID, permissionID uuid.UUID) error {
	// Verify role exists and belongs to tenant
	role, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("failed to verify role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// Verify permission exists
	permission, err := s.permissionRepo.GetByID(ctx, permissionID)
	if err != nil {
		return fmt.Errorf("failed to verify permission: %w", err)
	}
	if permission == nil {
		return fmt.Errorf("permission not found")
	}

	// Create the association
	rolePermission := &models.RolePermission{
		ID:           uuid.New(),
		RoleID:       roleID,
		PermissionID: permissionID,
	}

	return s.rolePermissionRepo.Create(ctx, tenantID, rolePermission)
}

func (s *rbacService) RevokePermissionFromRole(ctx context.Context, tenantID, roleID, permissionID uuid.UUID) error {
	return s.rolePermissionRepo.Delete(ctx, tenantID, roleID, permissionID)
}

func (s *rbacService) GetRolePermissions(ctx context.Context, tenantID, roleID uuid.UUID) ([]*models.Permission, error) {
	// Verify role exists and belongs to tenant
	_, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify role: %w", err)
	}

	// Get role-permission associations
	rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}

	// Get actual permissions
	var permissions []*models.Permission
	for _, rp := range rolePermissions {
		permission, err := s.permissionRepo.GetByID(ctx, rp.PermissionID)
		if err != nil {
			log.Printf("Warning: failed to get permission %s: %v", rp.PermissionID, err)
			continue
		}
		if permission != nil {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}