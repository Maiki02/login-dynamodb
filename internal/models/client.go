package models

import (
	"fmt"
	"myproject/pkg/structures"
	"myproject/pkg/validations"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Client representa la entidad de un cliente en el sistema.
type Client struct {
	ID                 primitive.ObjectID        `json:"_id" bson:"_id"`
	Name               string                    `json:"name" bson:"name" validate:"required,min=2"`
	LastName           string                    `json:"last_name" bson:"last_name" validate:"required,min=2"`
	Email              string                    `json:"email" bson:"email" validate:"required,email"`
	Identification     structures.Identification `json:"identification" bson:"identification"`
	DateOfBirth        time.Time                 `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	Phone              structures.Phone          `json:"phone" bson:"phone"`
	Address            structures.Address        `json:"address" bson:"address"`
	CreditBalanceCents int64                     `json:"credit_balance_cents,omitempty" bson:"credit_balance_cents,omitempty"` // Saldo a favor en centavos
	Status             int                       `json:"status" bson:"status" validate:"gte=0"`
	Notes              string                    `json:"notes,omitempty" bson:"notes,omitempty"`
	CreatedAt          time.Time                 `json:"created_at" bson:"created_at"`
	UpdatedAt          time.Time                 `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt          time.Time                 `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
}

// NewClient crea una nueva instancia de Client con validaciones básicas.
// Recibe los datos esenciales para crear un cliente funcional.
func NewClient(name, lastName, email *string, identification *structures.Identification, phone *structures.Phone, address *structures.Address) (*Client, error) {
	client := &Client{
		ID:        primitive.NewObjectID(),
		Status:    1, // Estado por defecto: Activo
		CreatedAt: time.Now().UTC(),
	}

	// 1. Validaciones basicas del cliente
	if name == nil {
		return nil, fmt.Errorf("%s es requerido", "Nombre")
	}

	if identification == nil {
		return nil, fmt.Errorf("%s es requerido", "Identificación")
	}

	nameToSave, err := validations.ValidateName(*name, "Nombre")
	if err != nil {
		return nil, err
	}
	client.Name = nameToSave

	if lastName != nil {
		lastNameToSave, err := validations.ValidateName(*lastName, "Apellido")
		if err != nil {
			return nil, err
		}
		client.LastName = lastNameToSave
	}

	if email != nil && *email != "" {
		isValidEmail := validations.IsValidEmail(*email)
		if !isValidEmail {
			return nil, validations.ErrInvalidClientEmail
		}
		client.Email = *email
	}

	if !validations.IsValidIdentification(identification) {
		return nil, validations.ErrInvalidClientIdentification
	}
	client.Identification = *identification

	if phone != nil && !phone.IsEmpty() {
		isValidPhone := validations.IsValidPhone(*phone)
		if !isValidPhone {
			return nil, validations.ErrInvalidClientPhone
		}
		client.Phone = *phone
	}

	if address != nil {
		isValidAddress := validations.IsValidAddress(*address)
		if !isValidAddress {
			return nil, validations.ErrInvalidClientAddress
		}
		client.Address = *address
	}

	// 3. Devolvemos el cliente creado y un error nulo
	return client, nil
}
