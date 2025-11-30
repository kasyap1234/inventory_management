package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/oauth2"
)

// GetGoogleAuthURL generates the Google OAuth2 login URL
func (s *authService) GetGoogleAuthURL(state string) string {
	return s.googleOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// HandleGoogleCallback handles the Google OAuth2 callback
func (s *authService) HandleGoogleCallback(ctx context.Context, code string) (*models.User, string, error) {
	token, err := s.googleOAuthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, "", fmt.Errorf("failed to exchange token: %v", err)
	}

	client := s.googleOAuthConfig.Client(ctx, token)
	userInfoResp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info: %v", err)
	}
	defer userInfoResp.Body.Close()

	var googleUser struct {
		ID            string `json:"id"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		GivenName     string `json:"given_name"`
		FamilyName    string `json:"family_name"`
		Picture       string `json:"picture"`
	}

	if err := json.NewDecoder(userInfoResp.Body).Decode(&googleUser); err != nil {
		return nil, "", fmt.Errorf("failed to decode user info: %v", err)
	}

	// Check if user exists by email (global search)
	user, err := s.userRepo.GetByEmailGlobal(ctx, googleUser.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, "", fmt.Errorf("failed to check existing user: %v", err)
	}

	if user != nil {
		// User exists, link Google ID if not linked
		if user.GoogleID == nil || *user.GoogleID != googleUser.ID {
			if err := s.userRepo.UpdateGoogleID(ctx, user.TenantID, user.ID, googleUser.ID); err != nil {
				return nil, "", fmt.Errorf("failed to link google account: %v", err)
			}
			// Update local user object
			user.GoogleID = &googleUser.ID
		}
		return user, "", nil
	}

	// User does not exist, return details for completion
	// We return a temporary token containing the Google user info
	tempTokenData := fmt.Sprintf("%s:%s:%s:%s", googleUser.Email, googleUser.ID, googleUser.GivenName, googleUser.FamilyName)
	tempToken := base64.URLEncoding.EncodeToString([]byte(tempTokenData))

	return nil, tempToken, nil
}

// CompleteGoogleSignup completes the registration for a Google user
func (s *authService) CompleteGoogleSignup(ctx context.Context, email, googleID, firstName, lastName, tenantName, subdomain string) (*models.User, error) {
	// 1. Create Tenant
	tenantID, err := s.createTenant(ctx, tenantName, subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %v", err)
	}

	// 2. Create User
	userID := uuid.New()
	user := &models.User{
		ID:        userID,
		TenantID:  tenantID,
		Email:     email,
		GoogleID:  &googleID,
		FirstName: firstName,
		LastName:  lastName,
		Status:    "active", // Auto-activate for Google users
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// We need a dummy password hash since it's required by DB usually, but here we can maybe set it to empty or a random string
	// Assuming DB allows null or we set a random un-matchable hash
	randomPass, _ := s.generateSecureToken()
	hashedPassword, _ := s.HashPassword(randomPass)
	user.PasswordHash = hashedPassword

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	// 3. Assign Roles
	adminRole, _, err := s.ensureTenantDefaults(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure default roles: %v", err)
	}

	if adminRole == nil {
		return nil, fmt.Errorf("admin role not found")
	}

	newUserRole := &models.UserRole{
		UserID: userID,
		RoleID: adminRole.ID,
	}
	if err := s.userRoleRepo.Create(ctx, tenantID, newUserRole); err != nil {
		return nil, fmt.Errorf("failed to assign role: %v", err)
	}

	return user, nil
}
