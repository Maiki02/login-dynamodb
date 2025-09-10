package services

import (
	"myproject/internal/models"
	"myproject/internal/repositories"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CompanyService define la lógica de negocio para las empresas.
type CompanyService interface {
	GetCompanyByID(id string) (*models.Company, error)
}

// companyService implementa la interfaz CompanyService.
type companyService struct {
	companyRepo repositories.CompanyRepository
}

// NewCompanyService crea una nueva instancia de CompanyService.
func NewCompanyService(repo repositories.CompanyRepository) CompanyService {
	return &companyService{
		companyRepo: repo,
	}
}

// GetCompanyByID obtiene una empresa por su ID.
func (s *companyService) GetCompanyByID(id string) (*models.Company, error) {
	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err // ID con formato inválido
	}

	filter := bson.M{"_id": idHex}

	// Usamos el repositorio inyectado
	return s.companyRepo.GetCompanyByFilter(filter)
}

/*
import (
	"myproject/internal/models"
	"myproject/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateCompany(company *models.Company) error {

	return repositories.CreateCompany(company)
}

func GetCompanyByUser() (*models.Company, error) {
	return repositories.GetCompanyByFilter(map[string]interface{}{})
}

func GetCompanyByID(id string) (*models.Company, error) {
	idHex, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	return repositories.GetCompanyByFilter(map[string]interface{}{"_id": idHex})
}*/

/*
func SetProductsInfo(nameDb string, companyID string, products *[]models.Product) error {
	//Obtenemos las secciones
	sections, err := GetAllSections(nameDb)
	if err != nil {
		return err
	}

	//Pasamos a Mapa
	sectionsMap := make(map[int32]models.Section)
	for _, section := range sections {
		sectionsMap[section.Code] = section
	}

	//Obtenemos los productos
	var productsInfo []structures.ProductInfo
	for _, product := range *products {
		section, ok := sectionsMap[product.Section]
		if !ok {
			return validations.ErrSectionNeedsExist
		}
		productsInfo = append(productsInfo, structures.ProductInfo{
			Text: product.GetText(section),
			ID:   product.ID,
		})
	}

	updates := map[string]interface{}{
		"products_info": productsInfo,
	}

	return repositories.UpdateCompany(nameDb, companyID, updates)
}
*/
