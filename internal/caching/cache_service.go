package caching

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"agromart2/internal/models"
)

type CacheService interface {
	// Product caching
	GetProduct(ctx context.Context, tenantID, productID uuid.UUID) (*models.Product, error)
	SetProduct(ctx context.Context, tenantID uuid.UUID, product *models.Product, ttl time.Duration) error
	DeleteProduct(ctx context.Context, tenantID, productID uuid.UUID) error

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

	log.Printf("DEBUG: Creating Redis client with address: %s (original: %s)", parsedAddr, addr)

	client := redis.NewClient(&redis.Options{
		Addr:     parsedAddr,
		Password: password,
		DB:       db,
	})

	// Test initial connectivity
	if pingErr := client.Ping(context.Background()).Err(); pingErr != nil {
		log.Printf("WARN: Redis ping failed on initialization: %v (address: %s)", pingErr, parsedAddr)
	} else {
		log.Printf("DEBUG: Redis connection established successfully")
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
	key := fmt.Sprintf("agromart:analytics:%s", tenantID.String())
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
	key := fmt.Sprintf("agromart:analytics:%s", tenantID.String())
	data, err := json.Marshal(analytics)
	if err != nil {
		return err
	}
	return r.client.Set(ctx, key, data, ttl).Err()
}

func (r *redisCacheService) InvalidateTenantCache(ctx context.Context, tenantID uuid.UUID) error {
	pattern := fmt.Sprintf("agromart:*:%s:*", tenantID.String())
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
}

func (r *redisCacheService) InvalidateAllCache(ctx context.Context) error {
	pattern := "agromart:*"
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}
	return nil
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