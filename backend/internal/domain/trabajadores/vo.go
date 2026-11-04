package trabajadores

import "fmt"

type Estado string

const (
	EstadoActivo Estado = "activo"
	EstadoBaja   Estado = "baja"
)

func ParseEstado(s string) (Estado, error) {
	switch Estado(s) {
	case EstadoActivo, EstadoBaja:
		return Estado(s), nil
	default:
		return "", fmt.Errorf("estado de trabajador inválido: %q", s)
	}
}

func (e Estado) IsActivo() bool     { return e == EstadoActivo }
func (e Estado) CanDarDeBaja() bool { return e == EstadoActivo }
func (e Estado) String() string     { return string(e) }
