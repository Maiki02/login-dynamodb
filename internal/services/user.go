package services

/*
import (
	"myproject/internal/models"
	"myproject/internal/repositories"
	"myproject/pkg/validations"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateUser(user *models.User) error {
	userDB, _ := GetUserByFilter(nil, nil, &user.ContactInfo.Email.Address)

	if userDB != nil {
		return validations.ErrDocumentAlreadyExists
	}

	return repositories.CreateUser(user)
}

func GetUserByFilter(id *primitive.ObjectID, name, email *string) (*models.User, error) {
	filter := map[string]interface{}{}

	if id != nil {
		filter["_id"] = *id
	}

	if name != nil {
		filter["personal_info.name"] = *name
	}

	if email != nil {
		filter["contact_info.email.address"] = *email
	}

	user, err := repositories.GetUserByFilter(filter)

	if err != nil {
		return nil, validations.ErrDocumentNotFound
	}

	return user, nil
}

func UpdateUser(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return repositories.UpdateUser(id, updates)
}
*/
/*
func GetClientByPhone(phone string) (*models.Client, error) {
	filter := map[string]interface{}{
		"phone": phone,
	}

	clients, err := repositories.GetClientsByFilter(filter)

	if err != nil || len(*clients) == 0 {
		return nil, validations.ErrDocumentNotFound
	}

	return &(*clients)[0], nil
}

func DeleteClient(id string) error {
	updates := map[string]interface{}{
		"deleted_at": time.Now(),
		"status":     0,
	}
	return repositories.UpdateClient(id, updates)
}*/
