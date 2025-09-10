package main

import (
	"context"
	"log"
	"myproject/cmd/routes"
	"myproject/internal/db"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("Could not load .env file, assuming production environment")
	}
}

// NO es necesaria la variable global 'httpAdapter'

func main() {
	db.ConnectDynamoDB()
	defer db.DisconnectDynamoDB()

	router := routes.InitRoutes()

	if _, ok := os.LookupEnv("LAMBDA_SERVER_PORT"); ok {
		// ESTAMOS EN ENTORNO LAMBDA 🚀
		log.Println("Running on AWS Lambda")

		// CORRECCIÓN: Declaramos el adaptador localmente con el operador :=
		// Go infiere el tipo correcto automáticamente.
		adapter := httpadapter.New(router)

		lambda.Start(func(ctx context.Context, req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			return adapter.ProxyWithContext(ctx, req)
		})

	} else {
		// ESTAMOS EN ENTORNO LOCAL 💻
		log.Println("Running on Local Machine")
		port := os.Getenv("PORT")
		if port == "" {
			port = "9000"
		}

		srv := &http.Server{
			Addr:    ":" + port,
			Handler: router,
		}

		go func() {
			log.Printf("Starting server on port %s", port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("ListenAndServe error: %s\n", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}
		log.Println("Server exiting")
	}
}
