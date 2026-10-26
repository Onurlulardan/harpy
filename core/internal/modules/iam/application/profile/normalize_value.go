package profile

import (
	"bytes"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func normalizeFieldValue(
	field Field, document []byte, normalizer TextNormalizer,
) ([]byte, error) {
	if !field.valid() || len(document) < 1 || len(document) > field.validation.maximumBytes {
		return nil, ErrValidationInvalid
	}
	canonical, err := canonicaljsonv1.Canonicalize(document, field.validation.maximumBytes)
	if err != nil || bytes.Equal(canonical, []byte("null")) {
		return nil, ErrValidationInvalid
	}
	var normalized []byte
	if field.cardinality == CardinalityList {
		normalized, err = normalizeListValue(field, canonical, normalizer)
	} else {
		normalized, err = normalizeScalarValue(field, canonical, normalizer)
	}
	if err != nil || len(normalized) > field.validation.maximumBytes {
		return nil, ErrValidationInvalid
	}
	return normalized, nil
}

func validateStoredFieldValue(
	field Field, document []byte, normalizer TextNormalizer,
) ([]byte, error) {
	canonical, err := canonicaljsonv1.Canonicalize(document, field.validation.maximumBytes)
	if err != nil {
		return nil, ErrDocumentInvalid
	}
	normalized, err := normalizeFieldValue(field, canonical, normalizer)
	if err != nil || !bytes.Equal(canonical, normalized) {
		return nil, ErrDocumentInvalid
	}
	return normalized, nil
}
