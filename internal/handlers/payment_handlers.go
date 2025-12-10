package handlers

import (
	"net/http"

	"agromart2/internal/common"
	"agromart2/internal/middleware"
	"agromart2/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type PaymentHandlers struct {
	paymentService services.PaymentService
	rbacMiddleware *middleware.RBACMiddleware
	publicKey      string
}

func NewPaymentHandlers(paymentService services.PaymentService, rbacMiddleware *middleware.RBACMiddleware, publicKey string) *PaymentHandlers {
	return &PaymentHandlers{
		paymentService: paymentService,
		rbacMiddleware: rbacMiddleware,
		publicKey:      publicKey,
	}
}

// GetPaymentConfig exposes safe, client-usable config such as the public Razorpay key.
func (h *PaymentHandlers) GetPaymentConfig(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"razorpay_key_id":  h.publicKey,
		"default_currency": "INR",
	})
}

// CreateOrderPayment handles POST /payments/orders
func (h *PaymentHandlers) CreateOrderPayment(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		Amount   float64           `json:"amount"`
		Currency string            `json:"currency"`
		Receipt  string            `json:"receipt"`
		OrderID  *string           `json:"order_id"`
		Notes    map[string]string `json:"notes"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if req.Amount <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Amount must be greater than zero")
	}

	var orderIDPtr *uuid.UUID
	if req.OrderID != nil && *req.OrderID != "" {
		orderUUID, err := uuid.Parse(*req.OrderID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid order_id format")
		}
		orderIDPtr = &orderUUID
	}

	payment, razorpayOrder, err := h.paymentService.CreateOneTimePayment(ctx, tenantID, req.Amount, req.Currency, req.Receipt, orderIDPtr, req.Notes)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"payment": payment,
		"razorpay": map[string]interface{}{
			"order_id":    razorpayOrder.ID,
			"amount":      razorpayOrder.Amount,
			"currency":    razorpayOrder.Currency,
			"key_id":      h.publicKey,
			"status":      razorpayOrder.Status,
			"receipt":     razorpayOrder.Receipt,
			"amount_paid": razorpayOrder.AmountPaid,
		},
	})
}

// VerifyOrderPayment handles POST /payments/verify
func (h *PaymentHandlers) VerifyOrderPayment(c echo.Context) error {
	ctx := c.Request().Context()
	tenantID, ok := common.GetTenantIDFromContext(ctx)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "Tenant not found")
	}

	var req struct {
		RazorpayOrderID   string `json:"razorpay_order_id"`
		RazorpayPaymentID string `json:"razorpay_payment_id"`
		RazorpaySignature string `json:"razorpay_signature"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
	}
	if req.RazorpayOrderID == "" || req.RazorpayPaymentID == "" || req.RazorpaySignature == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "order_id, payment_id and signature are required")
	}

	payment, err := h.paymentService.ConfirmPayment(ctx, tenantID, req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Payment verified",
		"payment": payment,
	})
}
