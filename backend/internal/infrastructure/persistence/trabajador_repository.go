package persistence

import (
	"context"
	"errors"
	"fmt"
	"time"

	trabread "boilerplate-learner/internal/application/trabajadores/read"
	domaintrab "boilerplate-learner/internal/domain/trabajadores"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// ── Write side ───────────────────────────────────────────────────────────────

type TrabajadorRepository struct {
	db *gorm.DB
}

func NewTrabajadorRepository(db *gorm.DB) *TrabajadorRepository {
	return &TrabajadorRepository{db: db}
}

var _ domaintrab.Repository = (*TrabajadorRepository)(nil)

func (r *TrabajadorRepository) FindByID(ctx context.Context, id int32) (*domaintrab.Trabajador, error) {
	var m models.Trabajador
	if err := connFromContext(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domaintrab.ErrNotFound
		}
		return nil, err
	}
	return mapTrabajador(&m), nil
}

func (r *TrabajadorRepository) Save(ctx context.Context, t *domaintrab.Trabajador) error {
	m := domainToTrabajadorModel(t)
	if err := connFromContext(ctx, r.db).Save(m).Error; err != nil {
		return err
	}
	t.ID = m.ID
	return nil
}

func (r *TrabajadorRepository) Delete(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.Trabajador{}, id).Error
}

// ── Read side ────────────────────────────────────────────────────────────────

type TrabajadorReadRepository struct {
	db *gorm.DB
}

func NewTrabajadorReadRepository(db *gorm.DB) *TrabajadorReadRepository {
	return &TrabajadorReadRepository{db: db}
}

var _ trabread.Queries = (*TrabajadorReadRepository)(nil)

type trabajadorWithPhone struct {
	models.Trabajador
	TelID     *int32  `gorm:"column:tel_id"`
	TelIMEI   *string `gorm:"column:tel_imei"`
	TelNumero *string `gorm:"column:tel_numero"`
}

type trabajadorGlobalStats struct {
	Activos             int64 `gorm:"column:activos"`
	ConTelefono         int64 `gorm:"column:con_telefono"`
	IncorporadosEsteMes int64 `gorm:"column:incorporados_este_mes"`
	ConIMEI             int64 `gorm:"column:con_imei"`
}

