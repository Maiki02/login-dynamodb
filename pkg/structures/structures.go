package structures

// DocumentType define los tipos de documentos de identificación.
type DocumentType string

const (
	DNI            DocumentType = "DNI"
	CUIL           DocumentType = "CUIL"
	Passport       DocumentType = "PTE"
	DriversLicense DocumentType = "LNC"
)

// Identification representa un documento de identidad.
type Identification struct {
	Type   DocumentType `json:"type" bson:"type"`
	Number string       `json:"number" bson:"number"`
}

// Address representa la estructura de una dirección física.
type Address struct {
	Street    string `json:"street" bson:"street" validate:"required"`
	Number    string `json:"number" bson:"number" validate:"required"`
	Floor     string `json:"floor,omitempty" bson:"floor,omitempty"`
	Apartment string `json:"apartment,omitempty" bson:"apartment,omitempty"`
	City      string `json:"city" bson:"city" validate:"required"`
	State     string `json:"state" bson:"state" validate:"required"`
	ZipCode   string `json:"zip_code" bson:"zip_code" validate:"required"`
	Country   string `json:"country" bson:"country" validate:"required"`
}

// Phone representa la estructura de un número de teléfono.
// Es reutilizable por otras entidades como Client o User.
type Phone struct {
	CountryCode string `json:"country_code" bson:"country_code"`
	AreaCode    string `json:"area_code" bson:"area_code"`
	Number      string `json:"number" bson:"number"`
}

// Money representa una cantidad monetaria, guardando el valor como un entero.
type Money struct {
	AmountCents int64  `json:"amount_cents" bson:"amount_cents"`
	Coin        string `json:"coin" bson:"coin"`
}

func (p Phone) IsEmpty() bool {
	// Ajusta esta lógica según los campos que consideres esenciales.
	return p.CountryCode == "" && p.AreaCode == "" && p.Number == ""
}
