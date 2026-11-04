package persistence

import (
	"context"
	"errors"
	"fmt"

	telfread "boilerplate-learner/internal/application/telefonos/read"
	domaintelf "boilerplate-learner/internal/domain/telefonos"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// ── Write side ───────────────────────────────────────────────────────────────

type TelefonoRepository struct {
	db *gorm.DB
}

func NewTelefonoRepository(db *gorm.DB) *TelefonoRepository { return &TelefonoRepository{db: db} }

var _ domaintelf.Repository = (*TelefonoRepository)(nil)

func (r *TelefonoRepository) FindByID(ctx context.Context, id int32) (*domaintelf.Telefono, error) {
	var m models.Telefono
	if err := connFromContext(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domaintelf.ErrNotFound
		}
		return nil, err
	}
	return mapTelefono(&m), nil
}

func (r *TelefonoRepository) Save(ctx context.Context, t *domaintelf.Telefono) error {
	m := domainToTelefonoModel(t)
	if err := connFromContext(ctx, r.db).Save(m).Error; err != nil {
		return err
	}
	t.ID = m.ID
	return nil
}

func (r *TelefonoRepository) Delete(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.Telefono{}, id).Error
}

// ── Actuaciones ──────────────────────────────────────────────────────────────

type ActuacionTelefonoRepository struct {
	db *gorm.DB
}

func NewActuacionTelefonoRepository(db *gorm.DB) *ActuacionTelefonoRepository {
	return &ActuacionTelefonoRepository{db: db}
}

var _ domaintelf.ActuacionRepository = (*ActuacionTelefonoRepository)(nil)

func (r *ActuacionTelefonoRepository) Add(ctx context.Context, telefonoID int32, descripcion string, userID *int32) (*domaintelf.Actuacion, error) {
	m := &models.ActuacionTelefono{TelefonoID: telefonoID, Descripcion: descripcion, UserID: userID}
	if err := connFromContext(ctx, r.db).Create(m).Error; err != nil {
		return nil, err
	}
	return &domaintelf.Actuacion{ID: m.ID, Descripcion: m.Descripcion, Fecha: m.Fecha, UserID: m.UserID}, nil
}

func (r *ActuacionTelefonoRepository) Remove(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.ActuacionTelefono{}, id).Error
}

// ── Read side ────────────────────────────────────────────────────────────────

type TelefonoReadRepository struct {
	db *gorm.DB
}

func NewTelefonoReadRepository(db *gorm.DB) *TelefonoReadRepository {
	return &TelefonoReadRepository{db: db}
}

var _ telfread.Queries = (*TelefonoReadRepository)(nil)

type telefonoGlobalStats struct {
	Total      int64 `gorm:"column:total"`
	Activos    int64 `gorm:"column:activos"`
	Congelados int64 `gorm:"column:congelados"`
	Bajas      int64 `gorm:"column:bajas"`
	SinDatos   int64 `gorm:"column:sin_datos"`
}

