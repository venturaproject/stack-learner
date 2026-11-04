package trabajadores

import (
	"time"

	"boilerplate-learner/internal/domain/shared"
)

// Trabajador is the aggregate root.
type Trabajador struct {
	shared.AggregateRoot

	ID                 int32
	Nombre             string
	Apellido           string
	SynergyResID       *int32
	DNI                *string
	Email              *string
	Telefono           *string
	Estado             Estado
	EmpStat            *string
	Cargo              *string
	Departamento       *string
	Loc                *string
	Ubicacion          *string
	Ciudad             *string
	FechaIncorporacion *time.Time
	FechaBaja          *time.Time
	Observaciones      string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type CreateData struct {
	Nombre             string
	Apellido           string
	SynergyResID       *int32
	DNI                *string
	Email              *string
	Telefono           *string
	Estado             string // "" defaults to activo
	EmpStat            *string
	Cargo              *string
	Departamento       *string
	Loc                *string
	Ubicacion          *string
	Ciudad             *string
	FechaIncorporacion *time.Time
	Observaciones      string
}

type UpdateData struct {
	Nombre             *string
	Apellido           *string
	SynergyResID       *int32
	DNI                *string
	Email              *string
	Telefono           *string
	Estado             *string
	EmpStat            *string
	Cargo              *string
	Departamento       *string
	Loc                *string
	Ubicacion          *string
	Ciudad             *string
	FechaIncorporacion *time.Time
	FechaBaja          *time.Time
	Observaciones      *string
}

func NewTrabajador(d CreateData) (*Trabajador, error) {
	if d.Nombre == "" {
		return nil, ErrNombreObligatorio
	}
	estado := EstadoActivo
	if d.Estado != "" {
		parsed, err := ParseEstado(d.Estado)
		if err != nil {
			return nil, ErrEstadoInvalido
		}
		estado = parsed
	}
	now := time.Now()
	t := &Trabajador{
		Nombre: d.Nombre, Apellido: d.Apellido, SynergyResID: d.SynergyResID,
		DNI: d.DNI, Email: d.Email, Telefono: d.Telefono, Estado: estado,
		EmpStat: d.EmpStat, Cargo: d.Cargo, Departamento: d.Departamento,
		Loc: d.Loc, Ubicacion: d.Ubicacion, Ciudad: d.Ciudad,
		FechaIncorporacion: d.FechaIncorporacion, Observaciones: d.Observaciones,
		CreatedAt: now, UpdatedAt: now,
	}
	t.Record(NewTrabajadorCreated(0))
	return t, nil
}

// Update applies a partial change. An invalid Estado string is rejected
// (previously it was cast blindly, bypassing the value object).
func (t *Trabajador) Update(u UpdateData) error {
	if u.Nombre != nil {
		t.Nombre = *u.Nombre
	}
	if u.Apellido != nil {
		t.Apellido = *u.Apellido
	}
	if u.SynergyResID != nil {
		t.SynergyResID = u.SynergyResID
	}
	if u.DNI != nil {
		t.DNI = u.DNI
	}
	if u.Email != nil {
		t.Email = u.Email
	}
	if u.Telefono != nil {
		t.Telefono = u.Telefono
	}
	if u.Estado != nil {
		parsed, err := ParseEstado(*u.Estado)
		if err != nil {
			return ErrEstadoInvalido
		}
		t.Estado = parsed
	}
	if u.EmpStat != nil {
		t.EmpStat = u.EmpStat
	}
	if u.Cargo != nil {
		t.Cargo = u.Cargo
	}
	if u.Departamento != nil {
		t.Departamento = u.Departamento
	}
	if u.Loc != nil {
		t.Loc = u.Loc
	}
	if u.Ubicacion != nil {
		t.Ubicacion = u.Ubicacion
	}
	if u.Ciudad != nil {
		t.Ciudad = u.Ciudad
	}
	if u.FechaIncorporacion != nil {
		t.FechaIncorporacion = u.FechaIncorporacion
	}
	if u.FechaBaja != nil {
		t.FechaBaja = u.FechaBaja
	}
	if u.Observaciones != nil {
		t.Observaciones = *u.Observaciones
	}
	t.UpdatedAt = time.Now()
	t.Record(NewTrabajadorUpdated(t.ID))
	return nil
}

// DarDeBaja marks the worker as terminated on the given date.
func (t *Trabajador) DarDeBaja(fecha time.Time) error {
	if !t.Estado.CanDarDeBaja() {
		return ErrYaBaja
	}
	t.Estado = EstadoBaja
	t.FechaBaja = &fecha
	t.UpdatedAt = time.Now()
	t.Record(NewTrabajadorDadoDeBaja(t.ID))
	return nil
}

// Delete records TrabajadorDeleted; the caller then removes the row.
func (t *Trabajador) Delete() {
	t.Record(NewTrabajadorDeleted(t.ID))
}
