package db

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var dynamoClient *dynamodb.Client

// ConnectDynamoDB inicializa la conexión a DynamoDB
func ConnectDynamoDB() {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	dynamoClient = dynamodb.NewFromConfig(cfg)
	log.Println("Connected to DynamoDB successfully")
}

// GetDynamoClient retorna el cliente de DynamoDB
func GetDynamoClient() *dynamodb.Client {
	if dynamoClient == nil {
		log.Fatal("DynamoDB client not initialized. Call ConnectDynamoDB() first.")
	}
	return dynamoClient
}

// DisconnectDynamoDB - DynamoDB no requiere desconexión explícita
// Pero mantenemos la función para consistencia con la arquitectura
func DisconnectDynamoDB() {
	log.Println("DynamoDB connection closed")
	dynamoClient = nil
}
