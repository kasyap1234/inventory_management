package background

import (
	"context"
	"log"
	"sync"
	"time"

	"agromart2/internal/analytics"
	"agromart2/internal/caching"
	"agromart2/internal/repositories"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

// JobScheduler manages background jobs for distributed environment
type JobScheduler struct {
	scheduler   gocron.Scheduler
	analyticsSvc *analytics.AnalyticsService
	cacheSvc    caching.CacheService
	inventoryRepo repositories.InventoryRepository
	orderRepo   repositories.OrderRepository
	tenantRepo  repositories.TenantRepository
	jobJobs     map[string]gocron.Job
	mu          sync.RWMutex
}

// NewJobScheduler creates a new job scheduler
func NewJobScheduler(analyticsSvc *analytics.AnalyticsService, cacheSvc caching.CacheService,
	inventoryRepo repositories.InventoryRepository, orderRepo repositories.OrderRepository,
	tenantRepo repositories.TenantRepository) *JobScheduler {

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("Failed to create scheduler: %v", err)
	}

	js := &JobScheduler{
		scheduler:     scheduler,
		analyticsSvc: analyticsSvc,
		cacheSvc:      cacheSvc,
		inventoryRepo: inventoryRepo,
		orderRepo:     orderRepo,
		tenantRepo:    tenantRepo,
		jobJobs:       make(map[string]gocron.Job),
	}

	js.registerJobs()

	return js
}

// Start starts the job scheduler
func (js *JobScheduler) Start() error {
	log.Printf("Starting background job scheduler")
	js.scheduler.Start()
	return nil
}

// Stop stops the job scheduler
func (js *JobScheduler) Stop() error {
	log.Printf("Stopping background job scheduler")
	return js.scheduler.Shutdown()
}

// registerJobs registers all background jobs
func (js *JobScheduler) registerJobs() {
	// Analytics refresh job - every 5 minutes
	analyticsJob, err := js.scheduler.NewJob(
		gocron.DurationJob(5*time.Minute),
		gocron.NewTask(js.refreshTenantAnalytics, context.Background()),
		gocron.WithName("tenant-analytics-refresh"),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	if err != nil {
		log.Printf("Failed to create analytics job: %v", err)
	} else {
		js.jobJobs["analytics"] = analyticsJob
	}

	// Cache cleanup job - every hour
	cacheJob, err := js.scheduler.NewJob(
		gocron.DurationJob(1*time.Hour),
		gocron.NewTask(js.cleanupExpiredCache),
		gocron.WithName("cache-cleanup"),
	)
	if err != nil {
		log.Printf("Failed to create cache cleanup job: %v", err)
	} else {
		js.jobJobs["cache-cleanup"] = cacheJob
	}

	// Inventory alerts job - every 30 minutes
	alertsJob, err := js.scheduler.NewJob(
		gocron.DurationJob(30*time.Minute),
		gocron.NewTask(js.processInventoryAlerts),
		gocron.WithName("inventory-alerts"),
	)
	if err != nil {
		log.Printf("Failed to create inventory alerts job: %v", err)
	} else {
		js.jobJobs["inventory-alerts"] = alertsJob
	}

	// Performance metrics collection - every 15 minutes
	metricsJob, err := js.scheduler.NewJob(
		gocron.DurationJob(15*time.Minute),
		gocron.NewTask(js.collectPerformanceMetrics),
		gocron.WithName("performance-metrics"),
	)
	if err != nil {
		log.Printf("Failed to create metrics job: %v", err)
	} else {
		js.jobJobs["metrics"] = metricsJob
	}

	log.Printf("Registered %d background jobs", len(js.jobJobs))
}

// refreshTenantAnalytics refreshes analytics for all tenants
func (js *JobScheduler) refreshTenantAnalytics(ctx context.Context) error {
	log.Printf("Starting tenant analytics refresh")

	// Get all active tenants
	tenants, err := js.tenantRepo.List(ctx, 1000, 0) // Reasonable limit
	if err != nil {
		log.Printf("Failed to get tenants for analytics refresh: %v", err)
		return err
	}

	// Process tenants in parallel with concurrency control
	semaphore := make(chan struct{}, 5) // Limit to 5 concurrent operations
	var wg sync.WaitGroup

	for _, tenant := range tenants {
		if tenant.Status != "active" {
			continue
		}

		wg.Add(1)
		go func(tenantID uuid.UUID) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			// Invalidated cached analytics to force refresh
			if err := js.analyticsSvc.InvalidateTenantAnalyticsCache(ctx, tenantID); err != nil {
				log.Printf("Failed to invalidate analytics cache for tenant %s: %v", tenantID.String(), err)
			}

			// Refresh analytics by calling the method which will cache the results
			if _, err := js.analyticsSvc.CalculateTenantAnalytics(ctx, tenantID); err != nil {
				log.Printf("Failed to refresh analytics for tenant %s: %v", tenantID.String(), err)
			} else {
				log.Printf("Refreshed analytics for tenant %s", tenantID.String())
			}
		}(tenant.ID)
	}

	wg.Wait()
	log.Printf("Completed tenant analytics refresh for %d tenants", len(tenants))
	return nil
}

