package domain

import "errors"

var (
	ErrInvalid        = errors.New("usage input is invalid")
	ErrAuthentication = errors.New("usage authentication is required")
	ErrConflict       = errors.New("usage source identity conflicts")
	ErrForbidden      = errors.New("usage access is forbidden")
	ErrNotFound       = errors.New("usage subject was not found")
	ErrUnavailable    = errors.New("usage dependency is unavailable")
	ErrRetentionHeld  = errors.New("usage retention is held")
	ErrSourceDrift    = errors.New("usage adoption source drifted")
)
