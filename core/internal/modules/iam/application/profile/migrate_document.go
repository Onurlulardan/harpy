package profile

import (
	"encoding/json"

	"github.com/google/uuid"
)

func MigrateDocument(
	sourceSchema Schema,
	targetSchema Schema,
	mappings []MigrationMapping,
	sourceDocument []byte,
	expectedSourceSHA256 []byte,
	normalizer TextNormalizer,
) (Document, error) {
	if len(mappings) == 0 {
		if !validateEmptyMigrationSchemas(sourceSchema, targetSchema) {
			return Document{}, ErrSchemaInvalid
		}
	} else if !validateMigrationMappings(sourceSchema, targetSchema, mappings) {
		return Document{}, ErrSchemaInvalid
	}
	source, err := DecodeDocument(
		sourceSchema, sourceDocument, expectedSourceSHA256, normalizer,
	)
	if err != nil {
		return Document{}, err
	}
	values := make(map[uuid.UUID]json.RawMessage, len(mappings)+len(source.values))
	for id, value := range source.values {
		sourceField, sourceFound := sourceSchema.Field(id)
		targetField, targetFound := targetSchema.Field(id)
		if sourceFound && targetFound && targetField.State() == FieldRetired &&
			sourceField.SemanticSHA256() == targetField.SemanticSHA256() {
			values[id] = value
		}
	}
	for _, mapping := range mappings {
		value, found := source.Value(mapping.SourceFieldID)
		if !found {
			continue
		}
		if mapping.Converter == MigrationNormalize {
			values[mapping.SourceFieldID] = value
			targetField, _ := targetSchema.Field(mapping.TargetFieldID)
			value, err = normalizeFieldValue(targetField, value, normalizer)
			if err != nil {
				return Document{}, ErrDocumentInvalid
			}
		}
		values[mapping.TargetFieldID] = value
	}
	return sealDocument(targetSchema, values, normalizer)
}
