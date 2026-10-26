package profile

import (
	"bytes"

	"github.com/google/uuid"
)

func UpgradeMigratedDocument(
	source Schema,
	target Schema,
	mappings []MigrationMapping,
	current Document,
	patch []byte,
	normalizer TextNormalizer,
) (Document, error) {
	if !validateMigrationMappings(source, target, mappings) ||
		target.state != SchemaActive || !current.valid() ||
		current.schemaID != target.id || current.resource != target.resource ||
		!documentSemanticsCompatible(target, current) {
		return Document{}, ErrPatchInvalid
	}
	_, changes, err := decodeObject(patch)
	if err != nil {
		return Document{}, ErrPatchInvalid
	}
	bySource := make(map[uuid.UUID]MigrationMapping, len(mappings))
	for _, mapping := range mappings {
		bySource[mapping.SourceFieldID] = mapping
	}
	values := cloneValues(current.values)
	for rawID, rawValue := range changes {
		id, valid := parseStableID(rawID)
		field, found := source.Field(id)
		mapping, mapped := bySource[id]
		if !valid || !found || field.state != FieldActive || !mapped {
			return Document{}, ErrFieldUnavailable
		}
		if bytes.Equal(rawValue, []byte("null")) {
			if field.required {
				return Document{}, ErrRequiredFieldMissing
			}
			delete(values, mapping.SourceFieldID)
			delete(values, mapping.TargetFieldID)
			continue
		}
		sourceValue, valueErr := normalizeFieldValue(field, rawValue, normalizer)
		if valueErr != nil {
			return Document{}, ErrPatchInvalid
		}
		if mapping.Converter == MigrationNormalize {
			values[mapping.SourceFieldID] = sourceValue
			targetField, _ := target.Field(mapping.TargetFieldID)
			sourceValue, valueErr = normalizeFieldValue(
				targetField, sourceValue, normalizer,
			)
			if valueErr != nil {
				return Document{}, ErrPatchInvalid
			}
		}
		values[mapping.TargetFieldID] = sourceValue
	}
	result, err := sealDocument(target, values, normalizer)
	if err == ErrRequiredFieldMissing {
		return Document{}, err
	}
	if err != nil {
		return Document{}, ErrPatchInvalid
	}
	return result, nil
}
