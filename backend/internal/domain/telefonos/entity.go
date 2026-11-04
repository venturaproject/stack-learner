package telefonos

import (
	"time"

	"boilerplate-learner/internal/domain/shared"
)

// Telefono is the aggregate root. It carries behaviour and invariants; it has no
// knowledge of JSON, GORM or HTTP. Read-side projections live in
// internal/application/telefonos/read.
type Telefono struct {
	shared.AggregateRoot

	ID            int32
	Numero        *string
	Tipo          *string
	Plan          *string
	Nplan         *string
	Pin           *string
	Puk           *string
	IMEI          *string
	IMEI2         *string
	Marca         *string
	Modelo        *string
	Linea         *string
	Operadora     *string
	EstadoID      *int32
	TipologiaID   *int32
	TrabajadorID  *int32
	Observaciones string
	Activo        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CreateData is the command payload for NewTelefono.
type CreateData struct {
	Numero        *string
	Tipo          *string
	Plan          *string
	Nplan         *string
	Pin           *string
	Puk           *string
	IMEI          *string
	IMEI2         *string
	Marca         *string
	Modelo        *string
	Linea         *string
	Operadora     *string
	EstadoID      *int32
	TipologiaID   *int32
	TrabajadorID  *int32
	Observaciones string
	Activo        bool
}

// UpdateData carries a partial update: a non-nil field is applied, a nil field
// is left unchanged.
type UpdateData struct {
	Numero        *string
	Tipo          *string
	Plan          *string
	Nplan         *string
	Pin           *string
	Puk           *string
	IMEI          *string
	IMEI2         *string
	Marca         *string
	Modelo        *string
	Linea         *string
	Operadora     *string
	EstadoID      *int32
	TipologiaID   *int32
	TrabajadorID  *int32
	Observaciones *string
	Activo        *bool
}

// NewTelefono builds a valid aggregate and records TelefonoCreated.
func NewTelefono(d CreateData) (*Telefono, error) {
	now := time.Now()
	t := &Telefono{
		Numero:        d.Numero,
		Tipo:          d.Tipo,
		Plan:          d.Plan,
		Nplan:         d.Nplan,
		Pin:           d.Pin,
		Puk:           d.Puk,
		IMEI:          d.IMEI,
		IMEI2:         d.IMEI2,
		Marca:         d.Marca,
		Modelo:        d.Modelo,
		Linea:         d.Linea,
		Operadora:     d.Operadora,
		EstadoID:      d.EstadoID,
		TipologiaID:   d.TipologiaID,
		TrabajadorID:  d.TrabajadorID,
		Observaciones: d.Observaciones,
		Activo:        d.Activo,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	t.Record(NewTelefonoCreated(0)) // ID filled by the repository on insert
	return t, nil
}

// Update applies a partial change and records TelefonoUpdated.
func (t *Telefono) Update(d UpdateData) error {
	if d.Numero != nil {
		t.Numero = d.Numero
	}
	if d.Tipo != nil {
		t.Tipo = d.Tipo
	}
	if d.Plan != nil {
		t.Plan = d.Plan
	}
	if d.Nplan != nil {
		t.Nplan = d.Nplan
	}
	if d.Pin != nil {
		t.Pin = d.Pin
	}
	if d.Puk != nil {
		t.Puk = d.Puk
	}
	if d.IMEI != nil {
		t.IMEI = d.IMEI
	}
	if d.IMEI2 != nil {
		t.IMEI2 = d.IMEI2
	}
	if d.Marca != nil {
		t.Marca = d.Marca
	}
	if d.Modelo != nil {
		t.Modelo = d.Modelo
	}
	if d.Linea != nil {
		t.Linea = d.Linea
	}
	if d.Operadora != nil {
		t.Operadora = d.Operadora
	}
	if d.EstadoID != nil {
		t.EstadoID = d.EstadoID
	}
	if d.TipologiaID != nil {
		t.TipologiaID = d.TipologiaID
	}
	if d.TrabajadorID != nil {
		t.TrabajadorID = d.TrabajadorID
	}
	if d.Observaciones != nil {
		t.Observaciones = *d.Observaciones
	}
	if d.Activo != nil {
		t.Activo = *d.Activo
	}
	t.UpdatedAt = time.Now()
	t.Record(NewTelefonoUpdated(t.ID))
	return nil
}

// Asignar links the phone to a worker. It is idempotent-safe but rejects
// reassignment to a different worker and assignment of an inactive line.
func (t *Telefono) Asignar(trabajadorID int32) error {
	if !t.Activo {
		return ErrInactivo
	}
	if t.TrabajadorID != nil && *t.TrabajadorID != trabajadorID {
		return ErrYaAsignado
	}
	t.TrabajadorID = &trabajadorID
	t.UpdatedAt = time.Now()
	t.Record(NewTelefonoAsignado(t.ID, trabajadorID))
	return nil
}

// Desactivar deactivates the line and unlinks its worker.
func (t *Telefono) Desactivar() {
	t.Activo = false
	t.TrabajadorID = nil
	t.UpdatedAt = time.Now()
	t.Record(NewTelefonoDesactivado(t.ID))
}

// Delete records TelefonoDeleted; the caller then removes the row.
func (t *Telefono) Delete() {
	t.Record(NewTelefonoDeleted(t.ID))
}

// Actuacion is a sub-entity of the Telefono aggregate (its own repository).
type Actuacion struct {
	ID          int32
	Descripcion string
	Fecha       time.Time
	UserID      *int32
}
