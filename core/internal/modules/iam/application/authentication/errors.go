package authentication

import "errors"

var (
	ErrAuthenticationContract = errors.New("authentication contract is invalid")
	ErrInvalidCredentials     = errors.New("authentication credential is invalid")
	ErrAuthenticationStorage  = errors.New("authentication storage is unavailable")
	ErrAuthenticationCrypto   = errors.New("authentication cryptography is unavailable")
)
