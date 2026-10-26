package recordkernel

import (
	"github.com/google/uuid"
	"slices"
)

func buildField(
	definition FieldDefinition,
	registry registryCatalog,
	published map[uuid.UUID]PublishedField,
	policies map[uuid.UUID]WritePolicy,
	relations map[uuid.UUID]RelationProfile,
) (Field, error) {
	fieldID, err := parseStableID(definition.ID)
	typeID, role := TypeID(definition.TypeID), Role(definition.Role)
	descriptor, known := registry.types[typeID]
	publication, bound := published[fieldID]
	policy, hasPolicy := policies[fieldID]
	if err != nil || !validPublicName(definition.Name) || !known ||
		!slices.Contains(descriptor.Roles, role) || !bound ||
		!publishedMatches(publication, definition) || !hasPolicy ||
		role == RoleCalculation && policy != WriteEngine {
		return Field{}, ErrCatalogInvalid
	}
	field := Field{ID: fieldID, PublicName: definition.Name, TypeID: typeID,
		Role: role, ValueProfile: descriptor.ValueProfile,
		Nullable: definition.Nullable, WritePolicy: policy,
		Generation:  publication.Generation,
		EnumOptions: append([]string(nil), definition.EnumOptions...)}
	if definition.Decimal != nil {
		field.Decimal = &DecimalProfile{
			Precision: definition.Decimal.Precision, Scale: definition.Decimal.Scale,
		}
	}
	if relation, found := relations[fieldID]; found {
		copy := relation
		field.Relation = &copy
	}
	return field, nil
}

func countRole(fields []Field, role Role) int {
	count := 0
	for _, field := range fields {
		if field.Role == role {
			count++
		}
	}
	return count
}
