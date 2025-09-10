package models

import (
	security "myproject/pkg/session"
	"myproject/pkg/validations"
	"strings"
	"time"

	"github.com/google/uuid"
)

// User representa la estructura de un usuario en la aplicación para DynamoDB.
type User struct {
	ID string `json:"id" dynamodbav:"user_id"`
	// Información personal del usuario
	PersonalInfo PersonalInfo `json:"personal_info" dynamodbav:"personal_info"`
	// Información de contacto del usuario
	ContactInfo ContactInfo `json:"contact_info" dynamodbav:"contact_info"`

	//Contraseña del usuario
	Password string `json:"password" dynamodbav:"password"`

	CreatedAt   time.Time `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" dynamodbav:"updated_at,omitempty"`
	DeletedAt   time.Time `json:"deleted_at,omitempty" dynamodbav:"deleted_at,omitempty"`
	Status      int32     `json:"status" dynamodbav:"status"` // Por ejemplo, 1: activo, 0: inactivo, -1: baneado
	LastSession time.Time `json:"last_session,omitempty" dynamodbav:"last_session,omitempty"`
}

// PersonalInfo agrupa la información personal del usuario.
type PersonalInfo struct {
	Name      string     `json:"name" dynamodbav:"name"`
	LastName  string     `json:"last_name" dynamodbav:"last_name"`
	BirthDate *time.Time `json:"birth_date,omitempty" dynamodbav:"birth_date,omitempty"`
}

// ContactInfo agrupa la información de contacto del usuario.
type ContactInfo struct {
	Email EmailDetails `json:"email" dynamodbav:"email"`
	// Phone structures.Phone `json:"phone,omitempty" dynamodbav:"phone,omitempty"` // Commented out for now
}

// EmailDetails agrupa toda la información relacionada con el email.
type EmailDetails struct {
	Address         string    `json:"address" dynamodbav:"address"` // Nombre cambiado de Email a Address para evitar conflicto con la estructura EmailDetails
	IsVerified      bool      `json:"is_verified" dynamodbav:"is_verified"`
	VerifiedAt      time.Time `json:"verified_at" dynamodbav:"verified_at"`
	IsSentForVerify bool      `json:"is_sent_for_verify" dynamodbav:"is_sent_for_verify"`
	SentAt          time.Time `json:"sent_at" dynamodbav:"sent_at"`
}

func NewUser(name, lastName, email, password string, companyName string) (*User, error) {
	nameToSave, err := validations.ValidateName(name, "Nombre")
	if err != nil {
		return nil, validations.ErrInvalidName
	}

	lastNameToSave, err := validations.ValidateName(lastName, "Apellido")
	if err != nil {
		return nil, validations.ErrInvalidLastName
	}

	isValidEmail := validations.IsValidEmail(email)
	if !isValidEmail {
		return nil, validations.ErrInvalidEmail
	}

	hashPassword, err := security.ValidateAndHashPassword(password)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:           uuid.New().String(),
		PersonalInfo: PersonalInfo{Name: nameToSave, LastName: lastNameToSave},
		ContactInfo:  ContactInfo{Email: EmailDetails{Address: strings.ToLower(email)}},
		Password:     *hashPassword,
		CreatedAt:    time.Now(),
		Status:       1, // active status
	}, nil
}

//bcript;

func (user *User) IsUserVerified() bool {
	return true
	//return user.ContactInfo.Email.IsVerified
}

func (user *User) GetUserDB() string {
	return user.ID + "_DB"
}
