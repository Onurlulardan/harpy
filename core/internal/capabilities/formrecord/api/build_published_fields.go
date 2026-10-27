package recordkernel

import (
	"github.com/google/uuid"
)

func publishedFieldMap(
	binding CatalogBinding,
) (map[uuid.UUID]PublishedField, error) {
	result := make(map[uuid.UUID]PublishedField, len(binding.Fields))
	for _, field := range binding.Fields {
		if !stableUUIDv7(field.FieldID) || !validPublicName(field.PublicName) ||
			!knownType(TypeID(field.Type)) || !validRole(field.Role) ||
			field.Generation < 1 {
			return nil, ErrCatalogInvalid
		}
		if _, duplicate := result[field.FieldID]; duplicate {
			return nil, ErrCatalogInvalid
		}
		result[field.FieldID] = field
	}
	return result, nil
}

func publishedMatches(
	published PublishedField,
	definition FieldDefinition,
) bool {
	return published.PublicName == definition.Name &&
		published.Type == string(definition.TypeID) &&
		published.Role == string(definition.Role)
}

func validRole(role string) bool {
	return role == string(RoleStored) ||
		role == string(RoleRelation) ||
		role == string(RoleCalculation)
}
