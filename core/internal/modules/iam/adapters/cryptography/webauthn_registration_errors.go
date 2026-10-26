package cryptography

import "errors"

var (
	ErrRegistrationInvalid = errors.New("WebAuthn registration contract is invalid")
	ErrRegistrationDenied  = errors.New("WebAuthn registration was denied")
)
