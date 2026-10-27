package csrf

import "errors"

var (
	ErrInvalidPolicy     = errors.New("invalid CSRF origin policy")
	ErrInvalidKey        = errors.New("invalid CSRF key")
	ErrInvalidCredential = errors.New("invalid CSRF credential")
	ErrSignerClosed      = errors.New("CSRF signer is closed")
)
