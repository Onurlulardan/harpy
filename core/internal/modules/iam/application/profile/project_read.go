package profile

import (
	"encoding/json"

	"github.com/google/uuid"
)

func ProjectRead(
	schema Schema, document Document, selected []uuid.UUID,
) ([]byte, error) {
	if !schema.valid() || schema.state != SchemaActive || !document.valid() ||
		document.resource != schema.resource || selected == nil ||
		!documentSemanticsCompatible(schema, document) {
		return nil, ErrDocumentInvalid
	}
	projection := make(map[string]json.RawMessage, len(selected))
	seen := make(map[uuid.UUID]struct{}, len(selected))
	for _, id := range selected {
		if _, duplicate := seen[id]; duplicate {
			return nil, ErrFieldUnavailable
		}
		seen[id] = struct{}{}
		field, found := schema.Field(id)
		if !found || field.state != FieldActive {
			return nil, ErrFieldUnavailable
		}
		value, present := document.values[id]
		if !present {
			value = json.RawMessage("null")
		}
		projection[id.String()] = value
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		return nil, ErrDocumentInvalid
	}
	canonical, err := canonicalObject(encoded)
	if err != nil {
		return nil, ErrDocumentInvalid
	}
	return canonical, nil
}

func ProjectCompatibilityRead(
	schema Schema, document Document, selected []uuid.UUID,
) ([]byte, error) {
	if !schema.valid() || schema.state != SchemaActive || !document.valid() ||
		document.resource != schema.resource || selected == nil {
		return nil, ErrDocumentInvalid
	}
	projection := make(map[string]json.RawMessage, len(selected))
	seen := make(map[uuid.UUID]struct{}, len(selected))
	for _, id := range selected {
		if _, duplicate := seen[id]; duplicate {
			return nil, ErrFieldUnavailable
		}
		seen[id] = struct{}{}
		field, found := schema.Field(id)
		if !found || field.state != FieldActive {
			return nil, ErrFieldUnavailable
		}
		semantic, known := document.semantics[id]
		if known && semantic != field.semanticSHA {
			return nil, ErrDocumentInvalid
		}
		value, present := document.values[id]
		if present && !known {
			return nil, ErrDocumentInvalid
		}
		if !present {
			value = json.RawMessage("null")
		}
		projection[id.String()] = value
	}
	encoded, err := json.Marshal(projection)
	if err != nil {
		return nil, ErrDocumentInvalid
	}
	canonical, err := canonicalObject(encoded)
	if err != nil {
		return nil, ErrDocumentInvalid
	}
	return canonical, nil
}
