// Package read holds the query side of the trabajadores module: read models and
// the Queries port. These types are the JSON contract returned by the handlers.
package read

import "context"

type Details struct {
	ID              int32   `json:"id"`
	SynergyResID    *int32  `json:"synergy_res_id"`
	NombreCompleto  string  `json:"nombre_completo"`
	Nombre          string  `json:"nombre"`
	Apellido        string  `json:"apellido"`
	Email           *string `json:"email"`
	Ubicacion       *string `json:"ubicacion"`
	Ciudad          *string `json:"ciudad"`
	Loc             *string `json:"loc"`
	TelefonoSynergy *string `json:"telefono"`
	IMEI            *string `json:"imei"`
	TelefonosID     *int32  `json:"telefonos_id"`
	EmpStat         *string `json:"emp_stat"`
	Activo          bool    `json:"activo"`
	Estado          string  `json:"estado"`
	FechaAlta       *string `json:"fecha_alta"`
	FechaBaja       *string `json:"fecha_baja"`
	Cargo           *string `json:"cargo"`
	Departamento    *string `json:"departamento"`
}

type Filters struct {
	Search    *string
	Activo    *string
	Ubicacion *string
	Ciudad    *string
	DateFrom  *string
	DateTo    *string
	Page      int64
	PerPage   int64
}

type ListStats struct {
	Activos             int64 `json:"activos"`
	ConTelefono         int64 `json:"con_telefono"`
	IncorporadosEsteMes int64 `json:"incorporados_este_mes"`
	ConIMEI             int64 `json:"con_imei"`
}

type ListResult struct {
	Results []Details `json:"results"`
	Total   int64     `json:"count"`
	Page    int64     `json:"page"`
	PerPage int64     `json:"per_page"`
	Stats   ListStats `json:"stats"`
}

// Queries is the read-side port. Not-found detail lookups return (nil, nil).
type Queries interface {
	List(ctx context.Context, f Filters) (*ListResult, error)
	Details(ctx context.Context, id int32) (*Details, error)
}
