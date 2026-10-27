package profile

import "github.com/google/uuid"

// SelectMigrationMappingsForSource derives the valid mapping subset for an
// immutable historical source schema. Mappings for fields that did not exist,
// or were already retired, in that source generation are intentionally skipped.
func SelectMigrationMappingsForSource(
	source Schema,
	target Schema,
	mappings []MigrationMapping,
) ([]MigrationMapping, error) {
	if source.ID() == uuid.Nil || target.ID() == uuid.Nil ||
		source.Resource() != target.Resource() || source.ID() == target.ID() ||
		source.Version() >= target.Version() || target.State() != SchemaActive ||
		(source.State() != SchemaActive && source.State() != SchemaRetired) ||
		len(mappings) == 0 || len(mappings) > MaximumActiveFields {
		return nil, ErrSchemaInvalid
	}
	sources := make(map[uuid.UUID]struct{}, len(mappings))
	targets := make(map[uuid.UUID]struct{}, len(mappings))
	selected := make([]MigrationMapping, 0, len(mappings))
	selectedTargets := make(map[uuid.UUID]struct{}, len(mappings))
	for _, mapping := range mappings {
		if !mapping.Converter.Valid() || mapping.SourceFieldID == uuid.Nil ||
			mapping.TargetFieldID == uuid.Nil {
			return nil, ErrSchemaInvalid
		}
		if _, duplicate := sources[mapping.SourceFieldID]; duplicate {
			return nil, ErrSchemaInvalid
		}
		if _, duplicate := targets[mapping.TargetFieldID]; duplicate {
			return nil, ErrSchemaInvalid
		}
		sources[mapping.SourceFieldID] = struct{}{}
		targets[mapping.TargetFieldID] = struct{}{}
		targetField, found := target.Field(mapping.TargetFieldID)
		if !found || targetField.State() != FieldActive {
			return nil, ErrSchemaInvalid
		}
		sourceField, found := source.Field(mapping.SourceFieldID)
		if !found || sourceField.State() == FieldRetired {
			continue
		}
		if sourceField.State() != FieldActive {
			return nil, ErrSchemaInvalid
		}
		selected = append(selected, mapping)
		selectedTargets[mapping.TargetFieldID] = struct{}{}
	}
	for _, field := range target.Fields() {
		if field.State() == FieldActive && field.Required() {
			if _, covered := selectedTargets[field.ID()]; !covered {
				return nil, ErrSchemaInvalid
			}
		}
	}
	if len(selected) == 0 {
		if !validateEmptyMigrationSchemas(source, target) {
			return nil, ErrSchemaInvalid
		}
		return selected, nil
	}
	if !validateMigrationMappings(source, target, selected) {
		return nil, ErrSchemaInvalid
	}
	return selected, nil
}
