package telefonos

import "context"

// Repository is the write side of the aggregate. Reads for list/detail views go
// through internal/application/telefonos/read.Queries.
type Repository interface {
	// FindByID loads the aggregate. It returns ErrNotFound when the row is absent.
	FindByID(ctx context.Context, id int32) (*Telefono, error)
	// Save inserts (ID == 0) or updates the aggregate, setting ID on insert.
	Save(ctx context.Context, t *Telefono) error
	Delete(ctx context.Context, id int32) error
}

// ActuacionRepository manages the Actuacion sub-entities of a Telefono.
type ActuacionRepository interface {
	Add(ctx context.Context, telefonoID int32, descripcion string, userID *int32) (*Actuacion, error)
	Remove(ctx context.Context, id int32) error
}
