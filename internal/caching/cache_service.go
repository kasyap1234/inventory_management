package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"agromart2/internal/models"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	// Product caching
	GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*models.Product, error)
	SetProduct(ctx context.Context, tenantID uuid.UUID, product *models.Product, ttl time.Duration) error
	DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error

	// Product List caching
	GetProductList(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error)
	SetProductList(ctx context.Context, tenantID uuid.UUID, limit, offset int, products []*models.Product, ttl time.Duration) error
	InvalidateProductList(ctx context.Context, tenantID uuid.UUID) error

	// Order List caching
	GetOrderList(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error)
	SetOrderList(ctx context.Context, tenantID uuid.UUID, limit, offset int, orders []*models.Order, ttl time.Duration) error
	InvalidateOrderList(ctx context.Context, tenantID uuid.UUID) error

	// Inventory caching
	GetInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error)
	SetInventory(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory, ttl time.Duration) error
	DeleteInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) error

	// Category caching
	GetCategory(ctx context.Context, tenantID, categoryID uuid.UUID) (*models.Category, error)
	SetCategory(ctx context.Context, tenantID uuid.UUID, category *models.Category, ttl time.Duration) error
	DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error

	// Analytics caching
	GetTenantAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error)
	SetTenantAnalytics(ctx context.Context, tenantID uuid.UUID, analytics map[string]interface{}, ttl time.Duration) error

	// Combined Analytics caching
	GetCombinedAnalytics(ctx context.Context, tenantID uuid.UUID) ([]byte, error)
	SetCombinedAnalytics(ctx context.Context, tenantID uuid.UUID, data []byte, ttl time.Duration) error

	// Cache invalidation
	InvalidateTenantCache(ctx context.Context, tenantID uuid.UUID) error
	InvalidateAllCache(ctx context.Context) error

	// Session management
	SetSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error
	GetSession(ctx context.Context, sessionID string) (string, error)
	DeleteSession(ctx context.Context, sessionID string) error

	// Rate limiting
	IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	IncrementRateLimit(ctx context.Context, key string, window time.Duration) error

	// Generic string operations for token management
	SetString(ctx context.Context, key string, value string, ttl time.Duration) error
	GetString(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error

	// Pattern-based deletion for token revocation
	DeleteByPattern(ctx context.Context, pattern string) error
}


type redisCacheService struct {
	client *redis.Client
}

func NewRedisCacheService(addr, password string, db int) CacheService {
	// Parse Redis URL to extract host:port if protocol is included
	parsedAddr := addr
	if strings.HasPrefix(addr, "redis://") || strings.HasPrefix(addr, "rediss://") {
		// Extract host:port from redis://host:port or rediss://host:port
		if hostPort := strings.TrimPrefix(strings.TrimPrefix(addr, "redis://"), "rediss://"); hostPort != addr {
			parsedAddr = hostPort
		}
	}

	log.Printf("Creating Redis client with address: %s", parsedAddr)

	client := redis.NewClient(&redis.Options{
		Addr:     parsedAddr,
		Password: password,
		DB:       db,

		// Connection timeouts (prevent hanging connections)
		DialTimeout:  5 * time.Second, // Maximum time to establish connection
		ReadTimeout:  3 * time.Second, // Maximum time to read response
		WriteTimeout: 3 * time.Second, // Maximum time to write request

		// Connection pool settings (optimize resource usage)
		PoolSize:     50,              // Maximum number of socket connections
		MinIdleConns: 10,              // Minimum idle connections to maintain
		PoolTimeout:  4 * time.Second, // Maximum time to wait for connection from pool

		// Retry configuration (handle transient failures)
		MaxRetries:      3,                      // Retry failed commands up to 3 times
		MinRetryBackoff: 8 * time.Millisecond,   // Minimum backoff between retries
		MaxRetryBackoff: 512 * time.Millisecond, // Maximum backoff between retries

		// Connection lifetime and cleanup
		ConnMaxIdleTime: 5 * time.Minute,  // Close idle connections after 5 minutes
		ConnMaxLifetime: 30 * time.Minute, // Recycle connections every 30 minutes
	})

	const maxAttempts = 5
	var pingErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		pingErr = client.Ping(ctx).Err()
		cancel()
		if pingErr == nil {
			log.Printf("INFO: Redis connection established on attempt %d", attempt)
			log.Printf("INFO: Redis pool configured: PoolSize=%d, MinIdleConns=%d", 50, 10)
			break
		}
		log.Printf("WARN: Redis ping attempt %d/%d failed: %v (address: %s)", attempt, maxAttempts, pingErr, parsedAddr)
		time.Sleep(time.Duration(attempt) * time.Second)
	}

	if pingErr != nil {
		log.Printf("WARN: Redis commands may fail until connection is established: %v", pingErr)
	}

	return &redisCacheService{client: client}
}

