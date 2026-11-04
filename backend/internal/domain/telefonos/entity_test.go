package telefonos_test

import (
	"errors"
	"testing"

	domaintelf "boilerplate-learner/internal/domain/telefonos"
)

func TestNewTelefonoRecordsCreatedEvent(t *testing.T) {
	tel, err := domaintelf.NewTelefono(domaintelf.CreateData{Activo: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	evs := tel.PullEvents()
	if len(evs) != 1 {
		t.Fatalf("want 1 event, got %d", len(evs))
	}
	if evs[0].EventName() != "telefono.created" {
		t.Fatalf("unexpected event name: %s", evs[0].EventName())
	}
	if got := tel.PullEvents(); len(got) != 0 {
		t.Fatalf("PullEvents should drain the buffer, got %d", len(got))
	}
}

func TestAsignarRejectsInactive(t *testing.T) {
	tel := &domaintelf.Telefono{Activo: false}
	if err := tel.Asignar(7); !errors.Is(err, domaintelf.ErrInactivo) {
		t.Fatalf("want ErrInactivo, got %v", err)
	}
}

func TestAsignarRejectsReassignment(t *testing.T) {
	other := int32(1)
	tel := &domaintelf.Telefono{Activo: true, TrabajadorID: &other}
	if err := tel.Asignar(2); !errors.Is(err, domaintelf.ErrYaAsignado) {
		t.Fatalf("want ErrYaAsignado, got %v", err)
	}
}

func TestAsignarSuccessRecordsEvent(t *testing.T) {
	tel := &domaintelf.Telefono{Activo: true}
	if err := tel.Asignar(42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.TrabajadorID == nil || *tel.TrabajadorID != 42 {
		t.Fatalf("worker not linked")
	}
	evs := tel.PullEvents()
	if len(evs) != 1 || evs[0].EventName() != "telefono.asignado" {
		t.Fatalf("unexpected events: %+v", evs)
	}
}

func TestDesactivarClearsWorker(t *testing.T) {
	w := int32(3)
	tel := &domaintelf.Telefono{Activo: true, TrabajadorID: &w}
	tel.Desactivar()
	if tel.Activo || tel.TrabajadorID != nil {
		t.Fatalf("desactivar did not reset state: %+v", tel)
	}
	if evs := tel.PullEvents(); len(evs) != 1 || evs[0].EventName() != "telefono.desactivado" {
		t.Fatalf("unexpected events: %+v", evs)
	}
}
