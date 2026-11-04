package dispositivos

import "boilerplate-learner/internal/domain/shared"

type DispositivoCreated struct {
	shared.BaseEvent
	DispositivoID int32
}

func NewDispositivoCreated(id int32) DispositivoCreated {
	return DispositivoCreated{BaseEvent: shared.NewBaseEvent(), DispositivoID: id}
}
func (DispositivoCreated) EventName() string { return "dispositivo.created" }

type DispositivoUpdated struct {
	shared.BaseEvent
	DispositivoID int32
}

func NewDispositivoUpdated(id int32) DispositivoUpdated {
	return DispositivoUpdated{BaseEvent: shared.NewBaseEvent(), DispositivoID: id}
}
func (DispositivoUpdated) EventName() string { return "dispositivo.updated" }

type DispositivoDeleted struct {
	shared.BaseEvent
	DispositivoID int32
}

func NewDispositivoDeleted(id int32) DispositivoDeleted {
	return DispositivoDeleted{BaseEvent: shared.NewBaseEvent(), DispositivoID: id}
}
func (DispositivoDeleted) EventName() string { return "dispositivo.deleted" }

type DispositivoEntregado struct {
	shared.BaseEvent
	DispositivoID int32
	TrabajadorID  int32
}

func NewDispositivoEntregado(id, trabajadorID int32) DispositivoEntregado {
	return DispositivoEntregado{BaseEvent: shared.NewBaseEvent(), DispositivoID: id, TrabajadorID: trabajadorID}
}
func (DispositivoEntregado) EventName() string { return "dispositivo.entregado" }
