package handlers

import (
	"encoding/json"
	"myproject/internal/services"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InvitationHandler struct {
	service *services.InvitationService
}

func NewInvitationHandler(service *services.InvitationService) *InvitationHandler {
	return &InvitationHandler{service: service}
}

func (h *InvitationHandler) SendInvitation(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	companyIDStr := vars["company_id"]

	if companyIDStr == "" {
		response.ResponseError(w, validations.ErrCompanyNotFound, http.StatusBadRequest)
		return
	}

	companyID, err := primitive.ObjectIDFromHex(companyIDStr)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	var req request.SendInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	if err := h.service.SendInvitation(r.Context(), companyID, &req); err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	response.ResponseSuccess(w, nil, http.StatusOK)
}

func (h *InvitationHandler) AcceptInvitation(w http.ResponseWriter, r *http.Request) {
	var req request.AcceptInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	_, err := h.service.AcceptInvitation(r.Context(), req.Token)
	if err != nil {
		if err.Error() == "user_not_found" {
			response.ResponseError(w, err, http.StatusConflict)
			return
		}
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	response.ResponseSuccess(w, nil, http.StatusOK)
}
