package services

import (
	"context"
	"math"
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/request"
	"myproject/pkg/response"
	"myproject/pkg/structures"
	"myproject/pkg/validations"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ClientService struct {
	clientRepo  *repositories.ClientRepository // <-- El especialista en clientes
	clientModel *models.Client
}

// NewClientService crea una nueva instancia del servicio de ventas.
func NewClientService(
	clientRepo *repositories.ClientRepository,
) *ClientService {
	return &ClientService{
		clientRepo: clientRepo,
	}
}

func (s *ClientService) CreateClient(nameDB string, clientReq *request.CreateClientRequest) (*models.Client, error) {
	client, err := models.NewClient(clientReq.Name, clientReq.LastName, clientReq.Email, clientReq.Identification, clientReq.Phone, clientReq.Address)
	if err != nil {
		return nil, err
	}

	return client, s.clientRepo.CreateClient(nameDB, client)
}

// Nuevo método para obtener clientes paginados
func (s *ClientService) GetPaginatedClients(ctx context.Context, nameDB string, page, limit int64, search, sortBy, sortOrder string) (*response.PaginatedResponse, error) {

	// Llamamos al repositorio para obtener los clientes y el conteo total
	clients, totalDocs, err := s.clientRepo.FindPaginated(ctx, nameDB, page, limit, search, sortBy, sortOrder)
	if err != nil {
		return nil, err
	}

	// Calcular metadatos de paginación
	totalPages := int64(math.Ceil(float64(totalDocs) / float64(limit)))

	// Construir la respuesta paginada
	paginatedResponse := &response.PaginatedResponse{
		Docs:        clients,
		TotalDocs:   totalDocs,
		Limit:       limit,
		TotalPages:  totalPages,
		Page:        page,
		HasNextPage: page < totalPages,
		HasPrevPage: page > 1,
	}

	return paginatedResponse, nil
}

func (s *ClientService) UpdateClient(ctx context.Context, nameDB, id string, updates *request.UpdateClientRequest) (*models.Client, error) {

	//TODO: toda esta lógica podría ir al request o modelo de cliente para que valide y limpiar el service
	updateMap := map[string]interface{}{}

	if updates.Name != nil {
		nameToSave, err := validations.ValidateName(*updates.Name, "Nombre")
		if err != nil {
			return nil, err
		}

		updateMap["name"] = nameToSave
	}

	if updates.LastName != nil {
		lastNameToSave, err := validations.ValidateName(*updates.LastName, "Apellido")
		if err != nil {
			return nil, err
		}

		updateMap["last_name"] = lastNameToSave
	}

	if updates.Email != nil {
		isValidEmail := validations.IsValidEmail(*updates.Email)
		if !isValidEmail {
			return nil, validations.ErrInvalidClientEmail
		}

		updateMap["email"] = strings.ToLower(*updates.Email)
	}

	if updates.Identification != nil {
		if !validations.IsValidIdentification(updates.Identification) {
			return nil, validations.ErrInvalidClientIdentification
		}
		updateMap["identification"] = *updates.Identification
	}

	if updates.Phone != nil {
		isValidPhone := validations.IsValidPhone(*updates.Phone)
		if !isValidPhone {
			return nil, validations.ErrInvalidClientPhone
		}

		updateMap["phone"] = *updates.Phone
	}

	if updates.Address != nil {
		isValidAddress := validations.IsValidAddress(*updates.Address)
		if !isValidAddress {
			return nil, validations.ErrInvalidClientAddress
		}

		updateMap["address"] = *updates.Address
	}

	if len(updateMap) == 0 {
		return nil, validations.ErrInvalidRequest
	}

	updateMap["updated_at"] = time.Now()

	return s.clientRepo.UpdateandFindClient(nameDB, id, updateMap)
}

func (s *ClientService) GetClientByPhone(nameDB string, phone structures.Phone) (*models.Client, error) {
	filter := map[string]interface{}{
		"phone": phone,
	}

	clients, err := s.clientRepo.GetClientsByFilter(nameDB, filter)

	if err != nil || len(*clients) == 0 {
		return nil, validations.ErrDocumentNotFound
	}

	return &(*clients)[0], nil
}

func (s *ClientService) GetClientByID(nameDB, id string) (*models.Client, error) {
	hexID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	return s.clientRepo.GetClientByFilter(nameDB, map[string]interface{}{"_id": hexID})
}

func (s *ClientService) DeleteClient(nameDB, id string) error {
	updates := map[string]interface{}{
		"deleted_at": time.Now(),
		"status":     0,
	}
	return s.clientRepo.UpdateClient(nameDB, id, updates)
}
