package security

import (
	"myproject/pkg/validations"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func IsValidPassword(password string) (bool, error) {
	if len(password) < 7 || len(password) > 30 {
		return false, validations.ErrPasswordChars
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#~$%^&*()_+={}\[\]:;"'<>,.?\/\\|-]`).MatchString(password)

	if !hasUpper || !hasLower || !hasNumber || !hasSpecial {
		return false, validations.ErrPasswordComplexity
	}

	return true, nil
}

func HashPassword(password string) (string, error) {
	// Hashea la contraseña usando bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	// Retorna el hash como string
	return string(hash), nil
}

func ValidateAndHashPassword(password string) (*string, error) {
	_, err := IsValidPassword(password)
	if err != nil {
		return nil, err
	}

	hashPassword, err := HashPassword(password)
	if err != nil {
		return nil, err
	}

	return &hashPassword, nil
}
