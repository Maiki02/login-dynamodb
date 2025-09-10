package dto

import (
	"myproject/internal/models"
	"myproject/pkg/structures"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//----------- SALES REPORTS -----------\\

// SaleInfoForReport contiene datos básicos de una venta para incluir en el reporte.
type SaleInfoForReport struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	SaleNumber int64              `json:"sale_number" bson:"sale_number"`
	SaleDate   time.Time          `json:"sale_date" bson:"sale_date"`
}

// QuotaInfoForReport representa una cuota pendiente con información de su venta.
type QuotaInfoForReport struct {
	ID              primitive.ObjectID `json:"_id" bson:"_id"`
	QuotaNumber     int                `json:"quota_number" bson:"quota_number"`
	ExpirationDate  time.Time          `json:"expiration_date" bson:"expiration_date"`
	AmountCents     int64              `json:"amount_cents" bson:"amount_cents"`
	Coin            string             `json:"coin" bson:"coin"` // Campo nuevo
	PaidAmountCents int64              `json:"paid_amount_cents" bson:"paid_amount_cents"`
	Status          models.QuotaStatus `json:"status" bson:"status"`
}

// ClientInfoForReport contiene datos básicos de un cliente para incluir en el reporte.
type ClientInfoForReport struct {
	ID             primitive.ObjectID        `json:"_id" bson:"_id"`
	Name           string                    `json:"name" bson:"name"`
	LastName       string                    `json:"last_name" bson:"last_name"`
	Identification structures.Identification `json:"identification" bson:"identification"`
}

// ClientSalesWithQuotasResponse es la nueva estructura de respuesta principal.
// Reemplaza a ClientPendingQuotasResponse.
type ClientSalesWithQuotasResponse struct {
	ClientInfo ClientInfoForReport  `json:"client_info" bson:"client_info"`
	SaleInfo   SaleInfoForReport    `json:"sale_info" bson:"sale_info"`
	Quotas     []QuotaInfoForReport `json:"quotas" bson:"quotas"`
}

//------------ PAYMENTS REPORTS ------------\\

type AffectedQuotaReport struct {
	Quota              QuotaInfoForReport `json:"quota" bson:"quota"`                               // Información de la cuota afectada
	AmountAppliedCents int64              `json:"amount_applied_cents" bson:"amount_applied_cents"` // Monto aplicado a las cuotas
}

// PaymentInfoForReport representa la información detallada de un pago.
type PaymentInfoForReport struct {
	ID             primitive.ObjectID    `json:"_id" bson:"_id"`
	PaymentNumber  int64                 `json:"payment_number" bson:"payment_number"`
	PaymentDate    time.Time             `json:"payment_date" bson:"payment_date"`
	AmountCents    int64                 `json:"amount_cents" bson:"amount_cents"`
	Coin           string                `json:"coin" bson:"coin"` // Campo nuevo para la moneda del pago
	Method         models.PaymentMethod  `json:"method" bson:"method"`
	Status         models.PaymentStatus  `json:"status" bson:"status"`
	AffectedQuotas []AffectedQuotaReport `json:"affected_quotas" bson:"affected_quotas"` // Lista de cuotas afectadas por este pago
}

// PaymentsReportResponse es la estructura de respuesta para el nuevo reporte de pagos.
type PaymentsReportResponse struct {
	ClientInfo ClientInfoForReport  `json:"client_info" bson:"client_info"`
	SaleInfo   SaleInfoForReport    `json:"sale_info" bson:"sale_info"`
	Payment    PaymentInfoForReport `json:"payment" bson:"payment"`
}

// ------------ PRODUCTS REPORTS ------------\\
// ProductResponse es un DTO para enviar una respuesta enriquecida del producto,
// incluyendo los detalles de la marca y el tipo de producto.
type ProductResponse struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	InternalCode int64              `json:"internal_code,omitempty" bson:"internal_code,omitempty"`
	Name         string             `json:"name" bson:"name"`
	Slug         string             `json:"slug" bson:"slug"`

	Description string `json:"description,omitempty" bson:"description,omitempty"`

	UnitOfMeasure string           `json:"unit_of_measure" bson:"unit_of_measure"`
	Status        string           `json:"status" bson:"status"`
	Variants      []models.Variant `json:"variants" bson:"variants"`
	CreatedAt     time.Time        `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at" bson:"updated_at"`

	// Campos "poblados" con $lookup
	Brand       *models.Brand       `json:"brand,omitempty" bson:"brand,omitempty"`
	ProductType *models.ProductType `json:"product_type,omitempty" bson:"product_type,omitempty"`
}
