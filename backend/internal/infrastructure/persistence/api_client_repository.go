package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	domain "boilerplate-learner/internal/domain/api_clients"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type ApiClientRepository struct {
	db *gorm.DB
}

func NewApiClientRepository(db *gorm.DB) *ApiClientRepository {
	return &ApiClientRepository{db: db}
}

var _ domain.Repository = (*ApiClientRepository)(nil)

func (r *ApiClientRepository) Create(ctx context.Context, client domain.ApiClient, secretHash string) error {
	scopesJSON, _ := json.Marshal(client.Scopes)
	m := models.ApiClient{
		ID:         client.ID,
		Name:       client.Name,
		ClientID:   client.ClientID,
		SecretHash: secretHash,
		Scopes:     string(scopesJSON),
		Active:     client.Active,
	}
	return connFromContext(ctx, r.db).Create(&m).Error
}

func (r *ApiClientRepository) List(ctx context.Context) ([]domain.ApiClient, error) {
	var ms []models.ApiClient
	if err := connFromContext(ctx, r.db).Order("created_at DESC").Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domain.ApiClient, 0, len(ms))
	for _, m := range ms {
		out = append(out, mapApiClient(&m))
	}
	return out, nil
}

func (r *ApiClientRepository) Revoke(ctx context.Context, id string) error {
	return connFromContext(ctx, r.db).
		Model(&models.ApiClient{}).
		Where("id = ?", id).
		Update("active", false).Error
}

func (r *ApiClientRepository) FindByClientIDAndSecret(ctx context.Context, clientID, secretHash string) (*domain.ApiClient, error) {
	var m models.ApiClient
	if err := connFromContext(ctx, r.db).
		Where("client_id = ? AND secret_hash = ?", clientID, secretHash).
		First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	c := mapApiClient(&m)
	return &c, nil
}

func (r *ApiClientRepository) FindTokenByAccessHash(ctx context.Context, accessHash string) (*domain.ApiClient, *domain.ApiClientToken, error) {
	return r.findTokenAndClient(ctx, "access_token_hash = ?", accessHash)
}

func (r *ApiClientRepository) FindTokenByRefreshHash(ctx context.Context, refreshHash string) (*domain.ApiClient, *domain.ApiClientToken, error) {
	return r.findTokenAndClient(ctx, "refresh_token_hash = ?", refreshHash)
}

func (r *ApiClientRepository) findTokenAndClient(ctx context.Context, condition, value string) (*domain.ApiClient, *domain.ApiClientToken, error) {
	var t models.ApiClientToken
	if err := connFromContext(ctx, r.db).Where(condition, value).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var c models.ApiClient
	if err := connFromContext(ctx, r.db).Where("client_id = ?", t.ClientID).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	client := mapApiClient(&c)
	token := mapApiClientToken(&t)
	return &client, &token, nil
}

func (r *ApiClientRepository) CreateToken(ctx context.Context, token domain.ApiClientToken, accessHash, refreshHash string) error {
	scopesJSON, _ := json.Marshal(token.Scopes)
	m := models.ApiClientToken{
		ID:               token.ID,
		ClientID:         token.ClientID,
		Scopes:           string(scopesJSON),
		AccessTokenHash:  accessHash,
		RefreshTokenHash: refreshHash,
		AccessExpiresAt:  token.AccessExpiresAt,
		RefreshExpiresAt: token.RefreshExpiresAt,
	}
	return connFromContext(ctx, r.db).Create(&m).Error
}

func (r *ApiClientRepository) DeleteTokensForClient(ctx context.Context, clientID string) error {
	return connFromContext(ctx, r.db).
		Where("client_id = ?", clientID).
		Delete(&models.ApiClientToken{}).Error
}

func (r *ApiClientRepository) DeleteTokenByRefreshHash(ctx context.Context, refreshHash string) error {
	return connFromContext(ctx, r.db).
		Where("refresh_token_hash = ?", refreshHash).
		Delete(&models.ApiClientToken{}).Error
}

func (r *ApiClientRepository) TouchLastUsed(ctx context.Context, clientID string) error {
	return connFromContext(ctx, r.db).
		Model(&models.ApiClient{}).
		Where("client_id = ?", clientID).
		Update("last_used_at", time.Now()).Error
}

// ── Mappers ───────────────────────────────────────────────────────────────────

func mapApiClient(m *models.ApiClient) domain.ApiClient {
	var scopes []string
	_ = json.Unmarshal([]byte(m.Scopes), &scopes)
	return domain.ApiClient{
		ID:         m.ID,
		Name:       m.Name,
		ClientID:   m.ClientID,
		Scopes:     scopes,
		Active:     m.Active,
		LastUsedAt: m.LastUsedAt,
		CreatedAt:  m.CreatedAt,
	}
}

func mapApiClientToken(m *models.ApiClientToken) domain.ApiClientToken {
	var scopes []string
	_ = json.Unmarshal([]byte(m.Scopes), &scopes)
	return domain.ApiClientToken{
		ID:               m.ID,
		ClientID:         m.ClientID,
		Scopes:           scopes,
		AccessExpiresAt:  m.AccessExpiresAt,
		RefreshExpiresAt: m.RefreshExpiresAt,
		CreatedAt:        m.CreatedAt,
	}
}
