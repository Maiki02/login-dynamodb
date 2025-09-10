package validations

import (
	"fmt"
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
