package models

import (
	"myproject/pkg/validations"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Company struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Name      string             `json:"name" bson:"name"`
	Phone     string             `json:"phone" bson:"phone"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	Status    int32              `json:"status" bson:"status"`
}

func NewCompany(name string) (*Company, error) {
	companyName, err := validations.ValidateCompanyName(name)
	if err != nil {
		return nil, err
	}

	return &Company{
		ID:        primitive.NewObjectID(),
		Name:      companyName,
		Phone:     "",
		CreatedAt: time.Now(),
		Status:    1,
	}, nil
}
