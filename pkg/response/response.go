package response

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Data   interface{} `json:"data"`
	Error  interface{} `json:"error"`
	Status int         `json:"status"`
}

func ResponseError(w http.ResponseWriter, message error, status int) {
	response := Response{
		Data:   nil,
		Error:  message.Error(),
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)               // Establecer el código de estado HTTP
	json.NewEncoder(w).Encode(response) // Enviar el JSON como respuesta
}

func ResponseSuccess(w http.ResponseWriter, data interface{}, status int) {
	response := Response{
		Data:   data,
		Error:  nil,
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)               // Establecer el código de estado HTTP
	json.NewEncoder(w).Encode(response) // Enviar el JSON como respuesta
}
