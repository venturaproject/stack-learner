package apiclients

import "errors"

var (
	ErrNotFound           = errors.New("api client no encontrado")
	ErrInvalidCredentials = errors.New("credenciales de api client inválidas")
	ErrRevoked            = errors.New("api client revocado")
	ErrTokenExpired       = errors.New("token expirado")
	ErrTokenInvalid       = errors.New("token inválido")
)
