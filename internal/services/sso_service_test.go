package services

import (
	"context"
	"encoding/base64"
	"testing"

	"agromart2/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type mockTenantRepo struct {
	tenant *models.Tenant
	err    error
}

func (m *mockTenantRepo) Create(ctx context.Context, tenant *models.Tenant) error { return nil }
func (m *mockTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockTenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockTenantRepo) Update(ctx context.Context, tenant *models.Tenant) error { return nil }
func (m *mockTenantRepo) Delete(ctx context.Context, id uuid.UUID) error          { return nil }
func (m *mockTenantRepo) List(ctx context.Context, limit, offset int) ([]*models.Tenant, error) {
	return nil, nil
}
func (m *mockTenantRepo) FindSettingsByTenantID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	return m.tenant, m.err
}
func (m *mockTenantRepo) UpdateSettings(ctx context.Context, tenant *models.Tenant) error { return nil }

type mockUserRepo struct {
	user *models.User
	err  error
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error { return nil }
func (m *mockUserRepo) GetByID(ctx context.Context, tenantID, id uuid.UUID) (*models.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error      { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, tenantID, id uuid.UUID) error { return nil }
func (m *mockUserRepo) List(ctx context.Context, tenantID uuid.UUID, limit, offset int) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) GetTenantIDByUserID(ctx context.Context, userID uuid.UUID) (uuid.UUID, error) {
	return uuid.Nil, nil
}
func (m *mockUserRepo) GetByEmailGlobal(ctx context.Context, email string) (*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, tenantID, userID uuid.UUID, passwordHash string) error {
	return nil
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status string) error {
	return nil
}
func (m *mockUserRepo) FindUsersByTenantID(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdateGoogleID(ctx context.Context, tenantID, userID uuid.UUID, googleID string) error {
	return nil
}
func (m *mockUserRepo) ListByStatus(ctx context.Context, tenantID uuid.UUID, status string, limit, offset int) ([]*models.User, error) {
	return nil, nil
}
func (m *mockUserRepo) IsFirstUserInTenant(ctx context.Context, tenantID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) IsPlatformAdmin(ctx context.Context, userID uuid.UUID) (bool, error) {
	return false, nil
}
func (m *mockUserRepo) SetPlatformAdmin(ctx context.Context, userID uuid.UUID, isPlatformAdmin bool) error {
	return nil
}

func TestGetAuthURLBuildsOIDCUrl(t *testing.T) {
	tenantID := uuid.New()
	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Example",
		Subdomain: "example",
		License:   "lic",
		Status:    "active",
		SSOConfig: &models.SSOConfig{
			Provider:     "oidc",
			IssuerURL:    "https://idp.example.com",
			ClientID:     "client-123",
			ClientSecret: "secret",
		},
	}

	svc := NewSSOService(&mockTenantRepo{tenant: tenant}, &mockUserRepo{}, "https://api.localhost")
	authURL, err := svc.GetAuthURL(context.Background(), tenantID.String())
	assert.NoError(t, err)
	assert.Contains(t, authURL, "https://idp.example.com/authorize")
	assert.Contains(t, authURL, "client-123")
	assert.Contains(t, authURL, "https%3A%2F%2Fapi.localhost%2Fv1%2Fauth%2Fsso%2Fcallback")
}

func TestHandleCallbackParsesIDToken(t *testing.T) {
	tenantID := uuid.New()
	user := &models.User{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    "user@example.com",
	}
	tenant := &models.Tenant{
		ID:        tenantID,
		Name:      "Example",
		Subdomain: "example",
		License:   "lic",
		Status:    "active",
		SSOConfig: &models.SSOConfig{
			Provider:     "oidc",
			IssuerURL:    "https://idp.example.com",
			ClientID:     "client-123",
			ClientSecret: "secret",
		},
	}

	svc := NewSSOService(&mockTenantRepo{tenant: tenant}, &mockUserRepo{user: user}, "https://api.localhost")

	// Inject a fake exchanger to avoid network calls
	if impl, ok := svc.(*ssoService); ok {
		impl.exchangeFn = func(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
			payload := base64.RawURLEncoding.EncodeToString([]byte(`{"email":"user@example.com"}`))
			idToken := "eyJhbGciOiJIUzI1NiJ9." + payload + ".signature"
			token := &oauth2.Token{
				AccessToken: "dummy",
			}
			return token.WithExtra(map[string]interface{}{"id_token": idToken}), nil
		}
	}

	resolvedUser, err := svc.HandleCallback(context.Background(), tenantID.String(), "auth-code")
	assert.NoError(t, err)
	assert.Equal(t, user.Email, resolvedUser.Email)
}
