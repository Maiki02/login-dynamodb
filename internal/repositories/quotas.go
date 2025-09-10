package repositories

import (
	"context"
	"myproject/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const QUOTA_COLLECTION = "quotas"

// QuotaRepository maneja la persistencia de las cuotas.
// Almacena el cliente de MongoDB para poder acceder a diferentes bases de datos dinámicamente.
type QuotaRepository struct {
	client *mongo.Client
}

// NewQuotaRepository crea una nueva instancia del repositorio de cuotas.
// Recibe el cliente de mongo, no una base de datos específica.
func NewQuotaRepository(client *mongo.Client) *QuotaRepository {
	return &QuotaRepository{
		client: client,
	}
}

// CreateMany inserta múltiples cuotas en una base de datos específica.
// Recibe el nombre de la base de datos (nameDB) como parámetro.
// El contexto (ctx) puede ser el de una transacción.
func (r *QuotaRepository) CreateMany(ctx context.Context, nameDB string, quotas []interface{}) error {
	// Obtiene la colección correcta de la base de datos especificada.
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	// Realiza la operación de inserción.
	_, err := collection.InsertMany(ctx, quotas)
	return err
}

// FindByIDs busca múltiples cuotas por un array de IDs.
// Devuelve un slice de cuotas. Es crucial para validar que todas las cuotas solicitadas existen.
func (r *QuotaRepository) FindByIDs(ctx context.Context, nameDB string, ids []primitive.ObjectID) ([]models.Quota, error) {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	// Usamos el operador $in para buscar todos los documentos cuyos _id estén en el slice de IDs.
	filter := bson.M{"_id": bson.M{"$in": ids}}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotas []models.Quota
	if err = cursor.All(ctx, &quotas); err != nil {
		return nil, err
	}

	return quotas, nil
}

// --- NUEVA FUNCIÓN ---
// FindUnpaidBySaleID busca todas las cuotas de una venta que no están completamente pagadas.
// Las ordena por número de cuota para asegurar el procesamiento secuencial correcto.
func (r *QuotaRepository) FindUnpaidBySaleID(ctx context.Context, nameDB string, saleID primitive.ObjectID) ([]models.Quota, error) {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	// Filtro para buscar cuotas de una venta específica que no tengan el estado "pagada".
	filter := bson.M{
		"sale_id": saleID,
		"status":  bson.M{"$ne": models.QuotaPaid},
	}

	// Opciones para ordenar los resultados por el número de cuota en orden ascendente.
	// Esto es CRÍTICO para que el pago se aplique a las cuotas más antiguas primero.
	opts := options.Find().SetSort(bson.D{{Key: "quota_number", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var quotas []models.Quota
	if err = cursor.All(ctx, &quotas); err != nil {
		return nil, err
	}

	return quotas, nil
}

// UpdateAsPaid utiliza BulkWrite para actualizar múltiples cuotas de forma atómica y eficiente.
// Recibe las cuotas que ya fueron leídas para construir las operaciones de actualización.
func (r *QuotaRepository) UpdateAsPaid(ctx context.Context, nameDB string, quotas []models.Quota, paymentID primitive.ObjectID) error {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	// Creamos un slice para almacenar todas las operaciones de actualización.
	var writes []mongo.WriteModel

	// Iteramos sobre cada cuota que necesitamos actualizar.
	for _, quota := range quotas {
		// Creamos un modelo de actualización para cada cuota.
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": quota.ID}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"status":            models.QuotaPaid,
					"paid_amount_cents": quota.AmountCents, // <-- Usamos el valor numérico real, no una referencia de texto.
				},
				"$push": bson.M{"payment_ids": paymentID},
			})

		writes = append(writes, model)
	}

	// Si no hay operaciones para realizar, no hacemos nada.
	if len(writes) == 0 {
		return nil
	}

	// Ejecutamos todas las operaciones de actualización en un solo lote.
	// Esto es atómico dentro de una transacción.
	_, err := collection.BulkWrite(ctx, writes)
	return err
}

// BulkUpdateStatus actualiza un conjunto de cuotas con sus nuevos estados y montos pagados.
// Está diseñada para el pago secuencial, donde las cuotas pueden quedar parcialmente pagadas.
func (r *QuotaRepository) BulkUpdateStatus(ctx context.Context, nameDB string, quotas []models.Quota, paymentID primitive.ObjectID) error {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)
	var writes []mongo.WriteModel

	// Iteramos sobre cada cuota que ha sido modificada en la lógica del servicio.
	for _, quota := range quotas {
		// Creamos un modelo de actualización para cada cuota.
		// Este modelo es más flexible que UpdateAsPaid.
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": quota.ID}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"status":            quota.Status,          // El nuevo estado (puede ser "pagada" o seguir "pendiente")
					"paid_amount_cents": quota.PaidAmountCents, // El nuevo monto acumulado pagado
				},
				"$addToSet": bson.M{"payment_ids": paymentID}, // Usamos $addToSet para evitar duplicar el ID del pago si se aplica a varias cuotas
			})

		writes = append(writes, model)
	}

	if len(writes) == 0 {
		return nil
	}

	// Ejecutamos todas las operaciones de actualización en un solo lote.
	_, err := collection.BulkWrite(ctx, writes)
	return err
}

