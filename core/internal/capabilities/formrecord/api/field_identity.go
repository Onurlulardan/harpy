package recordkernel

import (
	"github.com/google/uuid"
)

func parseStableID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || !stableUUIDv7(value) || value.String() != raw {
		return uuid.Nil, ErrCatalogInvalid
	}
	return value, nil
}

func validPublicName(value string) bool {
	if len(value) < 1 || len(value) > 63 || value[0] < 'a' || value[0] > 'z' ||
		value[len(value)-1] == '_' || reservedPublicName(value) {
		return false
	}
	previousUnderscore := false
	for _, character := range []byte(value) {
		letter := character >= 'a' && character <= 'z'
		digit := character >= '0' && character <= '9'
		underscore := character == '_'
		if !letter && !digit && !underscore || underscore && previousUnderscore {
			return false
		}
		previousUnderscore = underscore
	}
	return true
}

func reservedPublicName(value string) bool {
	switch value {
	case "id", "created_at", "created_by", "updated_at", "updated_by",
		"row_version", "deleted_at", "deleted_by", "_included":
		return true
	default:
		return false
	}
}
