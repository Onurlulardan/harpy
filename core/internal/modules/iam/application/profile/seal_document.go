package profile

import (
	"encoding/json"

	"github.com/google/uuid"
)

func sealDocument(
	schema Schema, values map[uuid.UUID]json.RawMessage, normalizer TextNormalizer,
) (Document, error) {
	if err := validateRequiredFields(schema, values); err != nil {
		return Document{}, err
	}
	wire := make(map[string]json.RawMessage, len(values))
	for id, value := range values {
		field, found := schema.Field(id)
		if !found {
			return Document{}, ErrDocumentInvalid
		}
		normalized, err := validateStoredFieldValue(field, value, normalizer)
		if err != nil {
			return Document{}, ErrDocumentInvalid
		}
		wire[id.String()] = normalized
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return Document{}, ErrDocumentInvalid
	}
	canonical, err := canonicalObject(encoded)
	if err != nil {
		return Document{}, ErrDocumentInvalid
	}
	return Document{
		schemaID: schema.id, resource: schema.resource, values: cloneValues(values),
		semantics: schemaSemantics(schema), canonical: canonical,
		digest: canonicalDigest(canonical),
	}, nil
}

func validateRequiredFields(
	schema Schema, values map[uuid.UUID]json.RawMessage,
) error {
	for _, field := range schema.fields {
		if field.state == FieldActive && field.required {
			if _, found := values[field.id]; !found {
				return ErrRequiredFieldMissing
			}
		}
	}
	return nil
}
