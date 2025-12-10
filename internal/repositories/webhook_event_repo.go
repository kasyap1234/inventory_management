package repositories

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookEventRepository interface {
	AlreadyProcessed(ctx context.Context, provider, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, provider, eventID, signature string, payload interface{}) error
}

type webhookEventRepo struct {
	db *pgxpool.Pool
}

func NewWebhookEventRepo(db *pgxpool.Pool) WebhookEventRepository {
	return &webhookEventRepo{db: db}
}

func (r *webhookEventRepo) AlreadyProcessed(ctx context.Context, provider, eventID string) (bool, error) {
	query := `SELECT 1 FROM webhook_events WHERE provider = $1 AND event_id = $2`
	row := r.db.QueryRow(ctx, query, provider, eventID)
	var exists int
	err := row.Scan(&exists)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (r *webhookEventRepo) MarkProcessed(ctx context.Context, provider, eventID, signature string, payload interface{}) error {
	var payloadJSON []byte
	if payload != nil {
		if data, err := json.Marshal(payload); err == nil {
			payloadJSON = data
		}
	}
	query := `
		INSERT INTO webhook_events (id, provider, event_id, signature, payload, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		ON CONFLICT (provider, event_id) DO NOTHING
	`
	_, err := r.db.Exec(ctx, query, uuid.New(), provider, eventID, signature, payloadJSON)
	return err
}
