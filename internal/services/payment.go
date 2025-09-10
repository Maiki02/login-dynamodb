// services/payment_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/validations"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const DEFAULT_COIN = "ARS" // Definimos una moneda por defecto, puede ser modificada según la lógica de negocio

// PaymentService contiene la lógica de negocio para los pagos.
type PaymentService struct {
	dbClient    *mongo.Client // Para iniciar sesiones de transacción
	saleRepo    repositories.SaleRepository
	quotaRepo   repositories.QuotaRepository
	paymentRepo repositories.PaymentRepository
	userRepo    repositories.UserRepository
}

// NewPaymentService crea una nueva instancia del servicio de pagos.
func NewPaymentService(client *mongo.Client,
	sRepo *repositories.SaleRepository,
	qRepo *repositories.QuotaRepository,
	pRepo *repositories.PaymentRepository,
	uRepo *repositories.UserRepository) *PaymentService {
	return &PaymentService{
		dbClient:    client,
		saleRepo:    *sRepo,
		quotaRepo:   *qRepo,
		paymentRepo: *pRepo,
		userRepo:    *uRepo,
	}
}

// ProcessBulkPayment procesa el pago de múltiples cuotas en una única transacción atómica.
func (s *PaymentService) ProcessBulkPayment(ctx context.Context, companyNameID string, saleID primitive.ObjectID, req request.PayQuotasRequest, collectorID primitive.ObjectID) (*models.Payment, error) {

	// 1. VALIDACIÓN PREVIA (FUERA DE LA TRANSACCIÓN)
	// Verificamos que la venta exista antes de iniciar una transacción costosa.
	_, err := s.saleRepo.FindByID(ctx, companyNameID, saleID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, validations.ErrSellNotFound
		}
		return nil, err
	}

	// 2. INICIO DE LA SESIÓN DE MONGODB
	session, err := s.dbClient.StartSession()
	if err != nil {
		return nil, fmt.Errorf("error al iniciar la sesión de MongoDB: %w", err)
	}
	defer session.EndSession(ctx)

	var createdPayment *models.Payment

	// 3. EJECUCIÓN DE LA TRANSACCIÓN
	// El driver de MongoDB se encarga del commit y rollback automáticamente.
	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		// --- Toda la lógica de aquí en adelante es ATÓMICA ---
		// A.1 OBTENEMOS LA VENTA PARA SABER SU ESTADO ACTUAL
		sale, err := s.saleRepo.FindByID(sessCtx, companyNameID, saleID)
		if err != nil {
			return nil, fmt.Errorf("error al buscar la venta: %w", err)
		}

		// A.2 BÚSQUEDA Y VALIDACIÓN DE CUOTAS
		quotas, err := s.quotaRepo.FindByIDs(sessCtx, companyNameID, req.QuotasIDs)
		if err != nil {
			return nil, fmt.Errorf("error al buscar las cuotas: %w", err)
		}
		if len(quotas) != len(req.QuotasIDs) {
			return nil, errors.New("una o más de las cuotas especificadas no fueron encontradas")
		}

		var totalAmountCents int64
		var affectedQuotas []models.AffectedQuota

		for _, q := range quotas {
			// Validación de Pertenencia: ¿Son todas de la misma venta?
			if q.SaleID != saleID {
				return nil, fmt.Errorf("la cuota con ID %s no pertenece a la venta indicada", q.ID.Hex())
			}
			// Validación de Estado: ¿Alguna ya fue pagada?
			if q.Status == models.QuotaPaid {
				return nil, fmt.Errorf("la cuota #%d ya ha sido pagada previamente", q.QuotaNumber)
			}

			pendingAmount := q.AmountCents - q.PaidAmountCents
			if pendingAmount <= 0 {
				// Esto no debería pasar si el estado no es "pagada", pero es una buena salvaguarda.
				continue
			}

			// Acumulamos solo el monto pendiente.
			totalAmountCents += pendingAmount

			// Añadimos el desglose correcto al slice.
			affectedQuotas = append(affectedQuotas, models.AffectedQuota{
				QuotaID:            q.ID,
				AmountAppliedCents: pendingAmount, // Se registra el monto que realmente se está pagando.
			})
		}

		// Si no hay nada que pagar, no tiene sentido crear un pago.
		if totalAmountCents <= 0 {
			return nil, errors.New("el monto total a pagar es cero o negativo, no se puede procesar el pago")
		}

		// B.1 OBTENER EL SIGUIENTE NÚMERO DE PAGO (AÑADIDO)
		paymentNumber, err := s.paymentRepo.GetNextPaymentNumber(sessCtx, companyNameID)
		if err != nil {
			return nil, fmt.Errorf("error al generar el número de pago: %w", err)
		}

		// B. CREACIÓN DEL DOCUMENTO DE PAGO
		paymentID := primitive.NewObjectID()
		newPayment := &models.Payment{
			ID:             paymentID,
			SaleID:         saleID,
			PaymentNumber:  paymentNumber,
			CollectorID:    collectorID,
			PaymentDate:    time.Now().UTC(),
			AmountCents:    totalAmountCents,
			Coin:           DEFAULT_COIN, // Usamos una moneda por defecto
			Method:         models.PaymentMethod(req.Method),
			Status:         models.PaymentCompleted,
			QuotasAffected: affectedQuotas, // Usamos una función helper
			CreatedAt:      time.Now().UTC(),
		}

		if err := s.paymentRepo.Create(sessCtx, companyNameID, newPayment); err != nil {
			return nil, fmt.Errorf("error al crear el documento de pago: %w", err)
		}

		// C. ACTUALIZACIÓN DE LAS CUOTAS
		// Pasamos el slice de cuotas que ya leímos y validamos.
		if err := s.quotaRepo.UpdateAsPaid(sessCtx, companyNameID, quotas, paymentID); err != nil {
			return nil, fmt.Errorf("error al actualizar las cuotas: %w", err)
		}

		// D. --- LÓGICA DE FINALIZACIÓN DE VENTA ---
		// Comprobamos si el monto acumulado MÁS el pago actual cubren el total de la venta.
		isSaleCompleted := (sale.CollectedAmountCents + totalAmountCents) >= sale.TotalAmountCents

		// Llamamos al nuevo método del repositorio de ventas.
		if err := s.saleRepo.UpdateSaleAfterPayment(sessCtx, companyNameID, saleID, paymentID, totalAmountCents, isSaleCompleted); err != nil {
			return nil, fmt.Errorf("error al actualizar la venta: %w", err)
		}

		createdPayment = newPayment
		return createdPayment, nil // ¡Todo OK! La transacción hará COMMIT.
	})

	if err != nil {
		return nil, err // Hubo un error, la transacción hizo ROLLBACK.
	}

	return result.(*models.Payment), nil
}

