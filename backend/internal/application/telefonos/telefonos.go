// Package telefonos holds the application (use-case) layer for the Telefono
// aggregate. Write use cases run inside a TxManager transaction and publish the
// aggregate's domain events through the outbox EventPublisher. Reads are served
// by the read.Queries port.
//
// Input structs carry json tags (they are bound directly from the request body)
// but no framework validation tags — required-field checks are explicit and
// return apperrors.Invalid.
package telefonos

import (
	appshared "boilerplate-learner/internal/application/shared"
	"boilerplate-learner/internal/application/telefonos/read"
	domaintelf "boilerplate-learner/internal/domain/telefonos"
)

// UseCases bundles every telefonos use case behind one dependency.
type UseCases struct {
	List            *ListUseCase
	Get             *GetUseCase
	ListEstados     *ListEstadosUseCase
	ListTipologias  *ListTipologiasUseCase
	Create          *CreateUseCase
	Update          *UpdateUseCase
	Delete          *DeleteUseCase
	Desactivar      *DesactivarUseCase
	Asignar         *AsignarUseCase
	AddActuacion    *AddActuacionUseCase
	RemoveActuacion *RemoveActuacionUseCase
}

func NewUseCases(
	repo domaintelf.Repository,
	actRepo domaintelf.ActuacionRepository,
	queries read.Queries,
	tx appshared.TxManager,
	pub appshared.EventPublisher,
) *UseCases {
	return &UseCases{
		List:            &ListUseCase{queries: queries},
		Get:             &GetUseCase{queries: queries},
		ListEstados:     &ListEstadosUseCase{queries: queries},
		ListTipologias:  &ListTipologiasUseCase{queries: queries},
		Create:          &CreateUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Update:          &UpdateUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Delete:          &DeleteUseCase{repo: repo, tx: tx, pub: pub},
		Desactivar:      &DesactivarUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Asignar:         &AsignarUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		AddActuacion:    &AddActuacionUseCase{repo: repo, actRepo: actRepo, tx: tx},
		RemoveActuacion: &RemoveActuacionUseCase{actRepo: actRepo},
	}
}

// ── Input DTOs ───────────────────────────────────────────────────────────────

type CreateInput struct {
	Numero        *string `json:"numero"`
	Tipo          *string `json:"tipo"`
	Plan          *string `json:"plan"`
	Nplan         *string `json:"nplan"`
	Pin           *string `json:"pin"`
	Puk           *string `json:"puk"`
	IMEI          *string `json:"imei"`
	IMEI2         *string `json:"imei2"`
	Marca         *string `json:"marca"`
	Modelo        *string `json:"modelo"`
	Linea         *string `json:"linea"`
	Operadora     *string `json:"operadora"`
	EstadoID      *int32  `json:"estado_telefonos_id"`
	TipologiaID   *int32  `json:"tipologias_id"`
	TrabajadorID  *int32  `json:"trabajador_id"`
	Observaciones string  `json:"observaciones"`
	Activo        bool    `json:"activo"`
}

func (in CreateInput) toDomain() domaintelf.CreateData {
	return domaintelf.CreateData{
		Numero: in.Numero, Tipo: in.Tipo, Plan: in.Plan, Nplan: in.Nplan,
		Pin: in.Pin, Puk: in.Puk, IMEI: in.IMEI, IMEI2: in.IMEI2,
		Marca: in.Marca, Modelo: in.Modelo, Linea: in.Linea, Operadora: in.Operadora,
		EstadoID: in.EstadoID, TipologiaID: in.TipologiaID, TrabajadorID: in.TrabajadorID,
		Observaciones: in.Observaciones, Activo: in.Activo,
	}
}

type UpdateInput struct {
	Numero        *string `json:"numero"`
	Tipo          *string `json:"tipo"`
	Plan          *string `json:"plan"`
	Nplan         *string `json:"nplan"`
	Pin           *string `json:"pin"`
	Puk           *string `json:"puk"`
	IMEI          *string `json:"imei"`
	IMEI2         *string `json:"imei2"`
	Marca         *string `json:"marca"`
	Modelo        *string `json:"modelo"`
	Linea         *string `json:"linea"`
	Operadora     *string `json:"operadora"`
	EstadoID      *int32  `json:"estado_telefonos_id"`
	TipologiaID   *int32  `json:"tipologias_id"`
	TrabajadorID  *int32  `json:"trabajador_id"`
	Observaciones *string `json:"observaciones"`
	Activo        *bool   `json:"activo"`
}

func (in UpdateInput) toDomain() domaintelf.UpdateData {
	return domaintelf.UpdateData{
		Numero: in.Numero, Tipo: in.Tipo, Plan: in.Plan, Nplan: in.Nplan,
		Pin: in.Pin, Puk: in.Puk, IMEI: in.IMEI, IMEI2: in.IMEI2,
		Marca: in.Marca, Modelo: in.Modelo, Linea: in.Linea, Operadora: in.Operadora,
		EstadoID: in.EstadoID, TipologiaID: in.TipologiaID, TrabajadorID: in.TrabajadorID,
		Observaciones: in.Observaciones, Activo: in.Activo,
	}
}

type AsignarInput struct {
	TrabajadorID int32 `json:"trabajador_id"`
}

type ActuacionInput struct {
	Descripcion string `json:"descripcion"`
}
