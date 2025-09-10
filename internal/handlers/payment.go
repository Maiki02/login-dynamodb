package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"myproject/internal/services"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PaymentHandler agrupa los manejadores de peticiones para las ventas.
// Contiene el servicio de ventas del cual depende.
type PaymentHandler struct {
	service *services.PaymentService
}

// NewPaymentHandler crea una nueva instancia de PaymentHandler.
func NewPaymentHandler(s *services.PaymentService) *PaymentHandler {
	return &PaymentHandler{
		service: s,
	}
}

// CreatePaymentHandler maneja la creación de un nuevo pago para un conjunto de cuotas.
func (h *PaymentHandler) CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. OBTENER DATOS DE LA URL Y AUTENTICACIÓN
	vars := mux.Vars(r)

	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	saleID, err := primitive.ObjectIDFromHex(vars["sale_id"]) // Validar el ID
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// Asumimos que un middleware de auth ya colocó el ID del usuario (cobrador) en el contexto.
	collectorID, ok := r.Context().Value(request.UserContextKey).(primitive.ObjectID)
	if !ok {
		response.ResponseError(w, validations.ErrInvalidUserID, http.StatusBadRequest)
		return
	}

	// 2. DECODIFICAR Y VALIDAR EL CUERPO (BODY) DE LA PETICIÓN
	var req request.PayQuotasRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Aquí iría la validación de la estructura `req` con una librería como "validator".

	// 3. LLAMAR AL SERVICIO PARA QUE HAGA EL TRABAJO PESADO
	newPayment, err := h.service.ProcessBulkPayment(r.Context(), companyNameID, saleID, req, collectorID)
	if err != nil {
		// Aquí podrías tener un manejador de errores más sofisticado
		// para devolver 404 si algo no se encuentra, 400 si hay un error de negocio, etc.
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// 4. RESPONDER CON ÉXITO
	response.ResponseSuccess(w, newPayment, http.StatusCreated)
}

// CreateSequentialPaymentHandler maneja la creación de un pago secuencial que afecta a las cuotas pendientes de una venta.
func (h *PaymentHandler) CreateSequentialPaymentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. OBTENER DATOS DE LA URL Y AUTENTICACIÓN
	vars := mux.Vars(r)

	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	saleID, err := primitive.ObjectIDFromHex(vars["sale_id"])
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// Asumimos que un middleware de auth ya colocó el ID del usuario (cobrador) en el contexto.
	collectorID, ok := r.Context().Value(request.UserContextKey).(primitive.ObjectID)
	if !ok {
		response.ResponseError(w, validations.ErrInvalidUserID, http.StatusBadRequest)
		return
	}

	// 2. DECODIFICAR Y VALIDAR EL CUERPO (BODY) DE LA PETICIÓN
	var req request.SequentialPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	// Aquí iría la validación de la estructura `req` con una librería como "validator".

	// 3. LLAMAR AL SERVICIO PARA QUE HAGA EL TRABAJO PESADO
	newPayment, err := h.service.ProcessSequentialPayment(r.Context(), companyNameID, saleID, req, collectorID)
	if err != nil {
		// Aquí se podría tener un manejador de errores más sofisticado
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// 4. RESPONDER CON ÉXITO
	response.ResponseSuccess(w, newPayment, http.StatusCreated)
}

// RevertPaymentHandler maneja la petición para revertir un pago existente.
func (h *PaymentHandler) RevertPaymentHandler(w http.ResponseWriter, r *http.Request) {
	// 1. OBTENER DATOS DE LA URL
	vars := mux.Vars(r)

	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	saleID, err := primitive.ObjectIDFromHex(vars["sale_id"])
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	paymentID, err := primitive.ObjectIDFromHex(vars["payment_id"])
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// 2. LLAMAR AL SERVICIO
	// El servicio se encargará de toda la lógica compleja.
	err = h.service.RevertPayment(r.Context(), companyNameID, saleID, paymentID)
	if err != nil {
		// Aquí podrías tener un manejador de errores más sofisticado
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// 3. RESPONDER CON ÉXITO
	response.ResponseSuccess(w, nil, http.StatusOK)
}

// GetPaymentsHandler maneja la petición para filtrar y obtener pagos.
func (h *PaymentHandler) GetPaymentsHandler(w http.ResponseWriter, r *http.Request) {
	// 1. OBTENER ID DE LA URL Y VALIDAR
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// 2. PARSEAR QUERY PARAMETERS (Filtros + Paginación)
	queryParams := r.URL.Query()
	var filter request.FilterPaymentsRequest

	const layoutISO = "2006-01-02"

	// Parsear fechas
	if startDateStr := queryParams.Get("startDate"); startDateStr != "" {
		startDate, err := time.Parse(layoutISO, startDateStr)
		if err != nil {
			response.ResponseError(w, validations.ErrInvalidFormatDate, http.StatusBadRequest)
			return
		}
		filter.StartDate = &startDate
	}
	if endDateStr := queryParams.Get("endDate"); endDateStr != "" {
		endDate, err := time.Parse(layoutISO, endDateStr)
		if err != nil {
			response.ResponseError(w, validations.ErrInvalidFormatDate, http.StatusBadRequest)
			return
		}
		filter.EndDate = &endDate
	}

	// Parsear estados
	if statusesStr := queryParams.Get("statuses"); statusesStr != "" {
		filter.Statuses = strings.Split(statusesStr, ",")
	}

	// Parsear parámetros de paginación
	page, _ := strconv.ParseInt(queryParams.Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(queryParams.Get("limit"), 10, 64)
	if limit < 1 {
		limit = 10
	}

	// 3. LLAMAR AL SERVICIO (con nuevos parámetros)
	paginatedPayments, err := h.service.GetPaymentsWithDetails(r.Context(), companyNameID, filter, page, limit)
	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// 4. RESPONDER CON ÉXITO
	response.ResponseSuccess(w, paginatedPayments, http.StatusOK)
}
