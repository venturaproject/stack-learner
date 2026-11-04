package apiclients

import "context"

type Repository interface {
	Create(ctx context.Context, client ApiClient, secretHash string) error
	List(ctx context.Context) ([]ApiClient, error)
	Revoke(ctx context.Context, id string) error
	FindByClientIDAndSecret(ctx context.Context, clientID, secretHash string) (*ApiClient, error)
	FindTokenByAccessHash(ctx context.Context, accessHash string) (*ApiClient, *ApiClientToken, error)
	FindTokenByRefreshHash(ctx context.Context, refreshHash string) (*ApiClient, *ApiClientToken, error)
	CreateToken(ctx context.Context, token ApiClientToken, accessHash, refreshHash string) error
	DeleteTokensForClient(ctx context.Context, clientID string) error
	DeleteTokenByRefreshHash(ctx context.Context, refreshHash string) error
	TouchLastUsed(ctx context.Context, clientID string) error
}
