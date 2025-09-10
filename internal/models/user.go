package models

import (
	"myproject/pkg/consts"
	security "myproject/pkg/session"
	"myproject/pkg/structures"
	"myproject/pkg/validations"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User representa la estructura de un usuario en la aplicación.
type User struct {
	ID primitive.ObjectID `json:"_id" bson:"_id"`
	// Información personal del usuario
	PersonalInfo PersonalInfo `json:"personal_info" bson:"personal_info"`
	// Información de contacto del usuario
	ContactInfo ContactInfo `json:"contact_info" bson:"contact_info"`

	//Contraseña del usuario
	Password string `json:"password" bson:"password"`

	CompaniesInfo []CompanyUserInfo `json:"companies_info" bson:"companies_info"`
	CreatedAt     time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
	DeletedAt     time.Time         `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Status        int32             `json:"status" bson:"status"` // Por ejemplo, 1: activo, 0: inactivo, -1: baneado
	LastSession   time.Time         `json:"last_session,omitempty" bson:"last_session,omitempty"`
}

// PersonalInfo agrupa la información personal del usuario.
type PersonalInfo struct {
	Name      string    `json:"name" bson:"name"`
	LastName  string    `json:"last_name" bson:"last_name"`
	BirthDate time.Time `json:"birth_date,omitempty" bson:"birth_date,omitempty"`
}

// ContactInfo agrupa la información de contacto del usuario.
type ContactInfo struct {
	Email EmailDetails     `json:"email" bson:"email"`
	Phone structures.Phone `json:"phone,omitempty" bson:"phone,omitempty"`
}

// EmailDetails agrupa toda la información relacionada con el email.
type EmailDetails struct {
	Address         string    `json:"address" bson:"address"` // Nombre cambiado de Email a Address para evitar conflicto con la estructura EmailDetails
	IsVerified      bool      `json:"is_verified" bson:"is_verified"`
	VerifiedAt      time.Time `json:"verified_at" bson:"verified_at"`
	IsSentForVerify bool      `json:"is_sent_for_verify" bson:"is_sent_for_verify"`
	SentAt          time.Time `json:"sent_at" bson:"sent_at"`
}

// CompanyInfo representa la información de una empresa asociada al usuario.
type CompanyUserInfo struct {
	CompanyID primitive.ObjectID `json:"company_id" bson:"company_id"`
	Name      string             `json:"name" bson:"name"`
	Roles     []string           `json:"roles" bson:"roles"` // Cambiado a slice de strings para múltiples roles
}

func NewUser(name, lastName, email, password string, company Company) (*User, error) {
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
		ID:            primitive.NewObjectID(),
		PersonalInfo:  PersonalInfo{Name: nameToSave, LastName: lastNameToSave},
		ContactInfo:   ContactInfo{Email: EmailDetails{Address: strings.ToLower(email)}},
		Password:      *hashPassword,
		CompaniesInfo: []CompanyUserInfo{{CompanyID: company.ID, Name: company.Name, Roles: []string{consts.ROLE_OWNER}}},
		CreatedAt:     time.Now(),
		Status:        1, // active status
	}, nil
}

//bcript;

func (user *User) IsUserVerified() bool {
	return true
	//return user.ContactInfo.Email.IsVerified
}

func (user *User) GetUserDB() string {
	return user.ID.Hex() + "_DB"
}
