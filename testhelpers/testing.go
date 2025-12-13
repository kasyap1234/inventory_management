package testhelpers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/testhelpers/containers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestDB holds the database connection for testing
type TestDB struct {
	Pool      *pgxpool.Pool
	Cleanup   func() error
	container *containers.PostgresContainer
}

// SetupTestDB creates a database connection for testing.
// If TEST_DATABASE_URL env var is set, it uses that connection.
// Otherwise, it spins up a PostgreSQL testcontainer with migrations.
func SetupTestDB(t *testing.T, connString string) *TestDB {
	t.Helper()

	// Check if we should use a testcontainer or existing database
	if connString == "" {
		connString = os.Getenv("TEST_DATABASE_URL")
	}

	// If no connection string, use testcontainer
	if connString == "" {
		return SetupTestDBWithContainer(t)
	}

	// Use provided connection string
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

// SetupTestDBWithContainer creates a PostgreSQL testcontainer for testing.
// The container is automatically terminated when the test completes.
func SetupTestDBWithContainer(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	// Get migrations path - find the project root
	migrationsPath := findMigrationsPath()
	if migrationsPath == "" {
		t.Logf("Warning: migrations path not found, container will have empty database")
	}

	config := containers.PostgresContainerConfig{
		ImageTag:       "16-alpine",
		Database:       "agromart2_test",
		User:           "postgres",
		Password:       "postgres",
		MigrationsPath: migrationsPath,
	}

	container, err := containers.NewPostgresContainer(ctx, config)
	if err != nil {
		t.Fatalf("Failed to create test tenant: %v", err)
	}

	// Register cleanup
	t.Cleanup(func() {
		if err := container.Close(ctx); err != nil {
			t.Logf("Warning: failed to close container: %v", err)
		}
	})

	return &TestDB{
		Pool:      container.Pool,
		container: container,
		Cleanup: func() error {
			return container.Close(ctx)
		},
	}
}

// findMigrationsPath attempts to find the migrations directory relative to the test file
func findMigrationsPath() string {
	// Get the directory of this source file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}

	// Navigate up to find the project root and migrations folder
	dir := filepath.Dir(filename)
	for i := 0; i < 5; i++ { // Walk up max 5 levels
		migrationsPath := filepath.Join(dir, "migrations")
		if info, err := os.Stat(migrationsPath); err == nil && info.IsDir() {
			return migrationsPath
		}
		dir = filepath.Dir(dir)
	}

	return ""
}

// SetupTestTenant creates a test tenant for testing
func SetupTestTenant(t *testing.T, db *TestDB) uuid.UUID {
	t.Helper()

	tenantID := uuid.New()
	subdomain := "test-" + uuid.New().String()[:8] // Unique subdomain per test
	query := `
		INSERT INTO tenants (id, name, subdomain, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := db.Pool.Exec(context.Background(), query, tenantID, "Test Tenant", subdomain, "active", time.Now())
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
		ID:            productID,
		TenantID:      tenantID,
		CategoryID:    &categoryID,
		Name:          "Test Product",
		BatchNumber:   &batchNum,
		ExpiryDate:    &expiry,
		Quantity:      100,
		UnitPrice:     10.99,
		Barcode:       &barcode,
		UnitOfMeasure: &unitMeasure,
		Description:   &description,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
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