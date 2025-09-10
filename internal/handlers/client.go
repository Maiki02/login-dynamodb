package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	// Asumo que tendrás un ClientService con métodos que reciben `nameDB`.
	"myproject/internal/services"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"

	"github.com/gorilla/mux"
)

// ClientHandler agrupa los manejadores de peticiones para clientes.
// Contiene el servicio de cliente del cual depende.
type ClientHandler struct {
	service *services.ClientService
}

// NewClientHandler crea una nueva instancia de ClientHandler.
func NewClientHandler(s *services.ClientService) *ClientHandler {
	return &ClientHandler{
		service: s,
	}
}

// CreateClientHandler se convierte en un método de ClientHandler.
func (h *ClientHandler) CreateClientHandler(w http.ResponseWriter, r *http.Request) {
	var clientReq request.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&clientReq); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// Llama al método del servicio a través de la instancia inyectada.
	client, err := h.service.CreateClient(companyNameID, &clientReq)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, client, http.StatusCreated)
}

// GetClientsHandler se convierte en un método de ClientHandler.
func (h *ClientHandler) GetClientsHandler(w http.ResponseWriter, r *http.Request) {
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// --- Lógica de Paginación ---
	queryParams := r.URL.Query()

	// Parsear 'page'
	page, err := strconv.ParseInt(queryParams.Get("page"), 10, 64)
	if err != nil || page < 1 {
		page = 1
	}

	// Parsear 'limit'
	limit, err := strconv.ParseInt(queryParams.Get("limit"), 10, 64)
	if err != nil || limit < 1 {
		limit = 10
	}

	// Obtener otros parámetros
	search := queryParams.Get("search")
	sortBy := queryParams.Get("sortBy")
	sortOrder := queryParams.Get("sortOrder")

	// Llama al método del servicio con los nuevos parámetros.
	paginatedClients, err := h.service.GetPaginatedClients(r.Context(), companyNameID, page, limit, search, sortBy, sortOrder)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, paginatedClients, http.StatusOK)
}

// UpdateClientHandler se convierte en un método de ClientHandler.
func (h *ClientHandler) UpdateClientHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updates request.UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	companyID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// Llama al método del servicio.
	updatedClient, err := h.service.UpdateClient(r.Context(), companyID, id, &updates)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, updatedClient, http.StatusOK)
}

// DeleteClientHandler se convierte en un método de ClientHandler.
func (h *ClientHandler) DeleteClientHandler(w http.ResponseWriter, r *http.Request) {
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	if id == "" {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Llama al método del servicio.
	if err := h.service.DeleteClient(companyNameID, id); err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, nil, http.StatusOK)
}
