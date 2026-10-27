package link

import "errors"

var (
	ErrInvalid     = errors.New("documents link input is invalid")
	ErrForbidden   = errors.New("documents link source is forbidden")
	ErrNotFound    = errors.New("documents link was not found")
	ErrConflict    = errors.New("documents link target conflicts")
	ErrUnavailable = errors.New("documents link dependency is unavailable")
)
