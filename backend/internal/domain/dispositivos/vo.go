package dispositivos

import "fmt"

type Estado string

const (
	EstadoDisponible Estado = "disponible"
	EstadoAsignado   Estado = "asignado"
)

func ParseEstado(s string) (Estado, error) {
	switch Estado(s) {
	case EstadoDisponible, EstadoAsignado:
		return Estado(s), nil
	default:
		return "", fmt.Errorf("estado de dispositivo inválido: %q", s)
	}
}

func (e Estado) IsAsignable() bool { return e == EstadoDisponible }
func (e Estado) String() string    { return string(e) }
