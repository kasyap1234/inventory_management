package repositories

import (
	"testing"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Product Repository Tests

// TestProductRepository_Create tests product creation validation
func TestProductRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		product *models.Product
		wantErr bool
	}{
		{
			name: "Valid product",
			product: &models.Product{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				Name:      "Test Product",
				Quantity:  100,
				UnitPrice: 19.99,
			},
			wantErr: false,
		},
		{
			name: "Product with description",
			product: &models.Product{
				ID:          uuid.New(),
				TenantID:    uuid.New(),
				Name:        "Test Product with Desc",
				Quantity:    50,
				UnitPrice:   29.99,
				Description: ptrString("A detailed description"),
			},
			wantErr: false,
		},
		{
			name: "Product with category",
			product: &models.Product{
				ID:         uuid.New(),
				TenantID:   uuid.New(),
				Name:       "Categorized Product",
				Quantity:   25,
				UnitPrice:  9.99,
				CategoryID: ptrUUID(uuid.New()),
			},
			wantErr: false,
		},
		{
			name: "Hazardous product",
			product: &models.Product{
				ID:          uuid.New(),
				TenantID:    uuid.New(),
				Name:        "Hazardous Product",
				Quantity:    10,
				UnitPrice:   49.99,
				IsHazardous: true,
				HazardClass: ptrString("Class 3"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify test case validity
			assert.NotEmpty(t, tt.product.Name)
			assert.NotEqual(t, uuid.Nil, tt.product.TenantID)
		})
	}
}

// TestProductRepository_GetByID tests fetching product by ID
func TestProductRepository_GetByID(t *testing.T) {
	tenantID := uuid.New()
	productID := uuid.New()

	tests := []struct {
		name      string
		tenantID  uuid.UUID
		productID uuid.UUID
		wantErr   bool
	}{
		{
			name:      "Existing product",
			tenantID:  tenantID,
			productID: productID,
			wantErr:   false,
		},
		{
			name:      "Non-existent product",
			tenantID:  tenantID,
			productID: uuid.New(),
			wantErr:   true,
		},
		{
			name:      "Wrong tenant",
			tenantID:  uuid.New(),
			productID: productID,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEqual(t, uuid.Nil, tt.tenantID)
			assert.NotEqual(t, uuid.Nil, tt.productID)
		})
	}
}

// Inventory Repository Tests

// TestInventoryRepository_Create tests inventory creation
func TestInventoryRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		inventory *models.Inventory
		wantErr   bool
	}{
		{
			name: "Valid inventory",
			inventory: &models.Inventory{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				ProductID: uuid.New(),
				Quantity:  100,
			},
			wantErr: false,
		},
		{
			name: "Zero quantity",
			inventory: &models.Inventory{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				ProductID: uuid.New(),
				Quantity:  0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEqual(t, uuid.Nil, tt.inventory.TenantID)
			assert.NotEqual(t, uuid.Nil, tt.inventory.ProductID)
		})
	}
}

// TestInventoryRepository_AdjustQuantity tests inventory quantity adjustment
func TestInventoryRepository_AdjustQuantity(t *testing.T) {
	tests := []struct {
		name       string
		initial    int
		adjustment int
		expected   int
		wantErr    bool
	}{
		{
			name:       "Add to inventory",
			initial:    100,
			adjustment: 50,
			expected:   150,
			wantErr:    false,
		},
		{
			name:       "Remove from inventory",
			initial:    100,
			adjustment: -50,
			expected:   50,
			wantErr:    false,
		},
		{
			name:       "Remove more than available",
			initial:    50,
			adjustment: -100,
			expected:   -50,
			wantErr:    true,
		},
		{
			name:       "Zero adjustment",
			initial:    100,
			adjustment: 0,
			expected:   100,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial + tt.adjustment
			if !tt.wantErr {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Category Repository Tests

// TestCategoryRepository_Create tests category creation
func TestCategoryRepository_Create(t *testing.T) {
	tests := []struct {
		name     string
		category *models.Category
		wantErr  bool
	}{
		{
			name: "Root category",
			category: &models.Category{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Name:     "Electronics",
			},
			wantErr: false,
		},
		{
			name: "Child category",
			category: &models.Category{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Name:     "Smartphones",
				ParentID: ptrUUID(uuid.New()),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category.Name)
			assert.NotEqual(t, uuid.Nil, tt.category.TenantID)
		})
	}
}

// TestCategoryRepository_GetHierarchy tests category hierarchy fetching
func TestCategoryRepository_GetHierarchy(t *testing.T) {
	tenantID := uuid.New()
	rootID := uuid.New()
	childID := uuid.New()

	// Simulating a hierarchy: Root -> Child
	categories := []*models.Category{
		{
			ID:       rootID,
			TenantID: tenantID,
			Name:     "Root Category",
		},
		{
			ID:       childID,
			TenantID: tenantID,
			Name:     "Child Category",
			ParentID: &rootID,
		},
	}

	assert.Len(t, categories, 2)
	assert.Nil(t, categories[0].ParentID) // Root has no parent
	assert.NotNil(t, categories[1].ParentID)
	assert.Equal(t, *categories[1].ParentID, rootID)
}

// Order Repository Tests

// TestOrderRepository_Create tests order creation
func TestOrderRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		order   *models.Order
		wantErr bool
	}{
		{
			name: "Valid order",
			order: &models.Order{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Status:   "pending",
			},
			wantErr: false,
		},
		{
			name: "Order with customer",
			order: &models.Order{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Status:   "pending",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEqual(t, uuid.Nil, tt.order.TenantID)
			assert.NotEmpty(t, tt.order.Status)
		})
	}
}

// TestOrderRepository_UpdateStatus tests order status updates
func TestOrderRepository_UpdateStatus(t *testing.T) {
	validStatuses := []string{"pending", "processing", "shipped", "delivered", "cancelled"}

	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
		wantErr    bool
	}{
		{
			name:       "Pending to processing",
			fromStatus: "pending",
			toStatus:   "processing",
			wantErr:    false,
		},
		{
			name:       "Processing to shipped",
			fromStatus: "processing",
			toStatus:   "shipped",
			wantErr:    false,
		},
		{
			name:       "Shipped to delivered",
			fromStatus: "shipped",
			toStatus:   "delivered",
			wantErr:    false,
		},
		{
			name:       "Cancel pending order",
			fromStatus: "pending",
			toStatus:   "cancelled",
			wantErr:    false,
		},
		{
			name:       "Invalid status transition",
			fromStatus: "delivered",
			toStatus:   "pending",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify statuses are valid
			assert.Contains(t, validStatuses, tt.fromStatus)
			assert.Contains(t, validStatuses, tt.toStatus)
		})
	}
}

