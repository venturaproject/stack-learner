// Package shared defines the outbound ports the application layer depends on.
// Infrastructure provides the implementations; application and domain never
// import infrastructure.
package shared

import (
	"context"

	domainshared "boilerplate-learner/internal/domain/shared"
)

// TxManager runs fn inside a single database transaction. Repositories resolve
// the active transaction from the context, so any repository call made from
// within fn participates in the same unit of work. If fn returns an error (or
// panics) the transaction is rolled back.
type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// EventPublisher persists domain events for asynchronous delivery. The
// transactional-outbox implementation writes them in the transaction carried by
// ctx, so events are never published unless the originating change commits.
type EventPublisher interface {
	Publish(ctx context.Context, events ...domainshared.DomainEvent) error
}
