package handlers

import (
	"encoding/json"
	"errors"
	"myproject/internal/services"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"
)

// QuotaHandler agrupa los manejadores para las cuotas.
type QuotaHandler struct {
	service *services.QuotaService
}

// NewQuotaHandler crea una nueva instancia de QuotaHandler.
func NewQuotaHandler(s *services.QuotaService) *QuotaHandler {
	return &QuotaHandler{
		service: s,
	}
}

// RescheduleQuotasHandler maneja la petición para reprogramar múltiples cuotas.
func (h *QuotaHandler) RescheduleQuotasHandler(w http.ResponseWriter, r *http.Request) {
	// 1. OBTENER ID DE LA URL
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// 2. DECODIFICAR Y VALIDAR EL CUERPO DE LA PETICIÓN
	var req request.RescheduleQuotasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Aquí podrías usar una librería como "validator" para las etiquetas del struct.
	if len(req.Updates) == 0 {
		response.ResponseError(w, errors.New("el arreglo de actualizaciones no puede estar vacío"), http.StatusBadRequest)
		return
	}

	// 3. LLAMAR AL SERVICIO
	err = h.service.RescheduleQuotas(r.Context(), companyNameID, req)
	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// 4. RESPONDER CON ÉXITO
	response.ResponseSuccess(w, map[string]string{"message": "Cuotas actualizadas exitosamente"}, http.StatusOK)
}
