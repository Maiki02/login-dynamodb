package request

import (
	"myproject/internal/models"
	"myproject/pkg/structures"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ProductInSaleRequest define la información que el frontend envía por cada producto.
// Permite un "override" opcional del precio y costo.
type ProductInSaleRequest struct {
	ProductID  primitive.ObjectID `json:"product_id" validate:"required"`
	VariantSKU string             `json:"variant_sku" validate:"required"`
	Quantity   int                `json:"quantity" validate:"required,gt=0"`

	// --- Overrides Opcionales ---
	// Si estos campos vienen en `null`, vacíos o no se envían, el backend usará los valores del catálogo.
	// Si se envían, se usarán estos valores para el snapshot de la venta.
	UnitPrice *structures.Money `json:"unit_price,omitempty"`
	UnitCost  *structures.Money `json:"unit_cost,omitempty"`
}

// CreateSaleRequest define el payload que se espera del frontend para crear una venta.
type CreateSaleRequest struct {
	ClientID        primitive.ObjectID     `json:"client_id" validate:"required"`
	Products        []ProductInSaleRequest `json:"products,omitempty"`
	Loan            *models.Loan           `json:"loan,omitempty"`
	Quotas          []models.Quota         `json:"quotas" validate:"required,min=1"`
	SaleDate        time.Time              `json:"sale_date" validate:"required"`
	ShippingDate    *time.Time             `json:"shipping_date,omitempty"`
	BillingAddress  *structures.Address    `json:"billing_address,omitempty"`
	ShippingAddress *structures.Address    `json:"shipping_address,omitempty"`
	Observations    *string                `json:"observations,omitempty"`
}
