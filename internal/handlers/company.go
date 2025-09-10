package handlers

import (
	"net/http"

	"myproject/internal/services"
	"myproject/pkg/response"
	"myproject/pkg/validations"

	"github.com/gorilla/mux"
)

// CompanyHandler maneja las solicitudes HTTP para las empresas.
type CompanyHandler struct {
	companyService services.CompanyService
}

// NewCompanyHandler crea una nueva instancia de CompanyHandler.
func NewCompanyHandler(cs services.CompanyService) *CompanyHandler {
	return &CompanyHandler{
		companyService: cs,
	}
}

// GetByID es el handler para obtener una empresa por su ID.
func (h *CompanyHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		response.ResponseError(w, validations.ErrDocumentNotFound, http.StatusBadRequest)
		return
	}

	// Usamos el servicio inyectado
	company, err := h.companyService.GetCompanyByID(id)
	if err != nil {
		response.ResponseError(w, err, http.StatusNotFound) // StatusNotFound es más apropiado aquí
		return
	}

	response.ResponseSuccess(w, company, http.StatusOK)
}
