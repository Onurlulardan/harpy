package profile

import "github.com/google/uuid"

func normalizeUUIDValue(document []byte) ([]byte, error) {
	value, err := decodeString(document)
	if err != nil {
		return nil, ErrValidationInvalid
	}
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.Variant() != uuid.RFC4122 ||
		parsed.String() != value {
		return nil, ErrValidationInvalid
	}
	return encodeScalar(value)
}
