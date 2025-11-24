package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/random"
	"golang.org/x/time/rate"

	"agromart2/internal/analytics"
	"agromart2/internal/caching"
	"agromart2/internal/common"
	"agromart2/internal/config"
	"agromart2/internal/handlers"
	"agromart2/internal/jobs"
	"agromart2/internal/jobs/background"
	"agromart2/internal/middleware"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"agromart2/internal/security"
	"agromart2/internal/services"
	"agromart2/internal/validation"
)

const version = "1.0.0"

func main() {
	// Load environment variables from .env if present
	if err := godotenv.Load(); err != nil {
		log.Printf("INFO: .env not found or could not be loaded: %v", err)
	}

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

	// Verify database schema before proceeding
	if err := verifySchema(pool); err != nil {
		log.Fatalf("Database schema verification failed: %v", err)
	}

	// Check environment for production vs development
	goEnv := os.Getenv("GO_ENV")
	isProduction := goEnv == "production" || goEnv == "prod"

	// JWT configuration with production validation
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		if isProduction {
			log.Fatal("PRODUCTION SECURITY VIOLATION: JWT_SECRET environment variable is required in production")
		}
		jwtSecret = random.String(32) // Generate random secret for development
		log.Printf("WARNING: Using generated JWT secret for development")
	}

	// Redis configuration
	redisAddr := tallyConfig.Queuing.RedisAddr
	redisPassword := tallyConfig.Queuing.RedisPassword
	redisDB := tallyConfig.Queuing.RedisDB

	// MinIO configuration with production validation
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		minioEndpoint = "localhost:9000" // Default MinIO endpoint
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		log.Fatal("MINIO_ACCESS_KEY environment variable is required")
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		log.Fatal("MINIO_SECRET_KEY environment variable is required")
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

	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8080"
	}

	enforceHTTPS := os.Getenv("ENFORCE_HTTPS") == "true"
	if isProduction && !enforceHTTPS {
		log.Fatal("PRODUCTION SECURITY VIOLATION: ENFORCE_HTTPS must be set to 'true' in production")
	}

	csrfSecret := os.Getenv("CSRF_SECRET")
	if csrfSecret == "" {
		if isProduction {
			log.Fatal("PRODUCTION SECURITY VIOLATION: CSRF_SECRET environment variable is required in production")
		}
		log.Printf("WARNING: CSRF_SECRET is not configured; generating ephemeral secret (tokens will reset on restart)")
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
	orderStatusHistoryRepo := repositories.NewOrderStatusHistoryRepo(pool)
	invoiceRepo := repositories.NewInvoiceRepo(pool)
	productImageRepo := repositories.NewProductImageRepo(pool)
	auditLogsRepo := repositories.NewAuditLogsRepo(pool)
	// notificationRepo := repositories.NewNotificationRepo(pool) // Removed unused
	// notificationConfigRepo := repositories.NewNotificationConfigRepo(pool) // Removed unused

	// Create cache service
	cacheSvc := caching.NewRedisCacheService(redisAddr, redisPassword, redisDB)

	// Create database optimizer
	dbOptimizer := common.NewDBOptimizer(pool)

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
	// Create subscription middleware service for feature enforcement
	subscriptionMiddleware := services.NewSubscriptionMiddlewareService(pool)

	// Create analytics service
	analyticsSvc := analytics.NewAnalyticsService(orderRepo, invoiceRepo, inventoryRepo, productRepo, cacheSvc, pool)

	// Create RBAC service with caching for improved performance
	rbacService := services.NewRBACServiceWithCache(userRoleRepo, rolePermissionRepo, permissionRepo, cacheSvc)

	// RBAC middleware
	rbacMiddleware := middleware.NewRBACMiddleware(rbacService)

	// Create user service
	userService := services.NewUserService(userRepo)

	// Create auth service
	authService := services.NewAuthService(
		cacheSvc,
		jwtSecret,
		3600,
		86400, // 1 hour access, 24 hour refresh
		userRepo,
		tenantRepo,
		roleRepo,
		userRoleRepo,
		rolePermissionRepo,
		permissionRepo,
		frontendURL,
		backendURL,
	)

	// Create product service
	productSvc := services.NewProductService(productRepo, inventoryRepo, categoryRepo, productImageRepo, minioSvc, cacheSvc)

	// Create batch service
	batchRepo := repositories.NewBatchRepository(pool)
	batchSvc := services.NewBatchService(batchRepo, productRepo)
	batchHandler := handlers.NewBatchHandler(batchSvc)

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
	userHandlers := handlers.NewUserHandlers(userRepo, tenantRepo, rbacMiddleware, userService)
	tenantHandlers := handlers.NewTenantHandlers(tenantService, rbacMiddleware)
	categoryHandlers := handlers.NewCategoryHandlers(categoryRepo, rbacMiddleware)
	warehouseHandlers := handlers.NewWarehouseHandlers(
		services.NewWarehouseService(warehouseRepo, subscriptionMiddleware),
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
	logger := common.GetGlobalLogger()
	inventoryAdapter := repositories.NewInventoryAdapter(inventoryRepo)
	inventoryService := services.NewInventoryService(inventoryAdapter, logger)
	auditLogsService := services.NewAuditLogsService(auditLogsRepo)

	orderSvc := services.NewOrderService(pool, orderRepo, inventoryRepo, inventoryService, orderStatusHistoryRepo, logger)

	invoiceSvc := services.NewInvoiceService(invoiceRepo, orderRepo, analyticsSvc, pool, tenantService, productSvc, supplierService, distributorService)
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
		tenantService,
	)

	tallyExporter := jobs.NewTallyExporter(invoiceRepo, orderRepo, productRepo, tallyConfig)
	tallyImporter := jobs.NewTallyImporter(orderRepo, invoiceRepo, tallyConfig)

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
	subscriptionHandlers := handlers.NewSubscriptionHandlers(subscriptionService, subscriptionMiddleware, rbacMiddleware)
	webhookHandlers := handlers.NewWebhookHandlers(subscriptionService, razorpayService, razorpayWebhookSecret, rbacMiddleware)

	notificationHandlers := handlers.NewNotificationHandlers(notificationService)

	// Create notification template and alert rule repositories and services
	notificationTemplateRepo := repositories.NewNotificationTemplateRepo(pool)
	webhookSubscriptionRepo := repositories.NewWebhookSubscriptionRepo(pool)
	alertRuleRepo := repositories.NewAlertRuleRepo(pool)

	notificationTemplateService := services.NewNotificationTemplateService(notificationTemplateRepo)
	webhookSubscriptionService := services.NewWebhookSubscriptionService(webhookSubscriptionRepo, logger)
	alertRuleService := services.NewAlertRuleService(alertRuleRepo, logger)

	// Device token repository for push notifications
	deviceTokenRepo := repositories.NewDeviceTokenRepository(pool)

	// Notification delivery repository & service (enhanced delivery pipeline)
	notificationDeliveryRepo := repositories.NewNotificationDeliveryRepo(pool)
	deliverySvc := services.NewNotificationDeliveryService(
		notificationDeliveryRepo,
		notificationTemplateService,
		webhookSubscriptionService,
		alertRuleService,
		common.GetGlobalLogger(),
		notificationService,
		userRepo,
		deviceTokenRepo,
		nil,   // Firebase app - configure when Firebase credentials are available
		false, // FCM enabled - set to true when Firebase is configured
	)

	// Register event handlers for critical errors and search usage
	common.RegisterEventHandler("critical_system_error", func(ctx context.Context, eventType string, data map[string]interface{}) error {
		tenantID, _ := common.GetTenantIDFromContext(ctx)
		var userIDPtr *uuid.UUID
		if uid, ok := common.GetUserIDFromContext(ctx); ok {
			userIDPtr = &uid
		}
		ev := &models.NotificationEvent{
			Type:      "critical_system_error",
			TenantID:  tenantID,
			UserID:    userIDPtr,
			Data:      data,
			Timestamp: time.Now(),
		}
		return deliverySvc.ProcessEvent(ctx, ev)
	})

	common.RegisterEventHandler("search_performed", func(ctx context.Context, eventType string, data map[string]interface{}) error {
		tenantID, _ := common.GetTenantIDFromContext(ctx)
		entityType, _ := data["entity_type"].(string)
		searchTerm, _ := data["search_term"].(string)
		filterCount, _ := data["filter_count"].(int)
		resultCount, _ := data["result_count"].(int)
		responseTimeMs, _ := data["response_time_ms"].(int64)
		var userID uuid.UUID
		if uid, ok := common.GetUserIDFromContext(ctx); ok {
			userID = uid
		}
		return analyticsSvc.RecordSearchUsage(ctx, tenantID, entityType, searchTerm, filterCount, resultCount, userID, responseTimeMs)
	})

	notificationTemplateHandlers := handlers.NewNotificationTemplateHandlers(notificationTemplateService, rbacMiddleware)
	alertRuleHandlers := handlers.NewAlertRuleHandlers(alertRuleService, rbacMiddleware)

	// Webhook test handlers (env-driven config)
	webhookTestCfg := config.LoadWebhookTestConfigFromEnv()
	webhookTestHandlers := handlers.NewWebhookTestHandlers(
		cacheSvc,
		webhookSubscriptionService,
		services.NewWebhookTestService(),
		webhookTestCfg,
	)

	// Create role and permission handlers
	roleHandlers := handlers.NewRoleHandlers(roleRepo, permissionRepo, rolePermissionRepo, rbacMiddleware)

	auditLogsHandlers := handlers.NewAuditLogsHandlers(auditLogsService, rbacMiddleware)

	// Create analytics handlers
	analyticsHandlers := handlers.NewAnalyticsHandlers(analyticsSvc, rbacMiddleware)
	securityHandlers := handlers.NewSecurityHandlers(csrfManager)
	dbOptimizationHandlers := handlers.NewDBOptimizationHandlers(dbOptimizer, rbacMiddleware)

	// Create Asynq client
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer asynqClient.Close()

	// Create Asynq inspector
	asynqInspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	// Create tally handlers
	tallyHandlers := handlers.NewTallyHandlers(tallyExporter, tallyImporter, asynqClient)

	// Create job handlers (must be created after asynqInspector)
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
		asynqInspector,
		asynqClient,
	)

	// Initialize background job scheduler
	jobScheduler := background.NewJobScheduler(
		analyticsSvc,
		cacheSvc,
		inventoryRepo,
		orderRepo,
		tenantRepo,
		productRepo,
		userRepo,
		notificationService,
	)

	// Start the job scheduler
	if err := jobScheduler.Start(); err != nil {
		log.Fatalf("Failed to start job scheduler: %v", err)
	}
	defer jobScheduler.Stop()

	// Schedule periodic materialized view refresh (every 5 minutes)
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			if err := dbOptimizer.RefreshMaterializedViews(ctx); err != nil {
				log.Printf("Failed to refresh materialized views: %v", err)
			} else {
				log.Println("Materialized views refreshed successfully")
			}
			cancel()
		}
	}()

	// Schedule periodic database maintenance (daily at 2 AM)
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
			time.Sleep(time.Until(next))

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
			tables := []string{"products", "inventory", "orders", "invoices", "audit_logs"}
			if err := dbOptimizer.VacuumAnalyze(ctx, tables); err != nil {
				log.Printf("Failed to vacuum analyze: %v", err)
			} else {
				log.Println("Database maintenance completed successfully")
			}
			cancel()
		}
	}()

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

	// Performance middleware
	perfMiddleware := middleware.NewPerformanceMiddleware()

	// Rate limiting middleware
	rateLimiter := middleware.NewRateLimiter(rate.Limit(10), 20)

	// Global middleware (order matters for performance)
	e.Use(echoMiddleware.Recover())
	e.Use(perfMiddleware.Gzip())                    // Enable gzip compression
	e.Use(perfMiddleware.Timeout(30 * time.Second)) // Request timeout
	e.Use(perfMiddleware.BodyLimit("10M"))          // Limit request body size
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.CORSWithConfig(echoMiddleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://localhost:3001", "https://app.agromart.com"},
		AllowMethods: []string{echo.GET, echo.PUT, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-CSRF-Token"},
	}))
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
	v1.Use(rateLimiter.Middleware())

	// Documentation routes (no auth required)
	v1.GET("/docs/guide", handlers.DocumentationGuideHandler)
	v1.GET("/docs/spec", handlers.DocumentationSpecHandler)

	// Security utilities
	v1.GET("/security/csrf", securityHandlers.GetCSRFToken)

	// Webhook routes (no auth required, verified by signature)
	webhooks := v1.Group("/webhooks")
	webhooks.POST("/razorpay", webhookHandlers.RazorpayWebhook)

	// Authentication routes (no JWT required for signup/login)
	auth := v1.Group("/auth")
	auth.POST("/signup", authHandlers.Signup)
	auth.POST("/login", authHandlers.Login)
	auth.POST("/refresh", authHandlers.Refresh)
	auth.POST("/password/forgot", authHandlers.ForgotPassword)
	auth.POST("/password/reset", authHandlers.ResetPassword)
	auth.POST("/verify", authHandlers.VerifyEmail)

	// Google Auth routes
	auth.GET("/google/login", authHandlers.GoogleLogin)
	auth.GET("/google/callback", authHandlers.GoogleCallback)
	auth.POST("/google/complete-signup", authHandlers.CompleteGoogleSignup)

	// 2FA routes
	auth.POST("/2fa/generate", authHandlers.Generate2FA)
	auth.POST("/2fa/enable", authHandlers.Enable2FA)
	auth.POST("/2fa/disable", authHandlers.Disable2FA)
	auth.POST("/2fa/verify", authHandlers.Verify2FA)

	// Protected routes (require JWT and RBAC)
	protected := v1.Group("")
	protected.Use(middleware.JWTMiddleware(userRepo, jwtSecret))

	// Protected auth routes
	protected.POST("/auth/logout", authHandlers.Logout)

	// User routes
	protected.GET("/me", authHandlers.Me)
	protected.GET("/users", userHandlers.ListUsers, rbacMiddleware.RequirePermission("users:list||read_users"))
	protected.GET("/users/:id", userHandlers.GetUser, rbacMiddleware.RequirePermission("users:read||read_users"))
	protected.POST("/users", userHandlers.CreateUser, rbacMiddleware.RequirePermission("users:create||create_users"))
	protected.PUT("/users/:id", userHandlers.UpdateUser, rbacMiddleware.RequirePermission("users:update||update_users"))
	protected.PUT("/users/me", userHandlers.UpdateUserProfile)
	protected.DELETE("/users/:id", userHandlers.DeleteUser, rbacMiddleware.RequirePermission("users:delete||delete_users"))

	// Tenant routes (restricted to system admins)
	// SECURITY: These routes manage the multi-tenant system itself and must be restricted
	protected.GET("/tenants", tenantHandlers.ListTenants, rbacMiddleware.RequirePermission("tenants:list||system:admin"))
	protected.POST("/tenants", tenantHandlers.CreateTenant, rbacMiddleware.RequirePermission("tenants:create||system:admin"))
	protected.GET("/tenants/:id", tenantHandlers.GetTenant, rbacMiddleware.RequirePermission("tenants:read||system:admin"))
	protected.PUT("/tenants/:id", tenantHandlers.UpdateTenant, rbacMiddleware.RequirePermission("tenants:update||system:admin"))
	protected.DELETE("/tenants/:id", tenantHandlers.DeleteTenant, rbacMiddleware.RequirePermission("tenants:delete||system:admin"))
	// Tenant settings routes (accessible by tenant admins)
	protected.GET("/tenant/settings", tenantHandlers.GetTenantSettings)
	protected.PUT("/tenant/settings", tenantHandlers.UpdateTenantSettings, rbacMiddleware.RequirePermission("tenant:manage_settings"))

	// Category routes
	protected.GET("/categories", categoryHandlers.ListCategories, rbacMiddleware.RequirePermission("categories:list||read_products"))
	protected.POST("/categories", categoryHandlers.CreateCategory, rbacMiddleware.RequirePermission("categories:create||create_products"))
	protected.GET("/categories/:id", categoryHandlers.GetCategory, rbacMiddleware.RequirePermission("categories:read||read_products"))
	protected.PUT("/categories/:id", categoryHandlers.UpdateCategory, rbacMiddleware.RequirePermission("categories:update||update_products"))
	protected.DELETE("/categories/:id", categoryHandlers.DeleteCategory, rbacMiddleware.RequirePermission("categories:delete||delete_products"))

	// Product routes
	protected.GET("/products", productHandlers.ListProducts, rbacMiddleware.RequirePermission("products:list||read_products"))
	protected.POST("/products", productHandlers.CreateProduct, rbacMiddleware.RequirePermission("products:create||create_products"))
	protected.GET("/products/:id", productHandlers.GetProduct, rbacMiddleware.RequirePermission("products:read||read_products"))
	protected.PUT("/products/:id", productHandlers.UpdateProduct, rbacMiddleware.RequirePermission("products:update||update_products"))
	protected.DELETE("/products/:id", productHandlers.DeleteProduct, rbacMiddleware.RequirePermission("products:delete||delete_products"))
	protected.GET("/products/search", productHandlers.SearchProducts, rbacMiddleware.RequirePermission("products:search||read_products"))
	protected.POST("/products/bulk/update", productHandlers.BulkUpdateProducts, rbacMiddleware.RequirePermission("products:bulk_update||update_products"))
	protected.POST("/products/bulk/create", productHandlers.BulkCreateProducts, rbacMiddleware.RequirePermission("products:bulk_create||create_products"))
	protected.POST("/products/bulk-price-update", productHandlers.BulkPriceUpdate, rbacMiddleware.RequirePermission("products:bulk_price_update||update_products"))

	// Batch routes
	protected.POST("/products/:productId/batches", batchHandler.CreateBatch)
	protected.GET("/products/:productId/batches", batchHandler.GetBatchesByProduct)
	protected.GET("/batches/:id", batchHandler.GetBatch)
	protected.PUT("/batches/:id", batchHandler.UpdateBatch)

	// Product image routes
	protected.POST("/products/:id/images", productHandlers.UploadProductImage)
	protected.GET("/products/:id/images", productHandlers.GetProductImages)
	protected.GET("/products/:id/images/:imageId/url", productHandlers.GetProductImageURL)
	protected.DELETE("/products/:id/images/:imageId", productHandlers.DeleteProductImage)

	// Warehouse routes
	protected.GET("/warehouses", warehouseHandlers.ListWarehouses, rbacMiddleware.RequirePermission("warehouses:list||read_products"))
	protected.POST("/warehouses", warehouseHandlers.CreateWarehouse, rbacMiddleware.RequirePermission("warehouses:create||create_products"))
	protected.GET("/warehouses/:id", warehouseHandlers.GetWarehouse, rbacMiddleware.RequirePermission("warehouses:read||read_products"))
	protected.PUT("/warehouses/:id", warehouseHandlers.UpdateWarehouse, rbacMiddleware.RequirePermission("warehouses:update||update_products"))
	protected.DELETE("/warehouses/:id", warehouseHandlers.DeleteWarehouse, rbacMiddleware.RequirePermission("warehouses:delete||delete_products"))

	// Distributor routes
	protected.GET("/distributors", distributorHandlers.ListDistributors, rbacMiddleware.RequirePermission("distributors:list||read_products"))
	protected.POST("/distributors", distributorHandlers.CreateDistributor, rbacMiddleware.RequirePermission("distributors:create||create_products"))
	protected.GET("/distributors/:id", distributorHandlers.GetDistributor, rbacMiddleware.RequirePermission("distributors:read||read_products"))
	protected.PUT("/distributors/:id", distributorHandlers.UpdateDistributor, rbacMiddleware.RequirePermission("distributors:update||update_products"))
	protected.DELETE("/distributors/:id", distributorHandlers.DeleteDistributor, rbacMiddleware.RequirePermission("distributors:delete||delete_products"))

	// Supplier routes
	protected.GET("/suppliers", supplierHandlers.ListSuppliers, rbacMiddleware.RequirePermission("suppliers:list||read_products"))
	protected.POST("/suppliers", supplierHandlers.CreateSupplier, rbacMiddleware.RequirePermission("suppliers:create||create_products"))
	protected.GET("/suppliers/:id", supplierHandlers.GetSupplier, rbacMiddleware.RequirePermission("suppliers:read||read_products"))
	protected.PUT("/suppliers/:id", supplierHandlers.UpdateSupplier, rbacMiddleware.RequirePermission("suppliers:update||update_products"))
	protected.DELETE("/suppliers/:id", supplierHandlers.DeleteSupplier, rbacMiddleware.RequirePermission("suppliers:delete||delete_products"))

	// Inventory routes
	protected.GET("/inventory", inventoryHandlers.ListInventories, rbacMiddleware.RequirePermission("inventory:list||read_products"))
	protected.POST("/inventory", inventoryHandlers.CreateInventory, rbacMiddleware.RequirePermission("inventory:create||create_products"))
	protected.POST("/inventory/adjust", inventoryHandlers.AdjustStock, rbacMiddleware.RequirePermission("inventory:adjust||update_products"))
	protected.GET("/inventory/:id", inventoryHandlers.GetInventory, rbacMiddleware.RequirePermission("inventory:read||read_products"))
	protected.GET("/inventory/:id/history", inventoryHandlers.GetInventoryHistory, rbacMiddleware.RequirePermission("inventory:read||read_products"))
	protected.PUT("/inventory/:id", inventoryHandlers.UpdateInventory, rbacMiddleware.RequirePermission("inventory:update||update_products"))
	protected.DELETE("/inventory/:id", inventoryHandlers.DeleteInventory, rbacMiddleware.RequirePermission("inventory:delete||delete_products"))
	protected.GET("/inventory/search", inventoryHandlers.SearchInventories, rbacMiddleware.RequirePermission("inventory:search||read_products"))
	// Start Asynq server in a goroutine
	go func() {
		if err := asynqSrv.Start(mux); err != nil {
			log.Fatalf("Could not start Asynq server: %v", err)
		}
		log.Println("Asynq server started")
	}()
	defer asynqSrv.Shutdown()

	// Periodic retry for failed notification deliveries with context cancellation and backoff
	go func() {
		const maxRetries = 5
		const baseDelay = 1 * time.Minute
		const maxDelay = 30 * time.Minute

		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			defer cancel()

			tenants, err := tenantRepo.List(ctx, 1000, 0)
			if err != nil {
				logger.ErrorWithContext(ctx, "Retry deliveries: failed to list tenants", err, nil)
				continue
			}

			for _, t := range tenants {
				if t.Status != "active" {
					continue
				}

				retryCount := 0
				delay := baseDelay

				for retryCount < maxRetries {
					if err := deliverySvc.RetryFailedDeliveries(ctx, t.ID); err != nil {
						retryCount++
						if retryCount < maxRetries {
							logger.WarnWithContext(ctx, "Retry failed, backing off",
								map[string]interface{}{
									"tenant_id":   t.ID.String(),
									"retry_count": retryCount,
									"delay":       delay.String(),
									"error":       err.Error(),
								})

							select {
							case <-time.After(delay):
								delay *= 2 // Exponential backoff
								if delay > maxDelay {
									delay = maxDelay
								}
								// Add jitter
								jitter := time.Duration(float64(delay) * 0.1)
								delay += time.Duration(time.Now().UnixNano() % int64(jitter))
							case <-ctx.Done():
								return
							}
							continue
						}

						logger.ErrorWithContext(ctx, "Exhausted retries for tenant notification delivery",
							errors.New("max retries exceeded"), map[string]interface{}{
								"tenant_id":   t.ID.String(),
								"max_retries": maxRetries,
								"final_error": err.Error(),
							})
					}
					break // Success, exit retry loop
				}
			}
		}
	}()

	// Order routes
	protected.GET("/orders", orderHandlers.GetOrders, rbacMiddleware.RequirePermission("orders:list||read_products"))
	protected.POST("/orders", orderHandlers.CreateOrder, rbacMiddleware.RequirePermission("orders:create||create_products"))
	protected.GET("/orders/:id", orderHandlers.GetOrder, rbacMiddleware.RequirePermission("orders:read||read_products"))
	protected.PUT("/orders/:id", orderHandlers.UpdateOrder, rbacMiddleware.RequirePermission("orders:update||update_products"))
	protected.DELETE("/orders/:id", orderHandlers.DeleteOrder, rbacMiddleware.RequirePermission("orders:delete||delete_products"))
	protected.POST("/orders/:id/approve", orderHandlers.ApproveOrder, rbacMiddleware.RequirePermission("orders:approve||update_products"))
	protected.POST("/orders/:id/process", orderHandlers.ProcessOrder, rbacMiddleware.RequirePermission("orders:process||update_products"))
	protected.POST("/orders/:id/receive", orderHandlers.ReceiveOrder, rbacMiddleware.RequirePermission("orders:receive||update_products"))
	protected.POST("/orders/:id/ship", orderHandlers.ShipOrder, rbacMiddleware.RequirePermission("orders:ship||update_products"))
	protected.POST("/orders/:id/deliver", orderHandlers.DeliverOrder, rbacMiddleware.RequirePermission("orders:deliver||update_products"))
	protected.POST("/orders/:id/cancel", orderHandlers.CancelOrder, rbacMiddleware.RequirePermission("orders:cancel||update_products"))

	// Invoice routes
	protected.GET("/invoices", invoiceHandlers.ListInvoices, rbacMiddleware.RequirePermission("invoices:list||read_products"))
	protected.POST("/invoices", invoiceHandlers.CreateInvoice, rbacMiddleware.RequirePermission("invoices:create||create_products"))
	protected.POST("/invoices/bulk-create", invoiceHandlers.BulkCreateInvoices, rbacMiddleware.RequirePermission("invoices:bulk_create||create_products"))
	protected.GET("/invoices/:id", invoiceHandlers.GetInvoice, rbacMiddleware.RequirePermission("invoices:read||read_products"))
	protected.PUT("/invoices/:id", invoiceHandlers.UpdateInvoice, rbacMiddleware.RequirePermission("invoices:update||update_products"))
	protected.PUT("/invoices/:id/status", invoiceHandlers.UpdateInvoiceStatus, rbacMiddleware.RequirePermission("invoices:update_status||update_products"))
	protected.GET("/invoices/unpaid", invoiceHandlers.GetUnpaidInvoices, rbacMiddleware.RequirePermission("invoices:list||read_products"))
	protected.POST("/invoices/:id/generate-pdf", invoiceHandlers.GenerateInvoicePDF, rbacMiddleware.RequirePermission("invoices:generate_pdf||read_products"))
	protected.DELETE("/invoices/:id", invoiceHandlers.DeleteInvoice, rbacMiddleware.RequirePermission("invoices:delete||delete_products"))

	// Tally routes
	protected.POST("/api/tally/export", tallyHandlers.ExportTallyData)
	protected.POST("/api/tally/import", tallyHandlers.ImportTallyData)

	// Subscription routes
	protected.GET("/subscriptions", subscriptionHandlers.ListSubscriptions)
	protected.POST("/subscriptions", subscriptionHandlers.CreateSubscription)
	protected.GET("/subscriptions/plans", subscriptionHandlers.GetAvailablePlans)
	protected.GET("/subscriptions/limits", subscriptionHandlers.GetSubscriptionLimits)
	protected.GET("/subscriptions/usage", subscriptionHandlers.GetSubscriptionUsage)
	protected.POST("/subscriptions/usage/refresh", subscriptionHandlers.RefreshUsageTracking)
	protected.GET("/subscriptions/:id", subscriptionHandlers.GetSubscriptionByID)
	protected.PUT("/subscriptions/:id", subscriptionHandlers.UpdateSubscriptionPlan)
	protected.POST("/subscriptions/:id/cancel", subscriptionHandlers.CancelSubscription)
	protected.POST("/subscriptions/:id/pause", subscriptionHandlers.PauseSubscription)
	protected.POST("/subscriptions/:id/resume", subscriptionHandlers.ResumeSubscription)
	protected.DELETE("/subscriptions/:id", subscriptionHandlers.DeleteSubscription)

	// Notification routes
	protected.POST("/notifications/send", notificationHandlers.SendNotification)
	protected.GET("/notifications", notificationHandlers.ListNotifications)
	protected.GET("/notifications/:id", notificationHandlers.GetNotification)
	protected.PUT("/notifications/:id/read", notificationHandlers.MarkNotificationAsRead)
	protected.PUT("/notifications/mark-all-read", notificationHandlers.MarkAllNotificationsAsRead)
	protected.PUT("/notifications/:id/archive", notificationHandlers.ArchiveNotification)
	protected.DELETE("/notifications/:id", notificationHandlers.DeleteNotification)
	protected.GET("/webhooks/subscriptions", notificationHandlers.ListWebhookSubscriptions)
	protected.POST("/webhooks/subscriptions", notificationHandlers.CreateWebhookSubscription)
	protected.PUT("/webhooks/subscriptions/:id", notificationHandlers.UpdateWebhookSubscription)
	protected.DELETE("/webhooks/subscriptions/:id", notificationHandlers.DeleteWebhookSubscription)

	// Test webhook route (protected + RBAC OR check webhooks:test||notifications:manage)
	protected.POST("/webhooks/test",
		webhookTestHandlers.TestWebhook,
		rbacMiddleware.RequirePermission("webhooks:test||notifications:manage"),
	)

	// Notification template routes
	protected.GET("/notification-templates", notificationTemplateHandlers.ListTemplates)
	protected.POST("/notification-templates", notificationTemplateHandlers.CreateTemplate)
	protected.GET("/notification-templates/:id", notificationTemplateHandlers.GetTemplate)
	protected.PUT("/notification-templates/:id", notificationTemplateHandlers.UpdateTemplate)
	protected.DELETE("/notification-templates/:id", notificationTemplateHandlers.DeleteTemplate)
	protected.POST("/notification-templates/:id/test", notificationTemplateHandlers.TestTemplate)

	// Alert rule routes
	protected.GET("/alert-rules", alertRuleHandlers.ListAlertRules)
	protected.POST("/alert-rules", alertRuleHandlers.CreateAlertRule)
	protected.GET("/alert-rules/:id", alertRuleHandlers.GetAlertRule)
	protected.PUT("/alert-rules/:id", alertRuleHandlers.UpdateAlertRule)
	protected.DELETE("/alert-rules/:id", alertRuleHandlers.DeleteAlertRule)
	protected.POST("/alert-rules/:id/test", alertRuleHandlers.TestAlertRule)

	// Role and permission management routes
	protected.GET("/roles", roleHandlers.ListRoles)
	protected.POST("/roles", roleHandlers.CreateRole)
	protected.GET("/roles/:id", roleHandlers.GetRole)
	protected.PUT("/roles/:id", roleHandlers.UpdateRole)
	protected.DELETE("/roles/:id", roleHandlers.DeleteRole)
	protected.GET("/roles/:id/permissions", roleHandlers.GetRolePermissions)
	protected.POST("/roles/:id/permissions", roleHandlers.AssignPermissionsToRole)
	protected.DELETE("/roles/:id/permissions/:permissionId", roleHandlers.RemovePermissionFromRole)
	protected.GET("/roles/:id/users", roleHandlers.GetRoleUsers)
	protected.GET("/permissions", roleHandlers.ListPermissions)

	// User role management routes
	protected.GET("/users/:id/roles", roleHandlers.GetUserRoles)
	protected.POST("/users/:id/roles", roleHandlers.AssignRolesToUser)
	protected.DELETE("/users/:id/roles/:roleId", roleHandlers.RemoveRoleFromUser)

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

	// Database optimization routes (admin only)
	dbOptimization := protected.Group("/db-optimization")
	dbOptimization.Use(rbacMiddleware.RequirePermission("system:admin"))
	dbOptimization.POST("/refresh-views", dbOptimizationHandlers.RefreshMaterializedViews)
	dbOptimization.GET("/slow-queries", dbOptimizationHandlers.GetSlowQueries)
	dbOptimization.GET("/unused-indexes", dbOptimizationHandlers.GetUnusedIndexes)
	dbOptimization.GET("/table-sizes", dbOptimizationHandlers.GetTableSizes)
	dbOptimization.GET("/cache-hit-ratio", dbOptimizationHandlers.GetCacheHitRatio)
	dbOptimization.GET("/stats", dbOptimizationHandlers.GetDatabaseStats)
	dbOptimization.POST("/vacuum", dbOptimizationHandlers.VacuumAnalyze)

	// WebSocket route
	wsHub := handlers.NewWebSocketHub()
	go wsHub.Run()
	wsHandlers := handlers.NewWebSocketHandlers(wsHub)
	protected.GET("/ws", wsHandlers.HandleWebSocket)

	// Export routes
	exportHandlers := handlers.NewExportHandlers()
	protected.POST("/export/excel", exportHandlers.ExportToExcel)

	// Search routes
	searchHandlers := handlers.NewSearchHandlers(pool)
	protected.POST("/search", searchHandlers.UnifiedSearch)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Validate that the port is a number.
	if _, err := strconv.Atoi(port); err != nil {
		log.Fatalf("Invalid port %s: %v", port, err)
	}

	log.Printf("🚀 Agromart2 server v%s starting on port %s", version, port)
	log.Printf("Database connected: %t", databaseURL != "") // Don't log the actual URL for security

	e.Logger.Fatal(e.Start(":" + port))
}

// verifySchema checks if all required tables exist in the database.
func verifySchema(pool *pgxpool.Pool) error {
	requiredTables := []string{
		"users",
		"tenants",
		"orders",
		"order_items",
		"products",
	}

	log.Println("Verifying database schema...")
	for _, table := range requiredTables {
		var exists bool
		query := `
			SELECT EXISTS (
				SELECT FROM information_schema.tables
				WHERE table_schema = 'public'
				AND table_name = $1
			)`
		err := pool.QueryRow(context.Background(), query, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check table %s: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("required table '%s' does not exist in database", table)
		}
		log.Printf("✓ Table '%s' exists", table)
	}

	log.Println("Database schema verification completed successfully")
	return nil
}
