//go:build integration

// Integration tests for the persistence layer. They need a real Postgres.
//
//	TEST_DATABASE_DSN='host=localhost port=5444 user=postgres password=postgres dbname=boilerplate_learner_test sslmode=disable' \
//	    go test -tags=integration ./internal/infrastructure/persistence/...
//
// If no database is reachable the tests are skipped, not failed.
package persistence

import (
	"context"
	"errors"
	"os"
	"testing"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	domaintelf "boilerplate-learner/internal/domain/telefonos"
	"boilerplate-learner/internal/infrastructure/events"
	"boilerplate-learner/internal/infrastructure/persistence/models"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_DSN")
	if dsn == "" {
		// A dedicated database: these tests DELETE rows on cleanup and must
		// never touch the dev/seeded `boilerplate_learner` database.
		dsn = "host=localhost port=5444 user=postgres password=postgres dbname=boilerplate_learner_test sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Skipf("no database: %v", err)
	}
	if err := db.AutoMigrate(&models.Telefono{}, &models.OutboxMessage{}); err != nil {
		t.Skipf("cannot migrate: %v", err)
	}
	t.Cleanup(func() {
		db.Exec("DELETE FROM actuaciones_telefono")
		db.Exec("DELETE FROM telefonos")
		db.Exec("DELETE FROM outbox_messages")
	})
	db.Exec("DELETE FROM telefonos")
	db.Exec("DELETE FROM outbox_messages")
	return db
}

func TestTelefonoRepository_SaveFindDelete(t *testing.T) {
	db := testDB(t)
	repo := NewTelefonoRepository(db)
	ctx := context.Background()

	num := "600000001"
	tel := &domaintelf.Telefono{Numero: &num, Activo: true}
	if err := repo.Save(ctx, tel); err != nil {
		t.Fatalf("save: %v", err)
	}
	if tel.ID == 0 {
		t.Fatal("Save did not set ID")
	}

	got, err := repo.FindByID(ctx, tel.ID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if got.Numero == nil || *got.Numero != num {
		t.Fatalf("round-trip mismatch: %+v", got)
	}

	if err := repo.Delete(ctx, tel.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByID(ctx, tel.ID); !errors.Is(err, domaintelf.ErrNotFound) {
		t.Fatalf("want ErrNotFound after delete, got %v", err)
	}
}

func TestOutbox_CommitPersistsBoth(t *testing.T) {
	db := testDB(t)
	tx := NewTxManager(db)
	repo := NewTelefonoRepository(db)
	pub := NewOutboxWriter(db)
	ctx := context.Background()

	err := tx.WithinTx(ctx, func(ctx context.Context) error {
		tel := &domaintelf.Telefono{Activo: true}
		if err := repo.Save(ctx, tel); err != nil {
			return err
		}
		return pub.Publish(ctx, domaintelf.NewTelefonoCreated(tel.ID))
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}

	var telCount, outboxCount int64
	db.Model(&models.Telefono{}).Count(&telCount)
	db.Model(&models.OutboxMessage{}).Where("event_name = ?", "telefono.created").Count(&outboxCount)
	if telCount != 1 || outboxCount != 1 {
		t.Fatalf("want 1 telefono + 1 outbox row, got %d / %d", telCount, outboxCount)
	}
}

func TestOutbox_RollbackDiscardsBoth(t *testing.T) {
	db := testDB(t)
	tx := NewTxManager(db)
	repo := NewTelefonoRepository(db)
	pub := NewOutboxWriter(db)
	ctx := context.Background()

	boom := errors.New("boom")
	err := tx.WithinTx(ctx, func(ctx context.Context) error {
		tel := &domaintelf.Telefono{Activo: true}
		if err := repo.Save(ctx, tel); err != nil {
			return err
		}
		if err := pub.Publish(ctx, domaintelf.NewTelefonoCreated(tel.ID)); err != nil {
			return err
		}
		return boom // force rollback
	})
	if !errors.Is(err, boom) {
		t.Fatalf("want boom, got %v", err)
	}

	var telCount, outboxCount int64
	db.Model(&models.Telefono{}).Count(&telCount)
	db.Model(&models.OutboxMessage{}).Count(&outboxCount)
	if telCount != 0 || outboxCount != 0 {
		t.Fatalf("rollback leaked rows: %d telefonos, %d outbox", telCount, outboxCount)
	}
}

func TestOutboxDispatcher_DrainMarksProcessed(t *testing.T) {
	db := testDB(t)
	pub := NewOutboxWriter(db)
	bus := events.NewBus(zap.NewNop())

	var delivered []string
	bus.Subscribe(func(e any) error {
		if env, ok := e.(OutboxEnvelope); ok {
			delivered = append(delivered, env.EventName)
		}
		return nil
	})

	if err := pub.Publish(context.Background(), domaintelf.NewTelefonoCreated(1)); err != nil {
		t.Fatalf("publish: %v", err)
	}

	d := NewOutboxDispatcher(db, bus, zap.NewNop())
	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("drain: %v", err)
	}

	if len(delivered) != 1 || delivered[0] != "telefono.created" {
		t.Fatalf("dispatcher did not deliver: %+v", delivered)
	}
	var unprocessed int64
	db.Model(&models.OutboxMessage{}).Where("processed_at IS NULL").Count(&unprocessed)
	if unprocessed != 0 {
		t.Fatalf("want 0 unprocessed rows, got %d", unprocessed)
	}
}

func TestOutboxDispatcher_RetriesThenDeadLetters(t *testing.T) {
	db := testDB(t)
	pub := NewOutboxWriter(db)
	bus := events.NewBus(zap.NewNop())
	bus.Subscribe(func(any) error { return errors.New("handler down") })

	if err := pub.Publish(context.Background(), domaintelf.NewTelefonoCreated(1)); err != nil {
		t.Fatalf("publish: %v", err)
	}

	d := NewOutboxDispatcher(db, bus, zap.NewNop())
	d.maxAttempts = 3

	for i := 0; i < d.maxAttempts; i++ {
		db.Model(&models.OutboxMessage{}).Where("failed_at IS NULL").Update("next_attempt_at", "now()")
		if err := d.Drain(context.Background()); err != nil {
			t.Fatalf("drain %d: %v", i, err)
		}
	}

	var m models.OutboxMessage
	db.First(&m)
	if m.ProcessedAt != nil {
		t.Fatal("a failing message must never be marked processed")
	}
	if m.FailedAt == nil {
		t.Fatalf("message should be dead-lettered after %d attempts, attempts=%d", d.maxAttempts, m.Attempts)
	}
	if m.LastError == nil || *m.LastError == "" {
		t.Fatal("dead-lettered message should record the last error")
	}

	// A dead-lettered row is no longer picked up.
	db.Model(&models.OutboxMessage{}).Update("next_attempt_at", "now()")
	before := m.Attempts
	if err := d.Drain(context.Background()); err != nil {
		t.Fatalf("drain after dead-letter: %v", err)
	}
	db.First(&m)
	if m.Attempts != before {
		t.Fatalf("dead-lettered row was retried: attempts %d -> %d", before, m.Attempts)
	}
}
