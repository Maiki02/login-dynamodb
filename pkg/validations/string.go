package validations

import (
	"fmt"
	"myproject/pkg/structures"
	"regexp"
	"strings"
)

const MAX_NAME_LENGTH = 30
const MAX_EMAIL_LENGTH = 100

// Expresión regular para nombres y apellidos.
// \p{L} -> Coincide con cualquier letra Unicode (incluye á, é, í, ó, ú, etc.).
// '     -> Apóstrofe.
// \s    -> Espacio en blanco.
var nameRegex = regexp.MustCompile(`^[\p{L}'\s]+$`)

// Expresión regular para nombres de empresa.
// \p{L}   -> Cualquier letra Unicode.
// 0-9     -> Números.
// '       -> Apóstrofe.
// \s      -> Espacio en blanco.
var companyNameRegex = regexp.MustCompile(`^[\p{L}0-9'\s]+$`)

// ValidateName limpia y valida un nombre o apellido.
// Devuelve el nombre sin espacios al inicio/final o un error.
func ValidateName(name string, fieldName string) (string, error) {
	cleanName := strings.TrimSpace(name)

	if len(cleanName) == 0 {
		return "", fmt.Errorf("%s es requerido", fieldName)
	}
	if len(cleanName) > MAX_NAME_LENGTH {
		return "", fmt.Errorf("%s es demasiado largo", fieldName)
	}
	if !nameRegex.MatchString(cleanName) {
		return "", fmt.Errorf("%s contiene caracteres inválidos", fieldName)
	}
	return cleanName, nil
}

// ValidateCompanyName limpia y valida el nombre de una empresa.
// Devuelve el nombre sin espacios al inicio/final o un error.
func ValidateCompanyName(name string) (string, error) {
	cleanName := strings.TrimSpace(name)

	if len(cleanName) == 0 {
		return "", ErrRequiredFieldName
	}
	if len(cleanName) > MAX_NAME_LENGTH {
		return "", ErrCompanyNameTooLong
	}
	if !companyNameRegex.MatchString(cleanName) {
		return "", ErrInvalidCompanyName
	}
	return cleanName, nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	if len(email) == 0 || len(email) > MAX_EMAIL_LENGTH {
		return false
	}
	return emailRegex.MatchString(email)
}

var digitsRegex = regexp.MustCompile(`^[0-9]+$`)

// IsValidIdentification valida la estructura de identificación completa.
func IsValidIdentification(ident *structures.Identification) bool {
	if ident == nil || ident.Number == "" || ident.Type == "" {
		return false
	}

	// Aquí puedes agregar validaciones específicas por tipo de documento.
	switch ident.Type {
	case structures.DNI:
		// Por ejemplo, para DNI solo permitimos números y un largo específico.
		return digitsRegex.MatchString(ident.Number) && len(ident.Number) >= 7 && len(ident.Number) <= 8
	case structures.CUIL:
		// El CUIL tiene un formato específico (ej: 11 dígitos).
		return digitsRegex.MatchString(ident.Number) && len(ident.Number) == 11
	case structures.Passport:
		// El pasaporte puede ser alfanumérico.
		return len(ident.Number) > 0
	default:
		// Si el tipo no es conocido, lo consideramos inválido.
		return false
	}
}

func IsValidPhone(p structures.Phone) bool {
	if !digitsRegex.MatchString(p.CountryCode) || len(p.CountryCode) < 1 || len(p.CountryCode) > 4 {
		return false
	}
	if !digitsRegex.MatchString(p.AreaCode) || len(p.AreaCode) < 1 || len(p.AreaCode) > 5 {
		return false
	}
	if !digitsRegex.MatchString(p.Number) || len(p.Number) < 4 || len(p.Number) > 15 {
		return false
	}
	return true
}
