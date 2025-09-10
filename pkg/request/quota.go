package request

import (
	"myproject/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type QuotaRequest struct {
	ExpirationDate time.Time `json:"expiration_date" validate:"required"`
	AmountCents    int64     `json:"amount_cents" validate:"required,gt=0"`
	Coin           string    `json:"coin" validate:"required"`
}

// QuotaUpdate define la información para actualizar una única cuota.
type QuotaUpdate struct {
	QuotaID           primitive.ObjectID `json:"quota_id" validate:"required"`
	NewExpirationDate time.Time          `json:"new_expiration_date" validate:"required"`
	NewStatus         models.QuotaStatus `json:"new_status" validate:"required"`
}

// RescheduleQuotasRequest es el cuerpo de la petición para reprogramar cuotas.
type RescheduleQuotasRequest struct {
	Updates []QuotaUpdate `json:"updates" validate:"required,min=1,dive"`
}