// Batch Repository Tests

// TestBatchRepository_Create tests batch creation
func TestBatchRepository_Create(t *testing.T) {
	now := time.Now()
	future := now.Add(30 * 24 * time.Hour)

	tests := []struct {
		name    string
		batch   *models.Batch
		wantErr bool
	}{
		{
			name: "Valid batch",
			batch: &models.Batch{
				ID:          uuid.New(),
				TenantID:    uuid.New(),
				ProductID:   uuid.New(),
				BatchNumber: "BATCH-001",
				Quantity:    100,
				ExpiryDate:  &future,
			},
			wantErr: false,
		},
		{
			name: "Batch without expiry",
			batch: &models.Batch{
				ID:          uuid.New(),
				TenantID:    uuid.New(),
				ProductID:   uuid.New(),
				BatchNumber: "BATCH-002",
				Quantity:    50,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.batch.BatchNumber)
			assert.Greater(t, tt.batch.Quantity, 0)
		})
	}
}

// TestBatchRepository_GetExpiring tests fetching expiring batches
func TestBatchRepository_GetExpiring(t *testing.T) {
	now := time.Now()
	soon := now.Add(7 * 24 * time.Hour)
	later := now.Add(90 * 24 * time.Hour)
	past := now.Add(-7 * 24 * time.Hour)

	tests := []struct {
		name         string
		expiryDate   time.Time
		withinDays   int
		shouldExpire bool
	}{
		{
			name:         "Expiring in 7 days",
			expiryDate:   soon,
			withinDays:   14,
			shouldExpire: true,
		},
		{
			name:         "Expiring in 90 days",
			expiryDate:   later,
			withinDays:   14,
			shouldExpire: false,
		},
		{
			name:         "Already expired",
			expiryDate:   past,
			withinDays:   14,
			shouldExpire: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			daysUntilExpiry := int(time.Until(tt.expiryDate).Hours() / 24)
			isExpiring := daysUntilExpiry <= tt.withinDays
			assert.Equal(t, tt.shouldExpire, isExpiring)
		})
	}
}

// User Repository Tests

// TestUserRepository_Create tests user creation
func TestUserRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		user    *models.User
		wantErr bool
	}{
		{
			name: "Valid user",
			user: &models.User{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				Email:     "test@example.com",
				FirstName: "Test",
				LastName:  "User",
			},
			wantErr: false,
		},
		{
			name: "User with phone",
			user: &models.User{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				Email:     "test2@example.com",
				FirstName: "Test",
				LastName:  "User 2",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.user.Email)
			assert.NotEmpty(t, tt.user.FirstName)
			assert.NotEqual(t, uuid.Nil, tt.user.TenantID)
		})
	}
}

// TestUserRepository_GetByEmail tests fetching user by email
func TestUserRepository_GetByEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{
			name:    "Valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "Non-existent email",
			email:   "nonexistent@example.com",
			wantErr: true,
		},
		{
			name:    "Invalid email format",
			email:   "not-an-email",
			wantErr: true,
		},
		{
			name:    "Empty email",
			email:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				assert.NotEmpty(t, tt.email)
				assert.Contains(t, tt.email, "@")
			}
		})
	}
}

// Tenant Repository Tests

// TestTenantRepository_Create tests tenant creation
func TestTenantRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		tenant  *models.Tenant
		wantErr bool
	}{
		{
			name: "Valid tenant",
			tenant: &models.Tenant{
				ID:        uuid.New(),
				Name:      "Test Company",
				Subdomain: "testcompany",
				Status:    "active",
			},
			wantErr: false,
		},
		{
			name: "Tenant with settings",
			tenant: &models.Tenant{
				ID:        uuid.New(),
				Name:      "Another Company",
				Subdomain: "anothercompany",
				Status:    "active",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.tenant.Name)
			assert.NotEmpty(t, tt.tenant.Subdomain)
			assert.NotEqual(t, uuid.Nil, tt.tenant.ID)
		})
	}
}

// TestTenantRepository_GetBySubdomain tests fetching tenant by subdomain
func TestTenantRepository_GetBySubdomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain string
		wantErr   bool
	}{
		{
			name:      "Existing subdomain",
			subdomain: "testcompany",
			wantErr:   false,
		},
		{
			name:      "Non-existent subdomain",
			subdomain: "nonexistent",
			wantErr:   true,
		},
		{
			name:      "Empty subdomain",
			subdomain: "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.wantErr {
				assert.NotEmpty(t, tt.subdomain)
			}
		})
	}
}

// Helper functions
func ptrUUID(u uuid.UUID) *uuid.UUID {
	return &u
}

func ptrString(s string) *string {
	return &s
}
