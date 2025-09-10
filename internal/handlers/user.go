package handlers

import (
	"myproject/pkg/response"
	"net/http"
)

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	response.ResponseSuccess(w, nil, http.StatusOK)
}

/*
import (
	"encoding/json"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"
)

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	//vars := mux.Vars(r)
	//id := vars["id"]

	//var updates RequestUserUpdate
	updates := map[string]interface{}{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
	}

	//TODO: validate updates

	response.ResponseSuccess(w, nil, http.StatusOK)
}
*/
