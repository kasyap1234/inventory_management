package services

import (
	"context"
	"errors"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type DistributorService interface {
	Create(ctx context.Context, tenantID uuid.UUID, distributor *models.Distributor) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Distributor, error)
	Update(ctx context.Context, tenantID uuid.UUID, distributor *models.Distributor) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Distributor, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Distributor, error)
}

type distributorService struct {
	distributorRepo repositories.DistributorRepository
}

func NewDistributorService(distributorRepo repositories.DistributorRepository) DistributorService {
	return &distributorService{
		distributorRepo: distributorRepo,
	}
}

func (s *distributorService) Create(ctx context.Context, tenantID uuid.UUID, distributor *models.Distributor) error {
	if distributor.Name == "" {
		return errors.New("distributor name is required")
	}

	// Check for duplicate name
	existing, err := s.distributorRepo.GetByName(ctx, tenantID, distributor.Name)
	if err == nil && existing != nil {
		return errors.New("distributor with this name already exists")
	}

	distributor.TenantID = tenantID
	distributor.ID = uuid.New()

	return s.distributorRepo.Create(ctx, distributor)
}

func (s *distributorService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Distributor, error) {
	return s.distributorRepo.GetByID(ctx, tenantID, id)
}

func (s *distributorService) Update(ctx context.Context, tenantID uuid.UUID, distributor *models.Distributor) error {
	if distributor.Name == "" {
		return errors.New("distributor name is required")
	}

	distributor.TenantID = tenantID
	return s.distributorRepo.Update(ctx, distributor)
}

func (s *distributorService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.distributorRepo.Delete(ctx, tenantID, id)
}

func (s *distributorService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Distributor, error) {
	return s.distributorRepo.List(ctx, tenantID, limit, offset)
}

func (s *distributorService) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Distributor, error) {
	return s.distributorRepo.GetByName(ctx, tenantID, name)
}