package request

// SendInvitationRequest define la estructura para la petición de enviar una invitación.
type SendInvitationRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role" binding:"required"`
}

// AcceptInvitationRequest define la estructura para la petición de aceptar una invitación.
type AcceptInvitationRequest struct {
	Token string `json:"token" binding:"required"`
}