// BulkRevertStatus revierte los cambios en un conjunto de cuotas.
// Nota: Este método es conceptualmente incorrecto para revertir montos parciales. Lo corregiré en el servicio.
func (r *QuotaRepository) BulkRevertStatus(ctx context.Context, nameDB string, quotasToRevert []models.Quota, paymentID primitive.ObjectID) error {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)
	var writes []mongo.WriteModel

	for _, quota := range quotasToRevert {
		// Determinamos el nuevo estado de la cuota.
		newStatus := models.QuotaPending
		if time.Now().UTC().After(quota.ExpirationDate) {
			newStatus = models.QuotaOverdue
		}

		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": quota.ID}).
			SetUpdate(bson.M{
				// Decrementamos el monto pagado.
				"$inc": bson.M{"paid_amount_cents": -quota.PaidAmountCents},
				// Actualizamos el estado.
				"$set": bson.M{"status": newStatus},
				// Quitamos el ID del pago del arreglo.
				"$pull": bson.M{"payment_ids": paymentID},
			})
		writes = append(writes, model)
	}

	if len(writes) == 0 {
		return nil
	}

	_, err := collection.BulkWrite(ctx, writes)
	return err
}

// QuotaRevertInstruction contiene los datos necesarios para revertir una única cuota.
// Se usa como parámetro para no pasar modelos complejos al repositorio.
type QuotaRevertInstruction struct {
	ID             primitive.ObjectID
	AmountToRevert int64
	NewStatus      models.QuotaStatus
}

// BulkRevert revierte los cambios en un conjunto de cuotas usando instrucciones específicas.
// Esta es la forma correcta, ya que el repositorio no toma decisiones de negocio.
func (r *QuotaRepository) BulkRevert(ctx context.Context, nameDB string, instructions []QuotaRevertInstruction, paymentID primitive.ObjectID) error {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	if len(instructions) == 0 {
		return nil // No hay nada que hacer
	}

	var writes []mongo.WriteModel

	for _, inst := range instructions {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": inst.ID}).
			SetUpdate(bson.M{
				// Decrementamos el monto pagado.
				"$inc": bson.M{"paid_amount_cents": -inst.AmountToRevert},
				// Actualizamos el estado.
				"$set": bson.M{"status": inst.NewStatus},
				// Quitamos el ID del pago del arreglo.
				"$pull": bson.M{"payment_ids": paymentID},
			})
		writes = append(writes, model)
	}

	_, err := collection.BulkWrite(ctx, writes)
	return err
}

// QuotaRescheduleInstruction contiene los datos necesarios para reprogramar una cuota.
type QuotaRescheduleInstruction struct {
	ID                primitive.ObjectID
	NewExpirationDate time.Time
	NewStatus         models.QuotaStatus
}

// BulkReschedule actualiza un conjunto de cuotas con nuevas fechas y estados.
func (r *QuotaRepository) BulkReschedule(ctx context.Context, nameDB string, instructions []QuotaRescheduleInstruction) error {
	collection := r.client.Database(nameDB).Collection(QUOTA_COLLECTION)

	if len(instructions) == 0 {
		return nil
	}

	var writes []mongo.WriteModel
	for _, inst := range instructions {
		model := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": inst.ID}).
			SetUpdate(bson.M{
				"$set": bson.M{
					"expiration_date": inst.NewExpirationDate,
					"status":          inst.NewStatus,
					"updated_at":      time.Now().UTC(),
				},
			})
		writes = append(writes, model)
	}

	_, err := collection.BulkWrite(ctx, writes)
	return err
}
