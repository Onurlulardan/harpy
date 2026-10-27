package entitlement

import "errors"

var (
	ErrKey        = errors.New("billing entitlement key is invalid")
	ErrDefinition = errors.New("billing entitlement definition is invalid")
	ErrValue      = errors.New("billing entitlement value is invalid")
	ErrResolution = errors.New("billing entitlement resolution is invalid")
)
