package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	"myproject/internal/dto"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/consts"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/slug"
	"myproject/pkg/validations"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductService interface {
	CreateProduct(ctx context.Context, nameDB string, req request.CreateProductRequest) (*models.Product, error)

	GetProductByID(ctx context.Context, nameDB, productID string) (*dto.ProductResponse, error)
	UpdateProduct(ctx context.Context, nameDB, productID string, req request.UpdateProductRequest) (*mongo.UpdateResult, error)
	UpdateVariant(ctx context.Context, nameDB, productID, sku string, req request.UpdateVariantRequest) (*mongo.UpdateResult, error)
	GetProducts(ctx context.Context, nameDB string, page, limit int64, brandID, typeID, search, sortBy, sortOrder string) (*response.PaginatedResponse, error)
}

type productService struct {
	productRepo     repositories.ProductRepository
	brandService    GenericServiceInterface
	prodTypeService GenericServiceInterface
	dbClient        *mongo.Client
}

func NewProductService(repo repositories.ProductRepository, brandSvc GenericServiceInterface, prodTypeSvc GenericServiceInterface, client *mongo.Client) ProductService {
	return &productService{
		productRepo:     repo,
		brandService:    brandSvc,
		prodTypeService: prodTypeSvc,
		dbClient:        client,
	}
}

// validateCreateRequest es una función privada que encapsula la lógica de validación.
func (s *productService) validateCreateRequest(req *request.CreateProductRequest) error {
	// 1. Validar el nombre del producto
	if req.Name == "" {
		return validations.ErrProductNameRequired
	}
	if len(req.Name) > 100 {
		return validations.ErrProductNameTooLong
	}

	// 2. Validar que haya al menos una variante
	if len(req.Variants) == 0 {
		return validations.ErrProductVariantRequired
	}

	// 3. Validar cada variante
	for _, v := range req.Variants {
		if v.SKU == "" {
			return validations.ErrVariantSKURequired
		}
		if len(v.SKU) > 50 {
			return validations.ErrVariantSKUTooLong
		}
		if len(v.Attributes) == 0 {
			return validations.ErrVariantAttributesRequired
		}

		// Validar Precio
		if v.Price.AmountCents < 0 {
			return validations.ErrVariantPriceInvalid
		}
		if v.Price.Coin == "" {
			return validations.ErrVariantCoinRequired
		}

		// Validar Costo (si existe)
		if v.Cost != nil && v.Cost.AmountCents < 0 {
			return validations.ErrVariantCostInvalid
		}

		// Validar Stock
		if v.Stock.CurrentQuantity < 0 {
			return validations.ErrVariantStockInvalid
		}
	}

	return nil
}