func (r *redisCacheService) GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*models.Product, error) {
	key := fmt.Sprintf("agromart:product:%s:%s", tenantID.String(), productID.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}

	var product models.Product
	if err := json.Unmarshal(data, &product); err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *redisCacheService) SetProduct(ctx context.Context, tenantID uuid.UUID, product *models.Product, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:product:%s:%s", tenantID.String(), product.ID.String())
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error {
	key := fmt.Sprintf("agromart:product:%s:%s", tenantID.String(), productID.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheService) GetInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) (*models.Inventory, error) {
	key := fmt.Sprintf("agromart:inventory:%s:%s:%s", tenantID.String(), warehouseID.String(), productID.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}

	var inventory models.Inventory
	if err := json.Unmarshal(data, &inventory); err != nil {
		return nil, err
	}
	return &inventory, nil
}

func (r *redisCacheService) SetInventory(ctx context.Context, tenantID uuid.UUID, inventory *models.Inventory, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:inventory:%s:%s:%s", tenantID.String(), inventory.WarehouseID.String(), inventory.ProductID.String())
	data, err := json.Marshal(inventory)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) DeleteInventory(ctx context.Context, tenantID, warehouseID, productID uuid.UUID) error {
	key := fmt.Sprintf("agromart:inventory:%s:%s:%s", tenantID.String(), warehouseID.String(), productID.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheService) GetCategory(ctx context.Context, tenantID, categoryID uuid.UUID) (*models.Category, error) {
	key := fmt.Sprintf("agromart:category:%s:%s", tenantID.String(), categoryID.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}

	var category models.Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *redisCacheService) SetCategory(ctx context.Context, tenantID uuid.UUID, category *models.Category, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:category:%s:%s", tenantID.String(), category.ID.String())
	data, err := json.Marshal(category)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error {
	key := fmt.Sprintf("agromart:category:%s:%s", tenantID.String(), categoryID.String())
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheService) GetTenantAnalytics(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	key := fmt.Sprintf("agromart:analytics:%s:dashboard", tenantID.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}

	var analytics map[string]interface{}
	if err := json.Unmarshal(data, &analytics); err != nil {
		return nil, err
	}
	return analytics, nil
}

func (r *redisCacheService) SetTenantAnalytics(ctx context.Context, tenantID uuid.UUID, analytics map[string]interface{}, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:analytics:%s:dashboard", tenantID.String())
	data, err := json.Marshal(analytics)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) InvalidateTenantCache(ctx context.Context, tenantID uuid.UUID) error {
	pattern := fmt.Sprintf("agromart:*:%s:*", tenantID.String())
	// Use SCAN instead of KEYS to avoid blocking Redis on large key sets
	return r.DeleteByPattern(ctx, pattern)
}

func (r *redisCacheService) InvalidateAllCache(ctx context.Context) error {
	pattern := "agromart:*"
	// Use SCAN instead of KEYS to avoid blocking Redis on large key sets
	return r.DeleteByPattern(ctx, pattern)
}

func (r *redisCacheService) SetSession(ctx context.Context, sessionID, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:session:%s", sessionID)
	return r.client.Set(ctx, key, userID, ttl).Err()
}

func (r *redisCacheService) GetSession(ctx context.Context, sessionID string) (string, error) {
	key := fmt.Sprintf("agromart:session:%s", sessionID)
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // not found
		}
		return "", err
	}
	return val, nil
}

func (r *redisCacheService) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("agromart:session:%s", sessionID)
	return r.client.Del(ctx, key).Err()
}

