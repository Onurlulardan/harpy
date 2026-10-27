package profile

import (
	"bytes"

	"github.com/google/uuid"
)

func Patch(
	schema Schema,
	current Document,
	patch []byte,
	normalizer TextNormalizer,
) (Document, error) {
	if !schema.valid() || schema.state != SchemaActive || !current.valid() ||
		current.resource != schema.resource {
		return Document{}, ErrPatchInvalid
	}
	if !documentSemanticsCompatible(schema, current) {
		return Document{}, ErrSchemaInvalid
	}
	_, changes, err := decodeObject(patch)
	if err != nil {
		return Document{}, ErrPatchInvalid
	}
	values := cloneValues(current.values)
	for id := range values {
		field, found := schema.Field(id)
		if found && field.state == FieldRetired {
			delete(values, id)
		}
	}
	for rawID, rawValue := range changes {
		id, valid := parseStableID(rawID)
		field, found := schema.Field(id)
		if !valid || !found || field.state != FieldActive {
			return Document{}, ErrFieldUnavailable
		}
		if bytes.Equal(rawValue, []byte("null")) {
			delete(values, id)
			continue
		}
		normalized, valueErr := normalizeFieldValue(field, rawValue, normalizer)
		if valueErr != nil {
			return Document{}, ErrPatchInvalid
		}
		values[id] = normalized
	}
	result, err := sealDocument(schema, values, normalizer)
	if err != nil {
		if err == ErrRequiredFieldMissing {
			return Document{}, err
		}
		return Document{}, ErrPatchInvalid
	}
	return result, nil
}

func documentSemanticsCompatible(schema Schema, document Document) bool {
	for id, source := range document.semantics {
		field, found := schema.Field(id)
		if found && field.semanticSHA != source {
			return false
		}
		if !found {
			if _, hasValue := document.values[id]; hasValue {
				return false
			}
		}
	}
	return true
}

func StablePatchKey(id uuid.UUID) string {
	if id == uuid.Nil || id.Variant() != uuid.RFC4122 {
		return ""
	}
	return id.String()
}
