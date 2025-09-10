package request

// Definimos un tipo personalizado para nuestra clave de contexto.
type ContextKey string

// UserContextKey es la clave que usaremos para almacenar y recuperar el userID del contexto.
// Al estar en un paquete compartido, nos aseguramos de que sea siempre la misma clave.
const UserContextKey = ContextKey("userID")
