package dispositivos

import (
	"time"

	"boilerplate-learner/internal/domain/shared"
)

// Dispositivo is the aggregate root.
type Dispositivo struct {
	shared.AggregateRoot

	ID            int32
	Numero        *string
	IMEI          *string
	Serie         *string
	SynergyResID  *int32
	BrandID       *int32
	ModelID       *int32
	Grupo         *string
	Estado        Estado
	TrabajadorID  *int32
	Observaciones string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type CreateData struct {
	Numero        *string
	IMEI          *string
	Serie         *string
	SynergyResID  *int32
	BrandID       *int32
	ModelID       *int32
	Grupo         *string
	Estado        string // "" defaults to disponible
	TrabajadorID  *int32
	Observaciones string
}

type UpdateData struct {
	Numero        *string
	IMEI          *string
	Serie         *string
	SynergyResID  *int32
	BrandID       *int32
	ModelID       *int32
	Grupo         *string
	Estado        *string
	TrabajadorID  *int32
	Observaciones *string
}

// NewDispositivo validates the command and records DispositivoCreated.
func NewDispositivo(d CreateData) (*Dispositivo, error) {
	estado := EstadoDisponible
	if d.Estado != "" {
		parsed, err := ParseEstado(d.Estado)
		if err != nil {
			return nil, ErrEstadoInvalido
		}
		estado = parsed
	}
	now := time.Now()
	dev := &Dispositivo{
		Numero: d.Numero, IMEI: d.IMEI, Serie: d.Serie, SynergyResID: d.SynergyResID,
		BrandID: d.BrandID, ModelID: d.ModelID, Grupo: d.Grupo, Estado: estado,
		TrabajadorID: d.TrabajadorID, Observaciones: d.Observaciones,
		CreatedAt: now, UpdatedAt: now,
	}
	dev.Record(NewDispositivoCreated(0))
	return dev, nil
}

// Update applies a partial change. An invalid Estado string is rejected
// (previously it was cast blindly, bypassing the value object).
func (d *Dispositivo) Update(u UpdateData) error {
	if u.Numero != nil {
		d.Numero = u.Numero
	}
	if u.IMEI != nil {
		d.IMEI = u.IMEI
	}
	if u.Serie != nil {
		d.Serie = u.Serie
	}
	if u.SynergyResID != nil {
		d.SynergyResID = u.SynergyResID
	}
	if u.BrandID != nil {
		d.BrandID = u.BrandID
	}
	if u.ModelID != nil {
		d.ModelID = u.ModelID
	}
	if u.Grupo != nil {
		d.Grupo = u.Grupo
	}
	if u.Estado != nil {
		parsed, err := ParseEstado(*u.Estado)
		if err != nil {
			return ErrEstadoInvalido
		}
		d.Estado = parsed
	}
	if u.TrabajadorID != nil {
		d.TrabajadorID = u.TrabajadorID
	}
	if u.Observaciones != nil {
		d.Observaciones = *u.Observaciones
	}
	d.UpdatedAt = time.Now()
	d.Record(NewDispositivoUpdated(d.ID))
	return nil
}

// Entregar assigns the device to a worker; only assignable devices may be handed
// over.
func (d *Dispositivo) Entregar(trabajadorID int32) error {
	if !d.Estado.IsAsignable() {
		return ErrYaAsignado
	}
	d.TrabajadorID = &trabajadorID
	d.Estado = EstadoAsignado
	d.UpdatedAt = time.Now()
	d.Record(NewDispositivoEntregado(d.ID, trabajadorID))
	return nil
}

// Delete records DispositivoDeleted; the caller then removes the row.
func (d *Dispositivo) Delete() {
	d.Record(NewDispositivoDeleted(d.ID))
}
