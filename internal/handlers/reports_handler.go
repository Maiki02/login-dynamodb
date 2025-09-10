package handlers

import (
	"myproject/internal/dto"
	"myproject/internal/services"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"
	"strconv"
	"time"
)

// ReportHandler agrupa los manejadores para los reportes.
// Depende de SaleService porque la lógica de negocio para este reporte está ahí.
type ReportHandler struct {
	saleService   *services.SaleService
	reportService *services.ReportService
}

// NewReportHandler crea una nueva instancia de ReportHandler.
func NewReportHandler(ss *services.SaleService, rs *services.ReportService) *ReportHandler {
	return &ReportHandler{
		saleService:   ss,
		reportService: rs,
	}
}

// GetPendingQuotasByClientHandler maneja la petición para obtener el reporte de cuotas pendientes por cliente.
func (h *ReportHandler) GetPendingQuotasByClientHandler(w http.ResponseWriter, r *http.Request) {
	// 1. OBTENER Y VALIDAR EL ID DE LA EMPRESA DESDE LA URL
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// 2. Obtenemos la query que se llama 'maxDate'
	maxDateStr := r.URL.Query().Get("maxDate")
	if maxDateStr == "" {
		response.ResponseError(w, validations.ErrInvalidDateToSearchReport, http.StatusBadRequest)
		return
	}

	// 2. CONVERTIMOS EL STRING A TIME.TIME
	// Definimos el formato esperado para la fecha. "2006-01-02" es el layout de Go para "YYYY-MM-DD".
	const layoutISO = "2006-01-02"
	maxDate, err := time.Parse(layoutISO, maxDateStr)
	if err != nil {
		// Si hay un error, significa que el formato de la fecha es incorrecto.
		response.ResponseError(w, validations.ErrInvalidFormatDate, http.StatusBadRequest)
		return
	}

	// 2. LLAMAR AL SERVICIO PARA GENERAR EL REPORTE
	// No hay cuerpo en la petición (es un GET), así que pasamos directamente al servicio.
	results, err := h.saleService.GetPendingQuotasReport(r.Context(), companyNameID, &maxDate)
	if err != nil {
		// Si hay un error en la capa de servicio, respondemos con un error interno del servidor.
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	// Si no se encontraron resultados, devolvemos un slice vacío en lugar de un error.
	// Esto es una práctica común en APIs REST para listas de recursos.
	if results == nil {
		results = make([]dto.ClientSalesWithQuotasResponse, 0)
	}

	// 3. RESPONDER CON LOS DATOS OBTENIDOS Y UN ESTADO 200 OK
	response.ResponseSuccess(w, results, http.StatusOK)
}

// GetPaymentsReportHandler maneja la petición para el reporte de pagos.
func (h *ReportHandler) GetPaymentsReportHandler(w http.ResponseWriter, r *http.Request) {
	companyNameID, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	// Parsear parámetros de paginación
	queryParams := r.URL.Query()
	page, _ := strconv.ParseInt(queryParams.Get("page"), 10, 64)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.ParseInt(queryParams.Get("limit"), 10, 64)
	if limit < 1 {
		limit = 10
	}

	response.ResponseSuccess(w, companyNameID, http.StatusOK)

	/*// Llamar al servicio de reportes
	results, err := h.reportService.GetPaymentsReport(r.Context(), companyNameID, page, limit)
	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	response.ResponseSuccess(w, results, http.StatusOK)*/
}
