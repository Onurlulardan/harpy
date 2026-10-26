package profile

import (
	"encoding/json"

	"github.com/google/uuid"
)

func ProjectLineageDocument(
	sourceSchema Schema,
	targetSchema Schema,
	sourceDocument []byte,
	expectedSourceSHA256 []byte,
	normalizer TextNormalizer,
) (Document, error) {
	if !sourceSchema.valid() || !targetSchema.valid() ||
		sourceSchema.Resource() != targetSchema.Resource() ||
		sourceSchema.Version() > targetSchema.Version() {
		return Document{}, ErrSchemaInvalid
	}
	source, err := DecodeDocument(
		sourceSchema, sourceDocument, expectedSourceSHA256, normalizer,
	)
	if err != nil {
		return Document{}, err
	}
	values := make(map[uuid.UUID]json.RawMessage, len(source.values))
	for id, value := range source.values {
		sourceField, sourceFound := sourceSchema.Field(id)
		targetField, targetFound := targetSchema.Field(id)
		if !sourceFound || !targetFound ||
			sourceField.SemanticSHA256() != targetField.SemanticSHA256() {
			return Document{}, ErrSchemaInvalid
		}
		values[id] = value
	}
	return sealDocument(targetSchema, values, normalizer)
}
