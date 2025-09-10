package handlers

import (
	"encoding/json"
	"myproject/internal/services"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// GenericHandler maneja las peticiones HTTP para entidades genéricas.
type GenericHandler struct {
	service services.GenericServiceInterface
}

// NewGenericHandler crea una instancia del handler genérico.
func NewGenericHandler(s services.GenericServiceInterface) *GenericHandler {
	return &GenericHandler{
		service: s,
	}
}

type CreateRequest struct {
	Name string `json:"name"`
}

func (h *GenericHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	entity, err := h.service.Create(r.Context(), nameDB, req.Name)
	if err != nil {
		// Aquí puedes mapear errores específicos a códigos HTTP
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, entity, http.StatusCreated)
}

func (h *GenericHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	entity, err := h.service.GetBySlug(r.Context(), nameDB, slug)
	if err != nil {
		response.ResponseError(w, err, http.StatusNotFound)
		return
	}

	response.ResponseSuccess(w, entity, http.StatusOK)
}

func (h *GenericHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	updatedEntity, err := h.service.Update(r.Context(), nameDB, slug, req.Name)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, updatedEntity, http.StatusOK)
}

// GetAll maneja la petición para obtener una lista paginada de entidades.
func (h *GenericHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// Extraer parámetros de la query
	queryParams := r.URL.Query()
	page, _ := strconv.ParseInt(queryParams.Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(queryParams.Get("limit"), 10, 64)
	if limit < 1 {
		limit = 10 // Valor por defecto
	}

	search := queryParams.Get("search")
	sortBy := queryParams.Get("sortBy")
	sortOrder := queryParams.Get("sortOrder")

	// Llamar al servicio
	paginatedResult, err := h.service.GetPaginated(r.Context(), nameDB, search, sortBy, sortOrder, page, limit)
	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	response.ResponseSuccess(w, paginatedResult, http.StatusOK)
}
