package services

import (
	"context"
	"math"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/response"
	"myproject/pkg/slug"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const ACTIVE_STATUS = "activo"

// GenericServiceInterface define la lógica de negocio para nuestras entidades.
type GenericServiceInterface interface {
	Create(ctx context.Context, nameDB, name string) (models.Entity, error)
	Update(ctx context.Context, nameDB, oldSlug, newName string) (interface{}, error)
	GetBySlug(ctx context.Context, nameDB, slug string) (interface{}, error)
	GetPaginated(ctx context.Context, nameDB, search, sortBy, sortOrder string, page, limit int64) (*response.PaginatedResponse, error)
}

type genericService struct {
	repoFactory func(string) repositories.GenericRepositoryInterface // Factory para obtener el repo
	newEntity   func() models.Entity                                 // Factory para crear una nueva entidad
	newSlice    func() interface{}                                   // Factory para crear un nuevo slice de entidades
}

// NewGenericService crea un servicio genérico.
func NewGenericService(repoFactory func(string) repositories.GenericRepositoryInterface, newEntity func() models.Entity, newSlice func() interface{}) GenericServiceInterface {
	return &genericService{
		repoFactory: repoFactory,
		newEntity:   newEntity,
		newSlice:    newSlice,
	}
}

func (s *genericService) Create(ctx context.Context, nameDB, name string) (models.Entity, error) {
	repo := s.repoFactory(nameDB)

	normalizedName, err := slug.NormalizeName(name)
	if err != nil {
		return nil, err
	}

	generatedSlug, err := slug.GenerateSlug(normalizedName)
	if err != nil {
		return nil, err
	}

	exists, err := repo.SlugExists(ctx, nameDB, generatedSlug)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, repositories.ErrDuplicateSlug
	}

	entity := s.newEntity()
	entity.SetID(primitive.NewObjectID())
	entity.SetName(normalizedName)
	entity.SetSlug(generatedSlug)
	entity.SetStatus(ACTIVE_STATUS)
	entity.SetCreatedAt(time.Now().UTC())
	entity.SetUpdatedAt(time.Now().UTC())

	_, err = repo.Create(ctx, nameDB, entity)
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (s *genericService) Update(ctx context.Context, nameDB, oldSlug, newName string) (interface{}, error) {
	repo := s.repoFactory(nameDB)

	normalizedName, err := slug.NormalizeName(newName)
	if err != nil {
		return nil, err
	}

	newSlug, err := slug.GenerateSlug(normalizedName)
	if err != nil {
		return nil, err
	}

	updates := bson.M{
		"name":       normalizedName,
		"updated_at": time.Now().UTC(),
	}

	if oldSlug != newSlug {
		exists, err := repo.SlugExists(ctx, nameDB, newSlug)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, repositories.ErrDuplicateSlug
		}
		updates["slug"] = newSlug
	}

	return repo.Update(ctx, nameDB, oldSlug, updates)
}

func (s *genericService) GetBySlug(ctx context.Context, nameDB, slug string) (interface{}, error) {
	repo := s.repoFactory(nameDB)
	result := s.newEntity()
	err := repo.GetBySlug(ctx, nameDB, slug, result)
	return result, err
}

func (s *genericService) GetPaginated(ctx context.Context, nameDB, search, sortBy, sortOrder string, page, limit int64) (*response.PaginatedResponse, error) {
	repo := s.repoFactory(nameDB)
	// Creamos un slice del tipo correcto (ej. *[]models.Brand) para pasar al repositorio
	results := s.newSlice()

	totalDocs, err := repo.FindPaginated(ctx, nameDB, search, sortBy, sortOrder, page, limit, results)
	if err != nil {
		return nil, err
	}

	// Calcular metadatos de paginación
	totalPages := int64(math.Ceil(float64(totalDocs) / float64(limit)))

	// Construir la respuesta paginada
	return &response.PaginatedResponse{
		Docs:        results,
		TotalDocs:   totalDocs,
		Limit:       limit,
		TotalPages:  totalPages,
		Page:        page,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}, nil
}