func (s *productService) CreateProduct(ctx context.Context, nameDB string, req request.CreateProductRequest) (*models.Product, error) {
	// --- Paso 1: Validaciones Robustas ---
	if err := s.validateCreateRequest(&req); err != nil {
		return nil, err
	}

	// --- Paso 2: Lógica de Creación Atómica con Transacción ---
	session, err := s.dbClient.StartSession()
	if err != nil {
		return nil, err
	}
	defer session.EndSession(ctx)

	var newProduct *models.Product

	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {

		var brandID, productTypeID primitive.ObjectID

		// 2.1. Manejar Marca (Find-or-Create)
		if req.BrandName != nil && *req.BrandName != "" {
			normalizedBrandName, err := slug.NormalizeName(*req.BrandName)
			if err != nil {
				return nil, err
			}
			brandSlug, _ := slug.GenerateSlug(normalizedBrandName)

			existingBrand, err := s.brandService.GetBySlug(sessCtx, nameDB, brandSlug)
			if err == nil {
				brandID = existingBrand.(*models.Brand).ID
			} else if err == repositories.ErrNotFound {
				entity, createErr := s.brandService.Create(sessCtx, nameDB, *req.BrandName)
				if createErr != nil {
					return nil, createErr
				}
				brandID = entity.(*models.Brand).ID
			} else {
				return nil, err
			}
		}

		// 2.2. Manejar Tipo de Producto (Find-or-Create)
		if req.ProductTypeName != nil && *req.ProductTypeName != "" {
			normalizedTypeName, err := slug.NormalizeName(*req.ProductTypeName)
			if err != nil {
				return nil, err
			}
			prodTypeSlug, _ := slug.GenerateSlug(normalizedTypeName)

			existingType, err := s.prodTypeService.GetBySlug(sessCtx, nameDB, prodTypeSlug)
			if err == nil {
				productTypeID = existingType.(*models.ProductType).ID
			} else if err == repositories.ErrNotFound {
				entity, createErr := s.prodTypeService.Create(sessCtx, nameDB, *req.ProductTypeName)
				if createErr != nil {
					return nil, createErr
				}
				productTypeID = entity.(*models.ProductType).ID
			} else {
				return nil, err
			}
		}

		// 2.3. Preparar el Producto
		normalizedName, _ := slug.NormalizeName(req.Name)
		productSlug, _ := slug.GenerateSlug(normalizedName)
		internalCode, err := s.productRepo.GetNextProductNumber(sessCtx, nameDB)
		if err != nil {
			return nil, err
		}

		productToCreate := &models.Product{
			ID:            primitive.NewObjectID(),
			InternalCode:  internalCode,
			Name:          normalizedName,
			Slug:          productSlug,
			BrandID:       brandID,
			ProductTypeID: productTypeID,
			Status:        consts.ACTIVE_STATUS,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
			Variants:      []models.Variant{},
		}

		for _, vReq := range req.Variants {
			variant := models.Variant{
				SKU:        vReq.SKU,
				Attributes: vReq.Attributes,
				Price:      vReq.Price,
				Cost:       vReq.Cost,
				Stock: models.Stock{
					CurrentQuantity:  vReq.Stock.CurrentQuantity,
					SellWithoutStock: vReq.Stock.SellWithoutStock,
				},
			}
			productToCreate.Variants = append(productToCreate.Variants, variant)
		}

		// 2.4. Insertar el Producto
		if _, err := s.productRepo.Create(sessCtx, nameDB, productToCreate); err != nil {
			return nil, err
		}

		newProduct = productToCreate
		return newProduct, nil
	})

	if err != nil {
		return nil, err
	}

	return newProduct, nil
}

func (s *productService) GetProductByID(ctx context.Context, nameDB, productID string) (*dto.ProductResponse, error) {
	id, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("ID de producto inválido")
	}
	return s.productRepo.FindByID(ctx, nameDB, id)
}

func (s *productService) UpdateProduct(ctx context.Context, nameDB, productID string, req request.UpdateProductRequest) (*mongo.UpdateResult, error) {
	id, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("ID de producto inválido")
	}

	updates := bson.M{}

	if req.Name != nil {
		normalizedName, err := slug.NormalizeName(*req.Name)
		if err != nil {
			return nil, err
		}
		productSlug, _ := slug.GenerateSlug(normalizedName)
		updates["name"] = normalizedName
		updates["slug"] = productSlug
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Status != nil {
		if *req.Status != consts.ACTIVE_STATUS && *req.Status != consts.INACTIVE_STATUS {
			return nil, errors.New("el estado del producto es inválido")
		}
		updates["status"] = *req.Status
	}

	// Lógica Find-or-Create para Brand usando el helper
	if req.BrandName != nil {
		brandID, err := s.findOrCreateEntity(ctx, nameDB, *req.BrandName, s.brandService)
		if err != nil {
			return nil, fmt.Errorf("error al procesar la marca: %w", err)
		}
		updates["brand_id"] = brandID
	}

	// Lógica Find-or-Create para ProductType usando el helper
	if req.ProductTypeName != nil {
		productTypeID, err := s.findOrCreateEntity(ctx, nameDB, *req.ProductTypeName, s.prodTypeService)
		if err != nil {
			return nil, fmt.Errorf("error al procesar el tipo de producto: %w", err)
		}
		updates["product_type_id"] = productTypeID
	}

	if len(updates) == 0 {
		return nil, validations.ErrProductNoFieldsToUpdate
	}

	updates["updated_at"] = time.Now().UTC()
	return s.productRepo.UpdateByID(ctx, nameDB, id, updates)
}

