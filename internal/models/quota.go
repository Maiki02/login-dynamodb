package models

import (
	"myproject/pkg/validations"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// QuotaStatus represents the possible statuses of a quota.
type QuotaStatus string

const (
	QuotaPending    QuotaStatus = "pendiente"
	QuotaPaid       QuotaStatus = "pagada"
	QuotaOverdue    QuotaStatus = "vencida"   // Vencida
	QuotaDelinquent QuotaStatus = "en_mora"   // En mora
	QuotaCancelled  QuotaStatus = "cancelada" // Cancelada y no se utiliza para calculos
)

// Quota represents one of the payments for a sale.
// This should be stored in its own "quotas" collection.
type Quota struct {
	ID              primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	SaleID          primitive.ObjectID   `json:"sale_id" bson:"sale_id"`
	PaymentsIDs     []primitive.ObjectID `json:"payment_ids,omitempty" bson:"payment_ids,omitempty"`
	QuotaNumber     int                  `json:"quota_number" bson:"quota_number"`
	ExpirationDate  time.Time            `json:"expiration_date" bson:"expiration_date"`
	AmountCents     int64                `json:"amount_cents" bson:"amount_cents"` // Costo de la cuota en centavos
	Coin            string               `json:"coin" bson:"coin"`
	PaidAmountCents int64                `json:"paid_amount_cents,omitempty" bson:"paid_amount_cents"` // Monto pagado en centavos
	Status          QuotaStatus          `json:"status" bson:"status"`
}

func (q *Quota) IsPaid() bool {
	return q.Status == QuotaPaid
}

func (q *Quota) GetExpirationDate() time.Time {
	return q.ExpirationDate
}

func (q *Quota) GetAmountCents() int64 {
	return q.AmountCents
}

func (q *Quota) GetCoin() string {
	return q.Coin
}

func (q *Quota) ValidateQuota() error {
	// 1. Validar que la fecha de expiración no sea anterior a la de hoy.
	// Normalizamos la fecha de hoy a medianoche para comparar solo las fechas, ignorando la hora.
	/*today := time.Now().Truncate(24 * time.Hour)
	if q.GetExpirationDate().Before(today) {
		return validations.ErrInvalidExpirationDate
	}*/
	//Si puede ser anterior a la de hoy

	// 2. Validar que el monto no sea negativo.
	if q.GetAmountCents() < 0 {
		return validations.ErrNegativeAmount
	}

	// 3. Validar que la moneda no tenga más de 10 caracteres.
	if len(q.GetCoin()) > 10 {
		return validations.ErrCoinTooLong
	}

	return nil
}
