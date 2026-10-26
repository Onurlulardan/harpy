package provider

import "errors"

var (
	ErrKey        = errors.New("billing provider key is invalid")
	ErrDescriptor = errors.New("billing provider descriptor is invalid")
	ErrRegistry   = errors.New("billing provider registry is invalid")
	ErrUnknown    = errors.New("billing provider is unknown")
)
