package services

import (
	"context"
	"fmt"

	"agromart2/internal/models"

	"github.com/google/uuid"
)

// DefaultRoleService handles automatic role assignment for new users
type DefaultRoleService interface {
	// AssignDefaultRole assigns the tenant's default role to a user
	// If the tenant has no default role configured, assigns the "user" role
	AssignDefaultRole(ctx context.Context, tenantID, userID uuid.UUID) error

	// AssignRolesFromInvitation assigns roles specified in an invitation
	// Falls back to default role if no roles specified in invitation
	AssignRolesFromInvitation(ctx context.Context, tenantID, userID uuid.UUID, invitation *models.Invitation) error

	// GetDefaultRole returns the default role for a tenant
	GetDefaultRole(ctx context.Context, tenantID uuid.UUID) (*models.Role, error)

	// SetDefaultRole sets the default role for a tenant
	SetDefaultRole(ctx context.Context, tenantID, roleID uuid.UUID) error
}

// TenantRepoForDefaultRole defines tenant repository dependencies
type TenantRepoForDefaultRole interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	Update(ctx context.Context, tenant *models.Tenant) error
}

// RoleRepoForDefaultRole defines role repository dependencies
type RoleRepoForDefaultRole interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Role, error)
	AssignUserToRole(ctx context.Context, tenantID uuid.UUID, userID, roleID uuid.UUID) error
}

type defaultRoleService struct {
	tenantRepo TenantRepoForDefaultRole
	roleRepo   RoleRepoForDefaultRole
}

// NewDefaultRoleService creates a new default role service
func NewDefaultRoleService(
	tenantRepo TenantRepoForDefaultRole,
	roleRepo RoleRepoForDefaultRole,
) DefaultRoleService {
	return &defaultRoleService{
		tenantRepo: tenantRepo,
		roleRepo:   roleRepo,
	}
}

// DefaultRoleName is the default role name to assign when no default is configured
const DefaultRoleName = "user"

// DefaultRolePermissions defines the basic permissions for the default "user" role
// These are suitable for a B2B inventory management app
var DefaultRolePermissions = []string{
	// Product access - read and create
	"product.list",
	"product.read",
	"product.create",
	"product.update",
	// Inventory access - read and basic operations
	"inventory.list",
	"inventory.read",
	"inventory.create",
	"inventory.update",
	// Order access - full CRUD for placing and managing orders
	"order.list",
	"order.read",
	"order.create",
	"order.update",
	// Read-only access to supporting entities
	"warehouse.list",
	"warehouse.read",
	"category.list",
	"category.read",
	"supplier.list",
	"supplier.read",
	"distributor.list",
	"distributor.read",
	"batch.list",
	"batch.read",
	"invoice.list",
	"invoice.read",
}

// AssignDefaultRole assigns the tenant's default role to a user
func (s *defaultRoleService) AssignDefaultRole(ctx context.Context, tenantID, userID uuid.UUID) error {
	// Get the default role for this tenant
	defaultRole, err := s.GetDefaultRole(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get default role: %w", err)
	}

	// Assign the role to the user
	if err := s.roleRepo.AssignUserToRole(ctx, tenantID, userID, defaultRole.ID); err != nil {
		return fmt.Errorf("failed to assign default role: %w", err)
	}

	return nil
}

// AssignRolesFromInvitation assigns roles specified in an invitation
func (s *defaultRoleService) AssignRolesFromInvitation(ctx context.Context, tenantID, userID uuid.UUID, invitation *models.Invitation) error {
	// If invitation has specific roles, assign those
	if len(invitation.RoleIDs) > 0 {
		for _, roleID := range invitation.RoleIDs {
			if err := s.roleRepo.AssignUserToRole(ctx, tenantID, userID, roleID); err != nil {
				return fmt.Errorf("failed to assign role %s from invitation: %w", roleID, err)
			}
		}
		return nil
	}

	// Fall back to the legacy single RoleID if present
	if invitation.RoleID != uuid.Nil {
		if err := s.roleRepo.AssignUserToRole(ctx, tenantID, userID, invitation.RoleID); err != nil {
			return fmt.Errorf("failed to assign role from invitation: %w", err)
		}
		return nil
	}

	// Fall back to the tenant's default role
	return s.AssignDefaultRole(ctx, tenantID, userID)
}

// GetDefaultRole returns the default role for a tenant
func (s *defaultRoleService) GetDefaultRole(ctx context.Context, tenantID uuid.UUID) (*models.Role, error) {
	// Get tenant to check for configured default role
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// If tenant has a default role configured, use it
	if tenant.DefaultRoleID != nil {
		role, err := s.roleRepo.GetByID(ctx, tenantID, *tenant.DefaultRoleID)
		if err == nil {
			return role, nil
		}
		// If configured role not found, fall back to "user" role
	}

	// Fall back to the standard "user" role
	role, err := s.roleRepo.GetByName(ctx, tenantID, DefaultRoleName)
	if err != nil {
		return nil, fmt.Errorf("default role '%s' not found for tenant: %w", DefaultRoleName, err)
	}

	return role, nil
}

// SetDefaultRole sets the default role for a tenant
func (s *defaultRoleService) SetDefaultRole(ctx context.Context, tenantID, roleID uuid.UUID) error {
	// Verify the role exists and belongs to the tenant
	_, err := s.roleRepo.GetByID(ctx, tenantID, roleID)
	if err != nil {
		return fmt.Errorf("role not found or doesn't belong to tenant: %w", err)
	}

	// Get tenant
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant: %w", err)
	}

	// Update the default role
	tenant.DefaultRoleID = &roleID
	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return fmt.Errorf("failed to update tenant default role: %w", err)
	}

	return nil
}
