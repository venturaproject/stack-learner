package telefonos

import "boilerplate-learner/internal/shared/apperrors"

// Domain errors. They carry an apperrors.Kind so the HTTP layer maps them to a
// status code without string matching.
var (
	ErrNotFound   = apperrors.NotFound("teléfono no encontrado")
	ErrYaAsignado = apperrors.Conflict("el teléfono ya está asignado a otro trabajador")
	ErrInactivo   = apperrors.Invalid("el teléfono está inactivo")
)
