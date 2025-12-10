package repositories

import (
	"context"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository interface {
	Create(ctx context.Context, payment *models.Payment) error
	UpdateStatus(ctx context.Context, tenantID uuid.UUID, razorpayOrderID, status string, paymentID *string, signature *string, paidAt *time.Time) error
	GetByRazorpayOrderID(ctx context.Context, tenantID uuid.UUID, razorpayOrderID string) (*models.Payment, error)
	MarkPaid(ctx context.Context, tenantID uuid.UUID, razorpayOrderID string, paymentID, signature *string, paidAt time.Time) error
}

type paymentRepo struct {
	db *pgxpool.Pool
}

func NewPaymentRepo(db *pgxpool.Pool) PaymentRepository {
	return &paymentRepo{db: db}
}

func (r *paymentRepo) Create(ctx context.Context, payment *models.Payment) error {
	if payment.ID == uuid.Nil {
		payment.ID = uuid.New()
	}
	query := `
		INSERT INTO payments (
			id, tenant_id, order_id, razorpay_order_id, razorpay_payment_id,
			currency, amount_paise, status, receipt, signature, created_at, updated_at, paid_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10, NOW(), NOW(), $11
		)
	`
	_, err := r.db.Exec(ctx, query,
		payment.ID,
		payment.TenantID,
		payment.OrderID,
		payment.RazorpayOrderID,
		payment.RazorpayPaymentID,
		payment.Currency,
		payment.AmountPaise,
		payment.Status,
		payment.Receipt,
		payment.Signature,
		payment.PaidAt,
	)
	return err
}

func (r *paymentRepo) UpdateStatus(ctx context.Context, tenantID uuid.UUID, razorpayOrderID, status string, paymentID *string, signature *string, paidAt *time.Time) error {
	query := `
		UPDATE payments
		SET status = $1,
			razorpay_payment_id = COALESCE($2, razorpay_payment_id),
			signature = COALESCE($3, signature),
			paid_at = COALESCE($4, paid_at),
			updated_at = NOW()
		WHERE tenant_id = $5 AND razorpay_order_id = $6
	`
	_, err := r.db.Exec(ctx, query, status, paymentID, signature, paidAt, tenantID, razorpayOrderID)
	return err
}

func (r *paymentRepo) MarkPaid(ctx context.Context, tenantID uuid.UUID, razorpayOrderID string, paymentID, signature *string, paidAt time.Time) error {
	status := "paid"
	return r.UpdateStatus(ctx, tenantID, razorpayOrderID, status, paymentID, signature, &paidAt)
}

func (r *paymentRepo) GetByRazorpayOrderID(ctx context.Context, tenantID uuid.UUID, razorpayOrderID string) (*models.Payment, error) {
	query := `
		SELECT id, tenant_id, order_id, razorpay_order_id, razorpay_payment_id, currency, amount_paise,
		       status, receipt, signature, created_at, updated_at, paid_at
		FROM payments
		WHERE tenant_id = $1 AND razorpay_order_id = $2
	`
	p := &models.Payment{}
	err := r.db.QueryRow(ctx, query, tenantID, razorpayOrderID).Scan(
		&p.ID,
		&p.TenantID,
		&p.OrderID,
		&p.RazorpayOrderID,
		&p.RazorpayPaymentID,
		&p.Currency,
		&p.AmountPaise,
		&p.Status,
		&p.Receipt,
		&p.Signature,
		&p.CreatedAt,
		&p.UpdatedAt,
		&p.PaidAt,
	)
	if err != nil {
		return nil, err
	}
	return p, nil
}
