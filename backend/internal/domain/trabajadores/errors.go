package trabajadores

import "boilerplate-learner/internal/shared/apperrors"

var (
	ErrNotFound          = apperrors.NotFound("trabajador no encontrado")
	ErrYaBaja            = apperrors.Conflict("el trabajador ya está dado de baja")
	ErrEstadoInvalido    = apperrors.Invalid("estado de trabajador inválido")
	ErrNombreObligatorio = apperrors.Invalid("el nombre es obligatorio")
)
