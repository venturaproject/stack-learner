package persistence

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	appshared "boilerplate-learner/internal/application/shared"
	domainshared "boilerplate-learner/internal/domain/shared"
	"boilerplate-learner/internal/infrastructure/events"
	"boilerplate-learner/internal/infrastructure/persistence/models"
)

// OutboxWriter implements appshared.EventPublisher by inserting events into the
// outbox_messages table using the transaction carried by the context.
type OutboxWriter struct {
	db *gorm.DB
}

func NewOutboxWriter(db *gorm.DB) *OutboxWriter { return &OutboxWriter{db: db} }

var _ appshared.EventPublisher = (*OutboxWriter)(nil)

func (w *OutboxWriter) Publish(ctx context.Context, evs ...domainshared.DomainEvent) error {
	if len(evs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]models.OutboxMessage, 0, len(evs))
	for _, e := range evs {
		payload, err := json.Marshal(e)
		if err != nil {
			return err
		}
		rows = append(rows, models.OutboxMessage{
			EventName:     e.EventName(),
			Payload:       payload,
			OccurredAt:    e.OccurredAt(),
			NextAttemptAt: now,
		})
	}
	return connFromContext(ctx, w.db).Create(&rows).Error
}

// OutboxEnvelope is what the dispatcher hands to the in-process bus. Subscribers
// switch on EventName and unmarshal Payload as needed.
type OutboxEnvelope struct {
	EventName  string
	Payload    json.RawMessage
	OccurredAt time.Time
}

// OutboxDispatcher polls outbox_messages and republishes unprocessed rows to the
// in-process bus. Single worker, at-least-once: a row is only stamped
// ProcessedAt once every subscriber handled it without error. Failed deliveries
// are retried with exponential backoff and dead-lettered after maxAttempts.
type OutboxDispatcher struct {
	db          *gorm.DB
	bus         *events.Bus
	log         *zap.Logger
	poll        time.Duration
	batch       int
	maxAttempts int
}

func NewOutboxDispatcher(db *gorm.DB, bus *events.Bus, log *zap.Logger) *OutboxDispatcher {
	// The dispatcher polls on a tight loop; keep its queries out of the SQL log.
	quiet := db.Session(&gorm.Session{Logger: logger.Discard})
	return &OutboxDispatcher{db: quiet, bus: bus, log: log, poll: 2 * time.Second, batch: 100, maxAttempts: 8}
}

// Start runs the dispatch loop until ctx is cancelled.
func (d *OutboxDispatcher) Start(ctx context.Context) {
	ticker := time.NewTicker(d.poll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := d.Drain(ctx); err != nil {
				d.log.Warn("outbox dispatch failed", zap.Error(err))
			}
		}
	}
}

// Drain dispatches a single batch of due outbox messages. Exported so tests
// (and one-off tooling) can flush synchronously.
func (d *OutboxDispatcher) Drain(ctx context.Context) error {
	var msgs []models.OutboxMessage
	if err := d.db.WithContext(ctx).
		Where("processed_at IS NULL AND failed_at IS NULL AND next_attempt_at <= ?", time.Now()).
		Order("id ASC").
		Limit(d.batch).
		Find(&msgs).Error; err != nil {
		return err
	}
	for i := range msgs {
		if err := d.dispatch(ctx, &msgs[i]); err != nil {
			return err
		}
	}
	return nil
}

func (d *OutboxDispatcher) dispatch(ctx context.Context, m *models.OutboxMessage) error {
	deliverErr := d.bus.Publish(OutboxEnvelope{
		EventName:  m.EventName,
		Payload:    m.Payload,
		OccurredAt: m.OccurredAt,
	})

	attempts := m.Attempts + 1
	updates := map[string]any{"attempts": attempts}

	switch {
	case deliverErr == nil:
		updates["processed_at"] = time.Now()
		updates["last_error"] = nil
	case attempts >= d.maxAttempts:
		updates["failed_at"] = time.Now()
		updates["last_error"] = deliverErr.Error()
		d.log.Error("outbox message dead-lettered",
			zap.Int64("id", m.ID), zap.String("event", m.EventName),
			zap.Int("attempts", attempts), zap.Error(deliverErr))
	default:
		updates["next_attempt_at"] = time.Now().Add(retryBackoff(attempts))
		updates["last_error"] = deliverErr.Error()
		d.log.Warn("outbox message delivery failed, will retry",
			zap.Int64("id", m.ID), zap.String("event", m.EventName),
			zap.Int("attempts", attempts), zap.Error(deliverErr))
	}

	return d.db.WithContext(ctx).Model(&models.OutboxMessage{}).
		Where("id = ?", m.ID).Updates(updates).Error
}

// retryBackoff grows exponentially (2s, 4s, 8s, …) capped at 5 minutes.
func retryBackoff(attempts int) time.Duration {
	if attempts > 8 {
		attempts = 8
	}
	return min(time.Duration(1<<uint(attempts))*time.Second, 5*time.Minute)
}
