package services

import (
	"context"
	"errors"
	"strings"

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

type tenantService struct {
	tenantRepo repositories.TenantRepository
}

func NewTenantService(tenantRepo repositories.TenantRepository) TenantService {
	return &tenantService{tenantRepo: tenantRepo}
}

type CreateTenantRequest struct {
	Name      string `json:"name" validate:"required"`
	Subdomain string `json:"subdomain" validate:"required"`
	License   string `json:"license"`
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