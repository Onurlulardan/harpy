package recordkernel

import (
	"github.com/google/uuid"
)

func decodeCatalogFields(
	documents []catalogFieldDocument,
	binding CatalogBinding,
	limits BudgetLimits,
) ([]Field, map[uuid.UUID]Field, map[string]Field, error) {
	if len(documents) < 1 || uint64(len(documents)) > limits.ActiveFields ||
		len(documents) != len(binding.Fields) {
		return nil, nil, nil, ErrCatalogInvalid
	}
	published, err := publishedFieldMap(binding)
	if err != nil {
		return nil, nil, nil, err
	}
	fields := make([]Field, len(documents))
	seenNames := make(map[string]struct{}, len(documents))
	for index, document := range documents {
		fieldID, parseErr := parseCatalogID(document.ID)
		typeID, role := TypeID(document.Type), Role(document.Role)
		expectedProfile, known := expectedValueProfile(typeID)
		publication, bound := published[fieldID]
		if parseErr != nil || !known || !catalogRoleAllowed(typeID, role) ||
			string(expectedProfile) != document.Profile || !bound ||
			publication.PublicName != document.PublicName ||
			publication.Type != document.Type || publication.Role != document.Role ||
			publication.Generation != document.Generation ||
			!validPublicName(document.PublicName) || !document.WritePolicy.valid() ||
			role == RoleCalculation && document.WritePolicy != WriteEngine ||
			index > 0 && document.ID <= documents[index-1].ID {
			return nil, nil, nil, ErrCatalogInvalid
		}
		if _, duplicate := seenNames[document.PublicName]; duplicate {
			return nil, nil, nil, ErrCatalogInvalid
		}
		seenNames[document.PublicName] = struct{}{}
		field := Field{ID: fieldID, PublicName: document.PublicName,
			TypeID: typeID, Role: role, ValueProfile: expectedProfile,
			Nullable: document.Nullable, WritePolicy: document.WritePolicy,
			Generation: document.Generation, Decimal: document.Decimal,
			EnumOptions: append([]string(nil), document.EnumOptions...),
			Relation:    document.Relation, Required: document.Required,
			Immutable: document.Immutable, TextMinimum: document.TextMinimum,
			TextMaximum: document.TextMaximum}
		if !validDecodedField(field, limits) {
			return nil, nil, nil, ErrCatalogInvalid
		}
		fields[index] = cloneField(field)
	}
	if uint64(countRole(fields, RoleRelation)) > limits.RelationFields ||
		uint64(countRole(fields, RoleCalculation)) > limits.CalculationFields {
		return nil, nil, nil, ErrCatalogInvalid
	}
	byID, byName := indexFields(fields)
	return fields, byID, byName, nil
}
