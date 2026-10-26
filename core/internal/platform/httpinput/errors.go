package httpinput

import "errors"

var (
	ErrInvalid              = errors.New("HTTP JSON request is invalid")
	ErrBodyTooLarge         = errors.New("HTTP JSON request body is too large")
	ErrMediaTypeUnsupported = errors.New("HTTP JSON media type is unsupported")
)
