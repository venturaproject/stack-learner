package auth

import "errors"

var (
	ErrNotFound        = errors.New("usuario no encontrado")
	ErrInvalidPassword = errors.New("credenciales inválidas")
	ErrTokenInvalid    = errors.New("token inválido o expirado")
	ErrTokenRevoked    = errors.New("token revocado")
)
