package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Brand struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name"`
	Slug      string             `json:"slug" bson:"slug"`
	Status    string             `json:"status" bson:"status"` // "activo", "inactivo"
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

func (b *Brand) GetName() string {
	if b == nil {
		return ""
	}

	if b.Name != "" {
		return b.Name
	}

	return ""
}
