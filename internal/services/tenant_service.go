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

func (s *tenantService) Create(ctx context.Context, req *CreateTenantRequest) (*models.Tenant, error) {
	if req.Name == "" || req.Subdomain == "" {
		return nil, errors.New("name and subdomain are required")
	}
	// Basic validation
	if strings.TrimSpace(req.Subdomain) != req.Subdomain {
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