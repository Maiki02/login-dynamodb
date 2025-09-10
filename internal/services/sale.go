// sale_service.go
package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"myproject/internal/dto"
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

// SaleService contiene la lógica de negocio para las ventas.
type SaleService struct {
	saleRepo    *repositories.SaleRepository
	quotaRepo   *repositories.QuotaRepository
	clientRepo  *repositories.ClientRepository  // <-- El especialista en clientes
	productRepo *repositories.ProductRepository // <-- El especialista en productos
	dbClient    *mongo.Client                   // Cliente de Mongo para manejar la sesión
}

// NewSaleService crea una nueva instancia del servicio de ventas.
func NewSaleService(
	saleRepo *repositories.SaleRepository,
	quotaRepo *repositories.QuotaRepository,
	clientRepo *repositories.ClientRepository,
	productRepo *repositories.ProductRepository,
	dbClient *mongo.Client,
) *SaleService {
	return &SaleService{
		saleRepo:    saleRepo,
		quotaRepo:   quotaRepo,
		clientRepo:  clientRepo,
		productRepo: productRepo,
		dbClient:    dbClient,
	}
}

// CreateSale orquesta la creación de una venta de forma atómica.
// Puede incluir productos, un préstamo, o ambos.
func (s *SaleService) CreateSale(ctx context.Context, nameDB string, req request.CreateSaleRequest) (*models.Sale, error) {
	// ====================================================================
	// PASO 1: VALIDACIONES PREVIAS (FUERA DE LA TRANSACCIÓN)
	// ====================================================================
	if req.ClientID.IsZero() {
		return nil, errors.New("el ID del cliente es requerido")
	}
	if len(req.Products) == 0 && (req.Loan == nil || !req.Loan.HasLoan()) {
		return nil, errors.New("la petición debe incluir al menos un producto o un préstamo")
	}
	if len(req.Quotas) == 0 {
		return nil, errors.New("la venta debe tener al menos una cuota")
	}

	// Validar existencia del cliente
	client, err := s.clientRepo.GetClientByFilter(nameDB, bson.M{"_id": req.ClientID})
	if err != nil || client == nil {
		return nil, validations.ErrClientNotFound
	}

	// Validar formato de cuotas
	for _, q := range req.Quotas {
		if err := q.ValidateQuota(); err != nil {
			return nil, err
		}
	}

	// Validar datos del préstamo si existe
	if req.Loan != nil && req.Loan.HasLoan() {
		if req.Loan.PrincipalAmount.AmountCents <= 0 {
			return nil, errors.New("el monto principal del préstamo debe ser mayor a cero")
		}
		if req.Loan.InterestRate < 0 {
			return nil, errors.New("la tasa de interés del préstamo no puede ser negativa")
		}
	}

	// ====================================================================
	// PASO 2: EJECUCIÓN DE LA TRANSACCIÓN ATÓMICA
	// ====================================================================
	session, err := s.dbClient.StartSession()
	if err != nil {
		return nil, fmt.Errorf("error al iniciar sesión en DB: %w", err)
	}
	defer session.EndSession(ctx)

	var newSale *models.Sale

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		var productsTotalCents int64
		soldProducts := make([]models.SoldProduct, 0, len(req.Products))

		// --- 2.1: Procesa y valida cada producto en la venta ---
		for _, item := range req.Products {
			productData, err := (*s.productRepo).FindByID(sessCtx, nameDB, item.ProductID)
			if err != nil {
				return nil, fmt.Errorf("producto con ID %s no encontrado", item.ProductID.Hex())
			}

			var variant *models.Variant
			for i := range productData.Variants {
				if productData.Variants[i].SKU == item.VariantSKU {
					variant = &productData.Variants[i]
					break
				}
			}
			if variant == nil {
				return nil, fmt.Errorf("variante con SKU %s no encontrada en producto %s", item.VariantSKU, productData.Name)
			}

			if !variant.Stock.SellWithoutStock && variant.Stock.CurrentQuantity < item.Quantity {
				return nil, fmt.Errorf("stock insuficiente para %s (SKU: %s)", productData.Name, variant.SKU)
			}
			if err := (*s.productRepo).DecrementVariantStock(sessCtx, nameDB, productData.ID, variant.SKU, item.Quantity); err != nil {
				return nil, fmt.Errorf("no se pudo actualizar el stock: %w", err)
			}

			finalUnitPrice := variant.Price
			if item.UnitPrice != nil {
				finalUnitPrice = *item.UnitPrice
			}

			finalUnitCost := variant.Cost
			if item.UnitCost != nil {
				finalUnitCost = item.UnitCost
			}

			subtotal := int64(item.Quantity) * finalUnitPrice.AmountCents
			soldProduct := models.SoldProduct{
				ProductID:       productData.ID,
				VariantSKU:      variant.SKU,
				InternalCode:    productData.InternalCode,
				Name:            productData.Name,
				BrandName:       productData.Brand.GetName(),
				ProductTypeName: productData.ProductType.GetName(),
				Attributes:      variant.Attributes,
				Quantity:        item.Quantity,
				UnitOfMeasure:   productData.UnitOfMeasure,
				UnitPrice:       finalUnitPrice,
				UnitCost:        finalUnitCost,
				SubtotalCents:   subtotal,
			}
			soldProducts = append(soldProducts, soldProduct)
			productsTotalCents += subtotal
		}

		// --- 2.2: Procesa el préstamo ---
		var finalLoan *models.Loan
		if req.Loan != nil && req.Loan.HasLoan() {

			finalLoan = &models.Loan{
				PrincipalAmount: req.Loan.PrincipalAmount,
				InterestRate:    req.Loan.InterestRate,
				TotalToRepay:    req.Loan.TotalToRepay,
				Observations:    req.Loan.Observations,
			}
		}

		// --- 2.3: Consolida totales y valida contra las cuotas ---
		saleTotalAmountCents := productsTotalCents + finalLoan.GetTotalToRepay()

		var quotaTotalAmountCents int64
		for _, q := range req.Quotas {
			quotaTotalAmountCents += q.AmountCents
		}

		if saleTotalAmountCents != quotaTotalAmountCents {
			return nil, fmt.Errorf("el total de las cuotas (%d) no coincide con el total de la venta (%d)", quotaTotalAmountCents, saleTotalAmountCents)
		}

		//TODO: Validar los centavos

		// --- 2.4: Prepara y crea los documentos finales ---
		saleID := primitive.NewObjectID()
		quotaIDs, quotasToInsert := s.prepareQuotas(req.Quotas, saleID)

		saleNumber, err := s.saleRepo.GetNextSaleNumber(sessCtx, nameDB)
		if err != nil {
			return nil, err
		}

		newSale = &models.Sale{
			ID:                   saleID,
			SaleNumber:           saleNumber,
			ClientID:             req.ClientID,
			Products:             soldProducts,
			Loan:                 finalLoan,
			QuotaIDs:             quotaIDs,
			Status:               models.StatusInProgress,
			TotalAmountCents:     saleTotalAmountCents,
			CollectedAmountCents: 0,
			PendingAmountCents:   saleTotalAmountCents,
			QuotaCount:           len(req.Quotas),
			SaleDate:             req.SaleDate,
			CreatedAt:            time.Now().UTC(),
		}

		if err := s.quotaRepo.CreateMany(sessCtx, nameDB, quotasToInsert); err != nil {
			return nil, err
		}
		if err := s.saleRepo.Create(sessCtx, nameDB, newSale); err != nil {
			return nil, err
		}

		return newSale, nil
	})

	if err != nil {
		return nil, err
	}

	return newSale, nil
}

