package request

import (
	"myproject/internal/models"
	"myproject/pkg/structures"
)

// VariantRequest define la estructura para una variante en la petición de creación.
type VariantRequest struct {
	SKU        string                 `json:"sku"`
	Attributes map[string]interface{} `json:"attributes"`
	Price      structures.Money       `json:"price"`
	Cost       *structures.Money      `json:"cost,omitempty"`
	Stock      models.Stock           `json:"stock"`
}

// CreateProductRequest define el payload mínimo para crear un producto.
type CreateProductRequest struct {
	Name            string           `json:"name"`
	BrandName       *string          `json:"brand_name,omitempty"`
	ProductTypeName *string          `json:"product_type_name,omitempty"`
	Variants        []VariantRequest `json:"variants"`
}

// UpdateProductRequest define el payload para actualizar los datos generales de un producto.
type UpdateProductRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	Status          *string `json:"status,omitempty"`
	BrandName       *string `json:"brand_name,omitempty"`
	ProductTypeName *string `json:"product_type_name,omitempty"`
	// Las variantes no se actualizan desde aquí para mantener la atomicidad.
}

// UpdateVariantRequest define el payload para actualizar una variante específica.
type UpdateVariantRequest struct {
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
	Price            *structures.Money      `json:"price,omitempty"`
	Cost             *structures.Money      `json:"cost,omitempty"`
	Stock            *models.Stock          `json:"stock,omitempty"`
	SellWithoutStock *bool                  `json:"sell_without_stock,omitempty"`
}
