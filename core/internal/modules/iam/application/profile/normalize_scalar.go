package profile

import "encoding/json"

func normalizeScalarValue(
	field Field, document []byte, normalizer TextNormalizer,
) ([]byte, error) {
	switch field.kind {
	case LogicalString:
		return normalizeStringValue(field, document, normalizer)
	case LogicalInteger:
		return normalizeIntegerValue(field, document)
	case LogicalDecimal:
		return normalizeDecimalValue(field, document)
	case LogicalBoolean:
		return normalizeBooleanValue(document)
	case LogicalDate, LogicalTimestamp:
		return normalizeTemporalValue(field.kind, document)
	case LogicalUUID:
		return normalizeUUIDValue(document)
	case LogicalEnum:
		return normalizeEnumValue(field, document)
	case LogicalJSON:
		return normalizeJSONValue(document)
	default:
		return nil, ErrValidationInvalid
	}
}

func decodeString(document []byte) (string, error) {
	if len(document) < 2 || document[0] != '"' {
		return "", ErrValidationInvalid
	}
	var value string
	if json.Unmarshal(document, &value) != nil {
		return "", ErrValidationInvalid
	}
	return value, nil
}

func encodeScalar(value any) ([]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return nil, ErrValidationInvalid
	}
	return document, nil
}
