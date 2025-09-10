package models

import (
	"myproject/pkg/structures"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SaleStatus represents the possible statuses of a sale.
type SaleStatus string

const (
	StatusPendingApproval SaleStatus = "pendiente_aprobacion"
	StatusPendingShipment SaleStatus = "pendiente_envio"
	StatusInProgress      SaleStatus = "en_progreso"
	StatusCompleted       SaleStatus = "completada"
	StatusCancelled       SaleStatus = "cancelada"
	StatusDelinquent      SaleStatus = "en_mora" // En mora
)

// SoldProduct representa un ítem dentro de una venta.
// Almacena una instantánea (snapshot) de la información relevante del producto
// y su variante en el momento exacto de la transacción.
type SoldProduct struct {
	// --- Referencias al Catálogo Original ---
	ProductID    primitive.ObjectID `json:"product_id" bson:"product_id"`
	VariantSKU   string             `json:"variant_sku" bson:"variant_sku"`
	InternalCode int64              `json:"internal_code,omitempty" bson:"internal_code,omitempty"`

	// --- Snapshot de Datos Descriptivos ---
	Name            string                 `json:"name" bson:"name"`
	BrandName       string                 `json:"brand_name,omitempty" bson:"brand_name,omitempty"`
	ProductTypeName string                 `json:"product_type_name,omitempty" bson:"product_type_name,omitempty"`
	Attributes      map[string]interface{} `json:"attributes" bson:"attributes"`

	// --- Snapshot de Datos Transaccionales (Contables) ---
	Quantity      int               `json:"quantity" bson:"quantity"`
	UnitOfMeasure string            `json:"unit_of_measure" bson:"unit_of_measure"`
	UnitPrice     structures.Money  `json:"unit_price" bson:"unit_price"`
	UnitCost      *structures.Money `json:"unit_cost,omitempty" bson:"unit_cost,omitempty"`
	SubtotalCents int64             `json:"subtotal_cents" bson:"subtotal_cents"`

	// --- Información Logística ---
	DeliveryDate  time.Time           `json:"delivery_date,omitempty" bson:"delivery_date,omitempty"`
	DeliveryPlace *structures.Address `json:"delivery_place,omitempty" bson:"delivery_place,omitempty"`
}

// Sale represents the complete commercial transaction.
// This should be stored in a "sales" collection.
type Sale struct {
	ID                   primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty"`
	SaleNumber           int64                `json:"sale_number" bson:"sale_number"`
	ClientID             primitive.ObjectID   `json:"client_id" bson:"client_id"`
	ShippingAddress      structures.Address   `json:"shipping_address,omitempty" bson:"shipping_address,omitempty"`
	BillingAddress       structures.Address   `json:"billing_address,omitempty" bson:"billing_address,omitempty"`
	Products             []SoldProduct        `json:"products" bson:"products"`
	Loan                 *Loan                `json:"loan,omitempty" bson:"loan,omitempty"`
	QuotaIDs             []primitive.ObjectID `json:"quota_ids" bson:"quota_ids"` // References to the quotas
	PaymentIDs           []primitive.ObjectID `json:"payment_ids,omitempty" bson:"payment_ids,omitempty"`
	Status               SaleStatus           `json:"status" bson:"status"`
	TotalAmountCents     int64                `json:"total_amount" bson:"total_amount"`
	CollectedAmountCents int64                `json:"collected_amount_cents" bson:"collected_amount_cents"`
	PendingAmountCents   int64                `json:"pending_amount_cents" bson:"pending_amount_cents"`
	QuotaCount           int                  `json:"quota_count" bson:"quota_count"`
	InterestPercentage   int64                `json:"interest_percentage,omitempty" bson:"interest_percentage,omitempty"` //En porcentaje
	SaleDate             time.Time            `json:"sale_date" bson:"sale_date"`
	EstimatedEndDate     time.Time            `json:"estimated_end_date,omitempty" bson:"estimated_end_date,omitempty"`
	CreatedAt            time.Time            `json:"created_at" bson:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type SaleResponse struct {
	ID               primitive.ObjectID `json:"_id" bson:"_id"`
	SaleNumber       int64              `json:"sale_number" bson:"sale_number"`
	Products         []SoldProduct      `json:"products,omitempty" bson:"products,omitempty"`
	Loan             *Loan              `json:"loan,omitempty" bson:"loan,omitempty"` // <-- CAMPO NUEVO
	Status           string             `json:"status" bson:"status"`
	TotalAmountCents int64              `json:"total_amount_cents" bson:"total_amount_cents"`
	SaleDate         time.Time          `json:"sale_date" bson:"sale_date"`
	CreatedAt        time.Time          `json:"created_at" bson:"created_at"`

	// Campos "poblados" o "unidos"
	Client      Client               `json:"client" bson:"client"`
	Quotas      []Quota              `json:"quotas" bson:"quotas"`
	PaymentsIDs []primitive.ObjectID `json:"payment_ids,omitempty" bson:"payment_ids,omitempty"`
}

// Loan representa los detalles de un préstamo otorgado dentro de una transacción.
// Se almacena como un sub-documento para mantener la claridad y escalabilidad del modelo.
type Loan struct {
	PrincipalAmount structures.Money `json:"principal_amount" bson:"principal_amount"` // El capital original prestado.
	InterestRate    float64          `json:"interest_rate" bson:"interest_rate"`       // Tasa de interés aplicada (ej: 0.20 para 20%).
	TotalToRepay    structures.Money `json:"total_to_repay" bson:"total_to_repay"`     // El monto total a devolver (capital + intereses).
	Observations    string           `json:"observations,omitempty" bson:"observations,omitempty"`
}

func (l *Loan) HasLoan() bool {
	if l == nil {
		return false
	}

	if l.PrincipalAmount.AmountCents == 0 {
		return false
	}
	return true
}

func (l *Loan) GetTotalToRepay() int64 {
	if l == nil {
		return 0
	}

	return l.TotalToRepay.AmountCents
}
