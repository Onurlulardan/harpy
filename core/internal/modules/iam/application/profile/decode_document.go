package profile

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"

	"github.com/google/uuid"
)

func DecodeDocument(
	schema Schema,
	document []byte,
	expectedSHA256 []byte,
	normalizer TextNormalizer,
) (Document, error) {
	if !schema.valid() {
		return Document{}, ErrSchemaInvalid
	}
	canonical, rawValues, err := decodeObject(document)
	if err != nil {
		return Document{}, ErrDocumentInvalid
	}
	values := make(map[uuid.UUID]json.RawMessage, len(rawValues))
	for rawID, rawValue := range rawValues {
		id, valid := parseStableID(rawID)
		field, found := schema.Field(id)
		if !valid || !found {
			return Document{}, ErrDocumentInvalid
		}
		normalized, valueErr := validateStoredFieldValue(field, rawValue, normalizer)
		if valueErr != nil {
			return Document{}, ErrDocumentInvalid
		}
		values[id] = normalized
	}
	if err := validateRequiredFields(schema, values); err != nil {
		return Document{}, err
	}
	digest := canonicalDigest(canonical)
	if expectedSHA256 != nil && (len(expectedSHA256) != sha256.Size ||
		!bytes.Equal(expectedSHA256, digest[:])) {
		return Document{}, ErrDocumentInvalid
	}
	return Document{
		schemaID: schema.id, resource: schema.resource, values: values,
		semantics: schemaSemantics(schema), canonical: canonical, digest: digest,
	}, nil
}

func decodeObject(document []byte) ([]byte, map[string]json.RawMessage, error) {
	canonical, err := canonicalObject(document)
	if err != nil {
		return nil, nil, err
	}
	values := make(map[string]json.RawMessage)
	if json.Unmarshal(canonical, &values) != nil {
		return nil, nil, ErrDocumentInvalid
	}
	return canonical, values, nil
}

func parseStableID(value string) (uuid.UUID, bool) {
	id, err := uuid.Parse(value)
	return id, err == nil && id != uuid.Nil && id.Variant() == uuid.RFC4122 &&
		id.String() == value
}
