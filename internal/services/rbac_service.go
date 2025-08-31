package services

import (
	"context"

	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type RBACService interface {
	UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permissionName string) (bool, error)
	GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error)
}

type rbacService struct {
	userRoleRepo       repositories.UserRoleRepository
	rolePermissionRepo repositories.RolePermissionRepository
	permissionRepo     repositories.PermissionRepository
}

func NewRBACService(userRoleRepo repositories.UserRoleRepository, rolePermissionRepo repositories.RolePermissionRepository, permissionRepo repositories.PermissionRepository) RBACService {
	return &rbacService{
		userRoleRepo:       userRoleRepo,
		rolePermissionRepo: rolePermissionRepo,
		permissionRepo:     permissionRepo,
	}
}

func (s *rbacService) UserHasPermission(ctx context.Context, userID, tenantID uuid.UUID, permissionName string) (bool, error) {
	userRoles, err := s.userRoleRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	for _, ur := range userRoles {
		rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, tenantID, ur.RoleID)
		if err != nil {
			return false, err
		}

		for _, rp := range rolePermissions {
			perm, err := s.permissionRepo.GetByID(ctx, rp.PermissionID)
			if err != nil {
				continue
			}
			if perm.Name == permissionName {
				return true, nil
			}
		}
	}

	return false, nil
}

func (s *rbacService) GetUserPermissions(ctx context.Context, userID, tenantID uuid.UUID) ([]string, error) {
	userRoles, err := s.userRoleRepo.ListByUser(ctx, tenantID, userID)
	if err != nil {
		return nil, err
	}

	permissionNames := make(map[string]bool)
	for _, ur := range userRoles {
		rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, tenantID, ur.RoleID)
		if err != nil {
			continue
		}

		for _, rp := range rolePermissions {
			perm, err := s.permissionRepo.GetByID(ctx, rp.PermissionID)
			if err != nil {
				continue
			}
			permissionNames[perm.Name] = true
		}
	}

	var perms []string
	for p := range permissionNames {
		perms = append(perms, p)
	}
	return perms, nil
}