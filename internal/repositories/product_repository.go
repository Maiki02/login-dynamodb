package repositories

import (
	"context"
	"errors"
	"myproject/internal/dto"
	"myproject/internal/models"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const productCollection = "products"

type ProductRepository interface {
	Create(ctx context.Context, nameDB string, product *models.Product) (*mongo.InsertOneResult, error)
	GetNextProductNumber(ctx context.Context, nameDB string) (int64, error)

	FindByID(ctx context.Context, nameDB string, id primitive.ObjectID) (*dto.ProductResponse, error)
	UpdateByID(ctx context.Context, nameDB string, id primitive.ObjectID, updates bson.M) (*mongo.UpdateResult, error)
	UpdateVariantBySKU(ctx context.Context, nameDB string, productID primitive.ObjectID, sku string, updates bson.M) (*mongo.UpdateResult, error)
	FindPaginated(ctx context.Context, nameDB string, page, limit int64, filters bson.M, sortBy, sortOrder string) ([]dto.ProductResponse, int64, error)
	DecrementVariantStock(ctx context.Context, nameDB string, productID primitive.ObjectID, sku string, quantity int) error
}

type productRepository struct {
	client *mongo.Client
}

func NewProductRepository(client *mongo.Client) ProductRepository {
	return &productRepository{client: client}
}

func (r *productRepository) Create(ctx context.Context, nameDB string, product *models.Product) (*mongo.InsertOneResult, error) {
	collection := r.client.Database(nameDB).Collection(productCollection)
	return collection.InsertOne(ctx, product)
}

func (r *productRepository) GetNextProductNumber(ctx context.Context, nameDB string) (int64, error) {
	collection := r.client.Database(nameDB).Collection("counters")
	filter := bson.M{"_id": "product_internal_code_seq"}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter struct {
		SequenceValue int64 `bson:"sequence_value"`
	}
	if err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter); err != nil {
		return 0, err
	}
	return counter.SequenceValue, nil
}

// FindByID ahora usa una agregación para "poblar" los datos de Brand y ProductType.
func (r *productRepository) FindByID(ctx context.Context, nameDB string, id primitive.ObjectID) (*dto.ProductResponse, error) {
	collection := r.client.Database(nameDB).Collection(productCollection)

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.D{{Key: "_id", Value: id}}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "brands"}, {Key: "localField", Value: "brand_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "brand"},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "product_types"}, {Key: "localField", Value: "product_type_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "product_type"},
		}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$brand"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$product_type"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []dto.ProductResponse
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	return &products[0], nil
}

func (r *productRepository) UpdateByID(ctx context.Context, nameDB string, id primitive.ObjectID, updates bson.M) (*mongo.UpdateResult, error) {
	collection := r.client.Database(nameDB).Collection(productCollection)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": updates}
	return collection.UpdateOne(ctx, filter, update)
}

func (r *productRepository) UpdateVariantBySKU(ctx context.Context, nameDB string, productID primitive.ObjectID, sku string, updates bson.M) (*mongo.UpdateResult, error) {
	collection := r.client.Database(nameDB).Collection(productCollection)
	filter := bson.M{"_id": productID, "variants.sku": sku}

	// Preparamos el update para que modifique solo los campos de la variante correcta
	updateFields := bson.M{}
	for key, value := range updates {
		updateFields["variants.$."+key] = value
	}
	update := bson.M{"$set": updateFields}

	return collection.UpdateOne(ctx, filter, update)
}

func (r *productRepository) FindPaginated(ctx context.Context, nameDB string, page, limit int64, filters bson.M, sortBy, sortOrder string) ([]dto.ProductResponse, int64, error) {
	collection := r.client.Database(nameDB).Collection(productCollection)

	// --- Pipeline de Conteo ---
	countPipeline := mongo.Pipeline{bson.D{{Key: "$match", Value: filters}}}
	countPipeline = append(countPipeline, bson.D{{Key: "$count", Value: "totalDocs"}})

	cursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var result []struct{ TotalDocs int64 }
	if err := cursor.All(ctx, &result); err != nil {
		return nil, 0, err
	}
	var totalDocs int64 = 0
	if len(result) > 0 {
		totalDocs = result[0].TotalDocs
	}

	// --- Pipeline de Datos ---
	pipeline := mongo.Pipeline{bson.D{{Key: "$match", Value: filters}}}

	// Ordenamiento
	sortStage := bson.D{{Key: "createdAt", Value: -1}} // Por defecto
	if sortBy != "" {
		order := 1
		if strings.ToLower(sortOrder) == "desc" {
			order = -1
		}
		sortStage = bson.D{{Key: sortBy, Value: order}}
	}
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortStage}})

	// Paginación
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: (page - 1) * limit}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})

	// Joins para poblar datos
	pipeline = append(pipeline,
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "brands"}, {Key: "localField", Value: "brand_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "brand"}}}},
		bson.D{{Key: "$lookup", Value: bson.D{{Key: "from", Value: "product_types"}, {Key: "localField", Value: "product_type_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "product_type"}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$brand"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
		bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$product_type"}, {Key: "preserveNullAndEmptyArrays", Value: true}}}},
	)

	dataCursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer dataCursor.Close(ctx)

	var products []dto.ProductResponse
	if err := dataCursor.All(ctx, &products); err != nil {
		return nil, 0, err
	}

	return products, totalDocs, nil
}

func (r *productRepository) DecrementVariantStock(ctx context.Context, nameDB string, productID primitive.ObjectID, sku string, quantity int) error {
	collection := r.client.Database(nameDB).Collection(productCollection)

	// El filtro asegura que actualizamos la variante correcta dentro del producto correcto.
	filter := bson.M{
		"_id":          productID,
		"variants.sku": sku,
	}

	// Usamos $inc para decrementar atómicamente la cantidad.
	// El valor negativo resta del total actual en la base de datos.
	update := bson.M{
		"$inc": bson.M{"variants.$.stock.current_quantity": -quantity},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// Opcional: Verificar si un documento fue realmente modificado.
	if result.ModifiedCount == 0 {
		return errors.New("no se pudo actualizar el stock para la variante " + sku)
	}

	return nil
}