// buildAffectedQuotas es una función de ayuda para crear el desglose del pago.
/*func buildAffectedQuotas(quotas []models.Quota) []models.AffectedQuota {
	affected := make([]models.AffectedQuota, len(quotas))
	for i, q := range quotas {
		affected[i] = models.AffectedQuota{
			QuotaID:            q.ID,
			AmountAppliedCents: q.AmountCents, // Asumimos que cada cuota se paga por completo.
		}
	}
	return affected
}*/

// ProcessSequentialPayment ahora maneja sobrepagos y los convierte en saldo a favor.
func (s *PaymentService) ProcessSequentialPayment(ctx context.Context, companyNameID string, saleID primitive.ObjectID, req request.SequentialPaymentRequest, collectorID primitive.ObjectID) (*models.Payment, error) {

	// 1. VALIDACIÓN PREVIA (FUERA DE LA TRANSACCIÓN)
	// No necesitamos traer toda la venta aquí, solo confirmar que existe.
	_, err := s.saleRepo.FindByID(ctx, companyNameID, saleID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("la venta especificada no existe")
		}
		return nil, err
	}

	session, err := s.dbClient.StartSession()
	if err != nil {
		return nil, fmt.Errorf("error al iniciar la sesión de MongoDB: %w", err)
	}
	defer session.EndSession(ctx)

	var createdPayment *models.Payment

	result, err := session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// --- Lógica atómica ---

		// A. BÚSQUEDA DE DOCUMENTOS NECESARIOS DENTRO DE LA TRANSACCIÓN
		// Traemos la venta para obtener el ID del cliente.
		sale, err := s.saleRepo.FindByID(sessCtx, companyNameID, saleID)
		if err != nil {
			return nil, fmt.Errorf("error al buscar la venta: %w", err)
		}

		unpaidQuotas, err := s.quotaRepo.FindUnpaidBySaleID(sessCtx, companyNameID, saleID)
		if err != nil {
			return nil, fmt.Errorf("error al buscar cuotas pendientes: %w", err)
		}

		// Si no hay cuotas pendientes, no tiene sentido aceptar un pago.
		if len(unpaidQuotas) == 0 {
			return nil, errors.New("la venta no tiene cuotas pendientes de pago")
		}

		// B. AGREGACIÓN DE LOS PAGOS RECIBIDOS
		var totalAmountCents int64
		methodSet := make(map[models.PaymentMethod]bool)
		for _, p := range req.Payments {
			totalAmountCents += p.AmountCents
			methodSet[models.PaymentMethod(p.Method)] = true
		}
		if totalAmountCents <= 0 {
			return nil, errors.New("el monto total del pago debe ser positivo")
		}

		// C. APLICACIÓN SECUENCIAL DEL PAGO A LAS CUOTAS
		var amountToApply = totalAmountCents
		var affectedQuotas []models.AffectedQuota
		var quotasToUpdate []models.Quota

		for _, quota := range unpaidQuotas {
			if amountToApply <= 0 {
				break
			}
			pendingOnQuota := quota.AmountCents - quota.PaidAmountCents
			if pendingOnQuota <= 0 {
				continue
			}
			paymentForThisQuota := min(amountToApply, pendingOnQuota)
			quota.PaidAmountCents += paymentForThisQuota
			if quota.PaidAmountCents >= quota.AmountCents {
				quota.Status = models.QuotaPaid
			}
			quotasToUpdate = append(quotasToUpdate, quota)
			affectedQuotas = append(affectedQuotas, models.AffectedQuota{
				QuotaID:            quota.ID,
				AmountAppliedCents: paymentForThisQuota,
			})
			amountToApply -= paymentForThisQuota
		}

		// --- LÓGICA DE SOBREPAGO ---
		// Si al final del bucle `amountToApply` es mayor que cero, tenemos un sobrepago.
		var overpaymentNotes string
		if amountToApply > 0 {
			clientCollection := s.dbClient.Database(companyNameID).Collection("clients")
			filter := bson.M{"_id": sale.ClientID}

			// Usamos $inc para incrementar atómicamente el saldo del cliente.
			update := bson.M{"$inc": bson.M{"credit_balance_cents": amountToApply}}

			_, err := clientCollection.UpdateOne(sessCtx, filter, update)
			if err != nil {
				return nil, fmt.Errorf("error al acreditar saldo a favor al cliente: %w", err)
			}
			// Preparamos una nota para el documento de pago que sirva de auditoría.
			overpaymentNotes = fmt.Sprintf("Se acreditó un saldo a favor de %.2f al cliente.", float64(amountToApply)/100)
		}

		// D. CREACIÓN DEL DOCUMENTO DE PAGO... (con notas adicionales)
		paymentID := primitive.NewObjectID()
		paymentNumber, err := s.paymentRepo.GetNextPaymentNumber(sessCtx, companyNameID)
		if err != nil {
			return nil, err
		}

		var paymentMethod models.PaymentMethod
		if len(methodSet) == 1 {
			for m := range methodSet {
				paymentMethod = m
			}
		} else {
			paymentMethod = models.MethodOther
		}

		finalNotes := req.Notes
		if overpaymentNotes != "" {
			if finalNotes != "" {
				finalNotes += ". " + overpaymentNotes
			} else {
				finalNotes = overpaymentNotes
			}
		}

		newPayment := &models.Payment{
			ID:             paymentID,
			SaleID:         saleID,
			PaymentNumber:  paymentNumber,
			CollectorID:    collectorID,
			Coin:           DEFAULT_COIN,
			PaymentDate:    time.Now().UTC(),
			AmountCents:    totalAmountCents,
			Method:         paymentMethod,
			Status:         models.PaymentCompleted,
			QuotasAffected: affectedQuotas,
			Notes:          finalNotes,
			CreatedAt:      time.Now().UTC(),
		}

		if err := s.paymentRepo.Create(sessCtx, companyNameID, newPayment); err != nil {
			return nil, fmt.Errorf("error al crear el documento de pago: %w", err)
		}

		// E. ACTUALIZACIÓN DE LAS CUOTAS
		if err := s.quotaRepo.BulkUpdateStatus(sessCtx, companyNameID, quotasToUpdate, paymentID); err != nil {
			return nil, fmt.Errorf("error al actualizar las cuotas: %w", err)
		}

		// F. --- LÓGICA DE FINALIZACIÓN DE VENTA ---
		// La misma lógica que en la otra función.
		isSaleCompleted := (sale.CollectedAmountCents + totalAmountCents) >= sale.TotalAmountCents

		// Llamamos al nuevo método del repositorio de ventas.
		if err := s.saleRepo.UpdateSaleAfterPayment(sessCtx, companyNameID, saleID, paymentID, totalAmountCents, isSaleCompleted); err != nil {
			return nil, fmt.Errorf("error al actualizar la venta: %w", err)
		}

		createdPayment = newPayment
		return createdPayment, nil
	})

	if err != nil {
		return nil, err
	}

	return result.(*models.Payment), nil
}

