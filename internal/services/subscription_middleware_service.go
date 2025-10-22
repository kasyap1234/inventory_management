package services

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubscriptionMiddlewareService handles subscription feature enforcement
type SubscriptionMiddlewareService interface {
	// Feature access checks
	CanAccessFeature(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error)
	
	// Resource limit checks
	CanCreateWarehouse(ctx context.Context, tenantID uuid.UUID) error
	CanCreateUser(ctx context.Context, tenantID uuid.UUID) error
	CanCreateProduct(ctx context.Context, tenantID uuid.UUID) error
	CanCreateOrder(ctx context.Context, tenantID uuid.UUID) error
	CanCreateSupplier(ctx context.Context, tenantID uuid.UUID) error
	CanCreateDistributor(ctx context.Context, tenantID uuid.UUID) error
	
	// Get limits and usage
	GetSubscriptionLimits(ctx context.Context, tenantID uuid.UUID) (*SubscriptionLimits, error)
	GetCurrentUsage(ctx context.Context, tenantID uuid.UUID) (*ResourceUsage, error)
	
	// Refresh usage tracking
	RefreshUsageTracking(ctx context.Context, tenantID uuid.UUID) error
}

type subscriptionMiddlewareService struct {
	db *pgxpool.Pool
}

// SubscriptionLimits represents the limits for a subscription plan
type SubscriptionLimits struct {
	PlanID                     string `db:"plan_id" json:"plan_id"`
	PlanName                   string `db:"plan_name" json:"plan_name"`
	MaxWarehouses              int    `db:"max_warehouses" json:"max_warehouses"`
	MaxUsers                   int    `db:"max_users" json:"max_users"`
	MaxProducts                int    `db:"max_products" json:"max_products"`
	MaxOrdersPerMonth          int    `db:"max_orders_per_month" json:"max_orders_per_month"`
	MaxSuppliers               int    `db:"max_suppliers" json:"max_suppliers"`
	MaxDistributors            int    `db:"max_distributors" json:"max_distributors"`
	AnalyticsEnabled           bool   `db:"analytics_enabled" json:"analytics_enabled"`
	AdvancedAnalyticsEnabled   bool   `db:"advanced_analytics_enabled" json:"advanced_analytics_enabled"`
	APIAccessEnabled           bool   `db:"api_access_enabled" json:"api_access_enabled"`
	CustomBrandingEnabled      bool   `db:"custom_branding_enabled" json:"custom_branding_enabled"`
	PrioritySupportEnabled     bool   `db:"priority_support_enabled" json:"priority_support_enabled"`
	MultiLocationEnabled       bool   `db:"multi_location_enabled" json:"multi_location_enabled"`
	CustomIntegrationsEnabled  bool   `db:"custom_integrations_enabled" json:"custom_integrations_enabled"`
	DedicatedAccountManager    bool   `db:"dedicated_account_manager" json:"dedicated_account_manager"`
	APIRateLimitPerMinute      int    `db:"api_rate_limit_per_minute" json:"api_rate_limit_per_minute"`
	StorageLimitGB             int    `db:"storage_limit_gb" json:"storage_limit_gb"`
}

// ResourceUsage represents current resource usage
type ResourceUsage struct {
	WarehousesCount          int     `db:"warehouses_count" json:"warehouses_count"`
	UsersCount               int     `db:"users_count" json:"users_count"`
	ProductsCount            int     `db:"products_count" json:"products_count"`
	OrdersCountCurrentMonth  int     `db:"orders_count_current_month" json:"orders_count_current_month"`
	SuppliersCount           int     `db:"suppliers_count" json:"suppliers_count"`
	DistributorsCount        int     `db:"distributors_count" json:"distributors_count"`
	StorageUsedGB            float64 `db:"storage_used_gb" json:"storage_used_gb"`
	APICallsCurrentMonth     int     `db:"api_calls_current_month" json:"api_calls_current_month"`
}

// NewSubscriptionMiddlewareService creates a new subscription middleware service
func NewSubscriptionMiddlewareService(db *pgxpool.Pool) SubscriptionMiddlewareService {
	return &subscriptionMiddlewareService{
		db: db,
	}
}

// CanAccessFeature checks if tenant can access a specific feature
func (s *subscriptionMiddlewareService) CanAccessFeature(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error) {
	var canAccess bool
	query := `SELECT can_access_feature($1, $2)`
	err := s.db.QueryRow(ctx, query, tenantID, feature).Scan(&canAccess)
	if err != nil {
		return false, fmt.Errorf("failed to check feature access: %w", err)
	}
	return canAccess, nil
}

