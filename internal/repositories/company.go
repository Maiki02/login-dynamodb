package repositories

import (
	"context"
	"myproject/internal/db" // <--- IMPORTACIÓN AÑADIDA
	"myproject/internal/models"
	"myproject/pkg/validations"

	"go.mongodb.org/mongo-driver/bson" // <--- IMPORTACIÓN AÑADIDA
)

// CompanyRepository define los métodos para interactuar con el almacenamiento de empresas.
type CompanyRepository interface {
	CreateCompany(company *models.Company) error
	GetCompanyByFilter(filter bson.M) (*models.Company, error) // Cambiado a bson.M
	UpdateCompany(id string, updates map[string]interface{}) error
	DeleteCompany(id string) error
}

// companyRepository implementa la interfaz CompanyRepository.
type companyRepository struct{}

// NewCompanyRepository crea una nueva instancia de companyRepository.
func NewCompanyRepository() CompanyRepository {
	return &companyRepository{}
}

const DB_COMPANY_NAME = "Companies_DB"
const COLLECTION_COMPANY = "company"

func (r *companyRepository) CreateCompany(company *models.Company) error {
	_, err := db.InsertDocument(DB_COMPANY_NAME, COLLECTION_COMPANY, company)
	return err
}

func (r *companyRepository) GetCompanyByFilter(filter bson.M) (*models.Company, error) {
	cursor, err := db.FindDocuments(DB_COMPANY_NAME, COLLECTION_COMPANY, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var company models.Company
	if cursor.Next(context.Background()) {
		if err = cursor.Decode(&company); err != nil {
			return nil, err
		}
		return &company, nil
	}

	return nil, validations.ErrCompanyNotFound
}

func (r *companyRepository) UpdateCompany(id string, updates map[string]interface{}) error {
	return db.UpdateDocumentByID(DB_COMPANY_NAME, COLLECTION_COMPANY, id, updates)
}

// DeleteCompany implementa la lógica de borrado llamando a la función del paquete db.
func (r *companyRepository) DeleteCompany(id string) error {
	return db.DeleteDocumentByID(DB_COMPANY_NAME, COLLECTION_COMPANY, id)
}
