package persistence

import (
	"context"
	"errors"
	"fmt"

	dispread "boilerplate-learner/internal/application/dispositivos/read"
	domaindisp "boilerplate-learner/internal/domain/dispositivos"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// ── Write side ───────────────────────────────────────────────────────────────

type DispositivoRepository struct {
	db *gorm.DB
}

func NewDispositivoRepository(db *gorm.DB) *DispositivoRepository {
	return &DispositivoRepository{db: db}
}

var _ domaindisp.Repository = (*DispositivoRepository)(nil)

func (r *DispositivoRepository) FindByID(ctx context.Context, id int32) (*domaindisp.Dispositivo, error) {
	var m models.Dispositivo
	if err := connFromContext(ctx, r.db).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domaindisp.ErrNotFound
		}
		return nil, err
	}
	return mapDispositivo(&m), nil
}

func (r *DispositivoRepository) Save(ctx context.Context, d *domaindisp.Dispositivo) error {
	m := domainToDispositivoModel(d)
	if err := connFromContext(ctx, r.db).Save(m).Error; err != nil {
		return err
	}
	d.ID = m.ID
	return nil
}

func (r *DispositivoRepository) Delete(ctx context.Context, id int32) error {
	return connFromContext(ctx, r.db).Delete(&models.Dispositivo{}, id).Error
}

// ── Catalog ──────────────────────────────────────────────────────────────────

type DispositivoCatalogRepository struct {
	db *gorm.DB
}

func NewDispositivoCatalogRepository(db *gorm.DB) *DispositivoCatalogRepository {
	return &DispositivoCatalogRepository{db: db}
}

var _ domaindisp.CatalogRepository = (*DispositivoCatalogRepository)(nil)

func (r *DispositivoCatalogRepository) ListBrands(ctx context.Context) ([]domaindisp.Brand, error) {
	var ms []models.DeviceBrand
	if err := r.db.WithContext(ctx).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domaindisp.Brand, 0, len(ms))
	for _, m := range ms {
		out = append(out, domaindisp.Brand{ID: m.ID, Nombre: m.Nombre})
	}
	return out, nil
}

func (r *DispositivoCatalogRepository) FindBrandByID(ctx context.Context, id int32) (*domaindisp.Brand, error) {
	var m models.DeviceBrand
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domaindisp.Brand{ID: m.ID, Nombre: m.Nombre}, nil
}

func (r *DispositivoCatalogRepository) CreateBrand(ctx context.Context, nombre string) (*domaindisp.Brand, error) {
	m := &models.DeviceBrand{Nombre: nombre}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return &domaindisp.Brand{ID: m.ID, Nombre: m.Nombre}, nil
}

func (r *DispositivoCatalogRepository) UpdateBrand(ctx context.Context, id int32, nombre string) (*domaindisp.Brand, error) {
	if err := r.db.WithContext(ctx).Model(&models.DeviceBrand{}).Where("id = ?", id).Update("nombre", nombre).Error; err != nil {
		return nil, err
	}
	return r.FindBrandByID(ctx, id)
}

func (r *DispositivoCatalogRepository) DeleteBrand(ctx context.Context, id int32) error {
	return r.db.WithContext(ctx).Delete(&models.DeviceBrand{}, id).Error
}

func (r *DispositivoCatalogRepository) ListModels(ctx context.Context, brandID *int32) ([]domaindisp.Model, error) {
	q := r.db.WithContext(ctx).Model(&models.DeviceModel{})
	if brandID != nil {
		q = q.Where("brand_id = ?", brandID)
	}
	var ms []models.DeviceModel
	if err := q.Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]domaindisp.Model, 0, len(ms))
	for _, m := range ms {
		out = append(out, domaindisp.Model{ID: m.ID, Nombre: m.Nombre, BrandID: m.BrandID})
	}
	return out, nil
}

func (r *DispositivoCatalogRepository) CreateModel(ctx context.Context, nombre string, brandID int32) (*domaindisp.Model, error) {
	m := &models.DeviceModel{Nombre: nombre, BrandID: brandID}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return &domaindisp.Model{ID: m.ID, Nombre: m.Nombre, BrandID: m.BrandID}, nil
}

func (r *DispositivoCatalogRepository) UpdateModel(ctx context.Context, id int32, nombre string, brandID int32) (*domaindisp.Model, error) {
	if err := r.db.WithContext(ctx).Model(&models.DeviceModel{}).Where("id = ?", id).
		Updates(map[string]any{"nombre": nombre, "brand_id": brandID}).Error; err != nil {
		return nil, err
	}
	var m models.DeviceModel
	if err := r.db.WithContext(ctx).First(&m, id).Error; err != nil {
		return nil, err
	}
	return &domaindisp.Model{ID: m.ID, Nombre: m.Nombre, BrandID: m.BrandID}, nil
}

func (r *DispositivoCatalogRepository) DeleteModel(ctx context.Context, id int32) error {
	return r.db.WithContext(ctx).Delete(&models.DeviceModel{}, id).Error
}

// ── Read side ────────────────────────────────────────────────────────────────

type DispositivoReadRepository struct {
	db *gorm.DB
}

func NewDispositivoReadRepository(db *gorm.DB) *DispositivoReadRepository {
	return &DispositivoReadRepository{db: db}
}

var _ dispread.Queries = (*DispositivoReadRepository)(nil)