// --- NUEVO HELPER ---
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// RevertPayment revierte un pago y todas sus consecuencias de forma atómica.
func (s *PaymentService) RevertPayment(ctx context.Context, companyNameID string, saleID, paymentID primitive.ObjectID) error {

	session, err := s.dbClient.StartSession()
	if err != nil {
		return fmt.Errorf("error al iniciar la sesión de MongoDB: %w", err)
	}
	defer session.EndSession(ctx)

	// Toda la lógica se ejecuta dentro de una transacción.
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		// 1. OBTENER DATOS NECESARIOS
		// Buscamos el pago que se quiere revertir.
		payment, err := s.paymentRepo.FindByID(sessCtx, companyNameID, paymentID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return nil, errors.New("El pago especificado no existe")
			}
			return nil, fmt.Errorf("Error al buscar el pago: %w", err)
		}

		// Buscamos la venta para acceder a su ClienteID
		sale, err := s.saleRepo.FindByID(sessCtx, companyNameID, saleID)
		if err != nil {
			return nil, fmt.Errorf("Error al buscar la venta: %w", err)
		}

		// 2. VALIDAR ESTADO
		if payment.Status == models.PaymentReverted {
			return nil, errors.New("Este pago ya ha sido revertido previamente")
		}
		if payment.SaleID != saleID {
			return nil, errors.New("El pago no pertenece a la venta indicada")
		}

		// 3. REVERTIR EL ESTADO DEL PAGO
		if err := s.paymentRepo.UpdateStatus(sessCtx, companyNameID, paymentID, models.PaymentReverted); err != nil {
			return nil, fmt.Errorf("error al actualizar el estado del pago: %w", err)
		}

		// 4. REVERTIR LAS CUOTAS AFECTADAS (--- REFACTORIZADO ---)

		// 4.1. El servicio prepara las instrucciones basado en las reglas de negocio.
		var instructions []repositories.QuotaRevertInstruction
		var totalAppliedInQuotas int64

		// Obtenemos las cuotas para saber su fecha de vencimiento.
		var quotaIDs []primitive.ObjectID
		for _, affected := range payment.QuotasAffected {
			quotaIDs = append(quotaIDs, affected.QuotaID)
		}
		quotas, err := s.quotaRepo.FindByIDs(sessCtx, companyNameID, quotaIDs)
		if err != nil {
			return nil, fmt.Errorf("error al buscar las cuotas afectadas: %w", err)
		}
		// Creamos un mapa para fácil acceso
		quotaMap := make(map[primitive.ObjectID]models.Quota)
		for _, q := range quotas {
			quotaMap[q.ID] = q
		}

		for _, affected := range payment.QuotasAffected {
			totalAppliedInQuotas += affected.AmountAppliedCents

			quota, ok := quotaMap[affected.QuotaID]
			if !ok {
				return nil, fmt.Errorf("inconsistencia de datos: la cuota %s afectada por el pago no fue encontrada", affected.QuotaID.Hex())
			}

			// Regla de negocio: determinar el nuevo estado.
			newStatus := models.QuotaPending
			if time.Now().UTC().After(quota.ExpirationDate) {
				newStatus = models.QuotaOverdue
			}

			instructions = append(instructions, repositories.QuotaRevertInstruction{
				ID:             affected.QuotaID,
				AmountToRevert: affected.AmountAppliedCents,
				NewStatus:      newStatus,
			})
		}

		// 4.2. El servicio le pasa las instrucciones claras al repositorio.
		if err := s.quotaRepo.BulkRevert(sessCtx, companyNameID, instructions, paymentID); err != nil {
			return nil, fmt.Errorf("error al revertir las cuotas: %w", err)
		}

		// 5. REVERTIR LOS MONTOS Y ESTADO DE LA VENTA
		if err := s.saleRepo.RevertPaymentUpdates(sessCtx, companyNameID, saleID, paymentID, payment.AmountCents); err != nil {
			return nil, fmt.Errorf("error al revertir los montos de la venta: %w", err)
		}

		// 6. REVERTIR EL SALDO A FAVOR (SI LO HUBO)
		creditToRevert := payment.AmountCents - totalAppliedInQuotas
		if creditToRevert > 0 {
			clientCollection := s.dbClient.Database(companyNameID).Collection("clients")
			filter := bson.M{"_id": sale.ClientID}
			update := bson.M{"$inc": bson.M{"credit_balance_cents": -creditToRevert}}
			if _, err := clientCollection.UpdateOne(sessCtx, filter, update); err != nil {
				return nil, fmt.Errorf("error al revertir el saldo a favor del cliente: %w", err)
			}
		}

		return nil, nil // ¡Todo OK! La transacción hará COMMIT.
	})

	return err
}

// GetPayments obtiene una lista de pagos basada en los filtros proporcionados.
func (s *PaymentService) GetPaymentsWithDetails(ctx context.Context, companyNameID string, filter request.FilterPaymentsRequest, page, limit int64) (*response.PaginatedResponse, error) {

	// 1. Obtener los pagos paginados desde el repositorio.
	paginatedResult, err := s.paymentRepo.GetByFilterWithDetails(ctx, companyNameID, filter, page, limit)
	if err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(paginatedResult.TotalDocs) / float64(limit)))

	// 2. Construye y devuelve la respuesta paginada.
	return &response.PaginatedResponse{
		Docs:        paginatedResult.Docs,
		TotalDocs:   paginatedResult.TotalDocs,
		Limit:       limit,
		TotalPages:  totalPages,
		Page:        page,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}, nil
}
