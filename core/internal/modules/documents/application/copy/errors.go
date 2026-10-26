package copy

import "errors"

var (
	ErrInvalid     = errors.New("documents copy input is invalid")
	ErrForbidden   = errors.New("documents copy operation is forbidden")
	ErrConflict    = errors.New("documents copy target conflicts")
	ErrUnavailable = errors.New("documents copy dependency is unavailable")
)
