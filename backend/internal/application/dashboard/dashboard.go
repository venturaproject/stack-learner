// Package dashboard is the application layer for the read-only dashboard
// summary. It owns no aggregate; it exposes a Queries port implemented by
// persistence so the HTTP handler never touches the database directly.
package dashboard

import "context"

type WorkerStats struct {
	Activos             int64 `json:"activos"`
	ConTelefono         int64 `json:"con_telefono"`
	IncorporadosEsteMes int64 `json:"incorporados_este_mes"`
	ConIMEI             int64 `json:"con_imei"`
}

type PhoneStats struct {
	Total      int64 `json:"total"`
	Activos    int64 `json:"activos"`
	Congelados int64 `json:"congelados"`
	Bajas      int64 `json:"bajas"`
	SinDatos   int64 `json:"sin_datos"`
}

type RecentDevice struct {
	ID             int32   `json:"id"`
	Marca          *string `json:"marca"`
	Modelo         *string `json:"modelo"`
	Numero         *string `json:"numero"`
	Grupo          *string `json:"grupo"`
	EmployeeNombre *string `json:"employee_nombre"`
	UpdatedAt      string  `json:"updated_at"`
}

type MonthPoint struct {
	Month          string `json:"month"` // YYYY-MM
	Incorporations int64  `json:"incorporations"`
}

type Summary struct {
	WorkerStats    WorkerStats    `json:"worker_stats"`
	PhoneStats     PhoneStats     `json:"phone_stats"`
	RecentDevices  []RecentDevice `json:"recent_devices"`
	MonthlyWorkers []MonthPoint   `json:"monthly_workers"`
	// Totals kept for backwards compatibility with simple cards.
	Totals struct {
		Trabajadores int64 `json:"trabajadores"`
		Telefonos    int64 `json:"telefonos"`
		Dispositivos int64 `json:"dispositivos"`
		Users        int64 `json:"users"`
	} `json:"totals"`
}

type Queries interface {
	Summary(ctx context.Context) (*Summary, error)
}

type GetUseCase struct{ queries Queries }

func NewGetUseCase(q Queries) *GetUseCase { return &GetUseCase{queries: q} }

func (uc *GetUseCase) Execute(ctx context.Context) (*Summary, error) {
	return uc.queries.Summary(ctx)
}
