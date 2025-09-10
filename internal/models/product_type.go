package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductType struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Slug      string             `json:"slug" bson:"slug"`
	Status    string             `json:"status" bson:"status"` // "activo", "inactivo"
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

func (p *ProductType) GetName() string {
	if p == nil {
		return ""
	}

	if p.Name != "" {
		return p.Name
	}

	return ""
}
