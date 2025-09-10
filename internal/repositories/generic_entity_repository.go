package repositories

import (
	"context"
	"errors"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrNotFound = errors.New("documento no encontrado")
var ErrDuplicateSlug = errors.New("el slug ya existe")

// GenericRepositoryInterface define los métodos de acceso a datos para entidades genéricas.
type GenericRepositoryInterface interface {
	Create(ctx context.Context, nameDB string, entity interface{}) (primitive.ObjectID, error)
	GetBySlug(ctx context.Context, nameDB, slug string, result interface{}) error
	Update(ctx context.Context, nameDB, slug string, updates bson.M) (interface{}, error)
	SlugExists(ctx context.Context, nameDB, slug string) (bool, error)
	FindPaginated(ctx context.Context, nameDB, search, sortBy, sortOrder string, page, limit int64, results interface{}) (int64, error)
}

// genericRepository implementa la interfaz con un cliente de MongoDB.
type genericRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewGenericRepository crea una instancia del repositorio genérico para una colección específica.
func NewGenericRepository(client *mongo.Client, collectionName string) func(string) GenericRepositoryInterface {
	return func(dbName string) GenericRepositoryInterface {
		return &genericRepository{
			client:     client,
			collection: client.Database(dbName).Collection(collectionName),
		}
	}
}

func (r *genericRepository) Create(ctx context.Context, nameDB string, entity interface{}) (primitive.ObjectID, error) {
	res, err := r.collection.InsertOne(ctx, entity)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return res.InsertedID.(primitive.ObjectID), nil
}

func (r *genericRepository) GetBySlug(ctx context.Context, nameDB, slug string, result interface{}) error {
	filter := bson.M{"slug": slug, "status": "activo"}
	err := r.collection.FindOne(ctx, filter).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (r *genericRepository) Update(ctx context.Context, nameDB, slug string, updates bson.M) (interface{}, error) {
	filter := bson.M{"slug": slug}
	update := bson.M{"$set": updates}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedDocument interface{}
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedDocument)

	return updatedDocument, err
}

func (r *genericRepository) SlugExists(ctx context.Context, nameDB, slug string) (bool, error) {
	filter := bson.M{"slug": slug}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindPaginated busca documentos de forma paginada y ordenada.
func (r *genericRepository) FindPaginated(ctx context.Context, nameDB, search, sortBy, sortOrder string, page, limit int64, results interface{}) (int64, error) {
	// 1. Construir el filtro de búsqueda
	filter := bson.M{"status": "activo"} // Solo traer entidades activas
	if search != "" {
		// Búsqueda "case-insensitive" por el campo 'name'
		filter["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	// 2. Contar el total de documentos que coinciden con el filtro
	totalDocs, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	// 3. Configurar opciones de paginación y ordenamiento
	findOptions := options.Find()
	findOptions.SetSkip((page - 1) * limit)
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

	// 4. Ejecutar la búsqueda
	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	// 5. Decodificar los resultados en el slice proporcionado
	if err = cursor.All(ctx, results); err != nil {
		return 0, err
	}

	return totalDocs, nil
}
