package dispositivos

import "context"

// Repository is the write side of the Dispositivo aggregate. Catalog
// (brands/models) lives in its own CatalogRepository; reads go through
// internal/application/dispositivos/read.Queries.
type Repository interface {
	FindByID(ctx context.Context, id int32) (*Dispositivo, error) // ErrNotFound when absent
	Save(ctx context.Context, d *Dispositivo) error               // insert (ID==0) or update
	Delete(ctx context.Context, id int32) error
}

// Brand and Model are the device catalog entities. They have no aggregate
// behaviour, so they are plain records managed through CatalogRepository.
type Brand struct {
	ID     int32
	Nombre string
}

type Model struct {
	ID      int32
	Nombre  string
	BrandID int32
}

type CatalogRepository interface {
	ListBrands(ctx context.Context) ([]Brand, error)
	FindBrandByID(ctx context.Context, id int32) (*Brand, error)
	CreateBrand(ctx context.Context, nombre string) (*Brand, error)
	UpdateBrand(ctx context.Context, id int32, nombre string) (*Brand, error)
	DeleteBrand(ctx context.Context, id int32) error

	ListModels(ctx context.Context, brandID *int32) ([]Model, error)
	CreateModel(ctx context.Context, nombre string, brandID int32) (*Model, error)
	UpdateModel(ctx context.Context, id int32, nombre string, brandID int32) (*Model, error)
	DeleteModel(ctx context.Context, id int32) error
}
