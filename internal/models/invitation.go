package models

import "github.com/golang-jwt/jwt/v5"

// InvitationClaims define la estructura de datos (payload) para el JWT de invitación.
type InvitationClaims struct {
	CompanyID   string `json:"company_id"`
	CompanyName string `json:"company_name"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	jwt.RegisteredClaims
}
