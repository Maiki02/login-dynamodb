package repositories

import (
	"context"
	"myproject/internal/dto"
	"myproject/internal/models"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const PAYMENT_COLLECTION = "payments"

//const COUNTERS_COLLECTION = "counters"

// PaymentRepository maneja la persistencia de los pagos.
// Almacena el cliente de MongoDB para poder acceder a diferentes bases de datos dinámicamente.
type PaymentRepository struct {
	client *mongo.Client
}

// NewPaymentRepository crea una nueva instancia del repositorio de ventas.
// Recibe el cliente de mongo, no una base de datos específica.
func NewPaymentRepository(client *mongo.Client) *PaymentRepository {
	return &PaymentRepository{
		client: client,
	}
}

// Create inserta un nuevo documento de pago en la base de datos.
// Está diseñado para ser usado dentro de una transacción.
func (r *PaymentRepository) Create(ctx context.Context, nameDB string, payment *models.Payment) error {
	collection := r.client.Database(nameDB).Collection(PAYMENT_COLLECTION)
	_, err := collection.InsertOne(ctx, payment)
	return err
}

// GetNextPaymentNumber obtiene un número de pago incremental y atómico.
// Utiliza una colección 'counters' para llevar la secuencia de forma segura.
func (r *PaymentRepository) GetNextPaymentNumber(ctx context.Context, nameDB string) (int64, error) {
	collection := r.client.Database(nameDB).Collection("counters")

	// Usamos un ID de documento diferente para el contador de pagos.
	filter := bson.M{"_id": "payment_number_seq"}
	update := bson.M{"$inc": bson.M{"sequence_value": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter struct {
		SequenceValue int64 `bson:"sequence_value"`
	}

	err := collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return 0, err
	}

	return counter.SequenceValue, nil
}

// FindByID busca un único pago por su ID.
func (r *PaymentRepository) FindByID(ctx context.Context, nameDB string, id primitive.ObjectID) (*models.Payment, error) {
	collection := r.client.Database(nameDB).Collection(PAYMENT_COLLECTION)
	var payment models.Payment
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	return &payment, err
}

// UpdateStatus actualiza el estado de un pago.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, nameDB string, id primitive.ObjectID, status models.PaymentStatus) error {
	collection := r.client.Database(nameDB).Collection(PAYMENT_COLLECTION)
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now().UTC()}}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}
func (r *PaymentRepository) GetByFilterWithDetails(ctx context.Context, nameDB string, filter request.FilterPaymentsRequest, page, limit int64) (*response.PaginatedResponse, error) {
	collection := r.client.Database(nameDB).Collection(PAYMENT_COLLECTION)

	// --- 1. Construir el filtro de búsqueda ---
	match := bson.M{}
	if len(filter.Statuses) > 0 {
		match["status"] = bson.M{"$in": filter.Statuses}
	}
	if filter.StartDate != nil || filter.EndDate != nil {
		dateFilter := bson.M{}
		if filter.StartDate != nil {
			dateFilter["$gte"] = filter.StartDate
		}
		if filter.EndDate != nil {
			dateFilter["$lt"] = filter.EndDate.AddDate(0, 0, 1)
		}
		match["payment_date"] = dateFilter
	}

	pipeline := mongo.Pipeline{}

	if len(match) > 0 {
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: match}})
	}

	// --- 2. Contar el total de documentos ---
	countPipeline := append(mongo.Pipeline{}, pipeline...)
	countPipeline = append(countPipeline, bson.D{{Key: "$count", Value: "totalDocs"}})
	cursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, err
	}
	var result []struct{ TotalDocs int64 }
	if err := cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	var totalDocs int64 = 0
	if len(result) > 0 {
		totalDocs = result[0].TotalDocs
	}
	cursor.Close(ctx)

	// --- 3. Paginación y Joins ---
	pipeline = append(pipeline,
		bson.D{{Key: "$sort", Value: bson.D{{Key: "payment_date", Value: -1}}}},
		bson.D{{Key: "$skip", Value: (page - 1) * limit}},
		bson.D{{Key: "$limit", Value: limit}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "sales"},
			{Key: "localField", Value: "sale_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sale_info"},
		}}},
		bson.D{{Key: "$unwind", Value: "$sale_info"}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "clients"},
			{Key: "localField", Value: "sale_info.client_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "client_info"},
		}}},
		bson.D{{Key: "$unwind", Value: "$client_info"}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "quotas"},
			{Key: "localField", Value: "quotas_affected.quota_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "quotas_details"},
		}}},
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "client_info", Value: bson.D{
				{Key: "_id", Value: "$client_info._id"},
				{Key: "name", Value: "$client_info.name"},
				{Key: "last_name", Value: "$client_info.last_name"},
				{Key: "identification", Value: "$client_info.identification"},
			}},
			{Key: "sale_info", Value: bson.D{
				{Key: "_id", Value: "$sale_info._id"},
				{Key: "sale_number", Value: "$sale_info.sale_number"},
				{Key: "sale_date", Value: "$sale_info.sale_date"},
			}},
			{Key: "payment", Value: bson.D{
				{Key: "_id", Value: "$_id"},
				{Key: "payment_number", Value: "$payment_number"},
				{Key: "payment_date", Value: "$payment_date"},
				{Key: "amount_cents", Value: "$amount_cents"},
				{Key: "coin", Value: "$coin"},
				{Key: "method", Value: "$method"},
				{Key: "status", Value: "$status"},
				{Key: "affected_quotas", Value: bson.D{
					{Key: "$map", Value: bson.D{
						{Key: "input", Value: "$quotas_affected"},
						{Key: "as", Value: "affected"},
						{Key: "in", Value: bson.D{
							{Key: "amount_applied_cents", Value: "$$affected.amount_applied_cents"},
							{Key: "quota", Value: bson.D{
								{Key: "$arrayElemAt", Value: bson.A{
									bson.D{
										{Key: "$filter", Value: bson.D{
											{Key: "input", Value: "$quotas_details"},
											{Key: "as", Value: "q_detail"},
											{Key: "cond", Value: bson.D{{Key: "$eq", Value: bson.A{"$$q_detail._id", "$$affected.quota_id"}}}},
										}},
									},
									0,
								}},
							}},
						}},
					}},
				}},
			}},
		}}},
	)

	cursor, err = collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payments []dto.PaymentsReportResponse
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}

	return &response.PaginatedResponse{
		Docs:      payments,
		TotalDocs: totalDocs,
	}, nil
}
