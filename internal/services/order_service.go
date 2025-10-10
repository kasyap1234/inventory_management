package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"agromart2/internal/common"
	"agromart2/internal/models"
	"agromart2/internal/repositories"
	"github.com/google/uuid"
)

// OrderServiceInterface defines the interface for order service operations
type OrderServiceInterface interface {
	CreateOrder(ctx context.Context, tenantID uuid.UUID, order *models.Order) error
	GetOrderByID(ctx context.Context, tenantID, orderID uuid.UUID) (*models.Order, error)
	ListOrders(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error)
	UpdateOrder(ctx context.Context, tenantID uuid.UUID, order *models.Order) error
	DeleteOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	GetOrderAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error)
	SearchOrders(ctx context.Context, tenantID uuid.UUID, filter *models.OrderSearchFilter) ([]*models.Order, error)
	ApproveOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	ProcessOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	ReceiveOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	ShipOrder(ctx context.Context, tenantID, orderID uuid.UUID, expectedDelivery *time.Time) error
	DeliverOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	CancelOrder(ctx context.Context, tenantID, orderID uuid.UUID) error
	GetOrderHistory(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.Order, error)
}

// OrderFilters defines filters for order queries
type OrderFilters struct {
	Status *string
	Limit  int
	Offset int
}

type orderService struct {
	orderRepo        repositories.OrderRepository
	inventoryRepo    repositories.InventoryRepository
	inventoryService InventoryService
}

// validOrderStatusTransitions defines allowed status transitions
var validOrderStatusTransitions = map[string][]string{
	"pending":    {"approved", "cancelled"},
	"approved":   {"processing", "cancelled"},
	"processing": {"shipped", "cancelled"},
	"shipped":    {"delivered", "cancelled"},
	"delivered":  {}, // Terminal state - no further transitions
	"cancelled":  {}, // Terminal state - no further transitions
}

// NewOrderService creates a new order service instance
func NewOrderService(orderRepo repositories.OrderRepository, inventoryRepo repositories.InventoryRepository, inventoryService InventoryService) OrderServiceInterface {
	return &orderService{
		orderRepo:        orderRepo,
		inventoryRepo:    inventoryRepo,
		inventoryService: inventoryService,
	}
}

// ValidateStatusTransition validates if a status transition is allowed
func (s *orderService) ValidateStatusTransition(currentStatus, newStatus string) error {
	if currentStatus == newStatus {
		return nil // Same status is allowed (no-op)
	}

	allowedTransitions, exists := validOrderStatusTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("unknown current status: %s", currentStatus)
	}

	// Check if new status is in allowed transitions
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("invalid status transition from '%s' to '%s'. Allowed transitions: %v",
		currentStatus, newStatus, allowedTransitions)
}

// UpdateOrderStatus updates order status with validation
func (s *orderService) UpdateOrderStatus(ctx context.Context, tenantID, orderID uuid.UUID, newStatus string) error {
	// Get current order
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Validate transition
	if err := s.ValidateStatusTransition(order.Status, newStatus); err != nil {
		return err
	}

	// Update status
	order.Status = newStatus
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// CreateOrder creates a new order with enhanced security and validation
func (s *orderService) CreateOrder(ctx context.Context, tenantID uuid.UUID, order *models.Order) error {
	// Sanitize input data to prevent XSS
	if err := common.SanitizeHTMLField(order.Notes, "order notes"); err != nil {
		return common.SecureErrorMessage("sanitize order notes", err)
	}

	// Validate business rules and data integrity
	if err := common.ValidateOrderBusinessRules(order.Quantity, order.UnitPrice, order.OrderType); err != nil {
		return common.SecureErrorMessage("validate order business rules", err)
	}

	// Set default values
	if order.ID == uuid.Nil {
		order.ID = uuid.New()
	}
	if order.OrderDate.IsZero() {
		order.OrderDate = time.Now()
	}
	order.CreatedAt = time.Now()
	order.UpdatedAt = time.Now()
	if order.Status == "" {
		order.Status = "pending"
	}

	// Data integrity checks for UUID fields
	if order.ProductID == uuid.Nil {
		return common.SecureErrorMessage("validate product ID", fmt.Errorf("product ID is required"))
	}
	if order.WarehouseID == uuid.Nil {
		return common.SecureErrorMessage("validate warehouse ID", fmt.Errorf("warehouse ID is required"))
	}
	if (order.OrderType == "purchase" && order.SupplierID == nil) ||
		(order.OrderType == "sales" && order.DistributorID == nil) {
		return common.SecureErrorMessage("validate order relationships",
			fmt.Errorf("supplier/distributor relationship validation failed"))
	}

	// Business validation: Check inventory based on order type
	if order.OrderType == "sales" {
		// For sales orders, check if sufficient inventory exists
		inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, order.WarehouseID, order.ProductID)
		if err != nil {
			if errors.Is(err, repositories.ErrInventoryNotFound) {
				return common.SecureErrorMessage("inventory validation",
					fmt.Errorf("insufficient inventory available for sales order"))
			}
			return common.SecureErrorMessage("check inventory availability", err)
		}
		if inventory.Quantity < order.Quantity {
			return common.SecureErrorMessage("inventory validation",
				fmt.Errorf("insufficient inventory available for sales order"))
		}
	}
	// For purchase orders, no inventory check is needed as they add inventory to stock

	// Save the order
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return common.SecureErrorMessage("save order", err)
	}

	return nil
}

