package trabajadores

import "context"

// Repository is the write side of the aggregate. Reads go through
// internal/application/trabajadores/read.Queries.
type Repository interface {
	FindByID(ctx context.Context, id int32) (*Trabajador, error) // ErrNotFound when absent
	Save(ctx context.Context, t *Trabajador) error               // insert (ID==0) or update
	Delete(ctx context.Context, id int32) error
}
