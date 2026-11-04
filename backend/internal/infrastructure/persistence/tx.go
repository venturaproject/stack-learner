package persistence

import (
	"context"

	"gorm.io/gorm"

	appshared "boilerplate-learner/internal/application/shared"
)

type txKey struct{}

// GormTxManager implements appshared.TxManager on top of GORM. It stores the
// active *gorm.DB transaction in the context so repositories (via connFromContext)
// automatically enlist in the same unit of work.
type GormTxManager struct {
	db *gorm.DB
}

func NewTxManager(db *gorm.DB) *GormTxManager { return &GormTxManager{db: db} }

var _ appshared.TxManager = (*GormTxManager)(nil)

func (m *GormTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		// Already inside a transaction — compose instead of nesting.
		return fn(ctx)
	}
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// connFromContext returns the transaction bound to ctx, or fallback when the
// call is not running inside TxManager.WithinTx. Every repository method should
// obtain its *gorm.DB through this helper.
func connFromContext(ctx context.Context, fallback *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return fallback.WithContext(ctx)
}