func (s *productService) UpdateVariant(ctx context.Context, nameDB, productID, sku string, req request.UpdateVariantRequest) (*mongo.UpdateResult, error) {
	id, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("ID de producto inválido")
	}

	updates := bson.M{}

	if req.Price != nil {
		if req.Price.AmountCents < 0 || req.Price.Coin == "" {
			return nil, validations.ErrVariantPriceInvalid
		}
		updates["price"] = *req.Price
	}
	if req.Cost != nil {
		if req.Cost.AmountCents < 0 {
			return nil, validations.ErrVariantCostInvalid
		}
		updates["cost"] = *req.Cost
	}
	if req.Stock != nil {
		if req.Stock.CurrentQuantity < 0 {
			return nil, validations.ErrVariantStockInvalid
		}
		updates["stock.current_quantity"] = req.Stock.CurrentQuantity
	}
	if req.SellWithoutStock != nil {
		updates["stock.sell_without_stock"] = *req.SellWithoutStock
	}
	if req.Attributes != nil {
		if len(req.Attributes) == 0 {
			return nil, validations.ErrVariantAttributesRequired
		}
		updates["attributes"] = req.Attributes
	}

	if len(updates) == 0 {
		return nil, validations.ErrProductNoFieldsToUpdate
	}

	// Añadir el timestamp de actualización al producto principal en una gorutina para no bloquear
	go s.productRepo.UpdateByID(context.Background(), nameDB, id, bson.M{"updated_at": time.Now().UTC()})

	return s.productRepo.UpdateVariantBySKU(ctx, nameDB, id, sku, updates)
}

// findOrCreateEntity es un helper privado para manejar la lógica de buscar o crear una entidad genérica.
func (s *productService) findOrCreateEntity(ctx context.Context, nameDB, name string, service GenericServiceInterface) (primitive.ObjectID, error) {
	if name == "" {
		return primitive.NilObjectID, nil
	}

	normalizedName, err := slug.NormalizeName(name)
	if err != nil {
		return primitive.NilObjectID, err
	}
	entitySlug, _ := slug.GenerateSlug(normalizedName)

	// Intenta obtener la entidad por su slug
	existingEntity, err := service.GetBySlug(ctx, nameDB, entitySlug)
	if err == nil {
		// La encontró, extraemos el ID
		if brand, ok := existingEntity.(*models.Brand); ok {
			return brand.ID, nil
		}
		if productType, ok := existingEntity.(*models.ProductType); ok {
			return productType.ID, nil
		}
		return primitive.NilObjectID, errors.New("tipo de entidad desconocido")
	}

	// Si no la encuentra, la crea
	if err == repositories.ErrNotFound {
		newEntity, createErr := service.Create(ctx, nameDB, name)
		if createErr != nil {
			return primitive.NilObjectID, createErr
		}
		// Extraemos el ID de la entidad recién creada
		if brand, ok := newEntity.(*models.Brand); ok {
			return brand.ID, nil
		}
		if productType, ok := newEntity.(*models.ProductType); ok {
			return productType.ID, nil
		}
		return primitive.NilObjectID, errors.New("tipo de entidad desconocido tras creación")
	}

	// Si hubo otro tipo de error
	return primitive.NilObjectID, err
}

func (s *productService) GetProducts(ctx context.Context, nameDB string, page, limit int64, brandSlug, typeSlug, search, sortBy, sortOrder string) (*response.PaginatedResponse, error) {
	filters := bson.M{"status": consts.ACTIVE_STATUS}

	if search != "" {
		filters["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	// Filtrar por brand_id si se provee el slug
	if brandSlug != "" {
		brand, err := s.brandService.GetBySlug(ctx, nameDB, brandSlug)
		if err != nil {
			// Si la marca no existe, no habrá productos que coincidan.
			if err == repositories.ErrNotFound {
				return &response.PaginatedResponse{Docs: []interface{}{}, TotalDocs: 0}, nil
			}
			return nil, fmt.Errorf("error al buscar la marca: %w", err)
		}
		filters["brand_id"] = brand.(*models.Brand).ID
	}

	// Filtrar por product_type_id si se provee el slug
	if typeSlug != "" {
		prodType, err := s.prodTypeService.GetBySlug(ctx, nameDB, typeSlug)
		if err != nil {
			if err == repositories.ErrNotFound {
				return &response.PaginatedResponse{Docs: []interface{}{}, TotalDocs: 0}, nil
			}
			return nil, fmt.Errorf("error al buscar el tipo de producto: %w", err)
		}
		filters["product_type_id"] = prodType.(*models.ProductType).ID
	}

	fmt.Println(search, filters, typeSlug)

	products, totalDocs, err := s.productRepo.FindPaginated(ctx, nameDB, page, limit, filters, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	totalPages := int64(math.Ceil(float64(totalDocs) / float64(limit)))
	return &response.PaginatedResponse{
		Docs:        products,
		TotalDocs:   totalDocs,
		Limit:       limit,
		TotalPages:  totalPages,
		Page:        page,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}, nil
}
