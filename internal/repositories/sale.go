package repositories

import (
	"context"
	"myproject/internal/dto"
	"myproject/internal/models"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const SALE_COLLECTION = "sales"
const COUNTERS_COLLECTION = "counters"

// SaleRepository maneja la persistencia de las ventas.
// Almacena el cliente de MongoDB para poder acceder a diferentes bases de datos dinámicamente.
type SaleRepository struct {
	client *mongo.Client
}

// NewSaleRepository crea una nueva instancia del repositorio de ventas.
// Recibe el cliente de mongo, no una base de datos específica.
func NewSaleRepository(client *mongo.Client) *SaleRepository {
	return &SaleRepository{
		client: client,
	}
}

// Create inserta una nueva venta en una base de datos específica.
// Se le debe pasar el contexto de la sesión si se usa en una transacción.
func (r *SaleRepository) Create(ctx context.Context, nameDB string, sale *models.Sale) error {
	collection := r.client.Database(nameDB).Collection(SALE_COLLECTION)
	_, err := collection.InsertOne(ctx, sale)
	return err
}

// GetNextSaleNumber obtiene un número de venta incremental y atómico de una base de datos específica.
func (r *SaleRepository) GetNextSaleNumber(ctx context.Context, nameDB string) (int64, error) {
	collection := r.client.Database(nameDB).Collection(COUNTERS_COLLECTION)

	filter := bson.M{"_id": "sale_number_seq"}
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

// GetSales recupera todas las ventas de una base de datos, incluyendo los datos completos del cliente y las cuotas.
func (r *SaleRepository) GetSales(ctx context.Context, nameDB string) ([]models.SaleResponse, error) {
	collection := r.client.Database(nameDB).Collection(SALE_COLLECTION)

	// Este es el "pipeline" de agregación, ahora con la sintaxis correcta.
	pipeline := mongo.Pipeline{
		// Etapa 1: Unir con la colección 'clients'
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "clients"},
				{Key: "localField", Value: "client_id"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "client_info"},
			}},
		},
		// Etapa 2: Unir con la colección 'quotas'
		bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "quotas"},
				{Key: "localField", Value: "quota_ids"},
				{Key: "foreignField", Value: "_id"},
				{Key: "as", Value: "quotas_info"},
			}},
		},
		// Etapa 3: "Desenrollar" el resultado del cliente.
		bson.D{
			{Key: "$unwind", Value: "$client_info"},
		},
		// Etapa 4: Proyectar los campos finales en el formato de SaleResponse
		bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 1},
				{Key: "sale_number", Value: 1},
				{Key: "products", Value: 1},
				{Key: "status", Value: 1},
				{Key: "total_amount_cents", Value: 1},
				{Key: "sale_date", Value: 1},
				{Key: "created_at", Value: 1},
				{Key: "client", Value: "$client_info"},
				{Key: "quotas", Value: "$quotas_info"},
			}},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sales []models.SaleResponse
	if err = cursor.All(ctx, &sales); err != nil {
		return nil, err
	}

	return sales, nil
}

// FindByID busca una única venta por su ID en una base de datos específica.
// Es útil para validar que una venta existe antes de realizar operaciones complejas.
func (r *SaleRepository) FindByID(ctx context.Context, nameDB string, id primitive.ObjectID) (*models.Sale, error) {
	collection := r.client.Database(nameDB).Collection(SALE_COLLECTION)

	var sale models.Sale
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&sale)
	if err != nil {
		return nil, err // Devuelve mongo.ErrNoDocuments si no se encuentra
	}

	return &sale, nil
}