// GetOrderByID retrieves an order by ID
func (s *orderService) GetOrderByID(ctx context.Context, tenantID, orderID uuid.UUID) (*models.Order, error) {
	return s.orderRepo.GetByID(ctx, tenantID, orderID)
}

// ListOrders lists orders with pagination
func (s *orderService) ListOrders(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Order, error) {
	return s.orderRepo.List(ctx, tenantID, limit, offset)
}

// UpdateOrder updates an order with enhanced security and validation
func (s *orderService) UpdateOrder(ctx context.Context, tenantID uuid.UUID, order *models.Order) error {
	// Sanitize input data to prevent XSS
	if err := common.SanitizeHTMLField(order.Notes, "order notes"); err != nil {
		return common.SecureErrorMessage("sanitize order notes", err)
	}

	// Get existing order for validation
	existingOrder, err := s.orderRepo.GetByID(ctx, tenantID, order.ID)
	if err != nil {
		return common.SecureErrorMessage("retrieve existing order", err)
	}
	if existingOrder == nil {
		return common.SecureErrorMessage("order lookup", fmt.Errorf("order not found"))
	}

	// Preserve critical fields that shouldn't be updated
	order.CreatedAt = existingOrder.CreatedAt
	order.TenantID = existingOrder.TenantID
	order.Status = existingOrder.Status // Status should be updated through specific methods

	// Validate business rules if quantity or price is being updated
	if order.Quantity != existingOrder.Quantity || order.UnitPrice != existingOrder.UnitPrice {
		if err := common.ValidateOrderBusinessRules(order.Quantity, order.UnitPrice, order.OrderType); err != nil {
			return common.SecureErrorMessage("validate updated order business rules", err)
		}

		// Check inventory if quantity is increasing for sales orders
		if existingOrder.OrderType == "sales" && order.Quantity > existingOrder.Quantity {
			additionalQuantity := order.Quantity - existingOrder.Quantity
			inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, order.WarehouseID, order.ProductID)
			if err != nil {
				if errors.Is(err, repositories.ErrInventoryNotFound) {
					return common.SecureErrorMessage("inventory validation",
						fmt.Errorf("insufficient additional inventory"))
				}
				return common.SecureErrorMessage("check updated inventory", err)
			}
			if inventory.Quantity < additionalQuantity {
				return common.SecureErrorMessage("inventory validation",
					fmt.Errorf("insufficient additional inventory"))
			}
		}
	}

	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return common.SecureErrorMessage("update order", err)
	}

	return nil
}

// DeleteOrder deletes an order
func (s *orderService) DeleteOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	return s.orderRepo.Delete(ctx, tenantID, orderID)
}

