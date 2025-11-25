package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

// BatchService handles business logic for batches
type BatchService struct {
	batchRepo   *repositories.BatchRepository
	productRepo repositories.ProductRepository
}

// NewBatchService creates a new BatchService
func NewBatchService(batchRepo *repositories.BatchRepository, productRepo repositories.ProductRepository) *BatchService {
	return &BatchService{
		batchRepo:   batchRepo,
		productRepo: productRepo,
	}
}

// CreateBatch creates a new batch and updates the product total quantity
func (s *BatchService) CreateBatch(ctx context.Context, tenantID uuid.UUID, batch *models.Batch) error {
	// Validate batch
	if batch.ProductID == uuid.Nil {
		return errors.New("product ID is required")
	}
	if batch.BatchNumber == "" {
		return errors.New("batch number is required")
	}
	if batch.Quantity < 0 {
		return errors.New("quantity cannot be negative")
	}

	// Set defaults
	if batch.Status == "" {
		batch.Status = "active"
	}
	now := time.Now()
	batch.CreatedAt = now
	batch.UpdatedAt = now
	batch.TenantID = tenantID

	// Create batch
	if err := s.batchRepo.Create(ctx, batch); err != nil {
		return err
	}

	// Update product total quantity
	// Note: In a real production system, this should be in a transaction
	// For now, we do best effort update
	return s.updateProductTotalQuantity(ctx, tenantID, batch.ProductID)
}

// UpdateBatch updates a batch and the product total quantity
func (s *BatchService) UpdateBatch(ctx context.Context, tenantID uuid.UUID, batch *models.Batch) error {
	batch.TenantID = tenantID
	if err := s.batchRepo.Update(ctx, batch); err != nil {
		return err
	}
	return s.updateProductTotalQuantity(ctx, tenantID, batch.ProductID)
}

// GetBatch retrieves a batch by ID
func (s *BatchService) GetBatch(ctx context.Context, tenantID, id uuid.UUID) (*models.Batch, error) {
	return s.batchRepo.GetByID(ctx, tenantID, id)
}

// GetBatchesByProduct retrieves all batches for a product
func (s *BatchService) GetBatchesByProduct(ctx context.Context, tenantID, productID uuid.UUID) ([]models.Batch, error) {
	return s.batchRepo.GetByProductID(ctx, tenantID, productID)
}

// updateProductTotalQuantity recalculates and updates the total quantity for a product
func (s *BatchService) updateProductTotalQuantity(ctx context.Context, tenantID, productID uuid.UUID) error {
	total, err := s.batchRepo.GetTotalQuantityByProductID(ctx, tenantID, productID)
	if err != nil {
		return fmt.Errorf("failed to calculate total quantity: %w", err)
	}

	// We need a method in ProductRepository to update just the quantity
	// For now, let's assume we can use a direct update or we need to add a method to ProductRepo
	return s.productRepo.UpdateQuantity(ctx, productID, total)
}
