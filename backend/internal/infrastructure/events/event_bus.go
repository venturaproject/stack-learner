package events

import (
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// Handler consumes an event. Returning an error — or panicking — tells the
// caller the event was not handled; the outbox dispatcher uses that signal to
// retry the message later instead of marking it delivered.
type Handler func(event any) error

type Bus struct {
	mu       sync.RWMutex
	handlers []Handler
	log      *zap.Logger
}

func NewBus(log *zap.Logger) *Bus {
	return &Bus{log: log}
}

func (b *Bus) Subscribe(h Handler) {
	b.mu.Lock()
	b.handlers = append(b.handlers, h)
	b.mu.Unlock()
}

// Publish delivers the event synchronously to every handler and returns the
// combined error of all handlers that failed or panicked. One failing handler
// does not stop the others.
func (b *Bus) Publish(event any) error {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers))
	copy(handlers, b.handlers)
	b.mu.RUnlock()

	errs := make([]error, 0, len(handlers))
	for _, h := range handlers {
		if err := b.deliver(h, event); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (b *Bus) deliver(h Handler, event any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			b.log.Error("event handler panicked", zap.Any("recover", r))
			err = fmt.Errorf("handler panicked: %v", r)
		}
	}()
	return h(event)
}
