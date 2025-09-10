package handlers

/*
import (
	"encoding/json"
	"net/http"
	"time"

	"myproject/internal/models"
	"myproject/internal/services"
	tokens "myproject/pkg/jwt"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"

	"github.com/gorilla/mux"
)
func CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var orderReq request.CreateOrderRequest

	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	nameDb, err := tokens.GetBdNameInToken(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	code := services.GetLastCodeOrder(nameDb)
	order, err := models.NewOrder(code, orderReq.Products, orderReq.Client,
		orderReq.DeliveryMethod, orderReq.DeliveryCost,
		orderReq.Address, orderReq.ClientObservation, orderReq.Adjustment,
		orderReq.PaymentMethod, orderReq.InternalObservation)

	if err != nil {
		response.ResponseError(w, err, http.StatusInternalServerError)
		return
	}

	if err := services.CreateOrder(nameDb, order); err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, order, http.StatusCreated)
}

func GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	nameDb, err := tokens.GetBdNameInToken(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	startDateParam := r.URL.Query().Get("startDate")
	endDateParam := r.URL.Query().Get("endDate")
	if startDateParam == "" && endDateParam == "" {
		response.ResponseError(w, validations.ErrInvalidQueryParams, http.StatusBadRequest)
		return
	}

	var parsedStartDate, parsedEndDate *time.Time

	if startDateParam != "" {
		startDate, err := time.Parse("2006-01-02", startDateParam)
		if err != nil {
			response.ResponseError(w, validations.ErrParsedDate, http.StatusBadRequest)
			return
		}
		startOfDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
		parsedStartDate = &startOfDay
	}

	if endDateParam != "" {
		endDate, err := time.Parse("2006-01-02", endDateParam)
		if err != nil {
			response.ResponseError(w, validations.ErrParsedDate, http.StatusBadRequest)
			return
		}
		endOfDay := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, endDate.Location())
		parsedEndDate = &endOfDay
	}

	orders, err := services.GetOrdersByDate(nameDb, parsedStartDate, parsedEndDate)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, orders, http.StatusOK)
}

func UpdateOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var updates request.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	nameDb, err := tokens.GetBdNameInToken(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	order, err := services.GetOrderByID(nameDb, id)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	err = order.ValidateWithUpdates(updates)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	updatesMap := validations.ParseStructToMap(updates)

	if err := services.UpdateOrder(nameDb, id, updatesMap); err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	response.ResponseSuccess(w, updates, http.StatusOK)
}

func ChangeStatusOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	var requestStatus request.ChangeStatusOrderRequest

	nameDb, err := tokens.GetBdNameInToken(r)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	if id == "" {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&requestStatus); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
		return
	}

	order, err := services.GetOrderByID(nameDb, id)
	if err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	if err := services.ChangeStatusOrder(nameDb, id, requestStatus.Status, order); err != nil {
		response.ResponseError(w, err, http.StatusBadRequest)
		return
	}

	order.Status = requestStatus.Status
	response.ResponseSuccess(w, order, http.StatusOK)
}
*/