// prepareQuotas es una función helper para transformar el request de cuotas
// en los documentos listos para ser insertados en la base de datos.
func (s *SaleService) prepareQuotas(quotaReqs []models.Quota, saleID primitive.ObjectID) ([]primitive.ObjectID, []interface{}) {
	quotasToInsert := make([]interface{}, len(quotaReqs))
	quotaIDs := make([]primitive.ObjectID, len(quotaReqs))

	for i, q := range quotaReqs {
		quotaID := primitive.NewObjectID()
		quota := models.Quota{
			ID:             quotaID,
			SaleID:         saleID,
			QuotaNumber:    i + 1,
			ExpirationDate: q.ExpirationDate,
			AmountCents:    q.AmountCents,
			Coin:           q.Coin,
			Status:         models.QuotaPending,
		}
		quotasToInsert[i] = quota
		quotaIDs[i] = quotaID
	}

	return quotaIDs, quotasToInsert
}

// Nuevo método para obtener ventas paginadas
func (s *SaleService) GetPaginatedSales(ctx context.Context, nameDB string, page, limit int64, search, sortBy, sortOrder string) (*response.PaginatedResponse, error) {

	sales, totalDocs, err := s.saleRepo.FindPaginated(ctx, nameDB, page, limit, search, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalDocs) / float64(limit)))

	paginatedResponse := &response.PaginatedResponse{
		Docs:        sales,
		TotalDocs:   totalDocs,
		Limit:       limit,
		TotalPages:  totalPages,
		Page:        page,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}

	return paginatedResponse, nil
}

// GetPendingQuotasReport encapsula la llamada al repositorio para generar el reporte.
// Aquí podrías añadir lógica de negocio extra en el futuro (ej: permisos, formateo).
/*func (s *SaleService) GetPendingQuotasReport(ctx context.Context, companyNameID string, maxDate *time.Time) ([]dto.ClientPendingQuotasResponse, error) {
	// Por ahora, solo llama al método del repositorio.
	// Mantiene la arquitectura limpia y desacoplada.

	if maxDate == nil {
		return nil, validations.ErrInvalidDateToSearchReport
	}

	return s.saleRepo.GetPendingQuotasByClient(ctx, companyNameID, *maxDate)
}*/

// GetPendingQuotasReport encapsula la llamada al repositorio para generar el reporte.
// Aquí podrías añadir lógica de negocio extra en el futuro (ej: permisos, formateo).
func (s *SaleService) GetPendingQuotasReport(ctx context.Context, companyNameID string, maxDate *time.Time) ([]dto.ClientSalesWithQuotasResponse, error) {
	// Por ahora, solo llama al método del repositorio.
	// Mantiene la arquitectura limpia y desacoplada.

	if maxDate == nil {
		return nil, validations.ErrInvalidDateToSearchReport
	}

	return s.saleRepo.GetClientSalesWithPendingQuotas(ctx, companyNameID, *maxDate)
}
