package persistence

import (
	"context"
	"time"

	appdash "boilerplate-learner/internal/application/dashboard"
	domaintelf "boilerplate-learner/internal/domain/telefonos"
	"boilerplate-learner/internal/infrastructure/persistence/models"

	"gorm.io/gorm"
)

// lastSixMonths returns the trailing 6 months (oldest first) as YYYY-MM points,
// filling gaps with 0.
func lastSixMonths(counts map[string]int64) []appdash.MonthPoint {
	now := time.Now()
	out := make([]appdash.MonthPoint, 0, 6)
	for i := 5; i >= 0; i-- {
		m := now.AddDate(0, -i, 0).Format("2006-01")
		out = append(out, appdash.MonthPoint{Month: m, Incorporations: counts[m]})
	}
	return out
}

type DashboardReadRepository struct {
	db *gorm.DB
}

func NewDashboardReadRepository(db *gorm.DB) *DashboardReadRepository {
	return &DashboardReadRepository{db: db}
}

var _ appdash.Queries = (*DashboardReadRepository)(nil)

func (r *DashboardReadRepository) Summary(ctx context.Context) (*appdash.Summary, error) {
	db := r.db.WithContext(ctx)
	var s appdash.Summary

	// Totals
	db.Model(&models.Trabajador{}).Count(&s.Totals.Trabajadores)
	db.Model(&models.Telefono{}).Count(&s.Totals.Telefonos)
	db.Model(&models.Dispositivo{}).Count(&s.Totals.Dispositivos)
	db.Model(&models.User{}).Count(&s.Totals.Users)

	// Worker stats (one aggregation with a LATERAL join to the active phone).
	db.Raw(`
		SELECT
			COUNT(CASE WHEN t.estado = 'activo' THEN 1 END)                                                     AS activos,
			COUNT(DISTINCT _tel.tel_id)                                                                          AS con_telefono,
			COUNT(CASE WHEN date_trunc('month', t.fecha_incorporacion) = date_trunc('month', NOW()) THEN 1 END)  AS incorporados_este_mes,
			COUNT(DISTINCT CASE WHEN _tel.tel_imei IS NOT NULL THEN t.id END)                                    AS con_imei
		FROM trabajadores t
		LEFT JOIN LATERAL (
			SELECT id AS tel_id, imei AS tel_imei
			FROM telefonos WHERE trabajador_id = t.id AND activo = true LIMIT 1
		) _tel ON true
	`).Scan(&s.WorkerStats)

	// Phone stats
	db.Raw(`
		SELECT
			COUNT(*) AS total,
			COUNT(CASE WHEN estado_id = ? THEN 1 END) AS activos,
			COUNT(CASE WHEN estado_id = ? THEN 1 END) AS congelados,
			COUNT(CASE WHEN estado_id = ? THEN 1 END) AS bajas,
			COUNT(CASE WHEN estado_id = ? OR estado_id IS NULL THEN 1 END) AS sin_datos
		FROM telefonos
	`, domaintelf.EstadoActivoID, domaintelf.EstadoCongeladoID,
		domaintelf.EstadoBajaID, domaintelf.EstadoSinDatosID).Scan(&s.PhoneStats)

	// Recent devices (5 most recently updated)
	var devices []models.Dispositivo
	db.Preload("Brand").Preload("Model").Preload("Trabajador").
		Order("updated_at DESC").Limit(5).Find(&devices)
	s.RecentDevices = make([]appdash.RecentDevice, 0, len(devices))
	for i := range devices {
		d := &devices[i]
		rd := appdash.RecentDevice{
			ID: d.ID, Numero: d.Numero, Grupo: d.Grupo,
			UpdatedAt: d.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
		if d.Brand != nil {
			rd.Marca = &d.Brand.Nombre
		}
		if d.Model != nil {
			rd.Modelo = &d.Model.Nombre
		}
		if d.Trabajador != nil {
			name := d.Trabajador.Nombre + " " + d.Trabajador.Apellido
			rd.EmployeeNombre = &name
		}
		s.RecentDevices = append(s.RecentDevices, rd)
	}

	// Monthly worker incorporations — last 6 calendar months, zero-filled.
	type monthRow struct {
		Month string
		N     int64
	}
	var rows []monthRow
	db.Raw(`
		SELECT to_char(date_trunc('month', fecha_incorporacion), 'YYYY-MM') AS month, COUNT(*) AS n
		FROM trabajadores
		WHERE fecha_incorporacion >= date_trunc('month', NOW()) - INTERVAL '5 months'
		GROUP BY 1
	`).Scan(&rows)
	byMonth := make(map[string]int64, len(rows))
	for _, r := range rows {
		byMonth[r.Month] = r.N
	}
	s.MonthlyWorkers = lastSixMonths(byMonth)

	return &s, nil
}
