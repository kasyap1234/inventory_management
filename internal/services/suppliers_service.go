package services

import (
	"context"
	"errors"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type SupplierService interface {
	Create(ctx context.Context, tenantID uuid.UUID, supplier *models.Supplier) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Supplier, error)
	Update(ctx context.Context, tenantID uuid.UUID, supplier *models.Supplier) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Supplier, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Supplier, error)
}

type supplierService struct {
	supplierRepo repositories.SupplierRepository
}

func NewSupplierService(supplierRepo repositories.SupplierRepository) SupplierService {
	return &supplierService{
		supplierRepo: supplierRepo,
	}
}

func (s *supplierService) Create(ctx context.Context, tenantID uuid.UUID, supplier *models.Supplier) error {
	if supplier.Name == "" {
		return errors.New("supplier name is required")
	}

	// Check for duplicate name
	existing, err := s.supplierRepo.GetByName(ctx, tenantID, supplier.Name)
	if err == nil && existing != nil {
		return errors.New("supplier with this name already exists")
	}

	supplier.TenantID = tenantID
	supplier.ID = uuid.New()

	return s.supplierRepo.Create(ctx, supplier)
}

func (s *supplierService) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Supplier, error) {
	return s.supplierRepo.GetByID(ctx, tenantID, id)
}

func (s *supplierService) Update(ctx context.Context, tenantID uuid.UUID, supplier *models.Supplier) error {
	if supplier.Name == "" {
		return errors.New("supplier name is required")
	}

	supplier.TenantID = tenantID
	return s.supplierRepo.Update(ctx, supplier)
}

func (s *supplierService) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	return s.supplierRepo.Delete(ctx, tenantID, id)
}

func (s *supplierService) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Supplier, error) {
	return s.supplierRepo.List(ctx, tenantID, limit, offset)
}

func (s *supplierService) GetByName(ctx context.Context, tenantID uuid.UUID, name string) (*models.Supplier, error) {
	return s.supplierRepo.GetByName(ctx, tenantID, name)
}