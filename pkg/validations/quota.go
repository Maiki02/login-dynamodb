package validations

import (
	"time"
)

// QuotaData es una interfaz para desacoplar la validación de la estructura concreta del request.
// Cualquier struct que implemente estos métodos puede ser validado.
type QuotaData interface {
	GetExpirationDate() time.Time
	GetAmountCents() int64
	GetCoin() string
}