// GetOrderAnalytics provides secure order analytics with date range validation
func (s *orderService) GetOrderAnalytics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (map[string]interface{}, error) {
	// Validate date range to prevent abuse
	if err := common.ValidateDateRange(startDate, endDate); err != nil {
		return nil, common.SecureErrorMessage("validate analytics date range", err)
	}

	// Get orders in validated date range
	orders, err := s.orderRepo.GetOrdersByTenantAndDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, common.SecureErrorMessage("retrieve order analytics data", err)
	}

	totalOrders := len(orders)
	totalValue := 0.0
	statusCounts := map[string]int{
		"pending": 0, "approved": 0, "processing": 0,
		"shipped": 0, "delivered": 0, "cancelled": 0,
	}

	for _, order := range orders {
		value, err := common.SafeMultiplyMonetary(float64(order.Quantity), order.UnitPrice)
		if err != nil {
			log.Printf("WARN: skipping order %s in analytics due to overflow: %v", order.ID, err)
			continue
		}
		totalValue += value
		statusCounts[order.Status]++
	}

	return map[string]interface{}{
		"total_orders":     totalOrders,
		"total_value":      totalValue,
		"status_breakdown": statusCounts,
		"period": map[string]string{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
	}, nil
}

// SearchOrders searches orders with secure query validation
func (s *orderService) SearchOrders(ctx context.Context, tenantID uuid.UUID, filter *models.OrderSearchFilter) ([]*models.Order, error) {
	if filter == nil {
		return nil, common.SecureErrorMessage("initialize search filter", fmt.Errorf("filter cannot be nil"))
	}

	// Sanitize search query to prevent injection through LIKE clauses
	if filter.Query != "" {
		filter.Query = common.SanitizeSearchQuery(filter.Query)
		if filter.Query == "" {
			return nil, common.SecureErrorMessage("sanitize search query", fmt.Errorf("invalid search query"))
		}
	}

	// Validate sorting parameters for safety
	if filter.SortBy != "" {
		filter.SortBy = strings.TrimSpace(filter.SortBy)
	}
	if filter.SortOrder != "" {
		filter.SortOrder = strings.ToLower(strings.TrimSpace(filter.SortOrder))
	}

	// Validate and limit pagination parameters
	var err error
	filter.Limit, filter.Offset, err = common.ValidatePaginationParams(filter.Limit, filter.Offset)
	if err != nil {
		return nil, common.SecureErrorMessage("validate pagination parameters", err)
	}

	// Validate date ranges if provided
	if filter.OrderDateFrom != nil && filter.OrderDateTo != nil {
		if err := common.ValidateDateRange(*filter.OrderDateFrom, *filter.OrderDateTo); err != nil {
			return nil, common.SecureErrorMessage("validate date range", err)
		}
	}
	if filter.ExpectedDeliveryAfter != nil && filter.ExpectedDeliveryBefore != nil {
		if err := common.ValidateDateRange(*filter.ExpectedDeliveryAfter, *filter.ExpectedDeliveryBefore); err != nil {
			return nil, common.SecureErrorMessage("validate delivery date range", err)
		}
	}

	orders, err := s.orderRepo.AdvancedSearch(ctx, tenantID, filter)
	if err != nil {
		return nil, common.SecureErrorMessage("perform order search", err)
	}

	return orders, nil
}

// ApproveOrder changes order status to approved with validation
func (s *orderService) ApproveOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	// Get the order
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Validate status transition
	if err := s.ValidateStatusTransition(order.Status, "approved"); err != nil {
		return err
	}

	order.Status = "approved"
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// ProcessOrder changes order status to processing and reserves inventory with security checks
func (s *orderService) ProcessOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return common.SecureErrorMessage("retrieve order for processing", err)
	}
	if order == nil {
		return common.SecureErrorMessage("order lookup", fmt.Errorf("order not found"))
	}

	// Validate status transition
	if err := s.ValidateStatusTransition(order.Status, "processing"); err != nil {
		return common.SecureErrorMessage("validate order status for processing", err)
	}

	// Additional validation: ensure data integrity
	if order.Quantity <= 0 || order.UnitPrice <= 0 {
		return common.SecureErrorMessage("validate order data", fmt.Errorf("invalid order data"))
	}

	// Reserve inventory with additional validation
	inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, order.WarehouseID, order.ProductID)
	if err != nil {
		if errors.Is(err, repositories.ErrInventoryNotFound) {
			return common.SecureErrorMessage("inventory validation", fmt.Errorf("insufficient inventory"))
		}
		return common.SecureErrorMessage("retrieve inventory for processing", err)
	}
	if inventory.Quantity < order.Quantity {
		return common.SecureErrorMessage("inventory validation", fmt.Errorf("insufficient inventory"))
	}

	// Calculate new quantity with overflow protection
	newQuantity := inventory.Quantity - order.Quantity
	if newQuantity < 0 {
		return common.SecureErrorMessage("inventory calculation", fmt.Errorf("negative inventory calculation"))
	}

	inventory.Quantity = newQuantity
	inventory.LastUpdated = time.Now()

	if err := s.inventoryRepo.Update(ctx, inventory); err != nil {
		return common.SecureErrorMessage("update inventory for order processing", err)
	}

	order.Status = "processing"
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return common.SecureErrorMessage("update order status", err)
	}

	return nil
}

