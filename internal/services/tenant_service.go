package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type TenantService interface {
	Create(ctx context.Context, req *CreateTenantRequest) (*models.Tenant, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error)
	Update(ctx context.Context, req *UpdateTenantRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]*models.Tenant, error)
	GetTenantSettings(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error)
	UpdateTenantSettings(ctx context.Context, req *UpdateTenantSettingsRequest) error
}

// PermissionLister abstracts permission listing for seeding defaults.
// Defined locally to avoid pulling the full repository interface into this service.
type PermissionLister interface {
	ListPermissions(ctx context.Context) ([]*models.Permission, error)
}

// RolePermissionAssigner abstracts creating role-permission mappings.
type RolePermissionAssigner interface {
	Create(ctx context.Context, tenantID uuid.UUID, rolePermission *models.RolePermission) error
}

// RoleTemplateApplier abstracts applying role templates
type RoleTemplateApplier interface {
	ApplyTemplateToTenant(ctx context.Context, tenantID uuid.UUID, templateName string) error
}

type tenantService struct {
	tenantRepo             repositories.TenantRepository
	invitationService      InvitationService
	roleRepo               repositories.RoleRepository
	permissionLister       PermissionLister
	rolePermissionAssigner RolePermissionAssigner
	roleTemplateApplier    RoleTemplateApplier
}

func NewTenantService(
	tenantRepo repositories.TenantRepository,
	invitationService InvitationService,
	roleRepo repositories.RoleRepository,
	permissionLister PermissionLister,
	rolePermissionAssigner RolePermissionAssigner,
	roleTemplateApplier RoleTemplateApplier,
) TenantService {
	return &tenantService{
		tenantRepo:             tenantRepo,
		invitationService:      invitationService,
		roleRepo:               roleRepo,
		permissionLister:       permissionLister,
		rolePermissionAssigner: rolePermissionAssigner,
		roleTemplateApplier:    roleTemplateApplier,
	}
}

type CreateTenantRequest struct {
	Name       string `json:"name" validate:"required"`
	Subdomain  string `json:"subdomain" validate:"required"`
	License    string `json:"license"`
	AdminEmail string `json:"admin_email" validate:"required,email"`
}

type UpdateTenantRequest struct {
	ID        uuid.UUID
	Name      string `json:"name" validate:"required"`
	Subdomain string `json:"subdomain" validate:"required"`
	License   string `json:"license"`
	Status    string `json:"status" validate:"required"`
}

type UpdateTenantSettingsRequest struct {
	TenantID  uuid.UUID
	Name      string `json:"name" validate:"required"`
	Subdomain string `json:"subdomain" validate:"required"`
	License   string `json:"license"`
}

func (s *tenantService) Create(ctx context.Context, req *CreateTenantRequest) (*models.Tenant, error) {
	if req.Name == "" || req.Subdomain == "" {
		return nil, errors.New("name and subdomain are required")
	}
	// Basic validation - check for any spaces in subdomain
	if strings.Contains(req.Subdomain, " ") {
		return nil, errors.New("subdomain cannot have spaces")
	}

	tenant := &models.Tenant{
		ID:        uuid.New(),
		Name:      req.Name,
		Subdomain: req.Subdomain,
		License:   req.License,
		Status:    "active",
	}

	if err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, err
	}

	rollbackOnError := func(origErr error) error {
		if rbErr := s.tenantRepo.Delete(ctx, tenant.ID); rbErr != nil {
			return fmt.Errorf("%w; rollback failed: %v", origErr, rbErr)
		}
		return origErr
	}

	// Apply comprehensive role templates from config
	// This ensures all standard roles (admin, manager, user, etc.) are available immediately
	roleTemplates := []string{"admin", "manager", "user", "viewer", "inventory_manager", "sales", "analyst"}
	for _, template := range roleTemplates {
		if err := s.roleTemplateApplier.ApplyTemplateToTenant(ctx, tenant.ID, template); err != nil {
			// Log warning but don't fail the whole creation - getting partial roles is better than none
			fmt.Printf("Warning: failed to apply role template %s for tenant %s: %v\n", template, tenant.ID, err)
		}
	}

	// Verify admin role exists (it should have been created by the template)
	adminRole, err := s.roleRepo.GetByName(ctx, tenant.ID, "admin")
	if err != nil {
		return nil, rollbackOnError(fmt.Errorf("tenant created but failed to find admin role: %w", err))
	}

	// If admin email is provided, create an invitation for the admin
	if req.AdminEmail != "" {
		// Find admin role (should exist now after seeding)
		inviteReq := &CreateInvitationRequest{
			TenantID: tenant.ID,
			Email:    req.AdminEmail,
			RoleID:   adminRole.ID,
		}

		// Get userID from context if available (e.g. super admin creating tenant)
		var invitedBy uuid.UUID
		if uid, ok := ctx.Value(common.UserIDKey).(uuid.UUID); ok {
			invitedBy = uid
		}

		if _, err := s.invitationService.CreateInvitation(ctx, inviteReq, invitedBy); err != nil {
			return nil, rollbackOnError(fmt.Errorf("tenant created but failed to invite admin: %w", err))
		}
	}

	return tenant, nil
}

