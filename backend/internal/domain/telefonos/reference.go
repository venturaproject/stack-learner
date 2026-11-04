package telefonos

// Estado reference-data IDs — the rows seeded into the estados_telefono catalog
// table. Kept here so the single source of truth for these magic numbers lives
// in the domain, not scattered through SQL in the infrastructure layer.
const (
	EstadoActivoID    int32 = 1
	EstadoCongeladoID int32 = 2
	EstadoBajaID      int32 = 3
	EstadoSinDatosID  int32 = 5
)
