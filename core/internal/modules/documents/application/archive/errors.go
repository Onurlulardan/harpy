package archive

import "errors"

var (
	ErrInvalid     = errors.New("documents archive input is invalid")
	ErrForbidden   = errors.New("documents archive operation is forbidden")
	ErrNotFound    = errors.New("documents archive was not found")
	ErrConflict    = errors.New("documents archive state conflicts")
	ErrLimit       = errors.New("documents archive limit exceeded")
	ErrCancelled   = errors.New("documents archive job was cancelled")
	ErrUnavailable = errors.New("documents archive dependency is unavailable")
)
