package handlers

import (
	"encoding/json"
	"errors"
	"myproject/internal/services"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductHandler struct {
	service services.ProductService
}

func NewProductHandler(s services.ProductService) *ProductHandler {
	return &ProductHandler{service: s}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	var req request.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	newProduct, err := h.service.CreateProduct(r.Context(), nameDB, req)
	if err != nil {
		// Aquí puedes mapear errores del servicio a códigos HTTP específicos
		// Por ejemplo, si err es ErrDuplicateSlug, podrías devolver 409 Conflict.
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, newProduct, http.StatusCreated)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]

	product, err := h.service.GetProductByID(r.Context(), nameDB, productID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			response.ResponseError(w, errors.New("producto no encontrado"), http.StatusNotFound)
		} else {
			response.ResponseError(w, err, http.StatusInternalServerError)
		}
		return
	}

	response.ResponseSuccess(w, product, http.StatusOK)
}

func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
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
		limit = 10
	}

	// Filtros y ordenamiento
	search := queryParams.Get("search")
	brandSlug := queryParams.Get("brand")
	typeSlug := queryParams.Get("type")
	sortBy := queryParams.Get("sortBy")
	sortOrder := queryParams.Get("sortOrder")

	paginatedResult, err := h.service.GetProducts(r.Context(), nameDB, page, limit, brandSlug, typeSlug, search, sortBy, sortOrder)
	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	response.ResponseSuccess(w, paginatedResult, http.StatusOK)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]

	var req request.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	_, err = h.service.UpdateProduct(r.Context(), nameDB, productID, req)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, map[string]string{"message": "Producto actualizado correctamente"}, http.StatusOK)
}

func (h *ProductHandler) UpdateVariant(w http.ResponseWriter, r *http.Request) {
	nameDB, err := validations.ValidateAndFormatMongoID(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	vars := mux.Vars(r)
	productID := vars["id"]
	sku := vars["sku"]

	var req request.UpdateVariantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	_, err = h.service.UpdateVariant(r.Context(), nameDB, productID, sku, req)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, map[string]string{"message": "Variante actualizada correctamente"}, http.StatusOK)
}
