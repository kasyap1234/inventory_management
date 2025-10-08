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
	"agromart2/internal/services"
)

const version = "1.0.0"

func main() {
	// Load Tally configuration
	tallyConfig, err := config.LoadTallyConfig("config/tally.toml")
	if err != nil {
		log.Fatalf("Failed to load tally config: %v", err)
	}

	// Database connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
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

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Verify database connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connection pool initialized successfully")

	// JWT configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = random.String(32) // Generate random secret for development
		log.Printf("WARNING: Using generated JWT secret: %s", jwtSecret)
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
		minioAccessKey = "minioadmin" // Default for development
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		minioSecretKey = "minioadmin" // Default for development
	}
	minioSSLStr := os.Getenv("MINIO_USE_SSL")
	useSSL := false
	if minioSSLStr == "true" {
		useSSL = true
	}

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

	// Create cache service
	cacheSvc := caching.NewRedisCacheService(redisAddr, redisPassword, redisDB)

	// Create services
	// Create analytics service
	analyticsSvc := analytics.NewAnalyticsService(orderRepo, invoiceRepo, inventoryRepo, productRepo, cacheSvc, pool)

	// Create RBAC service with caching for improved performance
	rbacService := services.NewRBACServiceWithCache(userRoleRepo, rolePermissionRepo, permissionRepo, cacheSvc)

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
		roleRepo,
		userRoleRepo,
		rbacMiddleware,
	)
	userHandlers := handlers.NewUserHandlers(userRepo, tenantRepo, rbacMiddleware)
	tenantHandlers := handlers.NewTenantHandlers(tenantService, rbacMiddleware)
	categoryHandlers := handlers.NewCategoryHandlers(categoryRepo, rbacMiddleware)
	warehouseHandlers := handlers.NewWarehouseHandlers(
		services.NewWarehouseService(warehouseRepo),
		rbacMiddleware,
	)
	distributorHandlers := handlers.NewDistributorHandlers(
		services.NewDistributorService(distributorRepo),
		rbacMiddleware,
	)
	supplierHandlers := handlers.NewSupplierHandlers(
		services.NewSupplierService(supplierRepo),
		rbacMiddleware,
	)
	inventoryService := services.NewInventoryService(inventoryRepo, productRepo, cacheSvc)

	orderSvc := services.NewOrderService(orderRepo, inventoryRepo, inventoryService)

	invoiceSvc := services.NewInvoiceService(invoiceRepo, orderRepo, analyticsSvc, pool)
	inventoryHandlers := handlers.NewInventoryHandlers(
		inventoryService,
		rbacMiddleware,
	)
	orderHandlers := handlers.NewOrderHandlers(orderSvc, rbacMiddleware)
	invoiceHandlers := handlers.NewInvoiceHandlers(invoiceSvc, orderSvc, productSvc, minioSvc, rbacMiddleware)

	jobHandlers := handlers.NewJobHandlers(
		tallyExporter,
		tallyImporter,
		jobs.NewInventoryAlertService(inventoryRepo, productRepo),
		jobs.NewAnalyticsRefreshService(analyticsSvc),
		analyticsSvc,
		orderRepo,
		invoiceRepo,
		productRepo,
		inventoryRepo,
	)

	// Create subscription, notification, and audit log handlers
	subscriptionRepo := repositories.NewSubscriptionRepo(pool)
	razorpayService := services.NewRazorpayService(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, razorpayService)
	subscriptionHandlers := handlers.NewSubscriptionHandlers(subscriptionService, rbacMiddleware)

	// Notification service configuration
	sendgridAPIKey := os.Getenv("SENDGRID_API_KEY")
	twilioAccountSID := os.Getenv("TWILIO_ACCOUNT_SID")
	twilioAuthToken := os.Getenv("TWILIO_AUTH_TOKEN")
	twilioPhone := os.Getenv("TWILIO_PHONE_NUMBER")

	notificationService := services.NewNotificationService(
		redisAddr,
		redisPassword,
		redisDB,
		sendgridAPIKey,
		twilioAccountSID,
		twilioAuthToken,
		twilioPhone,
	)
	notificationHandlers := handlers.NewNotificationHandlers(notificationService)

	auditLogsRepo := repositories.NewAuditLogsRepo(pool)
	auditLogsService := services.NewAuditLogsService(auditLogsRepo)
	auditLogsHandlers := handlers.NewAuditLogsHandlers(auditLogsService, rbacMiddleware)

	// Create analytics handlers
	analyticsHandlers := handlers.NewAnalyticsHandlers(analyticsSvc, rbacMiddleware)

	// Create Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer asynqClient.Close()

	// Create tally services and handlers
	tallyExporter := jobs.NewTallyExporter(invoiceRepo, orderRepo, productRepo, tallyConfig)
	tallyImporter := jobs.NewTallyImporter(orderRepo, invoiceRepo, tallyConfig)
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

	// Create Echo instance with optimized settings
	e := echo.New()
	e.HideBanner = true
	e.HidePort = false

	// Performance middleware
	perfMiddleware := middleware.NewPerformanceMiddleware()

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

	// Authentication routes (no JWT required for signup/login)
	auth := v1.Group("/auth")
	auth.POST("/signup", authHandlers.Signup)
	auth.POST("/login", authHandlers.Login)
	auth.POST("/refresh", authHandlers.Refresh)

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
	protected.GET("/tenants/:id", tenantHandlers.GetTenant)
	protected.PUT("/tenants/:id", tenantHandlers.UpdateTenant)
	protected.DELETE("/tenants/:id", tenantHandlers.DeleteTenant)

	// Business routes
	protected.GET("/categories", categoryHandlers.ListCategories)
	protected.POST("/categories", categoryHandlers.CreateCategory)
	protected.GET("/categories/:id", categoryHandlers.GetCategory)
	protected.PUT("/categories/:id", categoryHandlers.UpdateCategory)
	protected.DELETE("/categories/:id", categoryHandlers.DeleteCategory)

	// Product routes
	protected.GET("/products", productHandlers.ListProducts)
	protected.POST("/products", productHandlers.CreateProduct)
	protected.GET("/products/:id", productHandlers.GetProduct)
	protected.PUT("/products/:id", productHandlers.UpdateProduct)
	protected.DELETE("/products/:id", productHandlers.DeleteProduct)
	protected.GET("/products/search", productHandlers.SearchProducts)
	protected.POST("/products/bulk/update", productHandlers.BulkUpdateProducts)
	protected.POST("/products/bulk/create", productHandlers.BulkCreateProducts)
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
	protected.GET("/inventory/:id", inventoryHandlers.GetInventory)
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

	log.Printf("🚀 Agromart2 server v%s starting on port %d", version, port)
	log.Printf("Database connected: %t", databaseURL != "") // Don't log the actual URL for security

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}
