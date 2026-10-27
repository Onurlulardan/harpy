package adoption

import "github.com/google/uuid"

type profileFieldKey struct {
	resourceKind string
	fieldKey     string
}

var profileValueKinds = map[string]bool{
	"string": true, "integer": true, "number": true, "boolean": true,
	"date": true, "datetime": true, "uuid": true, "json": true,
}

func (state *validationState) validateProfileFields() map[profileFieldKey]uuid.UUID {
	mappings := make(map[profileFieldKey]ProfileFieldMapping)
	for _, mapping := range state.mapping.ProfileFields {
		key := profileKey(mapping.ResourceKind, mapping.SourceKey)
		if _, duplicate := mappings[key]; duplicate {
			state.add(issue("mapping_duplicate", "profiles", profileKeyText(key)))
		}
		mappings[key] = mapping
	}
	seen := make(map[profileFieldKey]bool)
	result := make(map[profileFieldKey]uuid.UUID)
	for _, source := range state.snapshot.ProfileFields {
		key := profileKey(source.ResourceKind, source.Key)
		if seen[key] || !profileKinds[key.resourceKind] ||
			!catalogKeyPattern.MatchString(key.fieldKey) ||
			!profileValueKinds[canonicalCatalogKey(source.ValueKind)] {
			state.add(issue("profile_field_invalid", "profiles", profileKeyText(key)))
		}
		seen[key] = true
		mapping, exists := mappings[key]
		if !exists || mapping.FieldID == uuid.Nil {
			state.add(issue("profile_field_mapping_missing", "profiles", profileKeyText(key)))
			continue
		}
		result[key] = mapping.FieldID
	}
	for key := range mappings {
		if !seen[key] {
			state.add(issue("mapping_source_missing", "profiles", profileKeyText(key)))
		}
	}
	return result
}

func profileKey(resourceKind, fieldKey string) profileFieldKey {
	return profileFieldKey{
		resourceKind: canonicalCatalogKey(resourceKind),
		fieldKey:     canonicalCatalogKey(fieldKey),
	}
}

func profileKeyText(key profileFieldKey) string {
	return key.resourceKind + ":" + key.fieldKey
}
