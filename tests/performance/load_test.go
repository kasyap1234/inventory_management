package performance

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// setupPerformanceTestDB creates a test database for performance tests
func setupPerformanceTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL environment variable not set, skipping performance tests")
	}

	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("Unable to connect to test database: %v", err)
	}

	// Ping the database to ensure a good connection
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return pool
}

// seedPerformanceData creates test data for performance tests
func seedPerformanceData(t *testing.T, pool *pgxpool.Pool, tenantID uuid.UUID, numProducts int) {
	t.Helper()

	// Create category first
	categoryRepo := repositories.NewCategoryRepo(pool)
	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "Performance Test Category",
		Description: "Category for performance testing",
	}
	err := categoryRepo.Create(context.Background(), category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// Create products
	productRepo := repositories.NewProductRepo(pool)
	for i := 0; i < numProducts; i++ {
		batchNum := fmt.Sprintf("BATCH_PERF_%d", i)
		barcode := fmt.Sprintf("PERF_BARCODE_%d", i)
		uom := "pcs"
		desc := fmt.Sprintf("Description for performance product %d", i)

		product := &models.Product{
			ID:             uuid.New(),
			TenantID:       tenantID,
			CategoryID:     &category.ID,
			Name:           fmt.Sprintf("Performance Product %d", i),
			BatchNumber:    &batchNum,
			Quantity:       rand.Intn(1000) + 1, // Random quantity 1-1000
			UnitPrice:      rand.Float64()*1000 + 10, // Random price 10-1010
			Barcode:        &barcode,
			UnitOfMeasure:  &uom,
			Description:    &desc,
		}

		err := productRepo.Create(context.Background(), product)
		if err != nil {
			t.Fatalf("Failed to create product %d: %v", i, err)
		}
	}
}

// BenchmarkProductRepository benchmarks product repository operations
func BenchmarkProductRepository(b *testing.B) {
	pool := setupPerformanceTestDB(&testing.T{})
	defer pool.Close()

	tenantID := uuid.New()

	// Setup performance data
	seedPerformanceData(&testing.T{}, pool, tenantID, 100)

	repo := repositories.NewProductRepo(pool)
	ctx := context.Background()

	b.ResetTimer()

	b.Run("GetByID", func(b *testing.B) {
		productID := uuid.New()
		// Create a test product for benchmarking
		testProduct := &models.Product{
			ID:        productID,
			TenantID:  tenantID,
			Name:      "Benchmark Product",
			Quantity:  100,
			UnitPrice: 29.99,
		}
		err := repo.Create(ctx, testProduct)
		if err != nil {
			b.Fatalf("Failed to create test product: %v", err)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := repo.GetByID(ctx, tenantID, productID)
				if err != nil {
					b.Errorf("GetByID failed: %v", err)
				}
			}
		})
	})

	b.Run("List", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := repo.List(ctx, tenantID, 50, 0)
				if err != nil {
					b.Errorf("List failed: %v", err)
				}
			}
		})
	})

	b.Run("Search", func(b *testing.B) {
		testCases := []struct {
			name  string
			query string
		}{
			{"ExactMatch", "Performance Product 50"},
			{"PartialMatch", "Performance"},
			{"NoMatch", "NonExistentProduct12345"},
		}

		for _, tc := range testCases {
			b.Run(tc.name, func(b *testing.B) {
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						_, err := repo.Search(ctx, tenantID, tc.query, nil, 50, 0)
						if err != nil {
							b.Errorf("Search failed: %v", err)
						}
					}
				})
			})
		}
	})
}

// BenchmarkProductService benchmarks product service operations
func BenchmarkProductService(b *testing.B) {
	pool := setupPerformanceTestDB(&testing.T{})
	defer pool.Close()

	tenantID := uuid.New()

	// Setup performance data
	seedPerformanceData(&testing.T{}, pool, tenantID, 50)

	productRepo := repositories.NewProductRepo(pool)
	inventoryRepo := repositories.NewInventoryRepo(pool)
	categoryRepo := repositories.NewCategoryRepo(pool)
	productImageRepo := repositories.NewProductImageRepo(pool)

	service := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, nil, nil)
	ctx := context.Background()

	b.ResetTimer()

	b.Run("Create", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				product := &models.Product{
					TenantID:  tenantID,
					Name:      fmt.Sprintf("Bench Create Product %d", rand.Int()),
					Quantity:  10,
					UnitPrice: 25.99,
				}

				err := service.Create(ctx, tenantID, product)
				if err != nil {
					b.Errorf("Create failed: %v", err)
				}
			}
		})
	})

	b.Run("GetByID", func(b *testing.B) {
		// Get existing product ID for benchmarking
		products, err := service.List(ctx, tenantID, 1, 0)
		if err != nil || len(products) == 0 {
			b.Skip("No products available for GetByID benchmark")
		}
		productID := products[0].ID

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := service.GetByID(ctx, tenantID, productID)
				if err != nil {
					b.Errorf("GetByID failed: %v", err)
				}
			}
		})
	})

	b.Run("List", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, err := service.List(ctx, tenantID, 10, 0)
				if err != nil {
					b.Errorf("List failed: %v", err)
				}
			}
		})
	})
}

