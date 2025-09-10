package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PaymentStatus string
type PaymentMethod string

const (
	PaymentCompleted PaymentStatus = "completed"
	PaymentReverted  PaymentStatus = "reverted"
)

const (
	MethodCash       PaymentMethod = "cash"
	MethodTransfer   PaymentMethod = "transfer"
	MethodDebitCard  PaymentMethod = "debit-card"
	MethosCreditCard PaymentMethod = "credit-card"
	MethodOther      PaymentMethod = "other"
)

// AffectedQuota detalla qué monto de un pago se aplicó a una cuota específica.
type AffectedQuota struct {
	QuotaID            primitive.ObjectID `json:"quota_id" bson:"quota_id"`
	AmountAppliedCents int64              `json:"amount_applied_cents" bson:"amount_applied_cents"`
}

// Payment representa una transacción de pago única.
type Payment struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SaleID         primitive.ObjectID `json:"sale_id" bson:"sale_id"`                               // Venta a la que pertenece el pago
	CollectorID    primitive.ObjectID `json:"collector_id,omitempty" bson:"collector_id,omitempty"` // Quién cobró
	PaymentNumber  int64              `json:"payment_number" bson:"payment_number"`                 // Número de pago secuencial
	PaymentDate    time.Time          `json:"payment_date" bson:"payment_date"`                     // Fecha en que se realizó el pago
	AmountCents    int64              `json:"amount_cents" bson:"amount_cents"`                     // Monto total de la transacción de pago
	Coin           string             `json:"coin" bson:"coin"`                                     // Moneda en la que se realizó el pago
	Method         PaymentMethod      `json:"method" bson:"method"`                                 // Método de pago
	Status         PaymentStatus      `json:"status" bson:"status"`
	QuotasAffected []AffectedQuota    `json:"quotas_affected" bson:"quotas_affected"` // Desglose de cómo se usó el dinero
	Notes          string             `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}
