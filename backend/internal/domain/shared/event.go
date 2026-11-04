// Package shared holds the domain-layer building blocks reused by every
// aggregate: the domain-event contract and the AggregateRoot mix-in.
package shared

import "time"

// DomainEvent is something meaningful that happened inside an aggregate. The
// aggregate itself records events (see AggregateRoot); the application layer
// pulls and publishes them after a successful write.
type DomainEvent interface {
	// EventName is a stable identifier used for the outbox / dispatcher, e.g.
	// "telefono.asignado".
	EventName() string
	OccurredAt() time.Time
}

// BaseEvent provides the OccurredAt bookkeeping. Embed it in concrete events
// and set At (use NewBaseEvent) when the event is recorded.
type BaseEvent struct {
	At time.Time
}

func NewBaseEvent() BaseEvent { return BaseEvent{At: time.Now()} }

func (e BaseEvent) OccurredAt() time.Time { return e.At }