func (r *redisCacheService) IsRateLimited(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	cacheKey := fmt.Sprintf("agromart:ratelimit:%s", key)
	count, err := r.client.Incr(ctx, cacheKey).Result()
	if err != nil {
		return true, err
	}

	// Set expiry on first request
	if count == 1 {
		r.client.Expire(ctx, cacheKey, window)
	}

	return count > int64(limit), nil
}

func (r *redisCacheService) IncrementRateLimit(ctx context.Context, key string, window time.Duration) error {
	cacheKey := fmt.Sprintf("agromart:ratelimit:%s", key)
	_, err := r.client.Incr(ctx, cacheKey).Result()
	if err != nil {
		return err
	}
	// Set expiry if not already set (only once)
	r.client.Expire(ctx, cacheKey, window)
	return nil
}

func (r *redisCacheService) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *redisCacheService) GetString(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // cache miss
		}
		return "", err
	}
	return val, nil
}

func (r *redisCacheService) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// DeleteByPattern deletes all keys matching the given pattern
// Uses SCAN instead of KEYS for production safety (non-blocking)
func (r *redisCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	var deletedCount int64

	for {
		// Use SCAN to iterate through keys matching the pattern
		keys, nextCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys with pattern %s: %w", pattern, err)
		}

		// Delete found keys in batches
		if len(keys) > 0 {
			deleted, err := r.client.Del(ctx, keys...).Result()
			if err != nil {
				log.Printf("WARN: Failed to delete some keys matching pattern %s: %v", pattern, err)
			}
			deletedCount += deleted
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if deletedCount > 0 {
		log.Printf("Deleted %d keys matching pattern: %s", deletedCount, pattern)
	}

	return nil
}

// Product List caching implementations
func (r *redisCacheService) GetProductList(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Product, error) {
	key := fmt.Sprintf("agromart:products:list:%s:%d:%d", tenantID.String(), limit, offset)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}

	var products []*models.Product
	if err := json.Unmarshal(data, &products); err != nil {
		return nil, err
	}
	return products, nil
}

func (r *redisCacheService) SetProductList(ctx context.Context, tenantID uuid.UUID, limit, offset int, products []*models.Product, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:products:list:%s:%d:%d", tenantID.String(), limit, offset)
	data, err := json.Marshal(products)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) InvalidateProductList(ctx context.Context, tenantID uuid.UUID) error {
	pattern := fmt.Sprintf("agromart:products:list:%s:*", tenantID.String())
	return r.DeleteByPattern(ctx, pattern)
}

// Order List caching implementations
func (r *redisCacheService) GetOrderList(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	key := fmt.Sprintf("agromart:orders:list:%s:%d:%d", tenantID.String(), limit, offset)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}

	var orders []*models.Order
	if err := json.Unmarshal(data, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *redisCacheService) SetOrderList(ctx context.Context, tenantID uuid.UUID, limit, offset int, orders []*models.Order, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:orders:list:%s:%d:%d", tenantID.String(), limit, offset)
	data, err := json.Marshal(orders)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) InvalidateOrderList(ctx context.Context, tenantID uuid.UUID) error {
	pattern := fmt.Sprintf("agromart:orders:list:%s:*", tenantID.String())
	return r.DeleteByPattern(ctx, pattern)
}

// Combined Analytics caching implementations
func (r *redisCacheService) GetCombinedAnalytics(ctx context.Context, tenantID uuid.UUID) ([]byte, error) {
	key := fmt.Sprintf("agromart:analytics:combined:%s", tenantID.String())
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // cache miss
		}
		return nil, err
	}
	return data, nil
}

func (r *redisCacheService) SetCombinedAnalytics(ctx context.Context, tenantID uuid.UUID, data []byte, ttl time.Duration) error {
	key := fmt.Sprintf("agromart:analytics:combined:%s", tenantID.String())
	return r.client.Set(ctx, key, data, ttl).Err()
}
