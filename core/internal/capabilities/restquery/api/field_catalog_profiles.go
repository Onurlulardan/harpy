package api

import "github.com/google/uuid"

func validCatalogProfiles(
	catalog FieldCatalog,
	fields map[uuid.UUID]FieldDefinition,
) bool {
	return validDefaultSelect(catalog.defaultSelect, fields) &&
		validStableID(catalog.stableID, fields) &&
		validSearchFields(catalog.searchFields, fields)
}

func validDefaultSelect(
	identifiers []uuid.UUID,
	fields map[uuid.UUID]FieldDefinition,
) bool {
	if len(identifiers) == 0 || len(identifiers) > maximumCatalogDefaultSelect {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		field, found := fields[identifier]
		if _, duplicate := seen[identifier]; !found || duplicate ||
			!field.capabilities.selectable {
			return false
		}
		seen[identifier] = struct{}{}
	}
	return true
}

func validStableID(
	identifier uuid.UUID,
	fields map[uuid.UUID]FieldDefinition,
) bool {
	field, found := fields[identifier]
	if !found || field.fieldType.kind != LogicalTypeUUID {
		return false
	}
	capabilities := field.capabilities
	return capabilities.selectable && capabilities.sortable &&
		capabilities.cursorCanonicalBytes >= minimumUUIDCursorCanonicalBytes
}

func validSearchFields(
	identifiers []uuid.UUID,
	fields map[uuid.UUID]FieldDefinition,
) bool {
	if len(identifiers) > maximumCatalogSearchFields {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		field, found := fields[identifier]
		if _, duplicate := seen[identifier]; !found || duplicate ||
			!field.capabilities.selectable || !field.capabilities.searchable {
			return false
		}
		seen[identifier] = struct{}{}
	}
	return true
}
