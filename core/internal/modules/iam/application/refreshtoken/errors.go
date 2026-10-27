package refreshtoken

import "errors"

var (
	ErrTokenInvalid      = errors.New("refresh token is invalid")
	ErrTokenDestroyed    = errors.New("refresh token is destroyed")
	ErrRandomUnavailable = errors.New("refresh token randomness is unavailable")
	ErrKeyInvalid        = errors.New("refresh token HMAC key is invalid")
)
