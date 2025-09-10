package request

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PayQuotasRequest struct {
	QuotasIDs []primitive.ObjectID `json:"quotas_ids" validate:"required,min=1"`
	Method    string               `json:"method" validate:"required"`
}

// PaymentDetail representa una sola acción de pago por parte de un cobrador.
type PaymentDetail struct {
	PaymentDate time.Time `json:"payment_date" validate:"required"`
	AmountCents int64     `json:"amount_cents" validate:"required,gt=0"`
	Method      string    `json:"method" validate:"required"`
}

// SequentialPaymentRequest es el cuerpo de la petición para crear un pago secuencial.
// Contiene un arreglo de pagos individuales.
type SequentialPaymentRequest struct {
	Payments []PaymentDetail `json:"payments" validate:"required,min=1,dive"`
	Notes    string          `json:"notes,omitempty"`
}

// FilterPaymentsRequest contiene los parámetros para filtrar pagos.
// Usamos punteros para poder diferenciar entre un valor no provisto (nil) y un valor vacío.
type FilterPaymentsRequest struct {
	StartDate *time.Time
	EndDate   *time.Time
	Statuses  []string
}