func (r *TelefonoReadRepository) List(ctx context.Context, f telfread.Filters) (*telfread.ListResult, error) {
	q := r.db.WithContext(ctx).Model(&models.Telefono{})

	if f.Search != nil && *f.Search != "" {
		like := "%" + *f.Search + "%"
		q = q.Where("numero ILIKE ? OR imei ILIKE ? OR marca ILIKE ? OR modelo ILIKE ?", like, like, like, like)
	}
	if f.Status != nil && *f.Status != "" {
		q = q.Where("estado_id = ?", f.Status)
	}
	if f.Plan != nil && *f.Plan != "" {
		q = q.Where("plan ILIKE ?", "%"+*f.Plan+"%")
	}

	var total int64
	q.Count(&total)

	var ms []models.Telefono
	if err := q.Preload("Estado").Preload("Trabajador").Preload("Actuaciones").
		Offset(int((f.Page - 1) * f.PerPage)).Limit(int(f.PerPage)).
		Order("id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}

	var gs telefonoGlobalStats
	r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(*) AS total,
			COUNT(CASE WHEN estado_id = ? THEN 1 END) AS activos,
			COUNT(CASE WHEN estado_id = ? THEN 1 END) AS congelados,
			COUNT(CASE WHEN estado_id = ? THEN 1 END) AS bajas,
			COUNT(CASE WHEN estado_id = ? OR estado_id IS NULL THEN 1 END) AS sin_datos
		FROM telefonos
	`, domaintelf.EstadoActivoID, domaintelf.EstadoCongeladoID,
		domaintelf.EstadoBajaID, domaintelf.EstadoSinDatosID).Scan(&gs)

	estados, _ := r.Estados(ctx)

	results := make([]telfread.Details, 0, len(ms))
	for i := range ms {
		results = append(results, mapTelefonoDetails(&ms[i]))
	}

	return &telfread.ListResult{
		Results: results,
		Total:   total,
		Page:    f.Page,
		PerPage: f.PerPage,
		Estados: estados,
		Stats: telfread.ListStats{
			Total: gs.Total, Activos: gs.Activos, Congelados: gs.Congelados,
			Bajas: gs.Bajas, SinDatos: gs.SinDatos,
		},
	}, nil
}

func (r *TelefonoReadRepository) Details(ctx context.Context, id int32) (*telfread.Details, error) {
	var m models.Telefono
	if err := r.db.WithContext(ctx).
		Preload("Estado").Preload("Trabajador").Preload("Actuaciones").
		First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := mapTelefonoDetails(&m)
	return &d, nil
}

func (r *TelefonoReadRepository) Estados(ctx context.Context) ([]telfread.EstadoSummary, error) {
	var ms []models.EstadoTelefono
	if err := r.db.WithContext(ctx).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]telfread.EstadoSummary, 0, len(ms))
	for _, m := range ms {
		out = append(out, telfread.EstadoSummary{ID: m.ID, Nombre: m.Nombre, Color: m.Color})
	}
	return out, nil
}

func (r *TelefonoReadRepository) Tipologias(ctx context.Context) ([]telfread.TipologiaSummary, error) {
	var ms []models.Tipologia
	if err := r.db.WithContext(ctx).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]telfread.TipologiaSummary, 0, len(ms))
	for _, m := range ms {
		out = append(out, telfread.TipologiaSummary{ID: m.ID, Nombre: m.Nombre})
	}
	return out, nil
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func mapTelefono(m *models.Telefono) *domaintelf.Telefono {
	return &domaintelf.Telefono{
		ID: m.ID, Numero: m.Numero, Tipo: m.Tipo, Plan: m.Plan, Nplan: m.Nplan,
		Pin: m.Pin, Puk: m.Puk, IMEI: m.IMEI, IMEI2: m.IMEI2, Marca: m.Marca,
		Modelo: m.Modelo, Linea: m.Linea, Operadora: m.Operadora,
		EstadoID: m.EstadoID, TipologiaID: m.TipologiaID, TrabajadorID: m.TrabajadorID,
		Observaciones: m.Observaciones, Activo: m.Activo,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func domainToTelefonoModel(t *domaintelf.Telefono) *models.Telefono {
	return &models.Telefono{
		ID: t.ID, Numero: t.Numero, Tipo: t.Tipo, Plan: t.Plan, Nplan: t.Nplan,
		Pin: t.Pin, Puk: t.Puk, IMEI: t.IMEI, IMEI2: t.IMEI2, Marca: t.Marca,
		Modelo: t.Modelo, Linea: t.Linea, Operadora: t.Operadora,
		EstadoID: t.EstadoID, TipologiaID: t.TipologiaID, TrabajadorID: t.TrabajadorID,
		Observaciones: t.Observaciones, Activo: t.Activo,
		CreatedAt: t.CreatedAt, UpdatedAt: t.UpdatedAt,
	}
}

func mapTelefonoDetails(m *models.Telefono) telfread.Details {
	d := telfread.Details{
		ID: m.ID, Numero: m.Numero, Plan: m.Plan, Nplan: m.Nplan, Tipo: m.Tipo,
		Pin: m.Pin, Puk: m.Puk, IMEI: m.IMEI, IMEI2: m.IMEI2, Marca: m.Marca,
		Modelo: m.Modelo, Linea: m.Linea, Operadora: m.Operadora,
		EstadoID: m.EstadoID, TipologiaID: m.TipologiaID,
		Observaciones: m.Observaciones, Activo: m.Activo,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
	if m.Estado != nil {
		d.Status = &telfread.StatusDetails{ID: m.Estado.ID, Nombre: m.Estado.Nombre, Color: m.Estado.Color}
	}
	if m.Trabajador != nil {
		d.Trabajador = &telfread.WorkerDetails{
			ID:             m.Trabajador.ID,
			NombreCompleto: fmt.Sprintf("%s %s", m.Trabajador.Nombre, m.Trabajador.Apellido),
		}
	}
	actuaciones := make([]telfread.Actuacion, 0, len(m.Actuaciones))
	for _, a := range m.Actuaciones {
		actuaciones = append(actuaciones, telfread.Actuacion{
			ID: a.ID, Descripcion: a.Descripcion, Fecha: a.Fecha, UserID: a.UserID,
		})
	}
	d.Actuaciones = actuaciones
	return d
}
