package application

import "errors"

var (
	ErrAuthentication = errors.New("audit authentication is required")
	ErrInvalid        = errors.New("audit request is invalid")
	ErrForbidden      = errors.New("audit operation is forbidden")
	ErrNotFound       = errors.New("audit entry was not found")
	ErrConflict       = errors.New("audit request conflicts with persisted evidence")
	ErrUnavailable    = errors.New("audit dependency is unavailable")
)
