package routes

import (
	"myproject/cmd/middlewares"
	"myproject/internal/db"
	"myproject/internal/handlers"
	"myproject/internal/repositories"
	"myproject/internal/services"
	"net/http"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	// 1. Configuramos dependencias siguiendo el patrón cebolla

	// Obtenemos la conexión principal a DynamoDB
	dynamoClient := db.GetDynamoClient()

	// A. Creamos instancias de los REPOSITORIOS (Repository Layer)
	userRepo := repositories.NewUserRepository(dynamoClient)

	// B. Creamos instancias de los SERVICIOS (Service Layer)
	sessionService := services.NewSessionService(userRepo)

	// C. Creamos instancias de los HANDLERS (Handler Layer)
	sessionHandler := handlers.NewSessionHandler(sessionService)

	// 2. REGISTRO DE RUTAS
	router := mux.NewRouter()

	// A. Configuración de middlewares
	router.Use(middlewares.BodySizeLimitMiddleware)
	router.Use(middlewares.LimitRequestsMiddleware)
	router.Use(middlewares.EnableCORSMiddleware)
	router.Use(middlewares.LoggingMiddleware)
	router.Use(middlewares.AuthMiddleware)

	// B. Configuración de rutas de autenticación
	router.HandleFunc("/auth/register", sessionHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/login", sessionHandler.LoginHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/refresh-token", sessionHandler.RefreshTokenHandler).Methods("GET", "OPTIONS")

	// C. Health check
	router.HandleFunc("/health", healthHandler).Methods("GET", "OPTIONS")

	return router
}

// healthHandler es un handler simple para verificar el estado de la API
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status": "healthy", "service": "login-dynamodb-api"}`))
}