// UpdateSaleAfterPayment actualiza una venta después de un pago y, opcionalmente, la finaliza.
// Esta función está diseñada para ser usada dentro de una transacción.
func (r *SaleRepository) UpdateSaleAfterPayment(ctx context.Context, nameDB string, saleID primitive.ObjectID, paymentID primitive.ObjectID, amountPaidCents int64, isCompleted bool) error {
	collection := r.client.Database(nameDB).Collection(SALE_COLLECTION)

	filter := bson.M{"_id": saleID}

	// Preparamos la actualización base que siempre se ejecuta
	update := bson.M{
		"$inc":  bson.M{"collected_amount": amountPaidCents},
		"$push": bson.M{"payment_ids": paymentID},
	}

	// Si el pago completa la venta, modificamos la actualización
	if isCompleted {
		// En lugar de decrementar el pendiente, lo seteamos a 0 para asegurar consistencia
		// y cambiamos el estado de la venta a "completada".
		update["$set"] = bson.M{
			"status":               models.StatusCompleted,
			"pending_amount_cents": 0,
		}
	} else {
		// Si la venta no se completa, simplemente decrementamos el monto pendiente
		update["$inc"].(bson.M)["pending_amount_cents"] = -amountPaidCents
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// RevertPaymentUpdates revierte los cambios de un pago en la venta.
func (r *SaleRepository) RevertPaymentUpdates(ctx context.Context, nameDB string, saleID primitive.ObjectID, paymentID primitive.ObjectID, amountToRevert int64) error {
	collection := r.client.Database(nameDB).Collection(SALE_COLLECTION)
	filter := bson.M{"_id": saleID}

	// Revertimos los montos y quitamos el ID del pago.
	// Forzamos el estado a "en_progreso" como un estado seguro.
	// Una lógica más avanzada podría re-evaluar si la venta queda "en_mora".
	update := bson.M{
		"$inc": bson.M{
			"collected_amount_cents": -amountToRevert,
			"pending_amount_cents":   amountToRevert,
		},
		"$pull": bson.M{"payment_ids": paymentID},
		"$set":  bson.M{"status": models.StatusInProgress},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// GetPendingQuotasByClient realiza una agregación para obtener las cuotas pendientes por cliente
// que vencen antes de la fecha especificada.
/*func (r *SaleRepository) GetPendingQuotasByClient(ctx context.Context, nameDB string, maxExpirationDate time.Time) ([]dto.ClientPendingQuotasResponse, error) {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	pendingStatuses := []models.QuotaStatus{
		models.QuotaPending,
		models.QuotaOverdue,
		models.QuotaDelinquent,
	}

	// Obtenemos la fecha actual para el nuevo filtro.
	// Usamos el inicio del día para asegurar que la comparación sea consistente.
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	pipeline := mongo.Pipeline{
		// ETAPA 1 (MODIFICADA): Filtrar cuotas usando $or para cumplir una de dos condiciones.
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "$or", Value: bson.A{
				// Condición 1: Lógica original (status específico Y fecha anterior a maxExpirationDate)
				bson.D{
					{Key: "status", Value: bson.D{{Key: "$in", Value: pendingStatuses}}},
					{Key: "expiration_date", Value: bson.D{{Key: "$lt", Value: maxExpirationDate}}},
				},
				// Condición 2: Nueva lógica (fecha entre hoy y maxExpirationDate, sin importar el status)
				bson.D{
					{Key: "expiration_date", Value: bson.D{
						{Key: "$gte", Value: today},
						{Key: "$lte", Value: maxExpirationDate},
					}},
				},
			}},
		}}},

		// Etapa 2: Unir con 'sales' para obtener detalles de la venta.
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "sales"},
			{Key: "localField", Value: "sale_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "sale_info"},
		}}},

		// Etapa 3: Desenrollar 'sale_info'.
		bson.D{{Key: "$unwind", Value: "$sale_info"}},

		// Etapa 4: Filtrar ventas que NO estén 'completada'.
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "sale_info.status", Value: bson.D{{Key: "$ne", Value: models.StatusCompleted}}},
		}}},

		// Etapa 5: Unir con 'clients'.
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "clients"},
			{Key: "localField", Value: "sale_info.client_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "client_info"},
		}}},

		// Etapa 6: Desenrollar 'client_info'.
		bson.D{{Key: "$unwind", Value: "$client_info"}},

		// ETAPA 7 (MODIFICADA): Agrupar por cliente, incluyendo los nuevos campos en el $push.
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$client_info._id"},
			{Key: "client_data", Value: bson.D{{Key: "$first", Value: "$client_info"}}},
			{Key: "pending_quotas", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "payment_ids", Value: "$payment_ids"}, // Se agrega 'payment_ids'
					{Key: "quota_number", Value: "$quota_number"},
					{Key: "expiration_date", Value: "$expiration_date"},
					{Key: "amount_cents", Value: "$amount_cents"},
					{Key: "coin", Value: "$coin"}, // Se agrega 'coin'
					{Key: "paid_amount_cents", Value: "$paid_amount_cents"},
					{Key: "status", Value: "$status"},
					{Key: "sale_info", Value: bson.D{
						{Key: "_id", Value: "$sale_info._id"},
						{Key: "sale_number", Value: "$sale_info.sale_number"},
						{Key: "sale_date", Value: "$sale_info.sale_date"},
					}},
				}},
			}},
		}}},

		// Etapa 8: Proyectar el resultado final para que coincida con nuestro DTO.
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "client_info", Value: bson.D{
				{Key: "_id", Value: "$client_data._id"},
				{Key: "name", Value: "$client_data.name"},
				{Key: "last_name", Value: "$client_data.last_name"},
				{Key: "identification", Value: "$client_data.identification"},
			}},
			{Key: "pending_quotas", Value: "$pending_quotas"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []dto.ClientPendingQuotasResponse
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}*/
// GetClientSalesWithPendingQuotas realiza una agregación para obtener las cuotas pendientes,
// agrupadas por cliente y por venta, que vencen antes de una fecha específica.
func (r *SaleRepository) GetClientSalesWithPendingQuotas(ctx context.Context, nameDB string, maxExpirationDate time.Time) ([]dto.ClientSalesWithQuotasResponse, error) {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	pendingStatuses := []models.QuotaStatus{
		models.QuotaPending,
		models.QuotaOverdue,
		models.QuotaDelinquent,
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	pipeline := mongo.Pipeline{
		// ETAPA 1: Filtrar cuotas por estado y fecha. Sin cambios.
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "$or", Value: bson.A{
				bson.D{
					{Key: "status", Value: bson.D{{Key: "$in", Value: pendingStatuses}}},
					{Key: "expiration_date", Value: bson.D{{Key: "$lt", Value: maxExpirationDate}}},
				},
				bson.D{
					{Key: "expiration_date", Value: bson.D{
						{Key: "$gte", Value: today},
						{Key: "$lte", Value: maxExpirationDate},
					}},
				},
			}},
		}}},

		// ETAPAS 2-6: Unir con ventas y clientes. Sin cambios.
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "sales"}, {Key: "localField", Value: "sale_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "sale_info"},
		}}},
		bson.D{{Key: "$unwind", Value: "$sale_info"}},
		bson.D{{Key: "$match", Value: bson.D{
			{Key: "sale_info.status", Value: bson.D{{Key: "$ne", Value: models.StatusCompleted}}},
		}}},
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "clients"}, {Key: "localField", Value: "sale_info.client_id"}, {Key: "foreignField", Value: "_id"}, {Key: "as", Value: "client_info"},
		}}},
		bson.D{{Key: "$unwind", Value: "$client_info"}},

		// ETAPA 7 (MODIFICADA): Agrupar por CLIENTE y VENTA.
		// El _id ahora es un objeto compuesto para crear un grupo por cada venta de cada cliente.
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "client_id", Value: "$client_info._id"},
				{Key: "sale_id", Value: "$sale_info._id"},
			}},
			{Key: "client_data", Value: bson.D{{Key: "$first", Value: "$client_info"}}},
			{Key: "sale_data", Value: bson.D{{Key: "$first", Value: "$sale_info"}}},
			{Key: "quotas", Value: bson.D{
				{Key: "$push", Value: bson.D{
					{Key: "_id", Value: "$_id"},
					{Key: "payment_ids", Value: "$payment_ids"},
					{Key: "quota_number", Value: "$quota_number"},
					{Key: "expiration_date", Value: "$expiration_date"},
					{Key: "amount_cents", Value: "$amount_cents"},
					{Key: "coin", Value: "$coin"},
					{Key: "paid_amount_cents", Value: "$paid_amount_cents"},
					{Key: "status", Value: "$status"},
					{Key: "sale_info", Value: bson.D{
						{Key: "_id", Value: "$sale_info._id"},
						{Key: "sale_number", Value: "$sale_info.sale_number"},
						{Key: "sale_date", Value: "$sale_info.sale_date"},
					}},
				}},
			}},
		}}},

		// ETAPA 8 (MODIFICADA): Proyectar el resultado final para que coincida con el nuevo DTO.
		bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "client_info", Value: bson.D{
				{Key: "_id", Value: "$client_data._id"},
				{Key: "name", Value: "$client_data.name"},
				{Key: "last_name", Value: "$client_data.last_name"},
				{Key: "identification", Value: "$client_data.identification"},
			}},
			{Key: "sale_info", Value: bson.D{
				{Key: "_id", Value: "$sale_data._id"},
				{Key: "sale_number", Value: "$sale_data.sale_number"},
				{Key: "sale_date", Value: "$sale_data.sale_date"},
			}},
			{Key: "quotas", Value: "$quotas"},
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// La variable de resultados ahora usa el nuevo tipo de DTO.
	var results []dto.ClientSalesWithQuotasResponse
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// FindPaginated para ventas
func (r *SaleRepository) FindPaginated(ctx context.Context, nameDB string, page, limit int64, search, sortBy, sortOrder string) ([]models.SaleResponse, int64, error) {
	collection := r.client.Database(nameDB).Collection(SALE_COLLECTION)

	// Pipeline de Agregación
	var pipeline mongo.Pipeline

	// --- Etapa 1: $lookup para unir con clientes ---
	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "clients"},
			{Key: "localField", Value: "client_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "client_info"},
		}},
	})

	// --- Etapa 2: $unwind para "desenvolver" el array de cliente ---
	// Usamos preserveNullAndEmptyArrays para no perder ventas si un cliente fue eliminado
	pipeline = append(pipeline, bson.D{
		{Key: "$unwind", Value: bson.D{
			{Key: "path", Value: "$client_info"},
			{Key: "preserveNullAndEmptyArrays", Value: true},
		}},
	})

	// --- Etapa 3: $match para el filtrado ---
	if search != "" {
		searchRegex := bson.M{"$regex": search, "$options": "i"}
		var saleNumber int64
		if num, err := strconv.ParseInt(search, 10, 64); err == nil {
			saleNumber = num
		}

		matchStage := bson.D{
			{Key: "$or", Value: bson.A{
				bson.M{"sale_number": saleNumber},
				bson.M{"client_info.name": searchRegex},
				bson.M{"client_info.last_name": searchRegex},
				bson.M{"client_info.identification.number": searchRegex},
			}},
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchStage}})
	}

	// --- Etapa 4: Contar documentos ANTES de la paginación ---
	// Creamos un pipeline de conteo que es una copia del principal hasta este punto
	countPipeline := append(mongo.Pipeline{}, pipeline...)
	countPipeline = append(countPipeline, bson.D{{Key: "$count", Value: "totalDocs"}})

	cursor, err := collection.Aggregate(ctx, countPipeline)
	if err != nil {
		return nil, 0, err
	}
	// El defer se encargará de cerrar el cursor al final de la función,
	// pasándole el contexto. Esto es más seguro y limpio.
	defer cursor.Close(ctx)

	var result []struct{ TotalDocs int64 }
	// AÑADIDO: Manejo de error para cursor.All()
	if err := cursor.All(ctx, &result); err != nil {
		return nil, 0, err
	}

	var totalDocs int64 = 0
	if len(result) > 0 {
		totalDocs = result[0].TotalDocs
	}

	// --- Etapa 5: Ordenamiento ($sort) ---
	sortStage := bson.D{{Key: "sale_date", Value: -1}} // Orden por defecto
	if sortBy != "" {
		order := 1
		if strings.ToLower(sortOrder) == "desc" {
			order = -1
		}
		sortStage = bson.D{{Key: sortBy, Value: order}}
	}
	pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortStage}})

	// --- Etapa 6 y 7: Paginación ($skip y $limit) ---
	pipeline = append(pipeline, bson.D{{Key: "$skip", Value: (page - 1) * limit}})
	pipeline = append(pipeline, bson.D{{Key: "$limit", Value: limit}})

	// --- Etapa 8: $lookup para unir con cuotas ---
	// Es importante hacer este lookup DESPUÉS de la paginación para mejorar el rendimiento
	pipeline = append(pipeline, bson.D{
		{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "quotas"},
			{Key: "localField", Value: "quota_ids"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "quotas"},
		}},
	})

	// --- Etapa 9: Proyección final para darle forma al resultado ---
	// Esto asegura que los campos coincidan con tu struct `SaleResponse`
	pipeline = append(pipeline, bson.D{
		{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "sale_number", Value: 1},
			{Key: "products", Value: 1},
			{Key: "loan", Value: 1}, // <-- CAMPO NUEVO AÑADIDO
			{Key: "status", Value: 1},
			{Key: "total_amount_cents", Value: "$total_amount_cents"}, // Corregido para usar el nombre correcto del campo
			{Key: "sale_date", Value: 1},
			{Key: "created_at", Value: 1},
			{Key: "client", Value: "$client_info"},
			{Key: "quotas", Value: "$quotas"},
			{Key: "payment_ids", Value: 1},
		}},
	})

	// --- Ejecutar el pipeline principal para obtener los datos ---
	finalCursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, 0, err
	}
	defer finalCursor.Close(ctx)

	var sales []models.SaleResponse
	if err = finalCursor.All(ctx, &sales); err != nil {
		return nil, 0, err
	}

	return sales, totalDocs, nil
}
