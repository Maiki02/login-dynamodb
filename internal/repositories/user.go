package repositories

import (
	"context"
	"myproject/internal/models"
	"myproject/pkg/validations"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// getUsersTableName retorna el nombre de la tabla de usuarios desde variables de entorno
func getUsersTableName() string {
	tableName := os.Getenv("DYNAMODB_TABLE_USERS")
	if tableName == "" {
		return "users" // nombre por defecto
	}
	return tableName
}

// UserRepository define los métodos para interactuar con el almacenamiento de usuarios en DynamoDB.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, user *models.User) error
}

// userRepository implementa la interfaz UserRepository usando DynamoDB.
type userRepository struct {
	dynamoClient *dynamodb.Client
}

// NewUserRepository crea una nueva instancia de userRepository.
func NewUserRepository(client *dynamodb.Client) UserRepository {
	return &userRepository{
		dynamoClient: client,
	}
}

// CreateUser crea un nuevo usuario en DynamoDB
func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	// Convertir el modelo a atributos de DynamoDB
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	// Realizar la operación PutItem
	_, err = r.dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(getUsersTableName()),
		Item:      item,
	})

	return err
}

// GetUserByID obtiene un usuario por su ID
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	result, err := r.dynamoClient.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(getUsersTableName()),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: id},
		},
	})

	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return nil, validations.ErrDocumentNotFound
	}

	var user models.User
	err = attributevalue.UnmarshalMap(result.Item, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByEmail obtiene un usuario por su email usando scan temporal
// TODO: Implementar GSI para mejor performance en producción
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	// Por ahora usamos Scan ya que no tenemos GSI configurado
	// En producción debería usarse un GSI para mejor performance
	result, err := r.dynamoClient.Scan(ctx, &dynamodb.ScanInput{
		TableName:        aws.String(getUsersTableName()),
		FilterExpression: aws.String("contact_info.#email.#address = :email"),
		ExpressionAttributeNames: map[string]string{
			"#email":   "email",
			"#address": "address",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})

	if err != nil {
		return nil, err
	}

	if len(result.Items) == 0 {
		return nil, validations.ErrDocumentNotFound
	}

	var user models.User
	err = attributevalue.UnmarshalMap(result.Items[0], &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser actualiza un usuario existente
func (r *userRepository) UpdateUser(ctx context.Context, id string, user *models.User) error {
	// Convertir el modelo a atributos de DynamoDB
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return err
	}

	// Realizar la operación PutItem (actualización completa)
	_, err = r.dynamoClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(getUsersTableName()),
		Item:      item,
	})

	return err
}
