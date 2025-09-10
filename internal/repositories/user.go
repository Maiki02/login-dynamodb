package repositories

import (
	"context"
	"myproject/internal/db"
	"myproject/internal/models"
	"myproject/pkg/validations"
)

const DB_USER_NAME = "Users_DB"
const COLLECTION_USER = "users"

// UserRepository define los métodos para interactuar con el almacenamiento de usuarios.
type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByFilter(filter map[string]interface{}) (*models.User, error)
	UpdateUser(id string, updates map[string]interface{}) error
}

// userRepository implementa la interfaz UserRepository.
type userRepository struct{}

// NewUserRepository crea una nueva instancia de userRepository.
func NewUserRepository() UserRepository {
	return &userRepository{}
}

func (r *userRepository) CreateUser(user *models.User) error {
	_, err := db.InsertDocument(DB_USER_NAME, COLLECTION_USER, user)
	return err
}

func (r *userRepository) UpdateUser(id string, updates map[string]interface{}) error {
	return db.UpdateDocumentByID(DB_USER_NAME, COLLECTION_USER, id, updates)
}

/*
func (r *userRepository) GetClientsByFilter(filter map[string]interface{}) (*[]models.Client, error) {
	cursor, err := db.FindDocuments(COLLECTION_CLIENT, filter)
	if err != nil {
		return nil, err
	}
	var clients []models.Client
	if err = cursor.All(context.Background(), &clients); err != nil {
		return nil, err
	}

	return &clients, nil
}*/

func (r *userRepository) GetUserByFilter(filter map[string]interface{}) (*models.User, error) {
	cursor, err := db.FindDocuments(DB_USER_NAME, COLLECTION_USER, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var user models.User
	if cursor.Next(context.Background()) {
		if err = cursor.Decode(&user); err != nil {
			return nil, err
		}
		return &user, nil
	}
	return nil, validations.ErrDocumentNotFound
}
