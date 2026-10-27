package api

import "github.com/google/uuid"

func catalogMemberFields(
	catalog FieldCatalog,
) (map[uuid.UUID]FieldDefinition, bool) {
	if len(catalog.fields) == 0 {
		return nil, false
	}
	fields := make(map[uuid.UUID]FieldDefinition, len(catalog.fields))
	identifiers := make(map[uuid.UUID]struct{}, len(catalog.fields)+len(catalog.relations))
	names := make(map[string]struct{}, len(catalog.fields)+len(catalog.relations))
	for _, field := range catalog.fields {
		if field.Validate() != nil || memberIdentityExists(
			field.id, field.publicName, identifiers, names,
		) {
			return nil, false
		}
		identifiers[field.id] = struct{}{}
		names[field.publicName] = struct{}{}
		fields[field.id] = field
	}
	for _, relation := range catalog.relations {
		if relation.Validate() != nil || memberIdentityExists(
			relation.relationID, relation.publicName, identifiers, names,
		) {
			return nil, false
		}
		identifiers[relation.relationID] = struct{}{}
		names[relation.publicName] = struct{}{}
	}
	return fields, true
}

func memberIdentityExists(
	id uuid.UUID,
	name string,
	identifiers map[uuid.UUID]struct{},
	names map[string]struct{},
) bool {
	if id == uuid.Nil || name == "" {
		return true
	}
	_, duplicateID := identifiers[id]
	_, duplicateName := names[name]
	return duplicateID || duplicateName
}
