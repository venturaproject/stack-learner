// Package trabajadores is the application layer for the Trabajador aggregate.
// Write use cases run in a TxManager transaction and publish domain events
// through the outbox EventPublisher.
package trabajadores

import (
	"time"

	appshared "boilerplate-learner/internal/application/shared"
	"boilerplate-learner/internal/application/trabajadores/read"
	domaintrab "boilerplate-learner/internal/domain/trabajadores"
)

type UseCases struct {
	List      *ListUseCase
	Get       *GetUseCase
	Create    *CreateUseCase
	Update    *UpdateUseCase
	Delete    *DeleteUseCase
	DarDeBaja *DarDeBajaUseCase
}

func NewUseCases(
	repo domaintrab.Repository,
	queries read.Queries,
	tx appshared.TxManager,
	pub appshared.EventPublisher,
) *UseCases {
	return &UseCases{
		List:      &ListUseCase{queries: queries},
		Get:       &GetUseCase{queries: queries},
		Create:    &CreateUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Update:    &UpdateUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
		Delete:    &DeleteUseCase{repo: repo, tx: tx, pub: pub},
		DarDeBaja: &DarDeBajaUseCase{repo: repo, queries: queries, tx: tx, pub: pub},
	}
}

const dateLayout = "2006-01-02"

func parseDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	t, err := time.Parse(dateLayout, *s)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// ── Input DTOs ───────────────────────────────────────────────────────────────

type CreateInput struct {
	Nombre             string  `json:"nombre"`
	Apellido           string  `json:"apellido"`
	SynergyResID       *int32  `json:"synergy_res_id"`
	DNI                *string `json:"dni"`
	Email              *string `json:"email"`
	Telefono           *string `json:"telefono"`
	Estado             string  `json:"estado"`
	EmpStat            *string `json:"emp_stat"`
	Cargo              *string `json:"cargo"`
	Departamento       *string `json:"departamento"`
	Loc                *string `json:"loc"`
	Ubicacion          *string `json:"ubicacion"`
	Ciudad             *string `json:"ciudad"`
	FechaIncorporacion *string `json:"fecha_incorporacion"`
	Observaciones      string  `json:"observaciones"`
}

func (in CreateInput) toDomain() (domaintrab.CreateData, error) {
	fecha, err := parseDate(in.FechaIncorporacion)
	if err != nil {
		return domaintrab.CreateData{}, err
	}
	return domaintrab.CreateData{
		Nombre: in.Nombre, Apellido: in.Apellido, SynergyResID: in.SynergyResID,
		DNI: in.DNI, Email: in.Email, Telefono: in.Telefono, Estado: in.Estado,
		EmpStat: in.EmpStat, Cargo: in.Cargo, Departamento: in.Departamento,
		Loc: in.Loc, Ubicacion: in.Ubicacion, Ciudad: in.Ciudad,
		FechaIncorporacion: fecha, Observaciones: in.Observaciones,
	}, nil
}

type UpdateInput struct {
	Nombre             *string `json:"nombre"`
	Apellido           *string `json:"apellido"`
	SynergyResID       *int32  `json:"synergy_res_id"`
	DNI                *string `json:"dni"`
	Email              *string `json:"email"`
	Telefono           *string `json:"telefono"`
	Estado             *string `json:"estado"`
	EmpStat            *string `json:"emp_stat"`
	Cargo              *string `json:"cargo"`
	Departamento       *string `json:"departamento"`
	Loc                *string `json:"loc"`
	Ubicacion          *string `json:"ubicacion"`
	Ciudad             *string `json:"ciudad"`
	FechaIncorporacion *string `json:"fecha_incorporacion"`
	FechaBaja          *string `json:"fecha_baja"`
	Observaciones      *string `json:"observaciones"`
}

func (in UpdateInput) toDomain() (domaintrab.UpdateData, error) {
	alta, err := parseDate(in.FechaIncorporacion)
	if err != nil {
		return domaintrab.UpdateData{}, err
	}
	baja, err := parseDate(in.FechaBaja)
	if err != nil {
		return domaintrab.UpdateData{}, err
	}
	return domaintrab.UpdateData{
		Nombre: in.Nombre, Apellido: in.Apellido, SynergyResID: in.SynergyResID,
		DNI: in.DNI, Email: in.Email, Telefono: in.Telefono, Estado: in.Estado,
		EmpStat: in.EmpStat, Cargo: in.Cargo, Departamento: in.Departamento,
		Loc: in.Loc, Ubicacion: in.Ubicacion, Ciudad: in.Ciudad,
		FechaIncorporacion: alta, FechaBaja: baja, Observaciones: in.Observaciones,
	}, nil
}
