package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"

	"agromart2/internal/analytics"
	"agromart2/internal/caching"
	"agromart2/internal/config"
	"agromart2/internal/handlers"
	"agromart2/internal/jobs"
	"agromart2/internal/middleware"
	"agromart2/internal/repositories"
	"agromart2/internal/security"
	"agromart2/internal/services"
	"agromart2/internal/validation"
)

const version = "1.0.0"

func isProduction(environment string) bool {
	return environment == "production" || environment == "prod"
}

func main() {
	// Load environment variables from .env if present
	if err := godotenv.Load(); err != nil {
		log.Printf("INFO: .env not found or could not be loaded: %v", err)
	}

	// Environment detection
	environment := os.Getenv("ENV")
	if environment == "" {
		environment = os.Getenv("GO_ENV")
	}
	if environment == "" {
		environment = "development"
	}
	log.Printf("INFO: Detected environment: %s", environment)

	// Load Tally configuration
	tallyConfig, err := config.LoadTallyConfig("config/tally.toml")
	if err != nil {
		log.Fatalf("Failed to load tally config: %v", err)
	}

	// Database connection
	var databaseURL string
	if os.Getenv("TEST_ENV") == "true" {
		databaseURL = tallyConfig.Tally.TestDatabaseURL
		log.Println("INFO: Using test database")
	} else {
		databaseURL = os.Getenv("DATABASE_URL")
	}

	if databaseURL == "" {
		log.Fatal("DATABASE_URL (or TEST_DATABASE_URL in test mode) environment variable is required")
	}
	
	// Don't log the actual database URL in production for security
	if !isProduction(environment) {
		log.Printf("DEBUG: Database URL configured (details hidden in production)")
	}

	// Create optimized database connection pool
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		log.Fatalf("Failed to parse database URL: %v", err)
	}

	// Optimize connection pool settings for production performance
	poolConfig.MaxConns = 50                       // Maximum connections in pool
	poolConfig.MinConns = 10                       // Minimum idle connections
	poolConfig.MaxConnLifetime = 30 * time.Minute  // Recycle connections every 30 min
	poolConfig.MaxConnIdleTime = 5 * time.Minute   // Close idle connections after 5 min
	poolConfig.HealthCheckPeriod = 1 * time.Minute // Health check interval

	var pool *pgxpool.Pool
	maxDBAttempts := 5
	for attempt := 1; attempt <= maxDBAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		cancel()
		if err != nil {
			log.Printf("Database connection attempt %d/%d failed: %v", attempt, maxDBAttempts, err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr := pool.Ping(pingCtx)
		pingCancel()
		if pingErr != nil {
			log.Printf("Database ping attempt %d/%d failed: %v", attempt, maxDBAttempts, pingErr)
			pool.Close()
			pool = nil
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		log.Printf("Database connection established on attempt %d", attempt)
		break
	}

	if pool == nil {
		log.Fatalf("Failed to establish database connection after %d attempts", maxDBAttempts)
	}
	defer pool.Close()

	// JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if isProduction(environment) {
			log.Fatalf("JWT_SECRET is required in production mode")
		} else {
			jwtSecret = random.String(32)
			log.Printf("WARNING: Using generated JWT secret: %s", jwtSecret)
		}
	} else if isProduction(environment) && len(jwtSecret) < 32 {
		log.Fatalf("JWT_SECRET must be at least 32 characters in production mode")
	}

	// Redis configuration
	redisAddr := tallyConfig.Queuing.RedisAddr
	redisPassword := tallyConfig.Queuing.RedisPassword
	redisDB := tallyConfig.Queuing.RedisDB

	// MinIO configuration
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000" // Default MinIO endpoint
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		if isProduction(environment) {
			log.Fatalf("MINIO_ACCESS_KEY is required in production mode")
		} else {
			log.Fatalf("MINIO_ACCESS_KEY is required. Set it in your environment or .env file")
		}
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		if isProduction(environment) {
			log.Fatalf("MINIO_SECRET_KEY is required in production mode")
		} else {
			log.Fatalf("MINIO_SECRET_KEY is required. Set it in your environment or .env file")
		}
	}
	minioSSLStr := os.Getenv("MINIO_USE_SSL")
	useSSL := false
	if minioSSLStr == "true" {
		useSSL = true
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	enforceHTTPS := os.Getenv("ENFORCE_HTTPS") == "true"

	csrfSecret := os.Getenv("CSRF_SECRET")
	if csrfSecret == "" {
		if isProduction(environment) {
			log.Fatalf("CSRF_SECRET is required in production mode")
		} else {
			csrfSecret = random.String(32)
			log.Printf("WARNING: CSRF_SECRET is not configured; generating ephemeral secret (tokens will reset on restart)")
		}
	} else if isProduction(environment) && len(csrfSecret) < 32 {
		log.Fatalf("CSRF_SECRET must be at least 32 characters in production mode")
	}
	csrfManager := security.NewCSRFTokenManager(csrfSecret, 2*time.Hour)

	// Initialize MinIO service
	minioSvc, err := services.NewMinioService(minioEndpoint, minioAccessKey, minioSecretKey, useSSL)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO service: %v", err)
	}

	// Create repositories
	userRepo := repositories.NewUserRepo(pool)
	tenantRepo := repositories.NewTenantRepo(pool)
	roleRepo := repositories.NewRoleRepo(pool)
	userRoleRepo := repositories.NewUserRoleRepo(pool)
	rolePermissionRepo := repositories.NewRolePermissionRepo(pool)
	permissionRepo := repositories.NewPermissionRepo(pool)
	categoryRepo := repositories.NewCategoryRepo(pool)
	productRepo := repositories.NewProductRepo(pool)
	warehouseRepo := repositories.NewWarehouseRepository(pool)
	supplierRepo := repositories.NewSupplierRepository(pool)
	distributorRepo := repositories.NewDistributorRepository(pool)
	inventoryRepo := repositories.NewInventoryRepo(pool)
	orderRepo := repositories.NewOrderRepo(pool)
	invoiceRepo := repositories.NewInvoiceRepo(pool)
	productImageRepo := repositories.NewProductImageRepo(pool)
	auditLogsRepo := repositories.NewAuditLogsRepo(pool)

	// Create cache service
	cacheSvc := caching.NewRedisCacheService(redisAddr, redisPassword, redisDB)

	// Notification service configuration
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	resendFromEmail := os.Getenv("RESEND_FROM_EMAIL")
	resendFromName := os.Getenv("RESEND_FROM_NAME")
	twilioAccountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")
	twilioPhone := os.Getenv("TWILIO_PHONE_NUMBER")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := 587
	if smtpPortStr := os.Getenv("SMTP_PORT"); smtpPortStr != "" {
		if parsed, err := strconv.Atoi(smtpPortStr); err != nil {
			log.Printf("WARNING: Invalid SMTP_PORT '%s', defaulting to %d", smtpPortStr, smtpPort)
		} else {
			smtpPort = parsed
		}
	}
	smtpUsername := os.Getenv("SMTP_USERNAME")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFromEmail := os.Getenv("SMTP_FROM_EMAIL")
	smtpFromName := os.Getenv("SMTP_FROM_NAME")

	notificationService := services.NewNotificationService(
		redisAddr,
		redisPassword,
		redisDB,
		resendAPIKey,
		resendFromEmail,
		resendFromName,
		twilioAccountSID,
		twilioAuthToken,
		twilioPhone,
		smtpHost,
		smtpPort,
		smtpUsername,
		smtpPassword,
		smtpFromEmail,
		smtpFromName,
	)

	// Create services
	// Create analytics service
	analyticsSvc := analytics.NewAnalyticsService(orderRepo, invoiceRepo, inventoryRepo, productRepo, cacheSvc, pool)

	// Create RBAC service with caching for improved performance
	rbacService := services.NewRBACServiceWithCache(userRoleRepo, roleRepo, rolePermissionRepo, permissionRepo, cacheSvc)

	// RBAC middleware
	rbacMiddleware := middleware.NewRBACMiddleware(rbacService)

	// Create auth service
	authService := services.NewAuthService(cacheSvc, jwtSecret, 3600, 86400) // 1 hour access, 24 hour refresh

	// Create product service
	productSvc := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, minioSvc, cacheSvc)

	// Create product handlers
	productHandlers := handlers.NewProductHandlers(productSvc, rbacMiddleware)

	// Create tenant service
	tenantService := services.NewTenantService(tenantRepo)

	// Create order service
	// orderSvc := services.NewOrderService(orderRepo, inventoryRepo, inventoryService) // moved after inventoryService

	// Create invoice service
	// invoiceSvc := services.NewInvoiceService(invoiceRepo, orderRepo, analyticsSvc, pool) // moved after inventoryService

	// Create handlers
	authHandlers := handlers.NewAuthHandlers(
		authService,
		userRepo,
		tenantRepo,
		roleRepo,
		userRoleRepo,
		rolePermissionRepo,
		permissionRepo,
		rbacMiddleware,
		notificationService,
		frontendURL,
	)
	userHandlers := handlers.NewUserHandlers(userRepo, tenantRepo, rbacMiddleware)
	tenantHandlers := handlers.NewTenantHandlers(tenantService, rbacMiddleware)
	categoryHandlers := handlers.NewCategoryHandlers(categoryRepo, rbacMiddleware)
	
	// Create role and permission handlers
	roleHandlers := handlers.NewRoleHandlers(rbacService, rbacMiddleware)
	permissionHandlers := handlers.NewPermissionHandlers(rbacService, rbacMiddleware)
	
	warehouseHandlers := handlers.NewWarehouseHandlers(
		services.NewWarehouseService(warehouseRepo),
		rbacMiddleware,
	)
	distributorService := services.NewDistributorService(distributorRepo)
	distributorHandlers := handlers.NewDistributorHandlers(
		distributorService,
		rbacMiddleware,
	)
	supplierService := services.NewSupplierService(supplierRepo)
	supplierHandlers := handlers.NewSupplierHandlers(
		supplierService,
		rbacMiddleware,
	)
	inventoryService := services.NewInventoryService(inventoryRepo, productRepo, cacheSvc)
	auditLogsService := services.NewAuditLogsService(auditLogsRepo)

	orderSvc := services.NewOrderService(orderRepo, inventoryRepo, inventoryService)

	invoiceSvc := services.NewInvoiceService(invoiceRepo, orderRepo, analyticsSvc, pool)
	inventoryHandlers := handlers.NewInventoryHandlers(
		inventoryService,
		auditLogsService,
		rbacMiddleware,
	)
	orderHandlers := handlers.NewOrderHandlers(orderSvc, rbacMiddleware)
	invoiceHandlers := handlers.NewInvoiceHandlers(
		invoiceSvc,
		orderSvc,
		productSvc,
		minioSvc,
		rbacMiddleware,
		supplierService,
		distributorService,
	)

	tallyExporter := jobs.NewTallyExporter(invoiceRepo, orderRepo, productRepo, tallyConfig)
	tallyImporter := jobs.NewTallyImporter(orderRepo, invoiceRepo, tallyConfig)

	// Create job inspector for monitoring background jobs
	jobInspector := jobs.NewAsynqJobInspector(redisAddr, redisPassword, redisDB)

	jobHandlers := handlers.NewJobHandlers(
		tallyExporter,
		tallyImporter,
		jobs.NewInventoryAlertService(inventoryRepo, productRepo),
		jobs.NewAnalyticsRefreshService(analyticsSvc, pool),
		analyticsSvc,
		orderRepo,
		invoiceRepo,
		productRepo,
		inventoryRepo,
		jobInspector,
	)

	// Create subscription, notification, and audit log handlers
	subscriptionRepo := repositories.NewSubscriptionRepo(pool)
	razorpayKeyID := os.Getenv("RAZORPAY_KEY_ID")
	razorpayKeySecret := os.Getenv("RAZORPAY_KEY_SECRET")
	razorpayWebhookSecret := os.Getenv("RAZORPAY_WEBHOOK_SECRET")
	if razorpayWebhookSecret == "" {
		log.Printf("WARNING: RAZORPAY_WEBHOOK_SECRET is not configured; Razorpay webhooks will be rejected")
	}
	razorpayService := services.NewRazorpayService(razorpayKeyID, razorpayKeySecret, razorpayWebhookSecret)
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, razorpayService)
	subscriptionHandlers := handlers.NewSubscriptionHandlers(subscriptionService, rbacMiddleware)

	notificationHandlers := handlers.NewNotificationHandlers(notificationService)

	auditLogsHandlers := handlers.NewAuditLogsHandlers(auditLogsService, rbacMiddleware)

	// Create analytics handlers
	analyticsHandlers := handlers.NewAnalyticsHandlers(analyticsSvc, rbacMiddleware)
	securityHandlers := handlers.NewSecurityHandlers(csrfManager)

	// Create Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer asynqClient.Close()

	// Create tally handlers
	tallyHandlers := handlers.NewTallyHandlers(tallyExporter, tallyImporter, asynqClient)

	// Create Asynq server
	asynqSrv := asynq.NewServer(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	}, asynq.Config{
		// Set concurrency level from config
		Concurrency: tallyConfig.Queuing.Concurrency,
		// Queues: if the task is not defined in this map, it would use the "default" queue
		Queues: tallyConfig.Queuing.QueuePriorities,
	})

	// Create Asynq mux and register handlers
	mux := asynq.NewServeMux()
	mux.HandleFunc(jobs.TypeTallyExport, tallyExporter.TallyExportHandler)
	mux.HandleFunc(jobs.TypeTallyImport, tallyImporter.TallyImportHandler)
	mux.HandleFunc(jobs.TypeTallyImportScan, tallyImporter.TallyImportScanHandler)

	// Optional: background scheduler for Tally CSV imports (file-based), controlled by env
	if os.Getenv("TALLY_IMPORT_ENABLE") == "true" {
		intervalMinutes := 15
		if v := os.Getenv("TALLY_IMPORT_INTERVAL_MINUTES"); v != "" {
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				intervalMinutes = parsed
			}
		}

		// Prefer Asynq Scheduler when CRON provided; otherwise fallback to ticker
		cron := os.Getenv("TALLY_IMPORT_SCHEDULER_CRON")
		if cron != "" {
			log.Printf("INFO: Tally scheduled import (Asynq) enabled: cron=%s", cron)
			scheduler := asynq.NewScheduler(asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword, DB: redisDB}, &asynq.SchedulerOpts{})
			if _, err := scheduler.Register(cron, jobs.NewTallyImportScanTask(), asynq.Queue("default")); err != nil {
				log.Printf("ERROR: Failed to register Tally import scan schedule: %v", err)
			} else {
				go func() {
					if err := scheduler.Run(); err != nil {
						log.Printf("ERROR: Tally scheduler stopped: %v", err)
					}
				}()
				defer scheduler.Shutdown()
			}
		} else {
			log.Printf("INFO: Tally scheduled import (ticker) enabled: every %d minute(s)", intervalMinutes)
			go func() {
				ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
				defer ticker.Stop()
				if err := tallyImporter.ScheduledImportJob(context.Background()); err != nil {
					log.Printf("WARN: Tally scheduled import run failed: %v", err)
				}
				for range ticker.C {
					if err := tallyImporter.ScheduledImportJob(context.Background()); err != nil {
						log.Printf("WARN: Tally scheduled import run failed: %v", err)
					}
				}
			}()
		}
	}

	// Create Echo instance with optimized settings
	e := echo.New()
	e.Validator = validation.NewValidator()
	e.HTTPErrorHandler = handlers.HTTPErrorHandler
	e.HideBanner = true
	e.HidePort = false

	e.Pre(middleware.EnforceHTTPS(enforceHTTPS))

	csrfSkipPaths := map[string]struct{}{
		"/v1/security/csrf":     {},
		"/v1/webhooks/razorpay": {},
	}
	e.Use(middleware.CSRFProtection(csrfManager, csrfSkipPaths))

	// Performance middleware with optional Redis rate limiting
	useRedisRateLimit := os.Getenv("USE_REDIS_RATE_LIMIT") == "true"
	perfMiddleware := middleware.NewPerformanceMiddlewareWithRedis(redisAddr, redisPassword, redisDB, useRedisRateLimit)
	if useRedisRateLimit {
		log.Printf("INFO: Redis-backed rate limiting enabled for cluster-aware enforcement")
	} else {
		log.Printf("INFO: In-memory rate limiting enabled (single instance mode)")
	}

	forgotPasswordRateLimiter := perfMiddleware.EndpointRateLimiter(5, 15*time.Minute, 5)
	resendVerificationRateLimiter := perfMiddleware.EndpointRateLimiter(3, 15*time.Minute, 3)
	log.Printf("INFO: Endpoint-specific rate limiting enabled for /auth/password/forgot (5 req/15min)")
	log.Printf("INFO: Endpoint-specific rate limiting enabled for /auth/verify/resend (3 req/15min)")

	// Global middleware (order matters for performance)
	e.Use(echoMiddleware.Recover())
	e.Use(perfMiddleware.Gzip())                    // Enable gzip compression
	e.Use(perfMiddleware.RateLimiter())             // Rate limiting (100 req/min per IP)
	e.Use(perfMiddleware.Timeout(30 * time.Second)) // Request timeout
	e.Use(perfMiddleware.BodyLimit("10M"))          // Limit request body size
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.RemoveTrailingSlash())

	// Request ID for tracing
	e.Use(echoMiddleware.RequestID())
	e.Use(middleware.SecurityHeaders())

	// Version middleware
	versionMiddleware := middleware.NewVersionMiddleware()
	e.Use(versionMiddleware.APIVersionResolver())

	// Health endpoints (no auth required)
	e.GET("/health", handlers.HealthCheck)
	e.GET("/health/ready", handlers.ReadinessCheck)
	e.GET("/health/detailed", func(c echo.Context) error {
		return handlers.HealthCheckDetailed(c, pool)
	})

	// Metrics endpoint (no auth required)
	e.GET("/metrics", handlers.MetricsHandler)

	// Documentation static files (no auth required)
	e.Static("/docs", "docs")

	// API routes
	v1 := e.Group("/v1")
	v1.Use(versionMiddleware.VersionHeader("v1"))

	// Documentation routes (no auth required)
	v1.GET("/docs/guide", handlers.DocumentationGuideHandler)
	v1.GET("/docs/spec", handlers.DocumentationSpecHandler)

	// Security utilities
	v1.GET("/security/csrf", securityHandlers.GetCSRFToken)

	// Authentication routes (no JWT required for signup/login)
	auth := v1.Group("/auth")
	auth.POST("/signup", authHandlers.Signup)
	auth.POST("/login", authHandlers.Login)
	auth.POST("/refresh", authHandlers.Refresh)
	auth.POST("/password/forgot", authHandlers.ForgotPassword, forgotPasswordRateLimiter)
	auth.POST("/password/reset", authHandlers.ResetPassword)
	auth.POST("/verify", authHandlers.VerifyEmail)
	auth.POST("/verify/resend", authHandlers.ResendVerificationEmail, resendVerificationRateLimiter)

	// Protected routes (require JWT and RBAC)
	protected := v1.Group("")
	protected.Use(middleware.JWTMiddleware(userRepo, jwtSecret))

	// Protected auth routes
	protected.POST("/auth/logout", authHandlers.Logout)

	// User routes
	protected.GET("/me", authHandlers.Me)
	protected.GET("/users", userHandlers.ListUsers)
	protected.GET("/users/:id", userHandlers.GetUser)
	protected.POST("/users", userHandlers.CreateUser)
	protected.PUT("/users/:id", userHandlers.UpdateUser)
	protected.DELETE("/users/:id", userHandlers.DeleteUser)

	// Tenant routes
	protected.GET("/tenants", tenantHandlers.ListTenants)
	protected.POST("/tenants", tenantHandlers.CreateTenant)
	protected.GET("/tenants/:id", tenantHandlers.GetTenant)
	protected.PUT("/tenants/:id", tenantHandlers.UpdateTenant)
	protected.DELETE("/tenants/:id", tenantHandlers.DeleteTenant)

	// Business routes
	protected.GET("/categories", categoryHandlers.ListCategories)
	protected.POST("/categories", categoryHandlers.CreateCategory)
	protected.GET("/categories/:id", categoryHandlers.GetCategory)
	protected.PUT("/categories/:id", categoryHandlers.UpdateCategory)
	protected.DELETE("/categories/:id", categoryHandlers.DeleteCategory)

	// Role routes
	protected.GET("/roles", roleHandlers.ListRoles)
	protected.POST("/roles", roleHandlers.CreateRole)
	protected.GET("/roles/:id", roleHandlers.GetRole)
	protected.PUT("/roles/:id", roleHandlers.UpdateRole)
	protected.DELETE("/roles/:id", roleHandlers.DeleteRole)

	// Permission routes
	protected.GET("/permissions", permissionHandlers.ListPermissions)
	
	// Role-Permission association routes
	protected.POST("/roles/:id/permissions", permissionHandlers.AssignPermissionsToRole)
	protected.DELETE("/roles/:id/permissions/:permissionId", permissionHandlers.RevokePermissionFromRole)
	protected.GET("/roles/:id/permissions", permissionHandlers.GetRolePermissions)

	// Product routes
	protected.GET("/products", productHandlers.ListProducts)
	protected.POST("/products", productHandlers.CreateProduct)
	protected.GET("/products/:id", productHandlers.GetProduct)
	protected.GET("/products/barcode/:barcode", productHandlers.GetProductByBarcode)
	protected.PUT("/products/:id", productHandlers.UpdateProduct)
	protected.DELETE("/products/:id", productHandlers.DeleteProduct)
	protected.GET("/products/search", productHandlers.SearchProducts)
	protected.POST("/products/bulk/update", productHandlers.BulkUpdateProducts)
	protected.POST("/products/bulk/create", productHandlers.BulkCreateProducts)
	protected.DELETE("/products/bulk/delete", productHandlers.BulkDeleteProducts)
	protected.POST("/products/bulk-price-update", productHandlers.BulkPriceUpdate)

	// Product image routes
	protected.POST("/products/:id/images", productHandlers.UploadProductImage)
	protected.GET("/products/:id/images", productHandlers.GetProductImages)
	protected.GET("/products/:id/images/:imageId/url", productHandlers.GetProductImageURL)
	protected.DELETE("/products/:id/images/:imageId", productHandlers.DeleteProductImage)

	protected.GET("/warehouses", warehouseHandlers.ListWarehouses)
	protected.POST("/warehouses", warehouseHandlers.CreateWarehouse)
	protected.GET("/warehouses/:id", warehouseHandlers.GetWarehouse)
	protected.PUT("/warehouses/:id", warehouseHandlers.UpdateWarehouse)
	protected.DELETE("/warehouses/:id", warehouseHandlers.DeleteWarehouse)

	protected.GET("/distributors", distributorHandlers.ListDistributors)
	protected.POST("/distributors", distributorHandlers.CreateDistributor)
	protected.GET("/distributors/:id", distributorHandlers.GetDistributor)
	protected.PUT("/distributors/:id", distributorHandlers.UpdateDistributor)
	protected.DELETE("/distributors/:id", distributorHandlers.DeleteDistributor)

	protected.GET("/suppliers", supplierHandlers.ListSuppliers)
	protected.POST("/suppliers", supplierHandlers.CreateSupplier)
	protected.GET("/suppliers/:id", supplierHandlers.GetSupplier)
	protected.PUT("/suppliers/:id", supplierHandlers.UpdateSupplier)
	protected.DELETE("/suppliers/:id", supplierHandlers.DeleteSupplier)

	protected.GET("/inventory", inventoryHandlers.ListInventories)
	protected.POST("/inventory", inventoryHandlers.CreateInventory)
	protected.POST("/inventory/adjust", inventoryHandlers.AdjustStock)
	protected.GET("/inventory/:id", inventoryHandlers.GetInventory)
	protected.GET("/inventory/:id/history", inventoryHandlers.GetInventoryHistory)
	protected.PUT("/inventory/:id", inventoryHandlers.UpdateInventory)
	protected.DELETE("/inventory/:id", inventoryHandlers.DeleteInventory)
	protected.GET("/inventory/search", inventoryHandlers.SearchInventories)
	// Start Asynq server in a goroutine
	go func() {
		if err := asynqSrv.Start(mux); err != nil {
			log.Fatalf("Could not start Asynq server: %v", err)
		}
		log.Println("Asynq server started")
	}()
	defer asynqSrv.Shutdown()

	protected.GET("/orders", orderHandlers.GetOrders)
	protected.POST("/orders", orderHandlers.CreateOrder)
	protected.GET("/orders/:id", orderHandlers.GetOrder)
	protected.PUT("/orders/:id", orderHandlers.UpdateOrder)
	protected.DELETE("/orders/:id", orderHandlers.DeleteOrder)
	protected.POST("/orders/:id/approve", orderHandlers.ApproveOrder)
	protected.POST("/orders/:id/process", orderHandlers.ProcessOrder)
	protected.POST("/orders/:id/receive", orderHandlers.ReceiveOrder)
	protected.POST("/orders/:id/ship", orderHandlers.ShipOrder)
	protected.POST("/orders/:id/deliver", orderHandlers.DeliverOrder)
	protected.POST("/orders/:id/cancel", orderHandlers.CancelOrder)

	protected.GET("/invoices", invoiceHandlers.ListInvoices)
	protected.POST("/invoices", invoiceHandlers.CreateInvoice)
	protected.POST("/invoices/bulk-create", invoiceHandlers.BulkCreateInvoices)
	protected.GET("/invoices/:id", invoiceHandlers.GetInvoice)
	protected.PUT("/invoices/:id", invoiceHandlers.UpdateInvoice)
	protected.PUT("/invoices/:id/status", invoiceHandlers.UpdateInvoiceStatus)
	protected.GET("/invoices/unpaid", invoiceHandlers.GetUnpaidInvoices)
	protected.POST("/invoices/:id/generate-pdf", invoiceHandlers.GenerateInvoicePDF)
	protected.DELETE("/invoices/:id", invoiceHandlers.DeleteInvoice)

	// Tally routes
	protected.POST("/api/tally/export", tallyHandlers.ExportTallyData)
	protected.POST("/api/tally/import", tallyHandlers.ImportTallyData)

	// Subscription routes
	protected.GET("/subscriptions", subscriptionHandlers.ListSubscriptions)
	protected.POST("/subscriptions", subscriptionHandlers.CreateSubscription)
	protected.GET("/subscriptions/:id", subscriptionHandlers.GetSubscriptionByID)
	protected.PUT("/subscriptions/:id", subscriptionHandlers.UpdateSubscriptionPlan)
	protected.POST("/subscriptions/:id/cancel", subscriptionHandlers.CancelSubscription)
	protected.POST("/subscriptions/:id/pause", subscriptionHandlers.PauseSubscription)
	protected.POST("/subscriptions/:id/resume", subscriptionHandlers.ResumeSubscription)
	protected.DELETE("/subscriptions/:id", subscriptionHandlers.DeleteSubscription)
	protected.GET("/subscriptions/plans", subscriptionHandlers.GetAvailablePlans)

	// Notification routes
	protected.POST("/notifications/send", notificationHandlers.SendNotification)
	protected.GET("/notifications", notificationHandlers.ListNotifications)
	protected.GET("/notifications/:id", notificationHandlers.GetNotification)
	protected.PUT("/notifications/:id/read", notificationHandlers.MarkNotificationAsRead)
	protected.DELETE("/notifications/:id", notificationHandlers.DeleteNotification)
	protected.GET("/webhooks/subscriptions", notificationHandlers.ListWebhookSubscriptions)
	protected.POST("/webhooks/subscriptions", notificationHandlers.CreateWebhookSubscription)
	protected.PUT("/webhooks/subscriptions/:id", notificationHandlers.UpdateWebhookSubscription)
	protected.DELETE("/webhooks/subscriptions/:id", notificationHandlers.DeleteWebhookSubscription)

	// Job management routes
	protected.GET("/jobs", jobHandlers.ListJobs)
	protected.GET("/jobs/:id", jobHandlers.GetJob)
	protected.POST("/jobs/:id/retry", jobHandlers.RetryJob)
	protected.POST("/jobs/:id/cancel", jobHandlers.CancelJob)
	protected.GET("/jobs/stats", jobHandlers.GetJobStats)

	// Audit logs routes
	protected.GET("/audit-logs", auditLogsHandlers.ListAuditLogs)
	protected.GET("/audit-logs/:id", auditLogsHandlers.GetAuditLog)
	protected.GET("/audit-logs/entity/:table/:record_id", auditLogsHandlers.GetEntityHistory)
	protected.GET("/audit-logs/user/:user_id", auditLogsHandlers.GetUserActivity)
	protected.GET("/audit-logs/summary", auditLogsHandlers.GetAuditSummary)
	protected.GET("/audit-logs/tables", auditLogsHandlers.GetTableNames)
	protected.GET("/audit-logs/actions", auditLogsHandlers.GetActions)

	// Advanced order management routes
	protected.GET("/orders/search", orderHandlers.SearchOrders)
	protected.GET("/orders/:id/history", orderHandlers.GetOrderHistory)
	protected.GET("/orders/analytics", orderHandlers.GetOrderAnalytics)

	// Analytics routes
	protected.GET("/analytics/dashboard", analyticsHandlers.GetDashboardAnalytics)
	protected.GET("/analytics/sales-trends", analyticsHandlers.GetSalesTrends)
	protected.GET("/analytics/gst-totals", analyticsHandlers.GetGSTTotals)
	protected.GET("/analytics/top-products", analyticsHandlers.GetTopProducts)
	protected.GET("/analytics/low-stock", analyticsHandlers.GetLowStockReport)
	protected.GET("/analytics/inventory-valuation", analyticsHandlers.GetInventoryValuation)
	protected.GET("/analytics/revenue-by-category", analyticsHandlers.GetRevenueByCategory)
	protected.GET("/analytics/order-status", analyticsHandlers.GetOrderStatusDistribution)
	protected.POST("/analytics/refresh", analyticsHandlers.RefreshAnalytics)

	// Start server
	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = "8080"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port %s: %v", portStr, err)
	}

	// Startup validation summary
	if isProduction(environment) {
		log.Printf("INFO: Production mode - all security secrets validated")
	} else {
		log.Printf("INFO: Development mode - using default or generated secrets")
	}

	log.Printf("🚀 Agromart2 server v%s starting on port %d", version, port)
	if isProduction(environment) {
		log.Printf("INFO: Production mode - database connection established")
	} else {
		log.Printf("DEBUG: Database connected: %t", databaseURL != "")
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
