package repositories

import (
	"context"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InvitationRepository interface {
	Create(ctx context.Context, invitation *models.Invitation) error
	GetByToken(ctx context.Context, token string) (*models.Invitation, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Invitation, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.InvitationStatus) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type invitationRepo struct {
	db *pgxpool.Pool
}

func NewInvitationRepo(db *pgxpool.Pool) InvitationRepository {
	return &invitationRepo{db: db}
}

func (r *invitationRepo) Create(ctx context.Context, invitation *models.Invitation) error {
	query := `
		INSERT INTO invitations (id, tenant_id, email, role_id, token, status, permissions, invited_by, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`
	_, err := r.db.Exec(ctx, query, invitation.ID, invitation.TenantID, invitation.Email, invitation.RoleID, invitation.Token, invitation.Status, invitation.Permissions, invitation.InvitedBy, invitation.ExpiresAt)
	return err
}

func (r *invitationRepo) GetByToken(ctx context.Context, token string) (*models.Invitation, error) {
	invitation := &models.Invitation{}
	query := `
		SELECT id, tenant_id, email, role_id, token, status, permissions, invited_by, expires_at, created_at, updated_at
		FROM invitations
		WHERE token = $1
	`
	err := r.db.QueryRow(ctx, query, token).Scan(&invitation.ID, &invitation.TenantID, &invitation.Email, &invitation.RoleID, &invitation.Token, &invitation.Status, &invitation.Permissions, &invitation.InvitedBy, &invitation.ExpiresAt, &invitation.CreatedAt, &invitation.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *invitationRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Invitation, error) {
	invitation := &models.Invitation{}
	query := `
		SELECT id, tenant_id, email, role_id, token, status, permissions, invited_by, expires_at, created_at, updated_at
		FROM invitations
		WHERE id = $1
	`
	err := r.db.QueryRow(ctx, query, id).Scan(&invitation.ID, &invitation.TenantID, &invitation.Email, &invitation.RoleID, &invitation.Token, &invitation.Status, &invitation.Permissions, &invitation.InvitedBy, &invitation.ExpiresAt, &invitation.CreatedAt, &invitation.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return invitation, nil
}

func (r *invitationRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status models.InvitationStatus) error {
	query := `
		UPDATE invitations
		SET status = $1, updated_at = NOW()
		WHERE id = $2
	`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *invitationRepo) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invitation, error) {
	query := `
		SELECT id, tenant_id, email, role_id, token, status, permissions, invited_by, expires_at, created_at, updated_at
		FROM invitations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invitations []*models.Invitation
	for rows.Next() {
		invitation := &models.Invitation{}
		if err := rows.Scan(&invitation.ID, &invitation.TenantID, &invitation.Email, &invitation.RoleID, &invitation.Token, &invitation.Status, &invitation.Permissions, &invitation.InvitedBy, &invitation.ExpiresAt, &invitation.CreatedAt, &invitation.UpdatedAt); err != nil {
			return nil, err
		}
		invitations = append(invitations, invitation)
	}
	return invitations, nil
}

func (r *invitationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM invitations WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
