package request

import (
	"myproject/pkg/structures"
)

//-------------- CLIENT ----------------\\

type CreateClientRequest struct {
	Name           *string                    `json:"name"`
	LastName       *string                    `json:"last_name"`
	Email          *string                    `json:"email"`
	Identification *structures.Identification `json:"identification"` // Cambiado
	Phone          *structures.Phone          `json:"phone"`
	Address        *structures.Address        `json:"address"`
}

type UpdateClientRequest struct {
	Name           *string                    `json:"name"`
	LastName       *string                    `json:"last_name"`
	Email          *string                    `json:"email"`
	Identification *structures.Identification `json:"identification"` // Cambiado
	Phone          *structures.Phone          `json:"phone"`
	Address        *structures.Address        `json:"address"`
}

//--------------------------------------\\

// -------------- SESSION ----------------\\
type RegisterUserRequest struct {
	Name            string `json:"name" binding:"required"`
	LastName        string `json:"last_name" binding:"required"`
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	CompanyName     string `json:"company_name"`               // Ya no es 'required'
	InvitationToken string `json:"invitation_token,omitempty"` // Nuevo campo opcional
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// -------------- PASSWORD ----------------\\
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
