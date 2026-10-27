package profile

import (
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

func normalizeEnumValue(field Field, document []byte) ([]byte, error) {
	value, err := decodeString(document)
	if err != nil || value == "" || !utf8.ValidString(value) ||
		strings.ContainsRune(value, '\x00') || !norm.NFC.IsNormalString(value) ||
		!field.validation.permitsEnum(value) {
		return nil, ErrValidationInvalid
	}
	return encodeScalar(value)
}
