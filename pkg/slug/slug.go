package slug

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	ErrNameTooShort = errors.New("el nombre es demasiado corto (mínimo 3 caracteres)")
	ErrNameTooLong  = errors.New("el nombre es demasiado largo (máximo 50 caracteres)")
)

// NormalizeName limpia y formatea un nombre para mostrar.
// Ej: "  remera de ALGODÓN  " -> "Remera de Algodón"
func NormalizeName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if len(trimmed) < 3 {
		return "", ErrNameTooShort
	}
	if len(trimmed) > 50 {
		return "", ErrNameTooLong
	}

	// Convierte a minúsculas y luego a formato de título
	// (la librería estándar `strings.Title` está obsoleta)
	words := strings.Fields(strings.ToLower(trimmed))
	for i, word := range words {
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}

	return strings.Join(words, " "), nil
}

// GenerateSlug crea un slug limpio y único a partir de un nombre, uniendo las palabras.
// Ej: "Coca Cola" o "Coca-Cola" -> "cocacola"
func GenerateSlug(name string) (string, error) {
	// 1. Convertir a minúsculas
	lower := strings.ToLower(name)

	// 2. Eliminar tildes y diacríticos (acentos) para normalizar.
	// Ej: "algodón" -> "algodon"
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	normalized, _, err := transform.String(t, lower)
	if err != nil {
		return "", err
	}

	// 3. Definir la expresión regular para encontrar cualquier caracter que NO sea
	// una letra minúscula (a-z) o un número (0-9).
	// Esto incluye espacios, guiones, y cualquier otro símbolo.
	nonAlphanumericRegex := regexp.MustCompile(`[^a-z0-9]`)

	// 4. Reemplazar todos esos caracteres con una cadena vacía, efectivamente eliminándolos.
	slug := nonAlphanumericRegex.ReplaceAllString(normalized, "")

	return slug, nil
}
