package adoption

import (
	"fmt"

	"github.com/google/uuid"
)

func (state *validationState) validateLayoutFields(fields map[profileFieldKey]uuid.UUID) {
	mappings := make(map[profileFieldKey]LayoutFieldMapping)
	for _, mapping := range state.mapping.LayoutFields {
		key := profileKey(mapping.ResourceKind, mapping.SourceKey)
		if _, duplicate := mappings[key]; duplicate {
			state.add(issue("mapping_duplicate", "layouts", profileKeyText(key)))
		}
		mappings[key] = mapping
	}
	seen := make(map[profileFieldKey]bool)
	positions := make(map[string]bool)
	layouts := make(map[string]uuid.UUID)
	for _, source := range state.snapshot.LayoutFields {
		key := profileKey(source.ResourceKind, source.FieldKey)
		position := fmt.Sprintf("%s:%s:%d", key.resourceKind,
			canonicalCatalogKey(source.SectionKey), source.Position)
		if seen[key] || fields[key] == uuid.Nil || source.Position < 0 ||
			!catalogKeyPattern.MatchString(canonicalCatalogKey(source.SectionKey)) {
			state.add(issue("layout_field_invalid", "layouts", profileKeyText(key)))
		}
		if positions[position] {
			state.add(issue("layout_position_collision", "layouts", position))
		}
		seen[key], positions[position] = true, true
		mapping, exists := mappings[key]
		if !exists || mapping.FieldID != fields[key] || mapping.LayoutID == uuid.Nil {
			state.add(issue("layout_field_mapping_missing", "layouts", profileKeyText(key)))
			continue
		}
		if prior := layouts[key.resourceKind]; prior != uuid.Nil && prior != mapping.LayoutID {
			state.add(issue("layout_id_collision", "layouts", key.resourceKind))
		}
		layouts[key.resourceKind] = mapping.LayoutID
	}
	for key := range mappings {
		if !seen[key] {
			state.add(issue("mapping_source_missing", "layouts", profileKeyText(key)))
		}
	}
}
