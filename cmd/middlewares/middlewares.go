package middlewares

import (
	"context"
	"log"
	tokens "myproject/pkg/jwt"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"net/http"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/time/rate"
)

var excludedRoutes = []string{
	"/auth/login",
	"/auth/register",
	"/auth/refresh-token",
	"/auth/forgot-password",
	"/auth/reset-password",
	"/auth/activate",

	"/health",

	"/webhook/wuzapi",
	"/accept-invitation",
}

/*var wuzapiRoutes = []string{
	"/webhook/wuzapi",
}*/

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verifica si la ruta actual está en las rutas excluidas
		for _, route := range excludedRoutes {
			if route == r.URL.Path {
				next.ServeHTTP(w, r) // Continúa sin verificar el token
				return
			}
		}

		println("Host:", r.Host)
		// Verificamos que si la ruta es de wuzapi, la solicitud, debe venir de un dominio especifico.
		/*for _, route := range wuzapiRoutes {
			if route == r.URL.Path {
				if !strings.Contains(r.Host, "192.168.100.3:9000") {
					response.ResponseError(w, validations.ErrInvalidRequest, http.StatusBadRequest)
					return
				}
			}
		}*/

		// Valida el token para las demás rutas
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.ResponseError(w, validations.ErrInvalidToken, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokens.GetClaims(tokenString)
		if err != nil || claims == nil {
			response.ResponseError(w, validations.ErrInvalidToken, http.StatusUnauthorized)
			return
		}

		// 3. Extrae el ID del claim usando la clave del mapa y conviértelo a string.
		idString, ok := (*claims)["id"].(string)
		if !ok {
			// Este error ocurre si el claim "id" no existe o no es un string.
			response.ResponseError(w, validations.ErrInvalidUserID, http.StatusUnauthorized)
			return
		}

		// 4. Convierte el ID de string a ObjectID
		userID, err := primitive.ObjectIDFromHex(idString)
		if err != nil {
			// Este error ocurre si el string no es un ObjectID válido.
			response.ResponseError(w, validations.ErrInvalidUserID, http.StatusUnauthorized)
			return
		}

		// 5. Crea un nuevo contexto con el ID del usuario
		ctx := context.WithValue(r.Context(), request.UserContextKey, userID)

		// 6. Llama al siguiente handler con la petición que incluye el nuevo contexto
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func EnableCORSMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Si es una solicitud OPTIONS, termina aquí
		if r.Method == "OPTIONS" {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Momento en que inicia el procesamiento
		start := time.Now()

		// Imprimimos la información clave de la petición entrante
		log.Printf(
			"Incoming Request -> Method: %s | URI: %s | Remote Addr: %s | User-Agent: %s",
			r.Method,
			r.URL.RequestURI(),
			r.RemoteAddr,
			r.Header.Get("User-Agent"),
		)

		// Pasamos la petición al siguiente middleware o al handler final
		next.ServeHTTP(w, r)

		// Esta línea se ejecuta después de que la petición fue respondida
		log.Printf(
			"Finished Request -> URI: %s | Duration: %s",
			r.URL.RequestURI(),
			time.Since(start),
		)
	})
}

// IPRateLimiter almacena los limitadores para cada IP
var (
	mu        sync.Mutex
	ipLimiter = make(map[string]*rate.Limiter)
)

// LimitRequestsMiddleware limita las peticiones por IP
func LimitRequestsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		// Obtiene la IP del cliente
		ip := r.RemoteAddr // O una versión más robusta que considere proxies

		// Crea un nuevo limitador si no existe para esta IP
		if _, found := ipLimiter[ip]; !found {
			// Permite 100 eventos por minuto (aprox 1.6 por segundo)
			ipLimiter[ip] = rate.NewLimiter(rate.Every(time.Minute/100), 100)
		}

		// Si la petición no está permitida, retorna un error
		if !ipLimiter[ip].Allow() {
			mu.Unlock()
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

//----------- BODY SIZE LIMIT MIDDLEWARE -----------\\

func BodySizeLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Limita el tamaño del body a 1MB (ajusta según tus necesidades)
		r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)
		next.ServeHTTP(w, r)
	})
}
