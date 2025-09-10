package routes

import (
	"myproject/cmd/middlewares"
	"myproject/internal/db" // Asumo que aquí tienes la conexión a la DB
	"myproject/internal/handlers"
	"myproject/internal/repositories"
	"myproject/internal/services"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	// 1. Configuramos dependencias

	// Obtenemos la conexión principal a la base de datos
	dbClient := db.GetDBClient()

	// A. Creamos instancias de los REPOSITORIOS
	userRepo := repositories.NewUserRepository()

	// B. Creamos instancias de los SERVICIOS, inyectando los repositorios
	sessionService := services.NewSessionService(userRepo, companyRepo)

	// C. Creamos instancias de los HANDLERS, inyectando los servicios
	sessionHandler := handlers.NewSessionHandler(sessionService)

	// 2. REGISTRO DE RUTAS
	router := mux.NewRouter()

	// A. Configuración de middlewares
	router.Use(middlewares.BodySizeLimitMiddleware)
	router.Use(middlewares.LimitRequestsMiddleware)
	router.Use(middlewares.EnableCORSMiddleware)
	router.Use(middlewares.LoggingMiddleware)
	router.Use(middlewares.AuthMiddleware)

	// B. Configuración de rutas
	//------------- AUTH ----------------\\
	// Estos handlers siguen siendo funciones simples por ahora.
	// En el futuro los actualizarás al mismo patrón.
	router.HandleFunc("/auth/register", sessionHandler.Register).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/login", sessionHandler.LoginHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/auth/refresh-token", sessionHandler.RefreshTokenHandler).Methods("GET", "OPTIONS")

	//------------- HEALTH CHECK ---------------\\
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET", "OPTIONS")

	return router
}