// LoadTest simulates concurrent load on the product service
func TestLoadProductService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping load test in short mode")
	}

	pool := setupPerformanceTestDB(t)
	defer pool.Close()

	tenantID := uuid.New()

	// Setup initial data
	seedPerformanceData(t, pool, tenantID, 20)

	productRepo := repositories.NewProductRepo(pool)
	inventoryRepo := repositories.NewInventoryRepo(pool)
	categoryRepo := repositories.NewCategoryRepo(pool)
	productImageRepo := repositories.NewProductImageRepo(pool)

	service := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, nil, nil)
	ctx := context.Background()

	// Load test configuration
	const (
		numGoroutines   = 50  // Concurrent users
		operationsPerGo = 20  // Operations per goroutine
	)

	start := time.Now()

	// Channel to track completion
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < operationsPerGo; j++ {
				// Mix of operations: 70% reads, 30% writes
				if rand.Float32() < 0.7 {
					// Read operation
					_, err := service.List(ctx, tenantID, 10, 0)
					if err != nil {
						t.Errorf("Goroutine %d: List failed: %v", goroutineID, err)
					}
				} else {
					// Write operation
					product := &models.Product{
						TenantID:  tenantID,
						Name:      fmt.Sprintf("Load Test Product G%d O%d", goroutineID, j),
						Quantity:  rand.Intn(100) + 1,
						UnitPrice: rand.Float64()*100 + 10,
					}

					err := service.Create(ctx, tenantID, product)
					if err != nil {
						t.Errorf("Goroutine %d: Create failed: %v", goroutineID, err)
					}
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	elapsed := time.Since(start)

	// Calculate metrics
	totalOperations := numGoroutines * operationsPerGo
	opsPerSecond := float64(totalOperations) / elapsed.Seconds()

	t.Logf("Load Test Results:")
	t.Logf("Total operations: %d", totalOperations)
	t.Logf("Elapsed time: %v", elapsed)
	t.Logf("Operations per second: %.2f", opsPerSecond)
	t.Logf("Average response time: %.2f ms", float64(elapsed.Milliseconds())/float64(totalOperations))

	// Performance assertions
	if opsPerSecond < 50 {
		t.Errorf("Throughput too low: %.2f ops/sec (minimum: 50)", opsPerSecond)
	}

	if elapsed > 30*time.Second {
		t.Errorf("Test took too long: %v (maximum: 30s)", elapsed)
	}
}

// StressTest simulates extreme load conditions
func TestStressProductService(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	pool := setupPerformanceTestDB(t)
	defer pool.Close()

	tenantID := uuid.New()

	// Create minimal initial data
	categoryRepo := repositories.NewCategoryRepo(pool)
	category := &models.Category{
		ID:          uuid.New(),
		TenantID:    tenantID,
		Name:        "Stress Test Category",
		Description: "Category for stress testing",
	}
	err := categoryRepo.Create(context.Background(), category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	productRepo := repositories.NewProductRepo(pool)
	inventoryRepo := repositories.NewInventoryRepo(pool)
	productImageRepo := repositories.NewProductImageRepo(pool)

	service := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, nil, nil)
	ctx := context.Background()

	// Stress test: rapid fire operations
	const numOperations = 200

	start := time.Now()

	// Create many products quickly
	for i := 0; i < numOperations; i++ {
		product := &models.Product{
			TenantID:   tenantID,
			Name:       fmt.Sprintf("Stress Product %d", i),
			Quantity:   100,
			UnitPrice:  29.99,
			CategoryID: &category.ID,
		}

		err := service.Create(ctx, tenantID, product)
		if err != nil {
			t.Errorf("Failed to create product %d: %v", i, err)
		}
	}

	elapsed := time.Since(start)
	opsPerSecond := float64(numOperations) / elapsed.Seconds()

	t.Logf("Stress Test Results:")
	t.Logf("Created %d products in %v", numOperations, elapsed)
	t.Logf("Operations per second: %.2f", opsPerSecond)

	// Verify data integrity
	products, err := service.List(ctx, tenantID, 1000, 0)
	if err != nil {
		t.Fatalf("Failed to verify created products: %v", err)
	}

	if len(products) != numOperations {
		t.Errorf("Data integrity check failed: expected %d products, got %d", numOperations, len(products))
	}

	// Performance thresholds for stress test
	if opsPerSecond < 20 {
		t.Errorf("Stress test throughput too low: %.2f ops/sec (minimum: 20)", opsPerSecond)
	}
}