// GetSubscriptionLimits gets the subscription limits for a tenant
func (s *subscriptionMiddlewareService) GetSubscriptionLimits(ctx context.Context, tenantID uuid.UUID) (*SubscriptionLimits, error) {
	query := `
		SELECT 
			sf.plan_id, sf.plan_name,
			sf.max_warehouses, sf.max_users, sf.max_products, 
			sf.max_orders_per_month, sf.max_suppliers, sf.max_distributors,
			sf.analytics_enabled, sf.advanced_analytics_enabled, sf.api_access_enabled,
			sf.custom_branding_enabled, sf.priority_support_enabled, sf.multi_location_enabled,
			sf.custom_integrations_enabled, sf.dedicated_account_manager,
			sf.api_rate_limit_per_minute, sf.storage_limit_gb
		FROM subscription_features sf
		INNER JOIN subscriptions s ON (
			(s.plan_name ILIKE '%basic%' AND sf.plan_id = 'basic') OR
			(s.plan_name ILIKE '%premium%' AND sf.plan_id = 'premium') OR
			(s.plan_name ILIKE '%enterprise%' AND sf.plan_id = 'enterprise')
		)
		WHERE s.tenant_id = $1 AND s.status = 'active'
		ORDER BY s.created_at DESC
		LIMIT 1
	`
	
	var limits SubscriptionLimits
	err := s.db.QueryRow(ctx, query, tenantID).Scan(
		&limits.PlanID, &limits.PlanName,
		&limits.MaxWarehouses, &limits.MaxUsers, &limits.MaxProducts,
		&limits.MaxOrdersPerMonth, &limits.MaxSuppliers, &limits.MaxDistributors,
		&limits.AnalyticsEnabled, &limits.AdvancedAnalyticsEnabled, &limits.APIAccessEnabled,
		&limits.CustomBrandingEnabled, &limits.PrioritySupportEnabled, &limits.MultiLocationEnabled,
		&limits.CustomIntegrationsEnabled, &limits.DedicatedAccountManager,
		&limits.APIRateLimitPerMinute, &limits.StorageLimitGB,
	)
	if err != nil {
		if err.Error() == "no rows in result set" {
			// No active subscription, return basic plan limits as default
			return s.getDefaultLimits(ctx)
		}
		return nil, fmt.Errorf("failed to get subscription limits: %w", err)
	}
	
	return &limits, nil
}

// getDefaultLimits returns default (basic) limits for tenants without active subscription
func (s *subscriptionMiddlewareService) getDefaultLimits(ctx context.Context) (*SubscriptionLimits, error) {
	query := `SELECT plan_id, plan_name, max_warehouses, max_users, max_products, max_orders_per_month, max_suppliers, max_distributors, analytics_enabled, advanced_analytics_enabled, api_access_enabled, custom_branding_enabled, priority_support_enabled, multi_location_enabled, custom_integrations_enabled, dedicated_account_manager, api_rate_limit_per_minute, storage_limit_gb FROM subscription_features WHERE plan_id = 'basic' LIMIT 1`
	var limits SubscriptionLimits
	err := s.db.QueryRow(ctx, query).Scan(
		&limits.PlanID, &limits.PlanName,
		&limits.MaxWarehouses, &limits.MaxUsers, &limits.MaxProducts,
		&limits.MaxOrdersPerMonth, &limits.MaxSuppliers, &limits.MaxDistributors,
		&limits.AnalyticsEnabled, &limits.AdvancedAnalyticsEnabled, &limits.APIAccessEnabled,
		&limits.CustomBrandingEnabled, &limits.PrioritySupportEnabled, &limits.MultiLocationEnabled,
		&limits.CustomIntegrationsEnabled, &limits.DedicatedAccountManager,
		&limits.APIRateLimitPerMinute, &limits.StorageLimitGB,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get default limits: %w", err)
	}
	return &limits, nil
}

// GetCurrentUsage gets the current resource usage for a tenant
func (s *subscriptionMiddlewareService) GetCurrentUsage(ctx context.Context, tenantID uuid.UUID) (*ResourceUsage, error) {
	query := `SELECT * FROM get_current_usage($1)`
	
	var usage ResourceUsage
	err := s.db.QueryRow(ctx, query, tenantID).Scan(
		&usage.WarehousesCount,
		&usage.UsersCount,
		&usage.ProductsCount,
		&usage.OrdersCountCurrentMonth,
		&usage.SuppliersCount,
		&usage.DistributorsCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get current usage: %w", err)
	}
	
	return &usage, nil
}

// CanCreateWarehouse checks if tenant can create a new warehouse
func (s *subscriptionMiddlewareService) CanCreateWarehouse(ctx context.Context, tenantID uuid.UUID) error {
	limits, err := s.GetSubscriptionLimits(ctx, tenantID)
	if err != nil {
		return err
	}
	
	// -1 means unlimited
	if limits.MaxWarehouses == -1 {
		return nil
	}
	
	usage, err := s.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if usage.WarehousesCount >= limits.MaxWarehouses {
		return fmt.Errorf("warehouse limit reached: %d/%d. Please upgrade your subscription to create more warehouses", 
			usage.WarehousesCount, limits.MaxWarehouses)
	}
	
	return nil
}

