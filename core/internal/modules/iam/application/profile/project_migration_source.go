package profile

import (
	"encoding/json"

	"github.com/google/uuid"
)

func ProjectMigrationSource(
	sourceSchema Schema,
	targetSchema Schema,
	mappings []MigrationMapping,
	targetDocument []byte,
	expectedTargetSHA256 []byte,
	normalizer TextNormalizer,
) (Document, error) {
	if !validateMigrationMappings(sourceSchema, targetSchema, mappings) {
		return Document{}, ErrSchemaInvalid
	}
	target, err := DecodeDocument(
		targetSchema, targetDocument, expectedTargetSHA256, normalizer,
	)
	if err != nil {
		return Document{}, err
	}
	values := make(map[uuid.UUID]json.RawMessage, len(target.values))
	for id, value := range target.values {
		sourceField, sourceFound := sourceSchema.Field(id)
		targetField, targetFound := targetSchema.Field(id)
		if sourceFound && targetFound &&
			sourceField.SemanticSHA256() == targetField.SemanticSHA256() {
			values[id] = value
		}
	}
	for _, mapping := range mappings {
		value, found := target.Value(mapping.SourceFieldID)
		if found {
			values[mapping.SourceFieldID] = value
		}
	}
	return sealDocument(sourceSchema, values, normalizer)
}
