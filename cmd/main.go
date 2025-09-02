package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

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

	// Create database connection pool
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

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
	analyticsSvc := analytics.NewAnalyticsService(orderRepo, invoiceRepo, inventoryRepo, productRepo, cacheSvc)

	rbacService := services.NewRBACService(userRoleRepo, rolePermissionRepo, permissionRepo)

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
	orderHandlers := handlers.NewOrderHandlers(orderSvc)
	invoiceHandlers := handlers.NewInvoiceHandlers(invoiceSvc, orderSvc, productSvc, minioSvc)

	// Create Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer asynqClient.Close()

	// Create tally services and handlers
	tallyExporter := jobs.NewTallyExporter(invoiceRepo, orderRepo, productRepo)
	tallyImporter := jobs.NewTallyImporter(orderRepo, invoiceRepo)
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

	// Create Echo instance
	e := echo.New()

	// Global middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoMiddleware.RemoveTrailingSlash())

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

	protected.GET("/invoices", invoiceHandlers.ListInvoices)
	protected.POST("/invoices", invoiceHandlers.CreateInvoice)
	protected.GET("/invoices/:id", invoiceHandlers.GetInvoice)
	protected.PUT("/invoices/:id", invoiceHandlers.UpdateInvoice)
	protected.PUT("/invoices/:id/status", invoiceHandlers.UpdateInvoiceStatus)
	protected.GET("/invoices/unpaid", invoiceHandlers.GetUnpaidInvoices)
	protected.POST("/invoices/:id/generate-pdf", invoiceHandlers.GenerateInvoicePDF)
	protected.DELETE("/invoices/:id", invoiceHandlers.DeleteInvoice)

	// Tally routes
	protected.POST("/api/tally/export", tallyHandlers.ExportTallyData)
	protected.POST("/api/tally/import", tallyHandlers.ImportTallyData)

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
	log.Printf("Database connected: %s", databaseURL != "") // Don't log the actual URL for security

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}