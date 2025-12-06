package services

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"agromart2/internal/models"
	"agromart2/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type SSOService interface {
	GetAuthURL(ctx context.Context, tenantID string) (string, error)
	HandleCallback(ctx context.Context, tenantID string, code string) (*models.User, error)
}

type ssoService struct {
	tenantRepo   repositories.TenantRepository
	userRepo     repositories.UserRepository
	redirectBase string
	exchangeFn   func(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error)
}

func NewSSOService(tenantRepo repositories.TenantRepository, userRepo repositories.UserRepository, redirectBase string) SSOService {
	return &ssoService{
		tenantRepo:   tenantRepo,
		userRepo:     userRepo,
		redirectBase: strings.TrimRight(redirectBase, "/"),
		exchangeFn: func(ctx context.Context, cfg *oauth2.Config, code string) (*oauth2.Token, error) {
			return cfg.Exchange(ctx, code)
		},
	}
}

func (s *ssoService) GetAuthURL(ctx context.Context, tenantID string) (string, error) {
	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return "", fmt.Errorf("invalid tenant id: %w", err)
	}

	tenant, err := s.tenantRepo.GetByID(ctx, tid)
	if err != nil {
		return "", fmt.Errorf("failed to load tenant: %w", err)
	}

	if tenant.SSOConfig == nil {
		return "", errors.New("tenant does not have SSO configured")
	}
	cfg := tenant.SSOConfig
	if strings.ToLower(cfg.Provider) != "oidc" {
		return "", fmt.Errorf("sso provider %s not supported", cfg.Provider)
	}
	if cfg.IssuerURL == "" || cfg.ClientID == "" || cfg.ClientSecret == "" {
		return "", errors.New("sso configuration incomplete")
	}

	redirectURI := s.buildRedirectURI(tenantID)
	authURL, err := url.Parse(strings.TrimRight(cfg.IssuerURL, "/") + "/authorize")
	if err != nil {
		return "", fmt.Errorf("invalid issuer url: %w", err)
	}

	query := url.Values{}
	query.Set("response_type", "code")
	query.Set("client_id", cfg.ClientID)
	query.Set("redirect_uri", redirectURI)
	query.Set("scope", "openid email profile")
	query.Set("state", uuid.New().String())

	authURL.RawQuery = query.Encode()
	return authURL.String(), nil
}

func (s *ssoService) HandleCallback(ctx context.Context, tenantID string, code string) (*models.User, error) {
	if code == "" {
		return nil, errors.New("authorization code is required")
	}

	tid, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant id: %w", err)
	}

	tenant, err := s.tenantRepo.GetByID(ctx, tid)
	if err != nil {
		return nil, fmt.Errorf("failed to load tenant: %w", err)
	}

	if tenant.SSOConfig == nil {
		return nil, errors.New("tenant does not have SSO configured")
	}
	cfg := tenant.SSOConfig
	if strings.ToLower(cfg.Provider) != "oidc" {
		return nil, fmt.Errorf("sso provider %s not supported", cfg.Provider)
	}

	redirectURI := s.buildRedirectURI(tenantID)
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  strings.TrimRight(cfg.IssuerURL, "/") + "/authorize",
			TokenURL: strings.TrimRight(cfg.IssuerURL, "/") + "/token",
		},
		RedirectURL: redirectURI,
		Scopes:      []string{"openid", "email", "profile"},
	}

	token, err := s.exchangeFn(ctx, oauthCfg, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange sso code: %w", err)
	}

	claims, err := parseIDTokenClaims(token)
	if err != nil {
		return nil, err
	}

	email := ""
	if val, ok := claims["email"].(string); ok {
		email = val
	}
	if email == "" {
		if val, ok := claims["sub"].(string); ok {
			email = val
		}
	}
	if email == "" {
		return nil, errors.New("sso response missing email")
	}

	user, err := s.userRepo.GetByEmail(ctx, tid, email)
	if err != nil {
		return nil, fmt.Errorf("user not found for SSO email %s: %w", email, err)
	}

	return user, nil
}

func (s *ssoService) buildRedirectURI(tenantID string) string {
	return fmt.Sprintf("%s/v1/auth/sso/callback?tenant_id=%s", s.redirectBase, tenantID)
}

func parseIDTokenClaims(token *oauth2.Token) (map[string]interface{}, error) {
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, errors.New("id_token missing in SSO response")
	}

	claims := jwt.MapClaims{}
	parser := jwt.Parser{}
	if _, _, err := parser.ParseUnverified(rawIDToken, claims); err != nil {
		return nil, fmt.Errorf("failed to parse id_token: %w", err)
	}

	return claims, nil
}
