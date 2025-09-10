package services

import (
	"context"
	"errors"
	"fmt"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/request"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// QuotaService contiene la lógica de negocio para las cuotas.
type QuotaService struct {
	dbClient  *mongo.Client
	quotaRepo *repositories.QuotaRepository
}

// NewQuotaService crea una nueva instancia del servicio de cuotas.
func NewQuotaService(client *mongo.Client, qr *repositories.QuotaRepository) *QuotaService {
	return &QuotaService{
		dbClient:  client,
		quotaRepo: qr,
	}
}

// RescheduleQuotas actualiza múltiples cuotas de forma atómica y validada.
func (s *QuotaService) RescheduleQuotas(ctx context.Context, companyNameID string, req request.RescheduleQuotasRequest) error {
	session, err := s.dbClient.StartSession()
	if err != nil {
		return fmt.Errorf("error al iniciar la sesión de MongoDB: %w", err)
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		// 1. EXTRAER IDS Y BUSCAR TODAS LAS CUOTAS DE UNA VEZ
		var quotaIDs []primitive.ObjectID
		for _, u := range req.Updates {
			quotaIDs = append(quotaIDs, u.QuotaID)
		}

		fetchedQuotas, err := s.quotaRepo.FindByIDs(sessCtx, companyNameID, quotaIDs)
		if err != nil {
			return nil, fmt.Errorf("error al buscar las cuotas: %w", err)
		}

		// 2. VALIDACIONES DE NEGOCIO

		// Validación 1: ¿Se encontraron todas las cuotas solicitadas?
		if len(fetchedQuotas) != len(req.Updates) {
			return nil, errors.New("una o más de las cuotas especificadas no existen")
		}

		// Para una búsqueda eficiente, creamos un mapa de las cuotas encontradas.
		quotaMap := make(map[primitive.ObjectID]models.Quota)
		for _, q := range fetchedQuotas {
			// Validación 2: ¿Alguna de las cuotas ya está pagada?
			if q.Status == models.QuotaPaid {
				return nil, fmt.Errorf("la cuota con ID %s ya está pagada y no puede ser modificada", q.ID.Hex())
			}
			quotaMap[q.ID] = q
		}

		var instructions []repositories.QuotaRescheduleInstruction
		today := time.Now().Truncate(24 * time.Hour)

		for _, update := range req.Updates {
			// Validación 3: La nueva fecha no puede ser anterior a hoy.
			if update.NewExpirationDate.Before(today) {
				return nil, fmt.Errorf("la nueva fecha de expiración para la cuota %s no puede ser en el pasado", update.QuotaID.Hex())
			}

			// Validación 4: El nuevo estado no puede ser "pagada".
			if update.NewStatus == models.QuotaPaid {
				return nil, errors.New("no se puede cambiar el estado de una cuota a 'pagada' desde este endpoint")
			}

			// Validación 5 (Extra): El estado debe ser uno de los predefinidos.
			switch update.NewStatus {
			case models.QuotaPending, models.QuotaOverdue, models.QuotaDelinquent, models.QuotaCancelled:
				// Estado válido, continuamos
			default:
				return nil, fmt.Errorf("el estado '%s' no es válido", update.NewStatus)
			}

			instructions = append(instructions, repositories.QuotaRescheduleInstruction{
				ID:                update.QuotaID,
				NewExpirationDate: update.NewExpirationDate,
				NewStatus:         update.NewStatus,
			})
		}

		// 3. EJECUTAR LA ACTUALIZACIÓN
		if err := s.quotaRepo.BulkReschedule(sessCtx, companyNameID, instructions); err != nil {
			return nil, fmt.Errorf("error al actualizar las cuotas en la base de datos: %w", err)
		}

		return nil, nil // ¡Todo OK! La transacción hará COMMIT.
	})

	return err
}