// CanCreateUser checks if tenant can create a new user
func (s *subscriptionMiddlewareService) CanCreateUser(ctx context.Context, tenantID uuid.UUID) error {
	limits, err := s.GetSubscriptionLimits(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if limits.MaxUsers == -1 {
		return nil
	}
	
	usage, err := s.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if usage.UsersCount >= limits.MaxUsers {
		return fmt.Errorf("user limit reached: %d/%d. Please upgrade your subscription to add more users", 
			usage.UsersCount, limits.MaxUsers)
	}
	
	return nil
}

// CanCreateProduct checks if tenant can create a new product
func (s *subscriptionMiddlewareService) CanCreateProduct(ctx context.Context, tenantID uuid.UUID) error {
	limits, err := s.GetSubscriptionLimits(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if limits.MaxProducts == -1 {
		return nil
	}
	
	usage, err := s.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if usage.ProductsCount >= limits.MaxProducts {
		return fmt.Errorf("product limit reached: %d/%d. Please upgrade your subscription to add more products", 
			usage.ProductsCount, limits.MaxProducts)
	}
	
	return nil
}

// CanCreateOrder checks if tenant can create a new order
func (s *subscriptionMiddlewareService) CanCreateOrder(ctx context.Context, tenantID uuid.UUID) error {
	limits, err := s.GetSubscriptionLimits(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if limits.MaxOrdersPerMonth == -1 {
		return nil
	}
	
	usage, err := s.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if usage.OrdersCountCurrentMonth >= limits.MaxOrdersPerMonth {
		return fmt.Errorf("monthly order limit reached: %d/%d. Please upgrade your subscription or wait for next billing cycle", 
			usage.OrdersCountCurrentMonth, limits.MaxOrdersPerMonth)
	}
	
	return nil
}

// CanCreateSupplier checks if tenant can create a new supplier
func (s *subscriptionMiddlewareService) CanCreateSupplier(ctx context.Context, tenantID uuid.UUID) error {
	limits, err := s.GetSubscriptionLimits(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if limits.MaxSuppliers == -1 {
		return nil
	}
	
	usage, err := s.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if usage.SuppliersCount >= limits.MaxSuppliers {
		return fmt.Errorf("supplier limit reached: %d/%d. Please upgrade your subscription to add more suppliers", 
			usage.SuppliersCount, limits.MaxSuppliers)
	}
	
	return nil
}

// CanCreateDistributor checks if tenant can create a new distributor
func (s *subscriptionMiddlewareService) CanCreateDistributor(ctx context.Context, tenantID uuid.UUID) error {
	limits, err := s.GetSubscriptionLimits(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if limits.MaxDistributors == -1 {
		return nil
	}
	
	usage, err := s.GetCurrentUsage(ctx, tenantID)
	if err != nil {
		return err
	}
	
	if usage.DistributorsCount >= limits.MaxDistributors {
		return fmt.Errorf("distributor limit reached: %d/%d. Please upgrade your subscription to add more distributors", 
			usage.DistributorsCount, limits.MaxDistributors)
	}
	
	return nil
}

// RefreshUsageTracking updates the usage tracking for a tenant
func (s *subscriptionMiddlewareService) RefreshUsageTracking(ctx context.Context, tenantID uuid.UUID) error {
	query := `
		INSERT INTO usage_tracking (
			tenant_id, period_start, period_end,
			warehouses_count, users_count, products_count,
			orders_count_current_month, suppliers_count, distributors_count
		)
		SELECT 
			$1,
			date_trunc('month', CURRENT_DATE),
			(date_trunc('month', CURRENT_DATE) + INTERVAL '1 month' - INTERVAL '1 day'),
			u.warehouses_count, u.users_count, u.products_count,
			u.orders_count_current_month, u.suppliers_count, u.distributors_count
		FROM get_current_usage($1) u
		ON CONFLICT (tenant_id, period_start) DO UPDATE SET
			warehouses_count = EXCLUDED.warehouses_count,
			users_count = EXCLUDED.users_count,
			products_count = EXCLUDED.products_count,
			orders_count_current_month = EXCLUDED.orders_count_current_month,
			suppliers_count = EXCLUDED.suppliers_count,
			distributors_count = EXCLUDED.distributors_count,
			updated_at = NOW()
	`
	
	_, err := s.db.Exec(ctx, query, tenantID)
	if err != nil {
		return fmt.Errorf("failed to refresh usage tracking: %w", err)
	}
	
	return nil
}
