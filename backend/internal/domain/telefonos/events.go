package telefonos

import "boilerplate-learner/internal/domain/shared"

// Every event embeds shared.BaseEvent (OccurredAt) and implements
// shared.DomainEvent via EventName.

type TelefonoCreated struct {
	shared.BaseEvent
	TelefonoID int32
}

func NewTelefonoCreated(id int32) TelefonoCreated {
	return TelefonoCreated{BaseEvent: shared.NewBaseEvent(), TelefonoID: id}
}
func (TelefonoCreated) EventName() string { return "telefono.created" }

type TelefonoUpdated struct {
	shared.BaseEvent
	TelefonoID int32
}

func NewTelefonoUpdated(id int32) TelefonoUpdated {
	return TelefonoUpdated{BaseEvent: shared.NewBaseEvent(), TelefonoID: id}
}
func (TelefonoUpdated) EventName() string { return "telefono.updated" }

type TelefonoDeleted struct {
	shared.BaseEvent
	TelefonoID int32
}

func NewTelefonoDeleted(id int32) TelefonoDeleted {
	return TelefonoDeleted{BaseEvent: shared.NewBaseEvent(), TelefonoID: id}
}
func (TelefonoDeleted) EventName() string { return "telefono.deleted" }

type TelefonoAsignado struct {
	shared.BaseEvent
	TelefonoID   int32
	TrabajadorID int32
}

func NewTelefonoAsignado(id, trabajadorID int32) TelefonoAsignado {
	return TelefonoAsignado{BaseEvent: shared.NewBaseEvent(), TelefonoID: id, TrabajadorID: trabajadorID}
}
func (TelefonoAsignado) EventName() string { return "telefono.asignado" }

type TelefonoDesactivado struct {
	shared.BaseEvent
	TelefonoID int32
}

func NewTelefonoDesactivado(id int32) TelefonoDesactivado {
	return TelefonoDesactivado{BaseEvent: shared.NewBaseEvent(), TelefonoID: id}
}
func (TelefonoDesactivado) EventName() string { return "telefono.desactivado" }
