package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/consts"
	tokens "myproject/pkg/jwt"
	"myproject/pkg/request"
	"myproject/pkg/validations"

	"golang.org/x/crypto/bcrypt"
)

const ACCESS_DURATION = 1
const REFRESH_DURATION = 24

// --- Definición de errores ---
var (
	ErrUserAlreadyExists = errors.New("el email ya está registrado")
)

// SessionService encapsula la lógica de negocio para las sesiones.
type SessionService interface {
	Register(ctx context.Context, req request.RegisterUserRequest) error
	Login(ctx context.Context, email, password string) (*tokens.Tokens, error)
	RefreshToken(ctx context.Context, token string) (*tokens.Tokens, error)
}

type sessionService struct {
	userRepo repositories.UserRepository
}

// NewSessionService crea una nueva instancia de SessionService.
func NewSessionService(userRepo repositories.UserRepository) SessionService {
	return &sessionService{
		userRepo: userRepo,
	}
}

// Register maneja la lógica de registro simple de usuarios.
func (s *sessionService) Register(ctx context.Context, req request.RegisterUserRequest) error {
	// 1. Validar si el usuario ya existe
	existingUser, _ := s.userRepo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return ErrUserAlreadyExists
	}

	// 2. Crear una empresa básica para el usuario
	company := models.CompanyUserInfo{
		CompanyID: "default-company-id",
		Name:      req.CompanyName,
		Roles:     []string{consts.ROLE_OWNER},
	}

	// 3. Crear el usuario (simplificamos NewUser para que no requiera Company)
	user := &models.User{
		ID: "", // Se generará en NewUser
		PersonalInfo: models.PersonalInfo{
			Name:     req.Name,
			LastName: req.LastName,
		},
		ContactInfo: models.ContactInfo{
			Email: models.EmailDetails{
				Address: strings.ToLower(req.Email),
			},
		},
		CompaniesInfo: []models.CompanyUserInfo{company},
		CreatedAt:     time.Now(),
		Status:        1, // active
	}

	// 4. Hash de la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 5. Generar ID único
	user.ID = generateUserID()

	// 6. Guardar el usuario en DynamoDB
	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return err
	}

	return nil
}

// Login maneja la autenticación de usuarios.
func (s *sessionService) Login(ctx context.Context, email, password string) (*tokens.Tokens, error) {
	emailLower := strings.ToLower(email)

	// 1. Buscar usuario por email
	user, err := s.userRepo.GetUserByEmail(ctx, emailLower)
	if err != nil {
		return nil, validations.ErrInvalidCredentials
	}

	// 2. Verificar que el usuario esté activo
	if !user.IsUserVerified() {
		return nil, validations.ErrUserInactive
	}

	// 3. Verificar contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, validations.ErrInvalidCredentials
	}

	// 4. Generar tokens
	accessToken, err := tokens.GenerateJWT(user, ACCESS_DURATION)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tokens.GenerateJWT(user, REFRESH_DURATION)
	if err != nil {
		return nil, err
	}

	// 5. Actualizar última sesión (async)
	go func() {
		updatedUser := *user
		updatedUser.LastSession = time.Now()
		s.userRepo.UpdateUser(context.Background(), user.ID, &updatedUser)
	}()

	return &tokens.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken maneja la renovación de tokens.
func (s *sessionService) RefreshToken(ctx context.Context, token string) (*tokens.Tokens, error) {
	// 1. Obtener claims del token
	claimsMap, err := tokens.GetClaims(token)
	if err != nil {
		return nil, validations.ErrInvalidToken
	}

	// 2. Extraer el ID del usuario del token
	userID, ok := (*claimsMap)["id"].(string)
	if !ok || userID == "" {
		return nil, validations.ErrInvalidToken
	}

	// 3. Buscar usuario en la base de datos
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, validations.ErrInvalidToken
	}

	// 4. Verificar que el usuario esté activo
	if !user.IsUserVerified() {
		return nil, validations.ErrUserInactive
	}

	// 5. Generar nuevos tokens
	accessToken, err := tokens.GenerateJWT(user, ACCESS_DURATION)
	if err != nil {
		return nil, err
	}

	refreshToken, err := tokens.GenerateJWT(user, REFRESH_DURATION)
	if err != nil {
		return nil, err
	}

	// 6. Actualizar última sesión (async)
	go func() {
		updatedUser := *user
		updatedUser.LastSession = time.Now()
		s.userRepo.UpdateUser(context.Background(), user.ID, &updatedUser)
	}()

	return &tokens.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// generateUserID genera un ID único para el usuario
func generateUserID() string {
	// Por simplicidad, usamos UUID v4
	// En producción podrías usar un patrón más específico
	return "user_" + time.Now().Format("20060102150405") + "_" + generateRandomString(6)
}

// generateRandomString genera una cadena aleatoria de la longitud especificada
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