type dispositivoGlobalStats struct {
	WithIMEI   int64 `gorm:"column:with_imei"`
	WithNumero int64 `gorm:"column:with_numero"`
	Assigned   int64 `gorm:"column:assigned"`
	Unassigned int64 `gorm:"column:unassigned"`
}

func (r *DispositivoReadRepository) List(ctx context.Context, f dispread.Filters) (*dispread.ListResult, error) {
	q := r.db.WithContext(ctx).Model(&models.Dispositivo{})

	if f.Search != nil && *f.Search != "" {
		like := "%" + *f.Search + "%"
		q = q.Where("numero ILIKE ? OR imei ILIKE ? OR serie ILIKE ?", like, like, like)
	}
	if f.Estado != nil && *f.Estado != "" {
		q = q.Where("estado = ?", f.Estado)
	}
	if f.Grupo != nil && *f.Grupo != "" {
		q = q.Where("grupo = ?", f.Grupo)
	}
	if f.TrabajadorID != nil {
		q = q.Where("trabajador_id = ?", f.TrabajadorID)
	}

	var total int64
	q.Count(&total)

	var ms []models.Dispositivo
	if err := q.Preload("Brand").Preload("Model").Preload("Trabajador").
		Offset(int((f.Page - 1) * f.PerPage)).Limit(int(f.PerPage)).
		Order("id DESC").Find(&ms).Error; err != nil {
		return nil, err
	}

	var gs dispositivoGlobalStats
	r.db.WithContext(ctx).Raw(`
		SELECT
			COUNT(CASE WHEN imei IS NOT NULL THEN 1 END) AS with_imei,
			COUNT(CASE WHEN numero IS NOT NULL THEN 1 END) AS with_numero,
			COUNT(CASE WHEN trabajador_id IS NOT NULL THEN 1 END) AS assigned,
			COUNT(CASE WHEN trabajador_id IS NULL THEN 1 END) AS unassigned
		FROM dispositivos
	`).Scan(&gs)

	marcas, _ := r.listBrandSummaries(ctx)

	results := make([]dispread.Details, 0, len(ms))
	for i := range ms {
		results = append(results, mapDispositivoDetails(&ms[i]))
	}

	return &dispread.ListResult{
		Results: results,
		Total:   total,
		Page:    f.Page,
		PerPage: f.PerPage,
		Marcas:  marcas,
		Stats: dispread.ListStats{
			Total: total, WithIMEI: gs.WithIMEI, WithNumero: gs.WithNumero,
			Assigned: gs.Assigned, Unassigned: gs.Unassigned,
		},
	}, nil
}

func (r *DispositivoReadRepository) Details(ctx context.Context, id int32) (*dispread.Details, error) {
	var m models.Dispositivo
	if err := r.db.WithContext(ctx).Preload("Brand").Preload("Model").Preload("Trabajador").
		First(&m, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	d := mapDispositivoDetails(&m)
	return &d, nil
}

func (r *DispositivoReadRepository) listBrandSummaries(ctx context.Context) ([]dispread.BrandSummary, error) {
	var ms []models.DeviceBrand
	if err := r.db.WithContext(ctx).Find(&ms).Error; err != nil {
		return nil, err
	}
	out := make([]dispread.BrandSummary, 0, len(ms))
	for _, m := range ms {
		out = append(out, dispread.BrandSummary{ID: m.ID, Nombre: m.Nombre})
	}
	return out, nil
}

// ── Mappers ──────────────────────────────────────────────────────────────────

func mapDispositivo(m *models.Dispositivo) *domaindisp.Dispositivo {
	return &domaindisp.Dispositivo{
		ID: m.ID, Numero: m.Numero, IMEI: m.IMEI, Serie: m.Serie,
		SynergyResID: m.SynergyResID, BrandID: m.BrandID, ModelID: m.ModelID,
		Grupo: m.Grupo, Estado: domaindisp.Estado(m.Estado), TrabajadorID: m.TrabajadorID,
		Observaciones: m.Observaciones, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func domainToDispositivoModel(d *domaindisp.Dispositivo) *models.Dispositivo {
	return &models.Dispositivo{
		ID: d.ID, Numero: d.Numero, IMEI: d.IMEI, Serie: d.Serie,
		SynergyResID: d.SynergyResID, BrandID: d.BrandID, ModelID: d.ModelID,
		Grupo: d.Grupo, Estado: string(d.Estado), TrabajadorID: d.TrabajadorID,
		Observaciones: d.Observaciones, CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	}
}

func mapDispositivoDetails(m *models.Dispositivo) dispread.Details {
	d := dispread.Details{
		ID: m.ID, Numero: m.Numero, IMEI: m.IMEI, Serie: m.Serie,
		SynergyResID: m.SynergyResID, BrandID: m.BrandID, ModelID: m.ModelID,
		Grupo: m.Grupo, Estado: m.Estado, Observaciones: m.Observaciones,
		TrabajadorID: m.TrabajadorID, CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
	if m.Brand != nil {
		d.Marca = &m.Brand.Nombre
	}
	if m.Model != nil {
		d.Modelo = &m.Model.Nombre
	}
	if m.Trabajador != nil {
		d.Employee = &dispread.EmployeeDetails{
			ID:             m.Trabajador.ID,
			NombreCompleto: fmt.Sprintf("%s %s", m.Trabajador.Nombre, m.Trabajador.Apellido),
			Email:          m.Trabajador.Email,
		}
		d.EmployeeSynergyResID = m.Trabajador.SynergyResID
	}
	return d
}
