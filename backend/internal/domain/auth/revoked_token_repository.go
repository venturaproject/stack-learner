package auth

import (
	"context"
	"time"
)

type RevokedTokenRepository interface {
	Revoke(ctx context.Context, jti string, exp time.Time) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
	CleanupExpired(ctx context.Context) (int64, error)
}
