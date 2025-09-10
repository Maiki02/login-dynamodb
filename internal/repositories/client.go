package repositories

import (
	"context"
	"myproject/internal/db"
	"myproject/internal/models"
	"myproject/pkg/validations"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const COLLECTION_CLIENT = "clients"

// ClientRepository maneja la persistencia de los clientes.
// Agrupa todos los métodos relacionados con la base de datos de clientes.
type ClientRepository struct {
	// Este struct puede estar vacío si las funciones de tu paquete `db`
	// ya manejan la conexión globalmente, como parece ser tu caso.
}

// NewClientRepository crea una nueva instancia del repositorio de clientes.
func NewClientRepository() *ClientRepository {
	return &ClientRepository{}
}

// CreateClient se convierte en un método de ClientRepository.
func (r *ClientRepository) CreateClient(nameDB string, client *models.Client) error {
	_, err := db.InsertDocument(nameDB, COLLECTION_CLIENT, client)
	return err
}

// GetAllClients se convierte en un método de ClientRepository.
func (r *ClientRepository) GetAllClients(nameDB string) ([]models.Client, error) {
	var clients []models.Client
	cursor, err := db.FindDocuments(nameDB, COLLECTION_CLIENT, nil)
	if err != nil {
		return clients, err
	}
	if err = cursor.All(context.Background(), &clients); err != nil {
		return clients, err
	}
	return clients, nil
}

// Nuevo método FindPaginated en el repositorio de cliente
func (r *ClientRepository) FindPaginated(ctx context.Context, nameDB string, page, limit int64, search, sortBy, sortOrder string) ([]models.Client, int64, error) {
	collection := db.GetDBClient().Database(nameDB).Collection(COLLECTION_CLIENT)

	// --- 1. Construir el filtro de búsqueda ---
	filter := bson.M{"deleted_at": bson.M{"$exists": false}} // No traer clientes borrados
	if search != "" {
		// Búsqueda "case-insensitive" por nombre o apellido
		searchRegex := bson.M{"$regex": search, "$options": "i"}
		filter["$or"] = []bson.M{
			{"name": searchRegex},
			{"last_name": searchRegex},
		}
	}

	// --- 2. Contar el total de documentos que coinciden con el filtro ---
	totalDocs, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// --- 3. Configurar opciones de paginación y ordenamiento ---
	findOptions := options.Find()
	// Skip: Cuántos documentos saltar
	findOptions.SetSkip((page - 1) * limit)
	// Limit: Cuántos documentos traer
	findOptions.SetLimit(limit)

	// Ordenamiento
	if sortBy != "" {
		order := 1 // Ascendente por defecto
		if strings.ToLower(sortOrder) == "desc" {
			order = -1 // Descendente
		}
		findOptions.SetSort(bson.D{{Key: sortBy, Value: order}})
	} else {
		// Orden por defecto si no se especifica
		findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})
	}

	// --- 4. Ejecutar la búsqueda ---
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var clients []models.Client
	if err = cursor.All(ctx, &clients); err != nil {
		return nil, 0, err
	}

	return clients, totalDocs, nil
}

func (r *ClientRepository) UpdateClient(nameDB, id string, updates map[string]interface{}) error {
	return db.UpdateDocumentByID(nameDB, COLLECTION_CLIENT, id, updates)
}

func (r *ClientRepository) UpdateandFindClient(nameDB, id string, updates map[string]interface{}) (*models.Client, error) {

	// Llamamos a la nueva función de la DB
	singleResult := db.UpdateAndFindDocumentByID(nameDB, COLLECTION_CLIENT, id, updates)

	if err := singleResult.Err(); err != nil {
		return nil, err
	}

	var client models.Client
	if err := singleResult.Decode(&client); err != nil {
		return nil, err
	}

	return &client, nil
}

func (r *ClientRepository) GetClientByID(nameDB, id string) (*models.Client, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	return r.GetClientByFilter(nameDB, map[string]interface{}{"_id": hexID})
}

// GetClientsByFilter se convierte en un método de ClientRepository.
func (r *ClientRepository) GetClientsByFilter(nameDB string, filter map[string]interface{}) (*[]models.Client, error) {
	cursor, err := db.FindDocuments(nameDB, COLLECTION_CLIENT, filter)
	if err != nil {
		return nil, err
	}
	var clients []models.Client
	if err = cursor.All(context.Background(), &clients); err != nil {
		return nil, err
	}

	return &clients, nil
}

// GetClientByFilter se convierte en un método de ClientRepository.
func (r *ClientRepository) GetClientByFilter(nameDB string, filter map[string]interface{}) (*models.Client, error) {
	cursor, err := db.FindDocuments(nameDB, COLLECTION_CLIENT, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var client models.Client
	if cursor.Next(context.Background()) {
		if err = cursor.Decode(&client); err != nil {
			return nil, err
		}
		return &client, nil
	}
	return nil, validations.ErrDocumentNotFound
}
