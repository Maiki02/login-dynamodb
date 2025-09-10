package validations

import (
	"myproject/pkg/structures"
	"regexp"
	"strings"
)

const MaxAddressFieldLength = 100

var zipRegex = regexp.MustCompile(`^[a-zA-Z0-9\-]{4,10}$`)

func IsValidAddress(a structures.Address) bool {
	if !isValidRequiredField(a.Street) {
		return false
	}
	if !isValidRequiredField(a.Number) {
		return false
	}
	if !isValidRequiredField(a.City) {
		return false
	}
	if !isValidRequiredField(a.State) {
		return false
	}
	if !isValidRequiredField(a.Country) {
		return false
	}
	if len(a.Floor) > MaxAddressFieldLength {
		return false
	}
	if len(a.Apartment) > MaxAddressFieldLength {
		return false
	}
	if !zipRegex.MatchString(a.ZipCode) {
		return false
	}
	return true
}

func isValidRequiredField(s string) bool {
	s = strings.TrimSpace(s)
	return s != "" && len(s) <= MaxAddressFieldLength
}
