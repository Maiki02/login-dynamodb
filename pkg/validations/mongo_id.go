package validations

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ValidateAndFormatMongoID toma una solicitud HTTP, extrae un ID de los vars de Gorilla Mux,
// verifica si es un ID hexadecimal válido de MongoDB y retorna el ID con el sufijo "_DB".
// Si el ID no es válido, retorna un error.
func ValidateAndFormatMongoID(r *http.Request) (string, error) {
	vars := mux.Vars(r)
	companyID := vars["company_id"]

	// Verificar si el ID es un hexadecimal válido para MongoDB
	_, err := primitive.ObjectIDFromHex(companyID)
	if err != nil {
		// Si hay un error, significa que el string no es un ObjectID válido
		return "", ErrCompanyNotFound
	}
	// Retornar el ID con el sufijo "_DB"
	return companyID + "_DB", nil
}
