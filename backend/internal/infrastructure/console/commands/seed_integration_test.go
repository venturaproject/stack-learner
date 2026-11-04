//go:build integration

package commands

import (
	"errors"
	"os"
	"testing"

	"boilerplate-learner/internal/infrastructure/persistence"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestSeedIsIdempotent(t *testing.T) {
	dsn := os.Getenv("SEED_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("SEED_TEST_DATABASE_DSN must point at an isolated database")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	log := zap.NewNop()
	rollback := errors.New("rollback seed test data")
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := persistence.Migrate(tx); err != nil { return err }
		if err := RunSeed(tx, log, false); err != nil { return err }
		count := func(model any) int64 {
			t.Helper()
			var n int64
			if err := tx.Model(model).Count(&n).Error; err != nil { t.Fatal(err) }
			return n
		}
		workers, phones, devices := count(&models.Trabajador{}), count(&models.Telefono{}), count(&models.Dispositivo{})
		if workers == 0 || phones == 0 || devices == 0 { t.Fatalf("seed data missing: workers=%d phones=%d devices=%d", workers, phones, devices) }
		if err := RunSeed(tx, log, false); err != nil { return err }
		if got := count(&models.Trabajador{}); got != workers { t.Fatalf("worker count changed after repeat: %d -> %d", workers, got) }
		if got := count(&models.Telefono{}); got != phones { t.Fatalf("phone count changed after repeat: %d -> %d", phones, got) }
		if got := count(&models.Dispositivo{}); got != devices { t.Fatalf("device count changed after repeat: %d -> %d", devices, got) }
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("seed test transaction error = %v, want rollback sentinel", err)
	}
}