// ReceiveOrder handles order receipt for purchase orders
func (s *orderService) ReceiveOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	if order.OrderType != "purchase" {
		return fmt.Errorf("receive operation only valid for purchase orders")
	}

	// Validate status transition (for purchase orders, processing -> delivered)
	if err := s.ValidateStatusTransition(order.Status, "delivered"); err != nil {
		return err
	}

	// Add quantity to inventory using AdjustStock (handles existing or new)
	err = s.inventoryService.AdjustStock(ctx, tenantID, order.WarehouseID, order.ProductID, order.Quantity)
	if err != nil {
		return fmt.Errorf("failed to update inventory: %w", err)
	}

	order.Status = "delivered"
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// ShipOrder changes status to shipped
func (s *orderService) ShipOrder(ctx context.Context, tenantID, orderID uuid.UUID, expectedDelivery *time.Time) error {
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Validate status transition
	if err := s.ValidateStatusTransition(order.Status, "shipped"); err != nil {
		return err
	}

	order.Status = "shipped"
	if expectedDelivery != nil {
		order.ExpectedDelivery = expectedDelivery
	}
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// DeliverOrder changes status to delivered
func (s *orderService) DeliverOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return fmt.Errorf("order not found")
	}

	// Validate status transition
	if err := s.ValidateStatusTransition(order.Status, "delivered"); err != nil {
		return err
	}

	order.Status = "delivered"
	order.UpdatedAt = time.Now()

	return s.orderRepo.Update(ctx, order)
}

// CancelOrder cancels an order and restores inventory if needed with secure validation
func (s *orderService) CancelOrder(ctx context.Context, tenantID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return common.SecureErrorMessage("retrieve order for cancellation", err)
	}
	if order == nil {
		return common.SecureErrorMessage("order lookup", fmt.Errorf("order not found"))
	}

	// Validate status transition (any status can be cancelled except terminal states)
	if err := s.ValidateStatusTransition(order.Status, "cancelled"); err != nil {
		return common.SecureErrorMessage("validate cancellation eligibility", err)
	}

	// Restore inventory if order was processing with validation
	if order.Status == "processing" || order.Status == "approved" {
		inventory, err := s.inventoryRepo.GetByWarehouseAndProduct(ctx, tenantID, order.WarehouseID, order.ProductID)
		if err != nil {
			if !errors.Is(err, repositories.ErrInventoryNotFound) {
				return common.SecureErrorMessage("retrieve inventory for cancellation", err)
			}
		} else {
			newQuantity := inventory.Quantity + order.Quantity
			if newQuantity < inventory.Quantity {
				return common.SecureErrorMessage("inventory restoration", fmt.Errorf("inventory would overflow"))
			}
			inventory.Quantity = newQuantity
			inventory.LastUpdated = time.Now()
			if updateErr := s.inventoryRepo.Update(ctx, inventory); updateErr != nil {
				return common.SecureErrorMessage("restore inventory for cancellation", updateErr)
			}
		}
	}

	order.Status = "cancelled"
	order.UpdatedAt = time.Now()

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return common.SecureErrorMessage("update order status for cancellation", err)
	}

	return nil
}

// GetOrderHistory returns order state changes (simplified implementation)
func (s *orderService) GetOrderHistory(ctx context.Context, tenantID, orderID uuid.UUID) ([]*models.Order, error) {
	// For now, just return the current order state
	order, err := s.orderRepo.GetByID(ctx, tenantID, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}
	if order == nil {
		return []*models.Order{}, nil
	}

	return []*models.Order{order}, nil
}
