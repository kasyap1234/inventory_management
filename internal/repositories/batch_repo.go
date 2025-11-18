package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// BatchRepository handles database operations for batches
type BatchRepository struct {
	db *pgxpool.Pool
}

// NewBatchRepository creates a new BatchRepository
func NewBatchRepository(db *pgxpool.Pool) *BatchRepository {
	return &BatchRepository{db: db}
}

// Create creates a new batch
func (r *BatchRepository) Create(ctx context.Context, batch *models.Batch) error {
	query := `
		INSERT INTO batches (
			product_id, batch_number, quantity, expiry_date, 
			manufacturing_date, location, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING id`

	err := r.db.QueryRow(ctx, query,
		batch.ProductID, batch.BatchNumber, batch.Quantity, batch.ExpiryDate,
		batch.ManufacturingDate, batch.Location, batch.Status, batch.CreatedAt, batch.UpdatedAt,
	).Scan(&batch.ID)

	return err
}

// GetByID retrieves a batch by ID
func (r *BatchRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Batch, error) {
	var batch models.Batch
	query := `SELECT id, product_id, batch_number, quantity, expiry_date, manufacturing_date, location, status, created_at, updated_at FROM batches WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(
		&batch.ID, &batch.ProductID, &batch.BatchNumber, &batch.Quantity, &batch.ExpiryDate,
		&batch.ManufacturingDate, &batch.Location, &batch.Status, &batch.CreatedAt, &batch.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// GetByProductID retrieves all batches for a product
func (r *BatchRepository) GetByProductID(ctx context.Context, productID uuid.UUID) ([]models.Batch, error) {
	var batches []models.Batch
	query := `SELECT id, product_id, batch_number, quantity, expiry_date, manufacturing_date, location, status, created_at, updated_at FROM batches WHERE product_id = $1 ORDER BY expiry_date ASC`
	rows, err := r.db.Query(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var batch models.Batch
		if err := rows.Scan(&batch.ID, &batch.ProductID, &batch.BatchNumber, &batch.Quantity, &batch.ExpiryDate,
			&batch.ManufacturingDate, &batch.Location, &batch.Status, &batch.CreatedAt, &batch.UpdatedAt); err != nil {
			return nil, err
		}
		batches = append(batches, batch)
	}
	return batches, nil
}

// Update updates an existing batch
func (r *BatchRepository) Update(ctx context.Context, batch *models.Batch) error {
	batch.UpdatedAt = time.Now()
	query := `
		UPDATE batches SET
			quantity = $1,
			expiry_date = $2,
			manufacturing_date = $3,
			location = $4,
			status = $5,
			updated_at = $6
		WHERE id = $7`

	result, err := r.db.Exec(ctx, query,
		batch.Quantity, batch.ExpiryDate, batch.ManufacturingDate,
		batch.Location, batch.Status, batch.UpdatedAt, batch.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update batch: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("batch not found")
	}

	return nil
}

// Delete deletes a batch
func (r *BatchRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM batches WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete batch: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("batch not found")
	}

	return nil
}

// GetTotalQuantityByProductID gets the sum of quantities of all active batches for a product
func (r *BatchRepository) GetTotalQuantityByProductID(ctx context.Context, productID uuid.UUID) (int, error) {
	var total int
	query := `
		SELECT COALESCE(SUM(quantity), 0) 
		FROM batches 
		WHERE product_id = $1 AND status = 'active'`

	if err := r.db.QueryRow(ctx, query, productID).Scan(&total); err != nil {
		return 0, fmt.Errorf("failed to get total quantity: %w", err)
	}
	return total, nil
}
