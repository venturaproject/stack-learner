package trabajadores

import "boilerplate-learner/internal/domain/shared"

type TrabajadorCreated struct {
	shared.BaseEvent
	TrabajadorID int32
}

func NewTrabajadorCreated(id int32) TrabajadorCreated {
	return TrabajadorCreated{BaseEvent: shared.NewBaseEvent(), TrabajadorID: id}
}
func (TrabajadorCreated) EventName() string { return "trabajador.created" }

type TrabajadorUpdated struct {
	shared.BaseEvent
	TrabajadorID int32
}

func NewTrabajadorUpdated(id int32) TrabajadorUpdated {
	return TrabajadorUpdated{BaseEvent: shared.NewBaseEvent(), TrabajadorID: id}
}
func (TrabajadorUpdated) EventName() string { return "trabajador.updated" }

type TrabajadorDeleted struct {
	shared.BaseEvent
	TrabajadorID int32
}

func NewTrabajadorDeleted(id int32) TrabajadorDeleted {
	return TrabajadorDeleted{BaseEvent: shared.NewBaseEvent(), TrabajadorID: id}
}
func (TrabajadorDeleted) EventName() string { return "trabajador.deleted" }

type TrabajadorDadoDeBaja struct {
	shared.BaseEvent
	TrabajadorID int32
}

func NewTrabajadorDadoDeBaja(id int32) TrabajadorDadoDeBaja {
	return TrabajadorDadoDeBaja{BaseEvent: shared.NewBaseEvent(), TrabajadorID: id}
}
func (TrabajadorDadoDeBaja) EventName() string { return "trabajador.dado_de_baja" }
