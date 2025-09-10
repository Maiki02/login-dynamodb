package handlers

import (
	"encoding/json"
	"net/http"

	"myproject/internal/services"
	tokens "myproject/pkg/jwt"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
)

// SessionHandler maneja las solicitudes HTTP relacionadas con la sesión.
type SessionHandler struct {
	sessionService services.SessionService
}

// NewSessionHandler crea una nueva instancia de SessionHandler.
func NewSessionHandler(ss services.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: ss,
	}
}

// Register es el handler para el registro de usuario.
func (h *SessionHandler) Register(w http.ResponseWriter, r *http.Request) {
	var sessionReq request.RegisterUserRequest

	if err := json.NewDecoder(r.Body).Decode(&sessionReq); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Pasar el contexto del request al service
	err := h.sessionService.Register(r.Context(), sessionReq)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, nil, http.StatusCreated)
}

func (h *SessionHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var sessionReq request.LoginUserRequest

	if err := json.NewDecoder(r.Body).Decode(&sessionReq); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Llamar al service con contexto
	tokens, err := h.sessionService.Login(r.Context(), sessionReq.Email, sessionReq.Password)
	if err != nil {
		response.ResponseError(w, err, http.StatusUnauthorized)
		return
	}

	// Se envía el token JWT en la respuesta
	response.ResponseSuccess(w, tokens, http.StatusOK)
}

func (h *SessionHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := tokens.GetTokenInHeader(r)
	if err != nil {
		response.ResponseError(w, validations.ErrInvalidToken, http.StatusUnauthorized)
		return
	}

	// Llamar al service con contexto
	newTokens, err := h.sessionService.RefreshToken(r.Context(), token)
	if err != nil {
		response.ResponseError(w, err, http.StatusUnauthorized)
		return
	}

	// Se envía el token JWT en la respuesta
	response.ResponseSuccess(w, newTokens, http.StatusOK)
}

/*
func ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var forgotPasswordReq request.ForgotPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&forgotPasswordReq); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	user, err := services.GetUserByFilter(nil, nil, &forgotPasswordReq.Email)
	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	go services.SendEmailToResetPassword(user.Email, user.Name, user.ID)

	response.ResponseSuccess(w, nil, http.StatusOK)
}

func ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var resetPasswordReq request.ResetPasswordRequest

	if err := json.NewDecoder(r.Body).Decode(&resetPasswordReq); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	link := r.URL.Query().Get("link")
	if link == "" {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	emailToken, err := tokens.GetFieldInToken(link, "email")
	if err != nil {
		response.ResponseError(w, validations.ErrInvalidToken, http.StatusUnauthorized)
		return
	}

	err = services.ActualizatePassword(emailToken, resetPasswordReq.Email, resetPasswordReq.Password)
	if err != nil {
		response.ResponseError(w, err, http.StatusTeapot)
		return
	}

	response.ResponseSuccess(w, nil, http.StatusOK)
}

func ActivateAccountHandler(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	if link == "" {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	emailToken, err := tokens.GetFieldInToken(link, "email")
	if err != nil {
		response.ResponseError(w, validations.ErrInvalidToken, http.StatusUnauthorized)
		return
	}

	err = services.ActivateAccount(emailToken)
	if err != nil {
		response.ResponseError(w, validations.ErrInvalidToken, http.StatusUnauthorized)
		return
	}

	response.ResponseSuccess(w, nil, http.StatusOK)
}
*/
