package profile

import (
	"encoding/json"
)

func normalizeListValue(
	field Field, document []byte, normalizer TextNormalizer,
) ([]byte, error) {
	var values []json.RawMessage
	if len(document) < 2 || document[0] != '[' || json.Unmarshal(document, &values) != nil ||
		len(values) > field.validation.maximumItems {
		return nil, ErrValidationInvalid
	}
	normalized := make([]json.RawMessage, len(values))
	for index, value := range values {
		item, err := normalizeScalarValue(field, value, normalizer)
		if err != nil {
			return nil, ErrValidationInvalid
		}
		normalized[index] = item
	}
	encoded, err := json.Marshal(normalized)
	if err != nil {
		return nil, ErrValidationInvalid
	}
	return canonicalMarshal(json.RawMessage(encoded), field.validation.maximumBytes)
}
