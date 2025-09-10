package routes

import (
	"myproject/cmd/middlewares"
	"myproject/internal/db" // Asumo que aquí tienes la conexión a la DB
	"myproject/internal/handlers"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/internal/services"
	"myproject/pkg/consts"

	"github.com/gorilla/mux"
)

func InitRoutes() *mux.Router {
	// 1. Configuramos dependencias

	// Obtenemos la conexión principal a la base de datos
	dbClient := db.GetDBClient()

	// A. Creamos instancias de los REPOSITORIOS
	clientRepo := repositories.NewClientRepository() // Este no necesitaba el dbClient
	saleRepo := repositories.NewSaleRepository(dbClient)
	quotaRepo := repositories.NewQuotaRepository(dbClient)
	paymentRepo := repositories.NewPaymentRepository(dbClient)
	userRepo := repositories.NewUserRepository()
	companyRepo := repositories.NewCompanyRepository()
	brandRepoFactory := repositories.NewGenericRepository(dbClient, consts.COLLECTION_BRANDS)
	productTypeRepoFactory := repositories.NewGenericRepository(dbClient, consts.COLLECTION_PRODUCT_TYPES)
	productRepo := repositories.NewProductRepository(dbClient)

	// B. Creamos instancias de los SERVICIOS, inyectando los repositorios
	clientService := services.NewClientService(clientRepo)
	saleService := services.NewSaleService(saleRepo, quotaRepo, clientRepo, &productRepo, dbClient)
	paymentService := services.NewPaymentService(dbClient, saleRepo, quotaRepo, paymentRepo, &userRepo)
	quotaService := services.NewQuotaService(dbClient, quotaRepo)
	//userService := services.NewUserService(userRepo, companyRepo)
	companyService := services.NewCompanyService(companyRepo)
	sessionService := services.NewSessionService(userRepo, companyRepo)
	reportService := services.NewReportService(paymentRepo) // B. Servicios genéricos
	brandService := services.NewGenericService(
		brandRepoFactory,
		func() models.Entity { return &models.Brand{} },
		func() interface{} { return &[]models.Brand{} },
	)
	productTypeService := services.NewGenericService(
		productTypeRepoFactory,
		func() models.Entity { return &models.ProductType{} },
		func() interface{} { return &[]models.ProductType{} },
	)
	productService := services.NewProductService(productRepo, brandService, productTypeService, dbClient)
	invitationService := services.NewInvitationService(userRepo, companyRepo)

	// C. Creamos instancias de los HANDLERS, inyectando los servicios
	clientHandler := handlers.NewClientHandler(clientService)
	saleHandler := handlers.NewSaleHandler(saleService)
	paymentHandler := handlers.NewPaymentHandler(paymentService)
	quotaHandler := handlers.NewQuotaHandler(quotaService)
	companyHandler := handlers.NewCompanyHandler(companyService)
	sessionHandler := handlers.NewSessionHandler(sessionService)
	reportsHandler := handlers.NewReportHandler(saleService, reportService)
	brandHandler := handlers.NewGenericHandler(brandService)
	productTypeHandler := handlers.NewGenericHandler(productTypeService)
	productHandler := handlers.NewProductHandler(productService)
	invitationHandler := handlers.NewInvitationHandler(invitationService)

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

	//------------ COMPANY ---------------\\
	router.HandleFunc("/companies/{id}", companyHandler.GetByID).Methods("GET", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/send-invitation", invitationHandler.SendInvitation).Methods("POST", "OPTIONS")
	router.HandleFunc("/accept-invitation", invitationHandler.AcceptInvitation).Methods("POST", "OPTIONS")

	//------------- CLIENT ----------------\\
	// Ahora las rutas llaman a los MÉTODOS de la INSTANCIA del handler
	router.HandleFunc("/companies/{company_id}/client", clientHandler.CreateClientHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/clients", clientHandler.GetClientsHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/client/{id}", clientHandler.UpdateClientHandler).Methods("PUT", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/client/{id}", clientHandler.DeleteClientHandler).Methods("DELETE", "OPTIONS")

	//------------- SALE ----------------\\
	router.HandleFunc("/companies/{company_id}/sale", saleHandler.CreateSaleHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/sales", saleHandler.GetSalesHandler).Methods("GET", "OPTIONS")

	//------------ PAYMENT ---------------\\
	router.HandleFunc("/companies/{company_id}/sales/{sale_id}/quotes/pay", paymentHandler.CreatePaymentHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/sales/{sale_id}/payments/sequential", paymentHandler.CreateSequentialPaymentHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/sales/{sale_id}/payments/{payment_id}/revert", paymentHandler.RevertPaymentHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/payments", paymentHandler.GetPaymentsHandler).Methods("GET", "OPTIONS")

	//------------ QUOTA ---------------\\
	router.HandleFunc("/companies/{company_id}/quotas/reschedule", quotaHandler.RescheduleQuotasHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/reports/pending-quotas-by-client", reportsHandler.GetPendingQuotasByClientHandler).Methods("GET", "OPTIONS")

	//------------ REPORT ---------------\\
	router.HandleFunc("/companies/{company_id}/reports/payments-by-client", reportsHandler.GetPaymentsReportHandler).Methods("GET", "OPTIONS")

	//------------- BRANDS ----------------\\
	router.HandleFunc("/companies/{company_id}/brands", brandHandler.Create).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/brands", brandHandler.GetAll).Methods("GET", "OPTIONS") // <-- RUTA NUEVA
	router.HandleFunc("/companies/{company_id}/brands/{slug}", brandHandler.GetBySlug).Methods("GET", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/brands/{slug}", brandHandler.Update).Methods("PUT", "OPTIONS")

	//------------- PRODUCT TYPES ----------------\\
	router.HandleFunc("/companies/{company_id}/product-types", productTypeHandler.Create).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/product-types", productTypeHandler.GetAll).Methods("GET", "OPTIONS") // <-- RUTA NUEVA
	router.HandleFunc("/companies/{company_id}/product-types/{slug}", productTypeHandler.GetBySlug).Methods("GET", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/product-types/{slug}", productTypeHandler.Update).Methods("PUT", "OPTIONS")

	//------------- PRODUCTS ----------------\\
	router.HandleFunc("/companies/{company_id}/product", productHandler.CreateProduct).Methods("POST", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/products", productHandler.GetProducts).Methods("GET", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/products/{id}", productHandler.GetProductByID).Methods("GET", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/products/{id}", productHandler.UpdateProduct).Methods("PUT", "OPTIONS")
	router.HandleFunc("/companies/{company_id}/products/{id}/variants/{sku}", productHandler.UpdateVariant).Methods("PUT", "OPTIONS")

	//------------- HEALTH CHECK ---------------\\
	router.HandleFunc("/health", handlers.HealthHandler).Methods("GET", "OPTIONS")

	return router
}
