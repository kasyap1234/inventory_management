package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"

	"github.com/google/uuid"

	"agromart2/internal/analytics"
	"agromart2/internal/caching"
	"agromart2/internal/common"
	"agromart2/internal/handlers"
	"agromart2/internal/jobs"
	"agromart2/internal/middleware"
	"agromart2/internal/repositories"
	"agromart2/internal/services"
)

const version = "1.0.0"

func main() {
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
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379" // Default Redis address
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0 // Default DB
	if redisDBStr != "" {
		if db, err := strconv.Atoi(redisDBStr); err == nil {
			redisDB = db
		}
	}

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
	userRepo := repositories.NewUserRepository(pool)
	tenantRepo := repositories.NewTenantRepository(pool)
	roleRepo := repositories.NewRoleRepository(pool)
	userRoleRepo := repositories.NewUserRoleRepository(pool)
	rolePermissionRepo := repositories.NewRolePermissionRepository(pool)
	permissionRepo := repositories.NewPermissionRepository(pool)
	categoryRepo := repositories.NewCategoryRepository(pool)
	productRepo := repositories.NewProductRepository(pool)
	warehouseRepo := repositories.NewWarehouseRepository(pool)
	supplierRepo := repositories.NewSupplierRepository(pool)
	distributorRepo := repositories.NewDistributorRepository(pool)
	inventoryRepo := repositories.NewInventoryRepository(pool)
	orderRepo := repositories.NewOrderRepository(pool)
	invoiceRepo := repositories.NewInvoiceRepository(pool)
	auditLogRepo := repositories.NewAuditLogsRepository(pool)
	tokenRepo := repositories.NewTokenRepository(pool)
	productImageRepo := repositories.NewProductImageRepository(pool)
	subscriptionRepo := repositories.NewSubscriptionRepository(pool)

	// Create cache service
	cacheSvc := caching.NewRedisCacheService(redisAddr, redisPassword, redisDB)

	// Create services
	// Create analytics service
	analyticsSvc := analytics.NewAnalyticsService(orderRepo, invoiceRepo, inventoryRepo, productRepo, cacheSvc)
	rbacService := services.NewRBACService(userRoleRepo, rolePermissionRepo, permissionRepo)

	// Create product service
	productSvc := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, minioSvc, cacheSvc)

	// Create product handlers
	productHandlers := handlers.NewProductHandlers(productSvc, rbacMiddleware)

	// Create order service
	// orderSvc := services.NewOrderService(orderRepo, inventoryRepo, inventoryService) // moved after inventoryService

	// Create invoice service
	// invoiceSvc := services.NewInvoiceService(invoiceRepo, orderRepo, analyticsSvc, pool) // moved after inventoryService

	// JWT middleware configuration (placeholder - will be fixed)
	jwtConfig := echojwt.Config{
		SigningKey: []byte(jwtSecret),
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(401, "Invalid token")
		},
	}

	jwtSaltedConfig := echojwt.Config{
		SigningKey: []byte(jwtSecret),
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(middleware.JWTCustomClaims)
		},
		ParsePayloadFunc: func(c echo.Context, dst jwt.Claims) error {
			claims, err := middleware.ParseJWTPayload(c, dst.(*middleware.JWTCustomClaims))
			if err != nil {
				return err
			}

			// Add user and tenant IDs to context
			ctx := context.WithValue(c.Request().Context(), common.UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, common.TenantIDKey, claims.TenantID)
			c.SetRequest(c.Request().WithContext(ctx))

			return nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return echo.NewHTTPError(401, "Invalid token")
		},
	}

	// RBAC middleware
	rbacMiddleware := middleware.NewRBACMiddleware(rbacService)

	// Create handlers
	authHandlers := handlers.NewAuthHandlers(
		userRepo,
		tenantRepo,
		roleRepo,
		userRoleRepo,
		rbacMiddleware,
	)
	userHandlers := handlers.NewUserHandlers(userRepo, tenantRepo, rbacMiddleware)
	tenantHandlers := handlers.NewTenantHandlers(tenantRepo, rbacMiddleware)
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

	// API routes
	v1 := e.Group("/v1")
	v1.Use(versionMiddleware.VersionHeader("v1"))

	// Authentication routes (no JWT required for signup/login)
	auth := v1.Group("/auth")
	auth.POST("/signup", authHandlers.Signup)
	auth.POST("/login", authHandlers.Login)
	auth.POST("/refresh", authHandlers.Refresh)

	// Protected routes (require JWT and RBAC)
	protected := v1.Group("")
	protected.Use(echojwt.WithConfig(jwtConfig))

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