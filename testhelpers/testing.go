package testhelpers

import (
	"context"
	"os"
	"testing"
	"time"

	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB holds the database connection for testing
type TestDB struct {
	Pool    *pgxpool.Pool
	Cleanup func() error
}

// SetupTestDB creates a pooled connection for testing
func SetupTestDB(t *testing.T, connString string) *TestDB {
	t.Helper()

	if connString == "" {
		connString = os.Getenv("TEST_DATABASE_URL")
		if connString == "" {
			connString = "host=localhost port=5432 user=postgres password=postgres dbname=agromart2_test sslmode=disable"
		}
	}

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return &TestDB{
		Pool: pool,
		Cleanup: func() error {
			pool.Close()
			return nil
		},
	}
}

// SetupTestTenant creates a test tenant for testing
func SetupTestTenant(t *testing.T, db *TestDB) uuid.UUID {
	t.Helper()

	tenantID := uuid.New()
	query := `
		INSERT INTO tenants (id, name, subdomain, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (subdomain) DO NOTHING
	`
	_, err := db.Pool.Exec(context.Background(), query, tenantID, "Test Tenant", "test-tenant", "active", time.Now())
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	return tenantID
}

// SetupTestCategory creates a test category for testing
func SetupTestCategory(t *testing.T, db *TestDB, tenantID uuid.UUID) uuid.UUID {
	t.Helper()

	categoryID := uuid.New()
	query := `
		INSERT INTO categories (id, tenant_id, name, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Pool.Exec(context.Background(), query, categoryID, tenantID, "Test Category", "Test description", time.Now())
	if err != nil {
		t.Fatalf("Failed to create test category: %v", err)
	}

	return categoryID
}

// SetupTestProduct creates a test product for testing
func SetupTestProduct(t *testing.T, db *TestDB, tenantID, categoryID uuid.UUID) *models.Product {
	t.Helper()

	productID := uuid.New()
	barcode := "123456789"
	unitMeasure := "kg"
	description := "Test product description"
	batchNum := "BATCH001"
	expiry := time.Now().Add(365 * 24 * time.Hour)

	product := &models.Product{
		ID:             productID,
		TenantID:       tenantID,
		CategoryID:     &categoryID,
		Name:           "Test Product",
		BatchNumber:    &batchNum,
		ExpiryDate:     &expiry,
		Quantity:       100,
		UnitPrice:      10.99,
		Barcode:        &barcode,
		UnitOfMeasure:  &unitMeasure,
		Description:    &description,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	query := `
		INSERT INTO products (id, tenant_id, category_id, name, batch_number, expiry_date, quantity, unit_price, barcode, unit_of_measure, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := db.Pool.Exec(context.Background(), query,
		product.ID, product.TenantID, product.CategoryID, product.Name, product.BatchNumber,
		product.ExpiryDate, product.Quantity, product.UnitPrice, product.Barcode,
		product.UnitOfMeasure, product.Description, product.CreatedAt, product.UpdatedAt)
	if err != nil {
		t.Fatalf("Failed to create test product: %v", err)
	}

	return product
}