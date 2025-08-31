package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, subscription *models.Subscription) error
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Subscription, error)
	Update(ctx context.Context, subscription *models.Subscription) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Subscription, error)
	GetByRazorpayID(ctx context.Context, tenantID uuid.UUID, razorpayID string) (*models.Subscription, error)
}

type subscriptionRepo struct {
	db *pgxpool.Pool
}

func NewSubscriptionRepo(db *pgxpool.Pool) SubscriptionRepository {
	return &subscriptionRepo{db: db}
}

func (r *subscriptionRepo) Create(ctx context.Context, subscription *models.Subscription) error {
	query := `
		INSERT INTO subscriptions (id, tenant_id, razorpay_subscription_id, plan_name, amount, currency, status, start_date, end_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, subscription.ID, subscription.TenantID, subscription.RazorpaySubscriptionID, subscription.PlanName, subscription.Amount, subscription.Currency, subscription.Status, subscription.StartDate, subscription.EndDate)
	return err
}

func (r *subscriptionRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.Subscription, error) {
	subscription := &models.Subscription{}
	query := `
		SELECT id, tenant_id, razorpay_subscription_id, plan_name, amount, currency, status, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE tenant_id = $1 AND id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, id).Scan(&subscription.ID, &subscription.TenantID, &subscription.RazorpaySubscriptionID, &subscription.PlanName, &subscription.Amount, &subscription.Currency, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (r *subscriptionRepo) GetByRazorpayID(ctx context.Context, tenantID uuid.UUID, razorpayID string) (*models.Subscription, error) {
	subscription := &models.Subscription{}
	query := `
		SELECT id, tenant_id, razorpay_subscription_id, plan_name, amount, currency, status, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE tenant_id = $1 AND razorpay_subscription_id = $2
	`
	err := r.db.QueryRow(ctx, query, tenantID, razorpayID).Scan(&subscription.ID, &subscription.TenantID, &subscription.RazorpaySubscriptionID, &subscription.PlanName, &subscription.Amount, &subscription.Currency, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.CreatedAt, &subscription.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return subscription, nil
}

func (r *subscriptionRepo) Update(ctx context.Context, subscription *models.Subscription) error {
	query := `
		UPDATE subscriptions
		SET razorpay_subscription_id = $1, plan_name = $2, amount = $3, currency = $4, status = $5, start_date = $6, end_date = $7, updated_at = NOW()
		WHERE tenant_id = $8 AND id = $9
	`
	_, err := r.db.Exec(ctx, query, subscription.RazorpaySubscriptionID, subscription.PlanName, subscription.Amount, subscription.Currency, subscription.Status, subscription.StartDate, subscription.EndDate, subscription.TenantID, subscription.ID)
	return err
}

func (r *subscriptionRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error {
	query := `DELETE FROM subscriptions WHERE tenant_id = $1 AND id = $2`
	_, err := r.db.Exec(ctx, query, tenantID, id)
	return err
}

func (r *subscriptionRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Subscription, error) {
	query := `
		SELECT id, tenant_id, razorpay_subscription_id, plan_name, amount, currency, status, start_date, end_date, created_at, updated_at
		FROM subscriptions
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subscriptions []*models.Subscription
	for rows.Next() {
		subscription := &models.Subscription{}
		if err := rows.Scan(&subscription.ID, &subscription.TenantID, &subscription.RazorpaySubscriptionID, &subscription.PlanName, &subscription.Amount, &subscription.Currency, &subscription.Status, &subscription.StartDate, &subscription.EndDate, &subscription.CreatedAt, &subscription.UpdatedAt); err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}