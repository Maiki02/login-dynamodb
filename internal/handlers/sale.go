package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"myproject/internal/services"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
)

// SaleHandler agrupa los manejadores de peticiones para las ventas.
// Contiene el servicio de ventas del cual depende.
type SaleHandler struct {
	service *services.SaleService
}

// NewSaleHandler crea una nueva instancia de SaleHandler.
func NewSaleHandler(s *services.SaleService) *SaleHandler {
	return &SaleHandler{
		service: s,
	}
}

// CreateSaleHandler maneja la petición para crear una nueva venta.
func (h *SaleHandler) CreateSaleHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Decodificar el cuerpo de la petición en el struct de request.
	var saleReq request.CreateSaleRequest
	if err := json.NewDecoder(r.Body).Decode(&saleReq); err != nil {
		println("Error al decodificar el cuerpo de la petición:", err)
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// 2. Obtener el nombre de la base de datos de la petición.
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// 3. Llamar al método del servicio, pasándole el contexto, el nombre de la DB y el request.
	//    El handler no sabe de transacciones ni de lógica de negocio, solo orquesta la llamada.
	newSale, err := h.service.CreateSale(r.Context(), companyNameID, saleReq)
	if err != nil {
		// El servicio ya manejó la lógica, aquí solo reportamos el error.
		// Podríamos tener un manejo más granular de errores aquí si quisiéramos.
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// 4. Si todo fue exitoso, responder con el objeto de venta creado y un status 201.
	response.ResponseSuccess(w, newSale, http.StatusCreated)
}

func (h *SaleHandler) GetSalesHandler(w http.ResponseWriter, r *http.Request) {
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// --- Lógica de Paginación ---
	queryParams := r.URL.Query()
	page, _ := strconv.ParseInt(queryParams.Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(queryParams.Get("limit"), 10, 64)
	if limit < 1 {
		limit = 10
	}

	search := queryParams.Get("search")
	sortBy := queryParams.Get("sortBy")
	sortOrder := queryParams.Get("sortOrder")

	// Llamar al nuevo método del servicio
	paginatedSales, err := h.service.GetPaginatedSales(r.Context(), companyNameID, page, limit, search, sortBy, sortOrder)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, paginatedSales, http.StatusOK)
}
