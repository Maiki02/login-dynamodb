package validations

import (
	"errors"
)

var (
	// DB
	ErrDocumentNotFound      = errors.New("document not found")
	ErrDocumentAlreadyExists = errors.New("document already exists")
	/*ErrQueryNotProcessed     = errors.New("query not processed")
	ErrQueryNotProvided      = errors.New("query not provided")
	ErrConnectionFailed      = errors.New("connection to MongoDB failed")
	ErrCursorFailed          = errors.New("decode cursor failed")
	ErrInsertDocumentFailed  = errors.New("insert document failed")
	ErrUpdateDocumentFailed  = errors.New("update document failed")*/
	ErrDeleteDocumentFailed = errors.New("delete document failed")

	// API
	ErrInvalidCode = errors.New("invalid code")
	/*ErrUnauthorized  = errors.New("unauthorized")
	ErrTokenNotFound = errors.New("token not found")
	ErrInvalidToken  = errors.New("invalid token")*/

	// Validations
	ErrInvalidRequest = errors.New("Invalid payload request")
	ErrFieldInvalid   = errors.New("Invalid field in JSON")
	ErrValueInvalid   = errors.New("Invalid value in JSON")
	//ErrMustName           = errors.New("%s is required")
	ErrInvalidQueryParams = errors.New("Invalid query params")
	ErrParsedDate         = errors.New("Invalid date")

	//Section
	ErrSectionNeedsExist = errors.New("Section must exist")

	//Client
	ErrCreatingClient = errors.New("Error creating client")
	//ErrClientName            = errors.New("Client name is required")
	ErrInvalidClientName           = errors.New("Invalid name")
	ErrInvalidClientLastName       = errors.New("Invalid last name")
	ErrInvalidClientEmail          = errors.New("Invalid email")
	ErrInvalidClientPhone          = errors.New("Invalid phone")
	ErrInvalidClientIdentification = errors.New("Invalid identification")
	ErrInvalidClientAddress        = errors.New("Invalid address")

	//Sale
	ErrClientNotFound     = errors.New("Cliente no encontrado")
	ErrSellNotFound       = errors.New("Venta no encontrada")
	ErrGenerateCode       = errors.New("Error al generar el codigo de venta")
	ErrInvalidSaleAddress = errors.New("El domicilio de la venta es inválido")

	// Quota
	ErrInvalidExpirationDate = errors.New("La fecha de expiración no puede ser anterior a la fecha actual")
	ErrNegativeAmount        = errors.New("El monto de la cuota no puede ser negativo")
	ErrCoinTooLong           = errors.New("La moneda de la cuota no puede tener más de 10 caracteres")

	// Reports
	ErrInvalidDateToSearchReport = errors.New("La fecha que se envió es inválida")
	ErrInvalidFormatDate         = errors.New("El formato de la fecha es inválido")

	//Company
	//ErrOwnerNeedsExist = errors.New("Owner need exist")
	ErrCompanyNotFound    = errors.New("Company not found")
	ErrRequiredFieldName  = errors.New("El campo de nombre es requerido")
	ErrCompanyNameTooLong = errors.New("El nombre de la empresa no puede tener más de 50 caracteres")
	ErrInvalidCompanyName = errors.New("Nombre de la empresa inválido")

	//Auth
	ErrInvalidCredentials = errors.New("Invalid credentials")
	ErrInvalidToken       = errors.New("Invalid token")
	ErrUserInactive       = errors.New("User is inactive")
	ErrInvalidUserID      = errors.New("Invalid user id")

	//Register
	ErrRequiredName       = errors.New("Name is required")
	ErrNameIsTooLong      = errors.New("Name is too long")
	ErrRequiredLastName   = errors.New("Last name is required")
	ErrLastNameIsTooLong  = errors.New("Last name is too long")
	ErrInvalidName        = errors.New("Invalid name")
	ErrInvalidLastName    = errors.New("Invalid last name")
	ErrInvalidEmail       = errors.New("Invalid email")
	ErrPasswordComplexity = errors.New("La contraseña debe contener una letra mayúscula, minúscula, número y caracter especial")
	ErrPasswordChars      = errors.New("La contraseña debe tener entre 7 y 30 caracteres")

	//Wuzapi
	/*ErrInvalidPhone   = errors.New("Invalid phone")
	ErrNotUrlWuzapi   = errors.New("Not URL in .env")
	ErrNotTokenWuzapi = errors.New("Not token in .env")

	//Open AI
	ErrNotApiKeyOpenAI  = errors.New("Not API key in .env")
	ErrNoResponseOpenAI = errors.New("No response from OpenAI")
	ErrNotMessageOpenAI = errors.New("No hay mensajes en el hilo de OpenAI")
	ErrThreadNotFound   = errors.New("Thread not found")
	ErrRunNotFound      = errors.New("Run not found")*/

	// Product
	ErrProductNameRequired    = errors.New("el nombre del producto es obligatorio")
	ErrProductNameTooLong     = errors.New("el nombre del producto es demasiado largo (máximo 100 caracteres)")
	ErrProductVariantRequired = errors.New("el producto debe tener al menos una variante")

	ErrProductNoFieldsToUpdate = errors.New("no hay campos para actualizar el producto")

	// Variant
	ErrVariantSKURequired        = errors.New("el SKU de la variante es obligatorio")
	ErrVariantSKUTooLong         = errors.New("el SKU es demasiado largo (máximo 50 caracteres)")
	ErrVariantAttributesRequired = errors.New("los atributos de la variante son obligatorios")
	ErrVariantPriceInvalid       = errors.New("el precio de la variante debe ser un monto válido y no negativo")
	ErrVariantCostInvalid        = errors.New("el costo de la variante, si se especifica, debe ser un monto válido y no negativo")
	ErrVariantStockInvalid       = errors.New("la cantidad en stock no puede ser negativa")
	ErrVariantCoinRequired       = errors.New("la moneda para el precio de la variante es obligatoria")
)
