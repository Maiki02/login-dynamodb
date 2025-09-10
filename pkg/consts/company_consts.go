package consts

// Definimos los estados que puede tener una entidad.
// Usamos iota para asignar valores incrementales automáticamente (0, 1, 2...).
const (
	STATUS_INACTIVE int32 = iota // 0
	STATUS_ACTIVE                // 1
	STATUS_PENDING               // 2 (Para el proceso de registro atómico)
)
