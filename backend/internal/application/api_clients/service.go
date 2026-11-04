package apiclients

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	appshared "boilerplate-learner/internal/application/shared"
	domain "boilerplate-learner/internal/domain/api_clients"
	"boilerplate-learner/internal/shared/apperrors"

	"github.com/google/uuid"
)

const (
	accessTokenTTLSecs  = 3600    // 1 hour
	refreshTokenTTLSecs = 2592000 // 30 days
)

// ── Crypto helpers ────────────────────────────────────────────────────────────

func generateHexToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func sha256Hex(input string) string {
	h := sha256.Sum256([]byte(input))
	return hex.EncodeToString(h[:])
}

// ── DTOs ─────────────────────────────────────────────────────────────────────

type CreateRequest struct {
	Name   string
	Scopes []string
}

type CreatedResponse struct {
	Client    ApiClientDTO
	RawSecret string
}

type ApiClientDTO struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	ClientID   string   `json:"client_id"`
	Scopes     []string `json:"scopes"`
	Active     bool     `json:"active"`
	LastUsedAt *string  `json:"last_used_at"`
	CreatedAt  string   `json:"created_at"`
}

type TokenResponse struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	TokenType    string   `json:"token_type"`
	ExpiresIn    int64    `json:"expires_in"`
	Scopes       []string `json:"scopes"`
}

// ── Service ───────────────────────────────────────────────────────────────────

type Service struct {
	repo domain.Repository
	tx   appshared.TxManager
}

func NewService(repo domain.Repository, tx appshared.TxManager) *Service {
	return &Service{repo: repo, tx: tx}
}

func (s *Service) Create(ctx context.Context, req CreateRequest) (*CreatedResponse, error) {
	rawSecret, err := generateHexToken(32)
	if err != nil {
		return nil, apperrors.Internal("error generando secret")
	}

	hexID, err := generateHexToken(8)
	if err != nil {
		return nil, apperrors.Internal("error generando client ID")
	}

	client := domain.ApiClient{
		ID:        uuid.NewString(),
		Name:      req.Name,
		ClientID:  fmt.Sprintf("cli_%s", hexID),
		Scopes:    req.Scopes,
		Active:    true,
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, client, sha256Hex(rawSecret)); err != nil {
		return nil, err
	}

	return &CreatedResponse{
		Client:    toDTO(client),
		RawSecret: rawSecret,
	}, nil
}

func (s *Service) List(ctx context.Context) ([]ApiClientDTO, error) {
	clients, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	dtos := make([]ApiClientDTO, 0, len(clients))
	for _, c := range clients {
		dtos = append(dtos, toDTO(c))
	}
	return dtos, nil
}

func (s *Service) Revoke(ctx context.Context, id string) error {
	return s.repo.Revoke(ctx, id)
}

func (s *Service) IssueToken(ctx context.Context, clientID, clientSecret string) (*TokenResponse, error) {
	client, err := s.repo.FindByClientIDAndSecret(ctx, clientID, sha256Hex(clientSecret))
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, apperrors.Unauthorized("credenciales inválidas")
	}
	if !client.Active {
		return nil, apperrors.Unauthorized("API client revocado")
	}

	scopes := domain.ExpandScopes(client.Scopes)
	return s.generateTokenPair(ctx, client.ClientID, scopes)
}

func (s *Service) RefreshToken(ctx context.Context, rawRefresh string) (*TokenResponse, error) {
	hash := sha256Hex(rawRefresh)
	client, token, err := s.repo.FindTokenByRefreshHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if client == nil || token == nil {
		return nil, apperrors.Unauthorized("refresh token inválido")
	}
	if !client.Active {
		return nil, apperrors.Unauthorized("API client revocado")
	}
	if time.Now().After(token.RefreshExpiresAt) {
		return nil, apperrors.Unauthorized("refresh token expirado")
	}

	if err := s.repo.DeleteTokenByRefreshHash(ctx, hash); err != nil {
		return nil, err
	}

	scopes := domain.ExpandScopes(client.Scopes)
	return s.generateTokenPair(ctx, client.ClientID, scopes)
}

func (s *Service) ValidateAccessToken(ctx context.Context, rawToken string) (*domain.ApiClient, *domain.ApiClientToken, error) {
	client, token, err := s.repo.FindTokenByAccessHash(ctx, sha256Hex(rawToken))
	if err != nil {
		return nil, nil, err
	}
	return client, token, nil
}

func (s *Service) TouchLastUsed(ctx context.Context, clientID string) {
	_ = s.repo.TouchLastUsed(ctx, clientID)
}

func (s *Service) generateTokenPair(ctx context.Context, clientID string, scopes []string) (*TokenResponse, error) {
	rawAccess, err := generateHexToken(40)
	if err != nil {
		return nil, apperrors.Internal("error generando access token")
	}
	rawRefresh, err := generateHexToken(40)
	if err != nil {
		return nil, apperrors.Internal("error generando refresh token")
	}

	now := time.Now()
	token := domain.ApiClientToken{
		ID:               uuid.NewString(),
		ClientID:         clientID,
		Scopes:           scopes,
		AccessExpiresAt:  now.Add(accessTokenTTLSecs * time.Second),
		RefreshExpiresAt: now.Add(refreshTokenTTLSecs * time.Second),
		CreatedAt:        now,
	}

	if err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.repo.DeleteTokensForClient(ctx, clientID); err != nil {
			return err
		}
		return s.repo.CreateToken(ctx, token, sha256Hex(rawAccess), sha256Hex(rawRefresh))
	}); err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  rawAccess,
		RefreshToken: rawRefresh,
		TokenType:    "Bearer",
		ExpiresIn:    accessTokenTTLSecs,
		Scopes:       scopes,
	}, nil
}

// ── Mapper ────────────────────────────────────────────────────────────────────

func toDTO(c domain.ApiClient) ApiClientDTO {
	dto := ApiClientDTO{
		ID:        c.ID,
		Name:      c.Name,
		ClientID:  c.ClientID,
		Scopes:    domain.ExpandScopes(c.Scopes),
		Active:    c.Active,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
	if c.LastUsedAt != nil {
		s := c.LastUsedAt.Format(time.RFC3339)
		dto.LastUsedAt = &s
	}
	return dto
}
