package models

import (
	"myproject/pkg/structures"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Stock maneja la información de inventario de una variante.
type Stock struct {
	CurrentQuantity  int  `json:"current_quantity" bson:"current_quantity"`
	ReorderPoint     int  `json:"reorder_point,omitempty" bson:"reorder_point,omitempty"`
	SellWithoutStock bool `json:"sell_without_stock" bson:"sell_without_stock"`
}

// Dimensions contiene las medidas físicas de una variante.
type Dimensions struct {
	WeightKG float64 `json:"weight_kg,omitempty" bson:"weight_kg,omitempty"`
	HeightCM float64 `json:"height_cm,omitempty" bson:"height_cm,omitempty"`
	WidthCM  float64 `json:"width_cm,omitempty" bson:"width_cm,omitempty"`
	LengthCM float64 `json:"length_cm,omitempty" bson:"length_cm,omitempty"`
}

// Variant representa una versión específica y vendible de un producto.
type Variant struct {
	SKU        string                 `json:"sku" bson:"sku"`
	Attributes map[string]interface{} `json:"attributes" bson:"attributes"`
	Price      structures.Money       `json:"price" bson:"price"`
	Cost       *structures.Money      `json:"cost,omitempty" bson:"cost,omitempty"`
	Stock      Stock                  `json:"stock" bson:"stock"`

	Images     []string    `json:"images,omitempty" bson:"images,omitempty"`
	Dimensions *Dimensions `json:"dimensions,omitempty" bson:"dimensions,omitempty"`
}

// Product es la estructura principal que agrupa toda la información de un producto.
type Product struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	InternalCode int64              `json:"internal_code,omitempty" bson:"internal_code,omitempty"` // Numérico
	Name         string             `json:"name" bson:"name"`
	Slug         string             `json:"slug" bson:"slug"`
	Description  string             `json:"description,omitempty" bson:"description,omitempty"`

	// Se guardan las referencias por ID para mantener la integridad de los datos.
	BrandID       primitive.ObjectID `json:"brand_id,omitempty" bson:"brand_id,omitempty"`
	ProductTypeID primitive.ObjectID `json:"product_type_id,omitempty" bson:"product_type_id,omitempty"`

	Tags                   []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	Images                 []string  `json:"images,omitempty" bson:"images,omitempty"`
	IsSellableFractionally bool      `json:"is_sellable_fractionally" bson:"is_sellable_fractionally"`
	UnitOfMeasure          string    `json:"unit_of_measure" bson:"unit_of_measure"`
	Status                 string    `json:"status" bson:"status"`
	Variants               []Variant `json:"variants" bson:"variants"`
	CreatedAt              time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" bson:"updated_at"`

	Brand       *Brand       `json:"brand,omitempty" bson:"brand,omitempty"`
	ProductType *ProductType `json:"product_type,omitempty" bson:"product_type,omitempty"`
}
