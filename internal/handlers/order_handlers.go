package handlers

import (
	"agromart2/internal/common"
	"net/http"
	"strconv"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// OrderHandlers handles HTTP requests for orders
type OrderHandlers struct {
	orderService services.OrderServiceInterface
}

// NewOrderHandlers creates a new order handlers instance
func NewOrderHandlers(orderService services.OrderServiceInterface) *OrderHandlers {
	return &OrderHandlers{
		orderService: orderService,
	}
}

// validateOrderType validates order type
func (h *OrderHandlers) validateOrderType(orderType string) error {
	if orderType != "purchase" && orderType != "sales" {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid order type")
	}
	return nil
}

// validateUUID validates UUID string
func (h *OrderHandlers) validateUUID(idStr string) (uuid.UUID, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, echo.NewHTTPError(http.StatusBadRequest, "Invalid UUID format")
	}
	return id, nil
}

// CreateOrder handles POST /orders
func (h *OrderHandlers) CreateOrder(c echo.Context) error {
	ctx := c.Request().Context()

	// Extract tenant ID from context
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		OrderType        string  `json:"order_type"`
		ProductID        string  `json:"product_id"`
		WarehouseID      string  `json:"warehouse_id"`
		Quantity         int     `json:"quantity"`
		UnitPrice        float64 `json:"unit_price"`
		ExpectedDelivery *string `json:"expected_delivery"`
		SupplierID       *string `json:"supplier_id"`
		DistributorID    *string `json:"distributor_id"`
		Notes            *string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return common.SendClientError(c, "Invalid request format")
	}

	// Validate required fields and types
	if err := common.ValidateOrderType(req.OrderType); err != nil {
		return common.SendValidationError(c, "order_type", err.Error())
	}

	productID, err := common.ValidateUUID(req.ProductID, "product_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	warehouseID, err := common.ValidateUUID(req.WarehouseID, "warehouse_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	// Set reasonable limits for quantity (max 10,000 units per order)
	if err := common.ValidatePositiveInteger(req.Quantity, "quantity", 10000); err != nil {
		return common.SendValidationError(c, "quantity", err.Error())
	}

	// Set reasonable limits for unit price (max $1,000,000)
	if err := common.ValidatePositiveFloat(req.UnitPrice, "unit_price", 1000000.0); err != nil {
		return common.SendValidationError(c, "unit_price", err.Error())
	}

	// Validate business logic
	if req.OrderType == "purchase" && (req.SupplierID == nil || common.SafeString(req.SupplierID) == "") {
		return common.SendValidationError(c, "supplier_id", "Supplier ID is required for purchase orders")
	}
	if req.OrderType == "sales" && (req.DistributorID == nil || common.SafeString(req.DistributorID) == "") {
		return common.SendValidationError(c, "distributor_id", "Distributor ID is required for sales orders")
	}

	order := &models.Order{
		ID:        uuid.New(),
		TenantID:  tenantID,
		OrderType: req.OrderType,
		ProductID: productID,
		WarehouseID: warehouseID,
		Quantity:  req.Quantity,
		UnitPrice: req.UnitPrice,
		Status:    "pending",
		Notes:     req.Notes,
	}

	if req.SupplierID != nil && common.SafeString(req.SupplierID) != "" {
		supplierID, err := common.ValidateUUID(common.SafeString(req.SupplierID), "supplier_id")
		if err != nil {
			return common.SendClientError(c, err.Error())
		}
		order.SupplierID = &supplierID
	}
	if req.DistributorID != nil && common.SafeString(req.DistributorID) != "" {
		distributorID, err := common.ValidateUUID(common.SafeString(req.DistributorID), "distributor_id")
		if err != nil {
			return common.SendClientError(c, err.Error())
		}
		order.DistributorID = &distributorID
	}
	if req.ExpectedDelivery != nil {
		expectedDate := common.SafeString(req.ExpectedDelivery)
		if expectedDate != "" {
			if err := common.ValidateDateFormat(expectedDate, "expected_delivery"); err != nil {
				return common.SendValidationError(c, "expected_delivery", err.Error())
			}
			deliveryDate, _ := time.Parse("2006-01-02", expectedDate)
			order.ExpectedDelivery = &deliveryDate
		}
	}

	if err := h.orderService.CreateOrder(ctx, tenantID, order); err != nil {
		return common.SendServerError(c, "Failed to create order: " + err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Order created successfully",
		"order":   order,
	})
}

// GetOrders handles GET /orders
func (h *OrderHandlers) GetOrders(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	limit := 10  // default
	offset := 0  // default

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	orders, err := h.orderService.ListOrders(ctx, tenantID, limit, offset)
	if err != nil {
		return common.SendServerError(c, "Failed to retrieve orders: " + err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"orders": orders,
		"limit":  limit,
		"offset": offset,
	})
}

