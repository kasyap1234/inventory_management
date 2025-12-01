package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/google/uuid"
)

type InvitationService interface {
	CreateInvitation(ctx context.Context, req *CreateInvitationRequest) (*models.Invitation, error)
	GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error)
	AcceptInvitation(ctx context.Context, token string, req *AcceptInvitationRequest) (*models.User, error)
	RevokeInvitation(ctx context.Context, id uuid.UUID) error
	ListInvitations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invitation, error)
}

type invitationService struct {
	invitationRepo      repositories.InvitationRepository
	userRepo            repositories.UserRepository
	userRoleRepo        repositories.UserRoleRepository
	notificationService NotificationService
	authService         AuthService
	frontendBaseURL     string
}

func NewInvitationService(
	invitationRepo repositories.InvitationRepository,
	userRepo repositories.UserRepository,
	userRoleRepo repositories.UserRoleRepository,
	notificationService NotificationService,
	authService AuthService,
	frontendBaseURL string,
) InvitationService {
	return &invitationService{
		invitationRepo:      invitationRepo,
		userRepo:            userRepo,
		userRoleRepo:        userRoleRepo,
		notificationService: notificationService,
		authService:         authService,
		frontendBaseURL:     frontendBaseURL,
	}
}

type CreateInvitationRequest struct {
	TenantID    uuid.UUID
	Email       string    `json:"email"`
	RoleID      uuid.UUID `json:"role_id"`
	Permissions []string  `json:"permissions"`
}

type AcceptInvitationRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `json:"password"`
}

func (s *invitationService) CreateInvitation(ctx context.Context, req *CreateInvitationRequest) (*models.Invitation, error) {
	// Check if user already exists in the tenant
	existingUser, err := s.userRepo.GetByEmail(ctx, req.TenantID, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.New("user already exists in this tenant")
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(tokenBytes)

	invitation := &models.Invitation{
		ID:          uuid.New(),
		TenantID:    req.TenantID,
		Email:       req.Email,
		RoleID:      req.RoleID,
		Token:       token,
		Status:      models.InvitationStatusPending,
		Permissions: req.Permissions,
		ExpiresAt:   time.Now().Add(48 * time.Hour), // 48 hours expiration
	}

	if err := s.invitationRepo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	// Send invitation email
	inviteURL := fmt.Sprintf("%s/accept-invite?token=%s", s.frontendBaseURL, token)
	emailBody := fmt.Sprintf(
		"<p>You have been invited to join Agromart. Click the link below to accept the invitation and set up your account.</p><p><a href=\"%s\">Accept Invitation</a></p><p>This link will expire in 48 hours.</p>",
		inviteURL,
	)

	if err := s.notificationService.SendEmail(ctx, req.TenantID, req.Email, "You have been invited to Agromart", emailBody); err != nil {
		// Log error but don't fail the request
		fmt.Printf("Failed to send invitation email: %v\n", err)
	}

	return invitation, nil
}

func (s *invitationService) GetInvitationByToken(ctx context.Context, token string) (*models.Invitation, error) {
	invitation, err := s.invitationRepo.GetByToken(ctx, token)
	if err != nil {
		return nil, errors.New("invalid invitation token")
	}

	if invitation.Status != models.InvitationStatusPending {
		return nil, errors.New("invitation is no longer valid")
	}

	if time.Now().After(invitation.ExpiresAt) {
		s.invitationRepo.UpdateStatus(ctx, invitation.ID, models.InvitationStatusExpired)
		return nil, errors.New("invitation has expired")
	}

	return invitation, nil
}

func (s *invitationService) AcceptInvitation(ctx context.Context, token string, req *AcceptInvitationRequest) (*models.User, error) {
	invitation, err := s.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Create user
	hashedPassword, err := s.authService.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New(),
		TenantID:     invitation.TenantID,
		Email:        invitation.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Status:       "active", // Auto-activate invited users
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Assign role
	userRole := &models.UserRole{
		UserID: user.ID,
		RoleID: invitation.RoleID,
	}
	if err := s.userRoleRepo.Create(ctx, invitation.TenantID, userRole); err != nil {
		// Log error, but user is created. Should probably rollback or handle better.
		fmt.Printf("Failed to assign role to user %s: %v\n", user.ID, err)
	}

	// Update invitation status
	if err := s.invitationRepo.UpdateStatus(ctx, invitation.ID, models.InvitationStatusAccepted); err != nil {
		fmt.Printf("Failed to update invitation status: %v\n", err)
	}

	return user, nil
}

func (s *invitationService) RevokeInvitation(ctx context.Context, id uuid.UUID) error {
	return s.invitationRepo.UpdateStatus(ctx, id, models.InvitationStatusRevoked)
}

func (s *invitationService) ListInvitations(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.Invitation, error) {
	return s.invitationRepo.ListByTenant(ctx, tenantID, limit, offset)
}
