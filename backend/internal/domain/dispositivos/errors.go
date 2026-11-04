package dispositivos

import "boilerplate-learner/internal/shared/apperrors"

var (
	ErrNotFound       = apperrors.NotFound("dispositivo no encontrado")
	ErrYaAsignado     = apperrors.Conflict("el dispositivo ya está asignado a un trabajador")
	ErrEstadoInvalido = apperrors.Invalid("estado de dispositivo inválido")
)
