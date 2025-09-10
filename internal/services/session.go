package services

import (
	"context"
	"errors"
	tokens "myproject/pkg/jwt"
	"myproject/pkg/validations"
	"os"
	"strings"
	"time"

	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/consts"
	"myproject/pkg/request"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

const ACCESS_DURATION = 1
const REFRESH_DURATION = 24

// --- Definición de errores ---
var (
	ErrInvitationTokenInvalid  = errors.New("token de invitación inválido o expirado")
	ErrInvitationEmailMismatch = errors.New("el email de la invitación no coincide con el proporcionado")
	ErrUserAlreadyExists       = errors.New("el email ya está registrado")
)

// SessionService encapsula la lógica de negocio para las sesiones.
type SessionService interface {
	Register(req request.RegisterUserRequest) error
	RegisterWithInvitation(ctx context.Context, req request.RegisterUserRequest) (*models.User, error)
	Login(email, password string) (*tokens.Tokens, error)
	RefreshToken(token string) (*tokens.Tokens, error)
}

type sessionService struct {
	userRepo    repositories.UserRepository
	companyRepo repositories.CompanyRepository
}

// NewSessionService crea una nueva instancia de SessionService.
func NewSessionService(userRepo repositories.UserRepository, companyRepo repositories.CompanyRepository) SessionService {
	return &sessionService{
		userRepo:    userRepo,
		companyRepo: companyRepo,
	}
}

// Register maneja la lógica de registro atómico de empresa y usuario.
func (s *sessionService) Register(req request.RegisterUserRequest) error {
	// 1. Validar si el usuario ya existe
	existingUser, _ := s.userRepo.GetUserByFilter(map[string]interface{}{"contact_info.email.address": req.Email})
	if existingUser != nil {
		return validations.ErrDocumentAlreadyExists
	}

	// 2. Crear la compañía con estado "pending"
	company, err := models.NewCompany(req.CompanyName)
	if err != nil {
		return err
	}
	company.Status = consts.STATUS_PENDING // Asumimos consts.STATUS_PENDING = 2
	if err := s.companyRepo.CreateCompany(company); err != nil {
		return err
	}

	// 3. Crear el usuario
	user, err := models.NewUser(req.Name, req.LastName, req.Email, req.Password, *company)
	if err != nil {
		// Si falla la creación del modelo de usuario, compensamos eliminando la empresa
		s.companyRepo.DeleteCompany(company.ID.Hex())
		return err
	}

	if err := s.userRepo.CreateUser(user); err != nil {
		// Si falla la inserción del usuario en la BD, compensamos eliminando la empresa
		s.companyRepo.DeleteCompany(company.ID.Hex())
		return err
	}

	// 4. Si todo fue exitoso, actualizamos el estado de la compañía a "active"
	updates := map[string]interface{}{"status": consts.STATUS_ACTIVE} // Asumimos consts.STATUS_ACTIVE = 1
	if err := s.companyRepo.UpdateCompany(company.ID.Hex(), updates); err != nil {
		// En un caso extremo, si esto falla, podrías tener un worker que limpie
		// empresas en estado "pending" después de cierto tiempo.
		//TODO: MENSAJE POR WHATSAPP PARA ARREGLARLO A MANO
		return err
	}

	// Opcional: Enviar email de activación
	// go SendEmailToActiveAccount(user.Email, user.Name, user.ID)

	return nil
}

// RegisterWithInvitation crea un nuevo usuario y lo asocia a una empresa mediante un token.
func (s *sessionService) RegisterWithInvitation(ctx context.Context, req request.RegisterUserRequest) (*models.User, error) {
	// 1. Validar el token de invitación
	claims := &models.InvitationClaims{}
	jwtSecret := os.Getenv("JWT_SECRET")
	_, err := jwt.ParseWithClaims(req.InvitationToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, ErrInvitationTokenInvalid
	}

	// 2. Seguridad: Verificar que el email del token coincida con el email del formulario
	if claims.Email != req.Email {
		return nil, ErrInvitationEmailMismatch
	}

	// 3. Verificar si el usuario ya existe
	existingUser, _ := s.userRepo.GetUserByFilter(bson.M{"contact_info.email.address": req.Email})
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 4. Validar que la empresa del token exista
	companyID, err := primitive.ObjectIDFromHex(claims.CompanyID)
	if err != nil {
		return nil, validations.ErrCompanyNotFound
	}
	company, err := s.companyRepo.GetCompanyByFilter(bson.M{"_id": companyID})
	if err != nil || company == nil {
		return nil, validations.ErrCompanyNotFound
	}

	// 5. Crear el usuario usando el modelo estándar
	user, err := models.NewUser(req.Name, req.LastName, req.Email, req.Password, *company)
	if err != nil {
		return nil, err
	}

	// 6. Sobrescribir la info de la empresa en CompaniesInfo con los datos del token
	user.CompaniesInfo = []models.CompanyUserInfo{{
		CompanyID: companyID,
		Name:      company.Name,
		Roles:     []string{claims.Role},
	}}

	// 7. Insertar el usuario en la base de datos
	if err := s.userRepo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login ahora es un método del servicio y usa el repositorio de usuarios.
func (s *sessionService) Login(email, password string) (*tokens.Tokens, error) {
	emailLower := strings.ToLower(email)
	// ¡CAMBIO CLAVE! Usamos el repositorio inyectado.
	user, err := s.userRepo.GetUserByFilter(bson.M{"contact_info.email.address": emailLower})
	if err != nil {
		return nil, err
	}

	if !user.IsUserVerified() {
		return nil, validations.ErrUserInactive
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, validations.ErrInvalidCredentials
	}

	token, err := tokens.GenerateJWT(user, ACCESS_DURATION)
	if err != nil {
		return nil, err
	}
	refreshToken, err2 := tokens.GenerateJWT(user, REFRESH_DURATION)
	if err2 != nil {
		return nil, err2
	}

	// ¡CAMBIO CLAVE! Usamos el repositorio inyectado.
	go s.userRepo.UpdateUser(user.ID.Hex(), map[string]interface{}{"last_session": time.Now()})

	return &tokens.Tokens{
		AccessToken:  token,
		RefreshToken: refreshToken,
	}, nil
}

// Obtenemos los claims del token
// Buscamos en la base de datos si el usuario existe
func (s *sessionService) RefreshToken(token string) (*tokens.Tokens, error) {
	claimsMap, err := tokens.GetClaims(token)
	if err != nil {
		return nil, err
	}

	// 1. Acceder a 'contact_info'
	contactInfoClaim, ok := (*claimsMap)["contact_info"].(map[string]interface{})
	if !ok {
		println("Error: 'contact_info' no encontrado o no es un mapa en los claims.")
		return nil, validations.ErrInvalidToken // O un error más específico si tienes
	}

	// 2. Acceder a 'email' dentro de 'contact_info'
	emailDetailsClaim, ok := contactInfoClaim["email"].(map[string]interface{})
	if !ok {
		println("Error: 'email' no encontrado o no es un mapa en 'contact_info'.")
		return nil, validations.ErrInvalidToken // O un error más específico
	}

	// 3. Acceder a 'address' dentro de 'email'
	emailAddress, ok := emailDetailsClaim["address"].(string)
	if !ok {
		println("Error: 'address' no encontrado o no es un string en 'email'.")
		return nil, validations.ErrInvalidToken // O un error más específico
	}

	user, err := s.userRepo.GetUserByFilter(bson.M{"contact_info.email.address": emailAddress})
	if err != nil {
		return nil, err
	}

	// Verificamos si el usuario está activo
	if !user.IsUserVerified() {
		return nil, validations.ErrUserInactive
	}

	// Buscamos la company en la BD
	/*company, err := GetCompanyByUser(user.ID.Hex() + "_DB")
	if err != nil {
		return nil, err
	}*/

	// Crear el JWT si la contraseña es válida
	accessToken, err := tokens.GenerateJWT(user, ACCESS_DURATION)
	if err != nil {
		return nil, err
	}

	// Crear el refresh token
	refreshToken, err := tokens.GenerateJWT(user, REFRESH_DURATION)
	if err != nil {
		return nil, err
	}

	// Actualizamos last_sesion en la BD
	go s.userRepo.UpdateUser(user.ID.Hex(), map[string]interface{}{"last_session": time.Now()})

	return &tokens.Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

/*
func ActualizatePassword(emailToken, email, password string) error {
	if emailToken == "" || email == "" || password == "" {
		return validations.ErrInvalidRequest
	}

	if emailToken != email {
		return validations.ErrInvalidRequest
	}

	userDB, err := GetUserByFilter(nil, nil, &email)
	if err != nil {
		return err
	}

	hashPassword, err := security.ValidateAndHashPassword(password)
	if err != nil {
		return err
	}

	idString := userDB.ID.Hex()

	err = repositories.UpdateUser(idString, map[string]interface{}{"password": *hashPassword})
	if err != nil {
		return err
	}

	return nil
}

func ActivateAccount(email string) error {
	if email == "" {
		return validations.ErrInvalidRequest
	}

	userDB, err := GetUserByFilter(nil, nil, &email)
	if err != nil {
		return err
	}

	idString := userDB.ID.Hex()
	updates := map[string]interface{}{
		"is_email_verified": true,
		"email_verify_at":   time.Now(),
	}

	err = repositories.UpdateUser(idString, updates)
	if err != nil {
		return err
	}

	return nil
}

/*func SendEmailToActiveAccount(email, name string, id primitive.ObjectID) error {
	statusCode, err := sendEmailGeneric(email, URL_INVITATION, name)

	if err != nil {
		return err
	}

	// Si se envía correctamente el email, actualizamos la bd
	if statusCode == http.StatusOK {
		updates := map[string]interface{}{
			"is_email_sent": true,
			"email_sent_at": time.Now(),
		}

		UpdateUser(id.Hex(), updates)
	}

	return nil
}

func SendEmailToResetPassword(email, name string, id primitive.ObjectID) error {
	_, err := sendEmailGeneric(email, URL_REMEMBER_PASSWORD, name)

	if err != nil {
		return err
	}
	return nil
}

func sendEmailGeneric(email, urlConst, name string) (int, error) {
	token, err := tokens.GenerateJWTEmail(email, 30)
	if err != nil {
		return 0, err
	}

	// Obtener la URL de la variable de entorno
	URL := os.Getenv(urlConst)

	// Crear el cuerpo de la petición
	body := map[string]string{
		"email":           email,
		"invitation_link": getInvitationLink(urlConst) + token,
		"name":            name,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	// Crear la solicitud HTTP POST
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

func getInvitationLink(URL string) string {
	if URL == URL_INVITATION {
		return "www.localhost:4200/activate?link="
	} else if URL == URL_REMEMBER_PASSWORD {
		return "www.localhost:4200/remember-password?link="
	}
	return ""
}*/
