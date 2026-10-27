package packageregistry

import "errors"

var (
	ErrInvalid     = errors.New("catalog package registry input is invalid")
	ErrForbidden   = errors.New("catalog package registry access is forbidden")
	ErrConflict    = errors.New("catalog package registry identity conflicts with immutable evidence")
	ErrNotFound    = errors.New("catalog package registry record was not found")
	ErrUnavailable = errors.New("catalog package registry is unavailable")
)
