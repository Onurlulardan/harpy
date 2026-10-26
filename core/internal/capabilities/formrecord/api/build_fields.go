package recordkernel

import (
	"github.com/google/uuid"
	"slices"
)

func buildFields(
	schema StorageSchema,
	registry registryCatalog,
	binding CatalogBinding,
	rules RecordRules,
) ([]Field, map[uuid.UUID]Field, map[string]Field, error) {
	if len(schema.Fields) < 1 || uint64(len(schema.Fields)) > registry.limits.ActiveFields {
		return nil, nil, nil, ErrCatalogInvalid
	}
	published, err := publishedFieldMap(binding)
	if err != nil || len(published) != len(schema.Fields) {
		return nil, nil, nil, ErrCatalogInvalid
	}
	relations, err := relationProfiles(schema)
	if err != nil {
		return nil, nil, nil, err
	}
	policies, err := recordFieldPolicies(rules, len(schema.Fields))
	if err != nil {
		return nil, nil, nil, err
	}
	fields := make([]Field, 0, len(schema.Fields))
	byID := make(map[uuid.UUID]Field, len(schema.Fields))
	byName := make(map[string]Field, len(schema.Fields))
	for _, definition := range schema.Fields {
		field, fieldErr := buildField(definition, registry, published, policies, relations)
		if fieldErr != nil {
			return nil, nil, nil, fieldErr
		}
		if _, duplicate := byID[field.ID]; duplicate {
			return nil, nil, nil, ErrCatalogInvalid
		}
		if _, duplicate := byName[field.PublicName]; duplicate {
			return nil, nil, nil, ErrCatalogInvalid
		}
		fields = append(fields, field)
		byID[field.ID], byName[field.PublicName] = field, field
	}
	if len(relations) != countRole(fields, RoleRelation) {
		return nil, nil, nil, ErrCatalogInvalid
	}
	slices.SortFunc(fields, func(left, right Field) int { return compareUUID(left.ID, right.ID) })
	return fields, byID, byName, nil
}
