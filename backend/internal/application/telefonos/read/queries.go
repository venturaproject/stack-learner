// Package read holds the query side of the telefonos module: the read models
// (query projections — no domain behaviour) and the Queries port implemented by
// persistence. These types are also the JSON contract returned by the HTTP
// handlers, hence the json tags.
package read

import (
	"context"
	"time"
)

type WorkerDetails struct {
	ID              int32   `json:"id"`
	NombreCompleto  string  `json:"nombre_completo"`
	TelefonoSynergy *string `json:"telefono"`
}

type StatusDetails struct {
	ID     int32   `json:"id"`
	Nombre string  `json:"nombre"`
	Color  *string `json:"color"`
}

type Actuacion struct {
	ID          int32     `json:"id"`
	Descripcion string    `json:"descripcion"`
	Fecha       time.Time `json:"fecha"`
	UserID      *int32    `json:"user_id"`
}

type Details struct {
	ID            int32          `json:"id"`
	Numero        *string        `json:"numero"`
	Plan          *string        `json:"plan"`
	Nplan         *string        `json:"nplan"`
	Tipo          *string        `json:"tipo"`
	Pin           *string        `json:"pin"`
	Puk           *string        `json:"puk"`
	IMEI          *string        `json:"imei"`
	IMEI2         *string        `json:"imei2"`
	Marca         *string        `json:"marca"`
	Modelo        *string        `json:"modelo"`
	Linea         *string        `json:"linea"`
	Operadora     *string        `json:"operadora"`
	EstadoID      *int32         `json:"estado_id"`
	TipologiaID   *int32         `json:"tipologia_id"`
	Observaciones string         `json:"observaciones"`
	Activo        bool           `json:"activo"`
	Status        *StatusDetails `json:"status"`
	Trabajador    *WorkerDetails `json:"trabajador"`
	Actuaciones   []Actuacion    `json:"actuaciones"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

type EstadoSummary struct {
	ID     int32   `json:"id"`
	Nombre string  `json:"nombre"`
	Color  *string `json:"color"`
}

type TipologiaSummary struct {
	ID     int32  `json:"id"`
	Nombre string `json:"nombre"`
}

type Filters struct {
	Search  *string
	Status  *string
	Plan    *string
	Page    int64
	PerPage int64
}

type ListStats struct {
	Total      int64 `json:"total"`
	Activos    int64 `json:"activos"`
	Congelados int64 `json:"congelados"`
	Bajas      int64 `json:"bajas"`
	SinDatos   int64 `json:"sin_datos"`
}

type ListResult struct {
	Results []Details       `json:"results"`
	Total   int64           `json:"count"`
	Page    int64           `json:"page"`
	PerPage int64           `json:"per_page"`
	Estados []EstadoSummary `json:"estados"`
	Stats   ListStats       `json:"stats"`
}

// Queries is the read-side port. Not-found detail lookups return (nil, nil).
type Queries interface {
	List(ctx context.Context, f Filters) (*ListResult, error)
	Details(ctx context.Context, id int32) (*Details, error)
	Estados(ctx context.Context) ([]EstadoSummary, error)
	Tipologias(ctx context.Context) ([]TipologiaSummary, error)
}
