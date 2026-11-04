package shared

// AggregateRoot is embedded (by value) in every aggregate entity. It collects
// domain events raised while executing behaviour so the application layer can
// publish them transactionally after persisting the aggregate.
//
//	type Telefono struct {
//	    shared.AggregateRoot
//	    ...
//	}
//
//	func (t *Telefono) Asignar(...) error {
//	    ...
//	    t.Record(TelefonoAsignado{...})
//	    return nil
//	}
type AggregateRoot struct {
	events []DomainEvent
}

// Record appends a domain event. Call it from aggregate behaviour, never from
// outside the aggregate.
func (a *AggregateRoot) Record(e DomainEvent) {
	a.events = append(a.events, e)
}

// PullEvents returns the recorded events and clears the buffer. The application
// layer calls this once, after a successful Save, and hands the result to the
// EventPublisher.
func (a *AggregateRoot) PullEvents() []DomainEvent {
	out := a.events
	a.events = nil
	return out
}
