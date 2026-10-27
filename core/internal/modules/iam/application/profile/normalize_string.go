package profile

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func normalizeStringValue(
	field Field, document []byte, normalizer TextNormalizer,
) ([]byte, error) {
	value, err := decodeString(document)
	if err != nil || !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') ||
		!norm.NFC.IsNormalString(value) {
		return nil, ErrValidationInvalid
	}
	value, err = normalizeProfileText(field, value, normalizer)
	if err != nil || !utf8.ValidString(value) || strings.ContainsRune(value, '\x00') ||
		!norm.NFC.IsNormalString(value) ||
		field.normalization != NormalizationNoneV1 && value == "" {
		return nil, ErrValidationInvalid
	}
	return encodeScalar(value)
}
