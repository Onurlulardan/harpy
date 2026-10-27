package profile

import (
	"bytes"
)

func normalizeBooleanValue(document []byte) ([]byte, error) {
	if bytes.Equal(document, []byte("true")) {
		return []byte("true"), nil
	}
	if bytes.Equal(document, []byte("false")) {
		return []byte("false"), nil
	}
	return nil, ErrValidationInvalid
}
