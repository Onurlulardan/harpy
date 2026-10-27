package profile

import "github.com/google/uuid"

type MigrationConverter string

const (
	MigrationCopy      MigrationConverter = "copy_v1"
	MigrationNormalize MigrationConverter = "copy_normalize_v1"
)

func (converter MigrationConverter) Valid() bool {
	return converter == MigrationCopy || converter == MigrationNormalize
}

type MigrationMapping struct {
	SourceFieldID uuid.UUID
	TargetFieldID uuid.UUID
	Converter     MigrationConverter
}

func validateMigrationMappings(
	source Schema,
	target Schema,
	mappings []MigrationMapping,
) bool {
	if !source.valid() || !target.valid() || source.resource != target.resource ||
		source.id == target.id || len(mappings) == 0 || len(mappings) > MaximumActiveFields {
		return false
	}
	sources := make(map[uuid.UUID]struct{}, len(mappings))
	targets := make(map[uuid.UUID]struct{}, len(mappings))
	for _, mapping := range mappings {
		if !mapping.Converter.Valid() || mapping.SourceFieldID == uuid.Nil ||
			mapping.TargetFieldID == uuid.Nil {
			return false
		}
		if _, duplicate := sources[mapping.SourceFieldID]; duplicate {
			return false
		}
		if _, duplicate := targets[mapping.TargetFieldID]; duplicate {
			return false
		}
		sourceField, sourceFound := source.Field(mapping.SourceFieldID)
		targetField, targetFound := target.Field(mapping.TargetFieldID)
		if !sourceFound || !targetFound || sourceField.State() != FieldActive ||
			targetField.State() != FieldActive {
			return false
		}
		switch mapping.Converter {
		case MigrationCopy:
			if sourceField.ID() != targetField.ID() ||
				sourceField.SemanticSHA256() != targetField.SemanticSHA256() {
				return false
			}
		case MigrationNormalize:
			if sourceField.ID() == targetField.ID() ||
				sourceField.Type() != targetField.Type() ||
				sourceField.Cardinality() != targetField.Cardinality() {
				return false
			}
			retained, found := target.Field(sourceField.ID())
			if !found || retained.State() != FieldRetired ||
				retained.SemanticSHA256() != sourceField.SemanticSHA256() {
				return false
			}
		}
		sources[mapping.SourceFieldID] = struct{}{}
		targets[mapping.TargetFieldID] = struct{}{}
	}
	return true
}

func validateEmptyMigrationSchemas(source Schema, target Schema) bool {
	if !source.valid() || !target.valid() || source.resource != target.resource ||
		source.id == target.id || source.version >= target.version ||
		target.state != SchemaActive ||
		(source.state != SchemaActive && source.state != SchemaRetired) {
		return false
	}
	for _, field := range target.fields {
		if field.state == FieldActive && field.required {
			return false
		}
	}
	for _, sourceField := range source.fields {
		targetField, found := target.Field(sourceField.id)
		if !found || targetField.state != FieldRetired ||
			targetField.semanticSHA != sourceField.semanticSHA {
			return false
		}
	}
	return true
}