func (r *TrabajadorReadRepository) List(ctx context.Context, f trabread.Filters) (*trabread.ListResult, error) {
	q := r.db.WithContext(ctx).Model(&models.Trabajador{})
	if f.Search != nil && *f.Search != "" {
		like := "%" + *f.Search + "%"
		q = q.Where("nombre ILIKE ? OR apellido ILIKE ? OR email ILIKE ? OR dni ILIKE ?", like, like, like, like)
	}
	if f.Activo != nil && *f.Activo != "" {
		if *f.Activo == "true" || *f.Activo == "1" {
			q = q.Where("estado = 'activo'")
		} else {
			q = q.Where("estado != 'activo'")
		}
	}
	if f.Ubicacion != nil && *f.Ubicacion != "" {
		q = q.Where("ubicacion = ?", f.Ubicacion)
	}
	if f.Ciudad != nil && *f.Ciudad != "" {
		q = q.Where("ciudad = ?", f.Ciudad)
	}
	if f.DateFrom != nil && *f.DateFrom != "" {
		q = q.Where("fecha_incorporacion >= ?", f.DateFrom)
	}
	if f.DateTo != nil && *f.DateTo != "" {
		q = q.Where("fecha_incorporacion <= ?", f.DateTo)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, err
	}

	const lateralPhone = `LEFT JOIN LATERAL (
		SELECT id AS tel_id, imei AS tel_imei, numero AS tel_numero
		FROM telefonos
		WHERE trabajador_id = trabajadores.id AND activo = true
		LIMIT 1
	) _tel ON true`

	var rows []trabajadorWithPhone
	if err := q.
		Select("trabajadores.*, _tel.tel_id, _tel.tel_imei, _tel.tel_numero").
		Joins(lateralPhone).
		Offset(int((f.Page - 1) * f.PerPage)).
		Limit(int(f.PerPage)).
		Order("apellido ASC, nombre ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	var gs trabajadorGlobalStats
	r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(CASE WHEN t.estado = 'activo' THEN 1 END)                                      AS activos,
			COUNT(DISTINCT _tel.tel_id)                                                           AS con_telefono,
			COUNT(CASE WHEN date_trunc('month', t.fecha_incorporacion) = date_trunc('month', NOW()) THEN 1 END) AS incorporados_este_mes,
			COUNT(DISTINCT CASE WHEN _tel.tel_imei IS NOT NULL THEN t.id END)                     AS con_imei
		FROM trabajadores t
		LEFT JOIN LATERAL (
			SELECT id AS tel_id, imei AS tel_imei
			FROM telefonos
			WHERE trabajador_id = t.id AND activo = true
			LIMIT 1
		) _tel ON true
	`).Scan(&gs)

	now := time.Now()
	details := make([]trabread.Details, 0, len(rows))
	for _, row := range rows {
		d := mapTrabajadorDetails(&row.Trabajador, now)
		d.TelefonoSynergy = row.TelNumero
		d.IMEI = row.TelIMEI
		if row.TelID != nil {
			d.TelefonosID = row.TelID
		}
		details = append(details, d)
	}

	return &trabread.ListResult{
		Results: details,
		Total:   total,
		Page:    f.Page,
		PerPage: f.PerPage,
		Stats: trabread.ListStats{
			Activos: gs.Activos, ConTelefono: gs.ConTelefono,
			IncorporadosEsteMes: gs.IncorporadosEsteMes, ConIMEI: gs.ConIMEI,
		},
	}, nil
}

func (r *TrabajadorReadRepository) Details(ctx context.Context, id int32) (*trabread.Details, error) {
	var m models.Trabajador
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var tel models.Telefono
	hasTel := r.db.WithContext(ctx).Where("trabajador_id = ? AND activo = true", id).First(&tel).Error == nil

	d := mapTrabajadorDetails(&m, time.Now())
	if hasTel {
		d.IMEI = tel.IMEI
		d.TelefonoSynergy = tel.Numero
		id32 := tel.ID
		d.TelefonosID = &id32
	}
	return &d, nil
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func mapTrabajador(m *models.Trabajador) *domaintrab.Trabajador {
	return &domaintrab.Trabajador{
		ID: m.ID, Nombre: m.Nombre, Apellido: m.Apellido, SynergyResID: m.SynergyResID,
		DNI: m.DNI, Email: m.Email, Telefono: m.Telefono,
		Estado: domaintrab.Estado(m.Estado), EmpStat: m.EmpStat, Cargo: m.Cargo,
		Departamento: m.Departamento, Loc: m.Loc, Ubicacion: m.Ubicacion, Ciudad: m.Ciudad,
		FechaIncorporacion: m.FechaIncorporacion, FechaBaja: m.FechaBaja,
		Observaciones: m.Observaciones, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func domainToTrabajadorModel(t *domaintrab.Trabajador) *models.Trabajador {
	return &models.Trabajador{
		ID: t.ID, Nombre: t.Nombre, Apellido: t.Apellido, SynergyResID: t.SynergyResID,
		DNI: t.DNI, Email: t.Email, Telefono: t.Telefono, Estado: string(t.Estado),
		EmpStat: t.EmpStat, Cargo: t.Cargo, Departamento: t.Departamento, Loc: t.Loc,
		Ubicacion: t.Ubicacion, Ciudad: t.Ciudad,
		FechaIncorporacion: t.FechaIncorporacion, FechaBaja: t.FechaBaja,
		Observaciones: t.Observaciones, CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

func mapTrabajadorDetails(m *models.Trabajador, now time.Time) trabread.Details {
	var fechaAlta, fechaBaja *string
	if m.FechaIncorporacion != nil {
		s := m.FechaIncorporacion.Format("2006-01-02")
		fechaAlta = &s
	}
	if m.FechaBaja != nil {
		s := m.FechaBaja.Format("2006-01-02")
		fechaBaja = &s
	}
	activo := m.Estado == "activo" && (m.FechaBaja == nil || m.FechaBaja.After(now))
	return trabread.Details{
		ID: m.ID, SynergyResID: m.SynergyResID,
		NombreCompleto: fmt.Sprintf("%s %s", m.Nombre, m.Apellido),
		Nombre:         m.Nombre, Apellido: m.Apellido, Email: m.Email,
		Ubicacion: m.Ubicacion, Ciudad: m.Ciudad, Loc: m.Loc,
		EmpStat: m.EmpStat, Activo: activo, Estado: m.Estado,
		FechaAlta: fechaAlta, FechaBaja: fechaBaja,
		Cargo: m.Cargo, Departamento: m.Departamento,
	}
}
