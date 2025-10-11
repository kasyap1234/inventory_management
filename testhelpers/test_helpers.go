package testhelpers

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

// TestDB represents a test database connection
type TestDB struct {
	Pool *pgxpool.Pool
	ctx  context.Context
}

// NewTestDB creates a new test database connection
// This should be called once per test suite
func NewTestDB(t *testing.T, databaseURL string) *TestDB {
	ctx := context.Background()
	
	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err, "Failed to connect to test database")
	
	// Verify connection
	err = pool.Ping(ctx)
	require.NoError(t, err, "Failed to ping test database")
	
	return &TestDB{
		Pool: pool,
		ctx:  ctx,
	}
}

// Close closes the test database connection
func (db *TestDB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// TruncateTables truncates all tables for clean test state
func (db *TestDB) TruncateTables(t *testing.T, tables []string) {
	for _, table := range tables {
		_, err := db.Pool.Exec(db.ctx, "TRUNCATE TABLE "+table+" CASCADE")
		require.NoError(t, err, "Failed to truncate table: "+table)
	}
}

// CreateTestTenant creates a test tenant and returns its ID
func (db *TestDB) CreateTestTenant(t *testing.T) uuid.UUID {
	tenantID := uuid.New()
	
	query := `
		INSERT INTO tenants (id, name, subdomain, status)
		VALUES ($1, $2, $3, $4)
	`
	
	_, err := db.Pool.Exec(db.ctx, query, tenantID, "Test Tenant", "test-"+tenantID.String()[:8], "active")
	require.NoError(t, err, "Failed to create test tenant")
	
	return tenantID
}

// CreateTestUser creates a test user and returns its ID
func (db *TestDB) CreateTestUser(t *testing.T, tenantID uuid.UUID, email string, passwordHash string) uuid.UUID {
	userID := uuid.New()
	
	query := `
		INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	_, err := db.Pool.Exec(db.ctx, query, userID, tenantID, email, passwordHash, "Test", "User", "active")
	require.NoError(t, err, "Failed to create test user")
	
	return userID
}

// CreateTestProduct creates a test product and returns its ID
func (db *TestDB) CreateTestProduct(t *testing.T, tenantID uuid.UUID, name string, unitPrice float64, quantity int) uuid.UUID {
	productID := uuid.New()
	
	query := `
		INSERT INTO products (id, tenant_id, name, unit_price, quantity)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := db.Pool.Exec(db.ctx, query, productID, tenantID, name, unitPrice, quantity)
	require.NoError(t, err, "Failed to create test product")
	
	return productID
}

// CreateTestWarehouse creates a test warehouse and returns its ID
func (db *TestDB) CreateTestWarehouse(t *testing.T, tenantID uuid.UUID, name string) uuid.UUID {
	warehouseID := uuid.New()
	
	query := `
		INSERT INTO warehouses (id, tenant_id, name, address, status)
		VALUES ($1, $2, $3, $4, $5)
	`
	
	_, err := db.Pool.Exec(db.ctx, query, warehouseID, tenantID, name, "Test Address", "active")
	require.NoError(t, err, "Failed to create test warehouse")
	
	return warehouseID
}

// CreateTestSupplier creates a test supplier and returns its ID
func (db *TestDB) CreateTestSupplier(t *testing.T, tenantID uuid.UUID, name string) uuid.UUID {
	supplierID := uuid.New()
	
	query := `
		INSERT INTO suppliers (id, tenant_id, name, email, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err := db.Pool.Exec(db.ctx, query, supplierID, tenantID, name, "test@example.com", "1234567890", "active")
	require.NoError(t, err, "Failed to create test supplier")
	
	return supplierID
}

// CreateTestDistributor creates a test distributor and returns its ID
func (db *TestDB) CreateTestDistributor(t *testing.T, tenantID uuid.UUID, name string) uuid.UUID {
	distributorID := uuid.New()
	
	query := `
		INSERT INTO distributors (id, tenant_id, name, email, phone, status)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	
	_, err := db.Pool.Exec(db.ctx, query, distributorID, tenantID, name, "test@example.com", "1234567890", "active")
	require.NoError(t, err, "Failed to create test distributor")
	
	return distributorID
}

// WaitForCondition waits for a condition to be true or times out
func WaitForCondition(t *testing.T, timeout time.Duration, condition func() bool, message string) {
	deadline := time.Now().Add(timeout)
	
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	t.Fatalf("Timeout waiting for condition: %s", message)
}

// AssertNoError is a helper to assert no error occurred
func AssertNoError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %v", message, err)
	}
}

// AssertError is a helper to assert an error occurred
func AssertError(t *testing.T, err error, message string) {
	if err == nil {
		t.Fatalf("%s: expected error but got nil", message)
	}
}

// AssertEqual is a helper to assert equality
func AssertEqual(t *testing.T, expected, actual interface{}, message string) {
	require.Equal(t, expected, actual, message)
}

// MockContext creates a context with test values
func MockContext() context.Context {
	return context.Background()
}

// MockTenantContext creates a context with tenant ID
func MockTenantContext(tenantID uuid.UUID) context.Context {
	// Note: In real implementation, this should use the actual context key
	// from your common package
	return context.WithValue(context.Background(), "tenant_id", tenantID)
}