func (s *tenantService) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return s.tenantRepo.GetByID(ctx, id)
}

func (s *tenantService) GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	if subdomain == "" {
		return nil, errors.New("subdomain is required")
	}
	return s.tenantRepo.GetBySubdomain(ctx, subdomain)
}

func (s *tenantService) Update(ctx context.Context, req *UpdateTenantRequest) error {
	if req.ID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(req.Subdomain) == "" {
		return errors.New("subdomain is required")
	}
	if strings.Contains(req.Subdomain, " ") {
		return errors.New("subdomain cannot contain spaces")
	}
	if len(req.Subdomain) < 3 {
		return errors.New("subdomain must be at least 3 characters long")
	}
	if strings.TrimSpace(req.Status) == "" {
		return errors.New("status is required")
	}

	// Get existing tenant
	existing, err := s.tenantRepo.GetByID(ctx, req.ID)
	if err != nil {
		return err
	}

	existing.Name = req.Name
	existing.Subdomain = req.Subdomain
	existing.License = req.License
	existing.Status = req.Status

	return s.tenantRepo.Update(ctx, existing)
}

func (s *tenantService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.tenantRepo.Delete(ctx, id)
}

func (s *tenantService) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return s.tenantRepo.List(ctx, limit, offset)
}

func (s *tenantService) GetTenantSettings(ctx context.Context, tenantID uuid.UUID) (*models.Tenant, error) {
	if tenantID == uuid.Nil {
		return nil, errors.New("tenant ID is required")
	}
	return s.tenantRepo.FindSettingsByTenantID(ctx, tenantID)
}

func (s *tenantService) UpdateTenantSettings(ctx context.Context, req *UpdateTenantSettingsRequest) error {
	if req.TenantID == uuid.Nil {
		return errors.New("tenant ID is required")
	}
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Subdomain == "" {
		return errors.New("subdomain is required")
	}
	if strings.Contains(req.Subdomain, " ") {
		return errors.New("subdomain cannot contain spaces")
	}
	if len(req.Subdomain) < 3 {
		return errors.New("subdomain must be at least 3 characters long")
	}

	existing, err := s.tenantRepo.FindSettingsByTenantID(ctx, req.TenantID)
	if err != nil {
		return err
	}

	existing.Name = req.Name
	existing.Subdomain = req.Subdomain
	existing.License = req.License

	return s.tenantRepo.UpdateSettings(ctx, existing)
}

// seedAdminPermissions assigns every permission to the tenant's admin role.
func (s *tenantService) seedAdminPermissions(ctx context.Context, tenantID, adminRoleID uuid.UUID) error {
	if s.permissionLister == nil || s.rolePermissionAssigner == nil {
		// No-op if dependencies are missing (e.g., in limited test setups)
		return nil
	}

	perms, err := s.permissionLister.ListPermissions(ctx)
	if err != nil {
		return err
	}

	var wildcardPermissionID uuid.UUID

	for _, perm := range perms {
		if perm.Name == "*" {
			wildcardPermissionID = perm.ID
		}

		rolePerm := &models.RolePermission{
			RoleID:       adminRoleID,
			PermissionID: perm.ID,
		}

		if err := s.rolePermissionAssigner.Create(ctx, tenantID, rolePerm); err != nil {
			// Ignore duplicates; surface any other error
			if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "23505") {
				return err
			}
		}
	}

	// Ensure wildcard permission is explicitly present for admins so future permissions are covered
	if wildcardPermissionID != uuid.Nil {
		rolePerm := &models.RolePermission{
			RoleID:       adminRoleID,
			PermissionID: wildcardPermissionID,
		}
		if err := s.rolePermissionAssigner.Create(ctx, tenantID, rolePerm); err != nil {
			if !strings.Contains(err.Error(), "duplicate") && !strings.Contains(err.Error(), "23505") {
				return err
			}
		}
	} else {
		fmt.Printf("Warning: wildcard permission '*' not found when seeding admin role for tenant %s\n", tenantID.String())
	}

	return nil
}
