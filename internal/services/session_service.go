package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"myproject/internal/models"
	"myproject/internal/repositories"
	tokens "myproject/pkg/jwt"
	"myproject/pkg/request"
	"myproject/pkg/validations"

	"github.com/google/uuid"
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

	// 2. Crear el usuario
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
		CreatedAt: time.Now(),
		Status:    1, // active
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

// generateUserID genera un ID único para el usuario usando UUID v4
func generateUserID() string {
	// UUID v4 garantiza distribución uniforme en DynamoDB
	// y unicidad global sin dependencia del tiempo
	return uuid.New().String()
}
