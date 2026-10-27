package attestation

import "errors"

var (
	errInvalidJSONNumber = errors.New("attestation JSON number is outside I-JSON")
	errJSONNull          = errors.New("attestation JSON null is forbidden")
)
