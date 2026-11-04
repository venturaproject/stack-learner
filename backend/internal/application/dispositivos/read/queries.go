// Package read holds the query side of the dispositivos module: read models and
// the Queries port. These types are the JSON contract returned by the handlers.
package read

import (
	"context"
	"time"
)

type EmployeeDetails struct {
	ID             int32   `json:"id"`
	NombreCompleto string  `json:"nombre_completo"`
	Email          *string `json:"email"`
}

type Details struct {
	ID                   int32            `json:"id"`
	Numero               *string          `json:"numero"`
	IMEI                 *string          `json:"imei"`
	Serie                *string          `json:"serie"`
	SynergyResID         *int32           `json:"synergy_res_id"`
	Marca                *string          `json:"marca"`
	Modelo               *string          `json:"modelo"`
	BrandID              *int32           `json:"brand_id"`
	ModelID              *int32           `json:"model_id"`
	Grupo                *string          `json:"grupo"`
	Estado               string           `json:"estado"`
	Observaciones        string           `json:"observaciones"`
	EmployeeSynergyResID *int32           `json:"employee_synergy_res_id"`
	Employee             *EmployeeDetails `json:"employee"`
	TrabajadorID         *int32           `json:"trabajador_id"`
	CreatedAt            time.Time        `json:"created_at"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

type BrandSummary struct {
	ID     int32  `json:"id"`
	Nombre string `json:"nombre"`
}

type ModelSummary struct {
	ID      int32  `json:"id"`
	Nombre  string `json:"nombre"`
	BrandID int32  `json:"brand_id"`
}

type Filters struct {
	Search       *string
	Marca        *string
	Tipo         *string
	Grupo        *string
	Estado       *string
	TrabajadorID *int32
	Page         int64
	PerPage      int64
}

type ListStats struct {
	Total      int64 `json:"total"`
	WithIMEI   int64 `json:"with_imei"`
	WithNumero int64 `json:"with_numero"`
	Assigned   int64 `json:"assigned"`
	Unassigned int64 `json:"unassigned"`
}

type ListResult struct {
	Results []Details      `json:"results"`
	Total   int64          `json:"count"`
	Page    int64          `json:"page"`
	PerPage int64          `json:"per_page"`
	Marcas  []BrandSummary `json:"marcas"`
	Stats   ListStats      `json:"stats"`
}

// Queries is the read-side port. Not-found detail lookups return (nil, nil).
type Queries interface {
	List(ctx context.Context, f Filters) (*ListResult, error)
	Details(ctx context.Context, id int32) (*Details, error)
}
