package events_test

import (
	"errors"
	"testing"

	"go.uber.org/zap"

	"boilerplate-learner/internal/infrastructure/events"
)

func TestBus_PublishReportsHandlerError(t *testing.T) {
	bus := events.NewBus(zap.NewNop())
	sentinel := errors.New("boom")

	bus.Subscribe(func(any) error { return nil })
	bus.Subscribe(func(any) error { return sentinel })

	err := bus.Publish("evt")
	if !errors.Is(err, sentinel) {
		t.Fatalf("want sentinel error, got %v", err)
	}
}

func TestBus_PublishRecoversPanicAsError(t *testing.T) {
	bus := events.NewBus(zap.NewNop())
	delivered := false

	bus.Subscribe(func(any) error { panic("kaboom") })
	bus.Subscribe(func(any) error { delivered = true; return nil })

	err := bus.Publish("evt")
	if err == nil {
		t.Fatal("want error from panicking handler")
	}
	if !delivered {
		t.Fatal("a panicking handler must not stop later handlers")
	}
}

func TestBus_PublishNilWhenAllHandlersSucceed(t *testing.T) {
	bus := events.NewBus(zap.NewNop())
	bus.Subscribe(func(any) error { return nil })
	if err := bus.Publish("evt"); err != nil {
		t.Fatalf("want nil, got %v", err)
	}
}
