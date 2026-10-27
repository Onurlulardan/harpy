package fieldregistry

import "errors"

var (
	ErrRegistryUnavailable = errors.New("forms field registry is unavailable")
	ErrTypeUnknown         = errors.New("forms field type is unknown")
	ErrRoleUnsupported     = errors.New("forms field role is unsupported")
)