// GetOrderByID handles GET /orders/:id
func (h *OrderHandlers) GetOrderByID(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	order, err := h.orderService.GetOrderByID(ctx, tenantID, orderID)
	if err != nil {
		return common.SendServerError(c, "Failed to retrieve order: " + err.Error())
	}
	if order == nil {
		return common.SendNotFoundError(c, "order")
	}

	return c.JSON(http.StatusOK, order)
}

// GetOrderAnalytics handles GET /orders/analytics
func (h *OrderHandlers) GetOrderAnalytics(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	startDateStr := c.QueryParam("start_date")
	endDateStr := c.QueryParam("end_date")

	var startDate, endDate time.Time

	if startDateStr == "" {
		startDate = time.Now().AddDate(-1, 0, 0) // Default to previous month
	} else {
		if err := common.ValidateDateFormat(startDateStr, "start_date"); err != nil {
			return common.SendValidationError(c, "start_date", err.Error())
		}
		startDate, _ = time.Parse("2006-01-02", startDateStr)
	}

	if endDateStr == "" {
		endDate = time.Now() // Default to today
	} else {
		if err := common.ValidateDateFormat(endDateStr, "end_date"); err != nil {
			return common.SendValidationError(c, "end_date", err.Error())
		}
		endDate, _ = time.Parse("2006-01-02", endDateStr)
	}

	analytics, err := h.orderService.GetOrderAnalytics(ctx, tenantID, startDate, endDate)
	if err != nil {
		return common.SendServerError(c, "Failed to generate order analytics: " + err.Error())
	}

	return c.JSON(http.StatusOK, analytics)
}

// SearchOrders handles GET /orders/search
func (h *OrderHandlers) SearchOrders(c echo.Context) error {
	ctx := c.Request().Context()

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var filters models.OrderSearchFilter
	limit := 10
	offset := 0

	// Parse query parameters
	status := c.QueryParam("status")
	if status != "" {
		// Validate status
		validStatuses := []string{"pending", "approved", "processing", "shipped", "delivered", "cancelled"}
		valid := false
		for _, s := range validStatuses {
			if status == s {
				valid = true
				break
			}
		}
		if valid {
			filters.Status = &status
		}
	}

	if limitParam := c.QueryParam("limit"); limitParam != "" {
		if l, err := strconv.Atoi(limitParam); err == nil && l > 0 {
			limit = l
		}
	}
	if offsetParam := c.QueryParam("offset"); offsetParam != "" {
		if o, err := strconv.Atoi(offsetParam); err == nil && o >= 0 {
			offset = o
		}
	}

	filters.Limit = limit
	filters.Offset = offset

	orders, err := h.orderService.SearchOrders(ctx, tenantID, &filters)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"orders": orders,
		"limit":  limit,
		"offset": offset,
	})
}

// ApproveOrder handles POST /orders/:id/approve
func (h *OrderHandlers) ApproveOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.orderService.ApproveOrder(ctx, tenantID, orderID); err != nil {
		return common.SendServerError(c, "Failed to approve order: " + err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order approved successfully",
	})
}

// ProcessOrder handles POST /orders/:id/process
func (h *OrderHandlers) ProcessOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.orderService.ProcessOrder(ctx, tenantID, orderID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order processed successfully",
	})
}
// ReceiveOrder handles POST /orders/:id/receive
func (h *OrderHandlers) ReceiveOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.orderService.ReceiveOrder(ctx, tenantID, orderID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order received successfully",
	})
}

// ShipOrder handles POST /orders/:id/ship
func (h *OrderHandlers) ShipOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		ExpectedDelivery *string `json:"expected_delivery"`
	}

	if err := c.Bind(&req); err != nil {
		return common.SendClientError(c, "Invalid request format")
	}

	var expectedDelivery *time.Time
	if req.ExpectedDelivery != nil && *req.ExpectedDelivery != "" {
		deliveryDate, err := time.Parse("2006-01-02", *req.ExpectedDelivery)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid expected delivery date format")
		}
		expectedDelivery = &deliveryDate
	}

	if err := h.orderService.ShipOrder(ctx, tenantID, orderID, expectedDelivery); err != nil {
		return common.SendServerError(c, "Failed to ship order: " + err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order shipped successfully",
	})
}

// DeliverOrder handles POST /orders/:id/deliver
func (h *OrderHandlers) DeliverOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.orderService.DeliverOrder(ctx, tenantID, orderID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order delivered successfully",
		"note": "Invoice will be automatically generated",
	})
}

// CancelOrder handles POST /orders/:id/cancel
func (h *OrderHandlers) CancelOrder(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	if err := h.orderService.CancelOrder(ctx, tenantID, orderID); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Order cancelled successfully",
	})
}

// GetOrderHistory handles GET /orders/:id/history
func (h *OrderHandlers) GetOrderHistory(c echo.Context) error {
	ctx := c.Request().Context()

	id := c.Param("id")
	orderID, err := common.ValidateUUID(id, "order_id")
	if err != nil {
		return common.SendClientError(c, err.Error())
	}

	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	history, err := h.orderService.GetOrderHistory(ctx, tenantID, orderID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, history)
}