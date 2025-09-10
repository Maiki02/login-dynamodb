package db

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dynamoClient *dynamodb.Client

// ConnectDynamoDB inicializa la conexión a DynamoDB
func ConnectDynamoDB() {
	ctx := context.Background()

	// Configurar opciones basadas en variables de entorno
	var optFns []func(*config.LoadOptions) error

	// Configurar región si está especificada
	if region := os.Getenv("AWS_REGION"); region != "" {
		optFns = append(optFns, config.WithRegion(region))
	}

	// Cargar configuración AWS
	cfg, err := config.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	// Crear cliente DynamoDB
	dynamoClient = dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		// Configurar endpoint personalizado si está especificado (para DynamoDB Local)
		if endpoint := os.Getenv("DYNAMODB_ENDPOINT"); endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	})

	// Log de información de configuración
	region := cfg.Region
	if region == "" {
		region = "default"
	}

	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "AWS DynamoDB"
	}

	log.Printf("DynamoDB client initialized successfully - Region: %s, Endpoint: %s", region, endpoint)
	log.Println("Note: DynamoDB uses HTTP requests, actual connection will be tested on first operation")
}

// GetDynamoClient retorna el cliente de DynamoDB
func GetDynamoClient() *dynamodb.Client {
	if dynamoClient == nil {
		log.Fatal("DynamoDB client not initialized. Call ConnectDynamoDB() first.")
	}
	return dynamoClient
}

// TestDynamoDBConnection verifica si realmente podemos conectarnos a DynamoDB
func TestDynamoDBConnection() error {
	if dynamoClient == nil {
		return errors.New("DynamoDB client not initialized")
	}

	// Intentar listar tablas como test de conectividad
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := dynamoClient.ListTables(ctx, &dynamodb.ListTablesInput{
		Limit: aws.Int32(1), // Solo necesitamos 1 para probar
	})

	if err != nil {
		log.Printf("❌ DynamoDB connection test failed: %v", err)
		return err
	}

	log.Println("✅ DynamoDB connection test successful")
	return nil
}

// DisconnectDynamoDB - DynamoDB no requiere desconexión explícita
// Pero mantenemos la función para consistencia con la arquitectura
func DisconnectDynamoDB() {
	log.Println("DynamoDB client cleaned up")
	dynamoClient = nil
}
