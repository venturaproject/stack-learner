package persistence

import (
	"context"
	"time"

	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

type RevokedTokenRepository struct {
	db *gorm.DB
}

func NewRevokedTokenRepository(db *gorm.DB) *RevokedTokenRepository {
	return &RevokedTokenRepository{db: db}
}

var _ domainauth.RevokedTokenRepository = (*RevokedTokenRepository)(nil)

func (r *RevokedTokenRepository) Revoke(ctx context.Context, jti string, exp time.Time) error {
	m := models.RevokedAccessToken{JTI: jti, Exp: exp}
	return connFromContext(ctx, r.db).
		Where(models.RevokedAccessToken{JTI: jti}).
		FirstOrCreate(&m).Error
}

func (r *RevokedTokenRepository) IsRevoked(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := connFromContext(ctx, r.db).
		Model(&models.RevokedAccessToken{}).
		Where("jti = ?", jti).
		Count(&count).Error
	return count > 0, err
}

func (r *RevokedTokenRepository) CleanupExpired(ctx context.Context) (int64, error) {
	result := connFromContext(ctx, r.db).
		Where("exp <= ?", time.Now()).
		Delete(&models.RevokedAccessToken{})
	return result.RowsAffected, result.Error
}