// cleanupExpiredCache performs cleanup of expired cache entries
func (js *JobScheduler) cleanupExpiredCache() error {
	log.Printf("Starting cache cleanup")

	// Redis handles TTL automatically, but we might need to clean up specific patterns
	// For now, just log that cleanup ran
	log.Printf("Cache cleanup completed (Redis handles TTL automatically)")

	return nil
}

// processInventoryAlerts checks for low stock and processes alerts
func (js *JobScheduler) processInventoryAlerts() error {
	log.Printf("Starting inventory alerts processing")

	// Get all tenants and process their inventory
	tenants, err := js.tenantRepo.List(context.Background(), 1000, 0)
	if err != nil {
		log.Printf("Failed to get tenants for inventory alerts: %v", err)
		return err
	}

	for _, tenant := range tenants {
		if tenant.Status != "active" {
			continue
		}

		// Check low stock levels
		inventories, err := js.inventoryRepo.List(context.Background(), tenant.ID, 1000, 0)
		if err != nil {
			log.Printf("Failed to get inventory for tenant %s: %v", tenant.ID.String(), err)
			continue
		}

		lowStockCount := 0
		for _, inv := range inventories {
			if inv.Quantity < 10 { // Threshold
				lowStockCount++
			}
		}

		if lowStockCount > 0 {
			log.Printf("ALERT: Tenant %s has %d inventory items with low stock", tenant.Name, lowStockCount)
			// TODO: Send notifications via email/SMS
		}
	}

	log.Printf("Completed inventory alerts processing")
	return nil
}

// collectPerformanceMetrics collects and stores performance metrics
func (js *JobScheduler) collectPerformanceMetrics() error {
	log.Printf("Collecting performance metrics")

	// Get cache hit rate from Redis
	// This would require Redis INFO command parsing
	// For now, just log a placeholder
	log.Printf("Performance metrics collection completed")

	return nil
}

// AddJob adds a custom job to the scheduler
func (js *JobScheduler) AddJob(name string, interval time.Duration, taskFn interface{}, params ...interface{}) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	job, err := js.scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(taskFn, params...),
		gocron.WithName(name),
	)

	if err != nil {
		return err
	}

	js.jobJobs[name] = job
	log.Printf("Added custom job: %s", name)
	return nil
}

// RemoveJob removes a job from the scheduler
func (js *JobScheduler) RemoveJob(name string) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	if job, exists := js.jobJobs[name]; exists {
		err := js.scheduler.RemoveJob(job.ID())
		delete(js.jobJobs, name)
		return err
	}

	return nil
}

// GetJobStatus returns information about scheduled jobs
func (js *JobScheduler) GetJobStatus() map[string]interface{} {
	js.mu.RLock()
	defer js.mu.RUnlock()

	status := make(map[string]interface{})
	status["total_jobs"] = len(js.jobJobs)
	jobs := make([]string, 0, len(js.jobJobs))

	for name := range js.jobJobs {
		jobs = append(jobs, name)
	}

	status["jobs"] = jobs

	return status
}