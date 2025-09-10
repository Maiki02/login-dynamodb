package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/request"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Definición de errores ---
var (
	ErrUserAlreadyMember = errors.New("el usuario ya es miembro de esta empresa")
	ErrCompanyNotFound   = errors.New("empresa no encontrada")
	ErrTokenInvalid      = errors.New("token inválido o expirado")
	ErrUserNotFound      = errors.New("user_not_found")
	ErrUpdateUser        = errors.New("no se pudo actualizar el usuario")
)

// --- Service ---
type InvitationService struct {
	userRepo    repositories.UserRepository
	companyRepo repositories.CompanyRepository
}

func NewInvitationService(userRepo repositories.UserRepository, companyRepo repositories.CompanyRepository) *InvitationService {
	return &InvitationService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
	}
}

// SendInvitation valida la invitación, crea un JWT y llama al servicio de correo.
func (s *InvitationService) SendInvitation(ctx context.Context, companyID primitive.ObjectID, req *request.SendInvitationRequest) error {
	// 1. Validar que el usuario no sea ya miembro de la empresa
	filter := bson.M{
		"contact_info.email.address": req.Email,
		"companies_info": bson.M{
			"$elemMatch": bson.M{"company_id": companyID},
		},
	}
	user, err := s.userRepo.GetUserByFilter(filter)
	if err == nil && user != nil {
		return ErrUserAlreadyMember
	}

	// 2. Obtener el nombre de la empresa
	company, err := s.companyRepo.GetCompanyByFilter(bson.M{"_id": companyID})
	if err != nil || company == nil {
		return ErrCompanyNotFound
	}

	// 3. Crear el JWT de invitación
	claims := models.InvitationClaims{
		CompanyID:   companyID.Hex(),
		CompanyName: company.Name,
		Email:       req.Email,
		Role:        req.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)), // El token expira en 3 días
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	tokenString, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(jwtSecret))
	if err != nil {
		return fmt.Errorf("error al generar el token de invitación: %w", err)
	}

	// 4. Preparar y enviar el email a través del servicio de Google App Script
	frontendURL := os.Getenv("FRONTEND_URL")
	invitationLink := fmt.Sprintf("%s/accept-invitation?token=%s", frontendURL, tokenString)

	// Si el usuario existe, usamos su nombre, de lo contrario, usamos el email
	name := req.Email
	if user != nil && user.PersonalInfo.Name != "" {
		name = user.PersonalInfo.Name
	}

	return s.sendEmailViaGoogle(req.Email, name, company.Name, invitationLink)
}

// sendEmailViaGoogle es un helper para llamar a tu webhook de Google App Script.
func (s *InvitationService) sendEmailViaGoogle(userEmail, userName, companyName, invitationLink string) error {
	googleAppScriptURL := os.Getenv("GOOGLE_APP_SCRIPT_URL")

	payload := map[string]string{
		"email":           userEmail,
		"name":            userName,
		"business_name":   companyName,
		"invitation_link": invitationLink,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(googleAppScriptURL, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Aquí podrías añadir una validación de la respuesta si tu App Script la devuelve
	return nil
}

// AcceptInvitation procesa un token de invitación para un usuario existente.
func (s *InvitationService) AcceptInvitation(ctx context.Context, tokenString string) (*models.User, error) {
	// 1. Validar el token y obtener las claims
	claims := &models.InvitationClaims{}
	jwtSecret := os.Getenv("JWT_SECRET")
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, ErrTokenInvalid
	}

	// 2. Buscar al usuario por el email del token
	user, err := s.userRepo.GetUserByFilter(bson.M{"contact_info.email.address": claims.Email})
	if err != nil || user == nil {
		return nil, ErrUserNotFound // Error específico para que el frontend redirija a registro
	}

	// 3. Agregar la empresa al usuario (si no es miembro ya)
	companyID, _ := primitive.ObjectIDFromHex(claims.CompanyID)
	for _, info := range user.CompaniesInfo {
		if info.CompanyID == companyID {
			return nil, ErrUserAlreadyMember
		}
	}

	newCompanyInfo := models.CompanyUserInfo{
		CompanyID: companyID,
		Name:      claims.CompanyName,
		Roles:     []string{claims.Role},
	}

	user.CompaniesInfo = append(user.CompaniesInfo, newCompanyInfo)

	// 4. Actualizar el usuario en la base de datos
	updates := map[string]interface{}{
		"companies_info": user.CompaniesInfo,
	}
	if err := s.userRepo.UpdateUser(user.ID.Hex(), updates); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUpdateUser, err)
	}

	return user, nil
}
